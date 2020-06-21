package hot_compiler

import (
	"crypto/sha256"
	gast "go/ast"
	"go/build"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"

	b58 "github.com/mr-tron/base58/base58"
	"github.com/sirupsen/logrus"
)

const buildTag = "controllerbus_hot_plugin"
const buildHashConstTag = "{buildHash}"

// CompilePluginFromFile compiles the plugin from the gfile.
//
// {buildHash} will be replaced with the build hash in the output filename.
func CompilePluginFromFile(
	le *logrus.Entry,
	gfile *gast.File,
	intermediateGoFile string,
	outFile string,
	preWriteOutFileHook func(nextOutFilePath, nextOutFileContentsPath string) error,
) error {
	builderCtx := build.Default
	builderCtx.BuildTags = append(builderCtx.BuildTags, buildTag)
	dat, err := FormatFile(gfile)
	if err != nil {
		return err
	}
	// write the intermediate go file
	if err := ioutil.WriteFile(intermediateGoFile, dat, 0644); err != nil {
		return err
	}
	// build the intermediate output dir
	tmpName, err := ioutil.TempDir("", "controllerbus-hot-compiler-tmpdir")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpName)

	// The Go plugin loader will recognize the binary as the identical library,
	// and do nothing when loading the library the second time, if the filename
	// is the same AND/OR if the "pluginpath" is the same.

	// The Go compiler is used here with a static pluginpath first. This is used
	// to generate a deterministic output hash. If the output filename with the
	// hash included does not already exist, then the go compiler is invoked a
	// second time with the plugin path including the hash.

	// start the go compiler excecution #1
	intermediateOutFile1 := path.Join(tmpName, "pass-1.cbus.so")
	ecmd := exec.Command(
		"go", "build",
		"-v",
		"-buildmode=plugin",
		"-tags",
		buildTag,
		"-o",
		intermediateOutFile1,
		intermediateGoFile,
	)
	ecmd.Env = make([]string, len(os.Environ()))
	copy(ecmd.Env, os.Environ())
	ecmd.Env = append(
		ecmd.Env,
		"GO111MODULE=on",
	)
	ecmd.Stderr = os.Stderr
	ecmd.Stdout = os.Stdout
	le.Debugf("running go compiler to detect changes: %s", ecmd.String())
	err = ecmd.Run()
	if err != nil {
		return err
	}

	// replace the hash in the output file.
	hasher := sha256.New()
	ofile, oerr := os.Open(intermediateOutFile1)
	if oerr != nil {
		return oerr
	}
	_, err = io.Copy(hasher, ofile)
	ofile.Close()
	if err != nil && err != io.EOF {
		return err
	}
	hashStr := b58.Encode(hasher.Sum(nil))

	if strings.Contains(outFile, buildHashConstTag) {
		outFile = strings.ReplaceAll(outFile, buildHashConstTag, hashStr[:16])
		if _, err := os.Stat(outFile); !os.IsNotExist(err) {
			le.Info("detected that output would be identical, skipping")
			return nil
		}
	}

	// add the build ID to the code to make it unique
	interFile, err := os.OpenFile(intermediateGoFile, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	interFile.Write([]byte("\nvar HotPluginBuildUUID = `" + hashStr + "`\n"))
	if err := ioutil.WriteFile(intermediateGoFile, dat, 0644); err != nil {
		return err
	}

	// start the go compiler execution #2
	intermediateOutFile2 := path.Join(tmpName, "pass-2.cbus.so")
	ecmd = exec.Command(
		"go", "build",
		"-v",
		"-buildmode=plugin",
		"-tags",
		buildTag,
		"-o",
		intermediateOutFile2,
		intermediateGoFile,
	)
	ecmd.Env = make([]string, len(os.Environ()))
	copy(ecmd.Env, os.Environ())
	ecmd.Env = append(
		ecmd.Env,
		"GO111MODULE=on",
	)
	ecmd.Stderr = os.Stderr
	ecmd.Stdout = os.Stdout
	le.Infof("running go compiler with updated unique build id: %s", ecmd.String())
	err = ecmd.Run()
	if err != nil {
		return err
	}

	if preWriteOutFileHook != nil {
		if err := preWriteOutFileHook(outFile, intermediateOutFile2); err != nil {
			return err
		}
	}

	ofile, oerr = os.Open(intermediateOutFile2)
	if oerr != nil {
		return oerr
	}
	defer ofile.Close()

	outFileFd, err := os.OpenFile(outFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer outFileFd.Close()
	defer outFileFd.Sync()

	_, err = io.Copy(outFileFd, ofile)
	return err
}

//go:build !js && !wasm
// +build !js,!wasm

package plugin_compiler

import (
	"context"
	"crypto/sha256"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	debounce_fswatcher "github.com/aperturerobotics/util/debounce-fswatcher"
	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Watcher watches a set of packages and re-generates an output plugin codegen
// and binary when the code files change.
type Watcher struct {
	le                *logrus.Entry
	packageLookupPath string
	packagePaths      []string
}

// NewWatcher constructs a new watcher.
//
// Recognizes and replaces {buildHash} in the output filename.
// The output path should be output-plugin-dir/output-plugin-{buildHash}.cbus.so
func NewWatcher(le *logrus.Entry, packageLookupPath string, packagePaths []string) *Watcher {
	return &Watcher{
		le:                le,
		packagePaths:      packagePaths,
		packageLookupPath: packageLookupPath,
	}
}

// WatchCompilePlugin watches and compiles package.
// Detects if the output with the same {buildHash} already exists.
// Replaces {buildHash} in output filename and in plugin binary version.
func (w *Watcher) WatchCompilePlugin(
	ctx context.Context,
	pluginCodegenPath string,
	pluginOutputPath string,
	pluginBinaryID string,
	pluginBinaryVersion string,
	compiledCb func(packages []string, outpPath string) error,
) error {
	le := w.le

	le.
		WithField("codegen-path", pluginCodegenPath).
		Info("hot: starting to build/watch plugin")

	// hhBaseKey is the seed key to use for blake3 (32 bytes)
	var hhBaseKey []byte
	{
		hs := sha256.New()
		_, _ = hs.Write([]byte(pluginBinaryID))
		_, _ = hs.Write([]byte(pluginOutputPath))
		hhBaseKey = hs.Sum(nil)
	}

	// passOutputPath may or may not contain {buildHash}
	compilePluginOnce := func(
		ctx context.Context,
		an *Analysis,
		buildPrefix string,
		passOutputPath string,
		pluginBinaryVersion string,
	) error {
		moduleCompiler, err := NewModuleCompiler(
			ctx,
			w.le,
			buildPrefix,
			pluginCodegenPath,
			pluginBinaryID,
		)
		if err != nil {
			return err
		}
		if err := moduleCompiler.GenerateModules(an, pluginBinaryVersion); err != nil {
			return err
		}
		err = moduleCompiler.CompilePlugin(passOutputPath)
		moduleCompiler.Cleanup()
		return err
	}

	compilePlugin := func() (*Analysis, error) {
		rctx := ctx
		ctx, compileCtxCancel := context.WithCancel(rctx)
		defer compileCtxCancel()

		le.
			WithField("plugin-output-filename", path.Base(pluginOutputPath)).
			Debugf("analyzing packages: %v", w.packagePaths)
		an, err := AnalyzePackages(ctx, w.le, w.packageLookupPath, w.packagePaths)
		if err != nil {
			return nil, err
		}

		// pass 1: codegen + build with static build prefix.
		passBinDir := filepath.Join(pluginCodegenPath, "bin")
		if err := os.MkdirAll(passBinDir, 0o755); err != nil {
			return nil, err
		}

		pass1OutputPath := filepath.Join(passBinDir, "pass1.cbus.so")
		if err := compilePluginOnce(ctx, an, "pass1", pass1OutputPath, pluginBinaryVersion); err != nil {
			return nil, err
		}

		buildHash, err := HashPluginForBuildID(hhBaseKey, pass1OutputPath)
		if err != nil {
			return nil, err
		}
		buildHashShort := buildHash[:16]
		passBuildPluginVersion := strings.ReplaceAll(pluginBinaryVersion, buildHashConstTag, buildHashShort)

		targetOutputPath := pluginOutputPath
		if strings.Contains(targetOutputPath, buildHashConstTag) {
			targetOutputPath = strings.ReplaceAll(targetOutputPath, buildHashConstTag, buildHashShort)
			if _, err := os.Stat(targetOutputPath); !os.IsNotExist(err) {
				le.Info("detected that output would be identical, skipping")
				return an, nil
			}
			if err := CleanupOldVersions(le, pluginOutputPath); err != nil {
				return an, err
			}
		}

		// pass 2: codegen + build with the build hash
		pass2OutputPath := filepath.Join(passBinDir, "pass2.cbus.so")
		buildPrefix := "cbus-plugin-" + buildHashShort
		if err := compilePluginOnce(ctx, an, buildPrefix, pass2OutputPath, passBuildPluginVersion); err != nil {
			return an, err
		}
		if err := copyFileFromTo(pass2OutputPath, targetOutputPath); err != nil {
			return an, err
		}
		if err == nil && compiledCb != nil {
			err = compiledCb(w.packagePaths, targetOutputPath)
		}

		return an, err
	}

	for {
		an, err := compilePlugin()
		if err != nil {
			return err
		}

		// ignoreWatchPrefixes are prefixes to ignore from watching
		ignoreWatchPrefixes := []string{"vendor", "node_modules", ".bldr", "(disabled)"}

		// build file watchlist
		watchedFiles := make(map[string]struct{})
		codefileMap := an.GetProgramCodeFiles(w.packagePaths, "")
		for _, filePaths := range codefileMap {
		FilePathLoop:
			for _, filePath := range filePaths {
				for _, prefix := range ignoreWatchPrefixes {
					if strings.HasPrefix(filePath, prefix) {
						continue FilePathLoop
					}
				}
				watchedFiles[filePath] = struct{}{}
			}
		}

		watchedSourcePaths := make(map[string]struct{}, len(watchedFiles))
		watchedSourceDirs := make(map[string]struct{}, len(watchedFiles))
		for filePath := range watchedFiles {
			watchedSourcePaths[filePath] = struct{}{}
			sourceDir := filepath.Dir(filePath)
			if _, ok := watchedSourceDirs[sourceDir]; !ok {
				watchedSourceDirs[sourceDir] = struct{}{}
			}
		}

		le.Debugf(
			"hot: watching %d packages with %d files",
			len(w.packagePaths),
			len(watchedFiles),
		)

		// It's best to watch the entire directory tree and filter the events.
		//
		// This is both more efficient on the kernel side and avoids nasty quriks
		// with git and other editors deleting and re-creating files.
		//
		// See fsnotify comments:
		//   Watching individual files (rather than directories) is generally not
		//   recommended as many programs (especially editors) update files atomically: it
		//   will write to a temporary file which is then moved to to destination,
		//   overwriting the original (or some variant thereof). The watcher on the
		//   original file is now lost, as that no longer exists.
		//
		// It's necessary to create one watcher per directory:
		//   https://github.com/fsnotify/fsnotify/issues/18
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			return err
		}

		for watchedDirPath := range watchedSourceDirs {
			le.Infof("watching dir path: %s", watchedDirPath)
			err = watcher.Add(watchedDirPath)
			if err != nil {
				_ = watcher.Close()
				return err
			}
		}

		// wait for a file change
		happened, err := debounce_fswatcher.DebounceFSWatcherEvents(
			ctx,
			watcher,
			time.Millisecond*500,
			func(event fsnotify.Event) (match bool, err error) {
				if _, ok := watchedSourcePaths[event.Name]; !ok {
					return false, nil
				}
				return true, nil
			},
		)
		_ = watcher.Close()
		if err != nil {
			return err
		}

		le.Infof("re-analyzing packages after %d filesystem events", len(happened))
	}
}

// CleanupOldVersions cleans up old versions from the target path.
func CleanupOldVersions(le *logrus.Entry, pluginOutputPath string) error {
	pluginOutputFilename := path.Base(pluginOutputPath)
	pluginOutputDirectory := path.Dir(pluginOutputPath)
	// cleanup old versions
	if !strings.Contains(pluginOutputFilename, buildHashConstTag) {
		return nil
	}
	pts := strings.Split(pluginOutputFilename, buildHashConstTag)
	if len(pts) != 2 {
		return errors.Errorf(
			"expected one instance of %s but found %d",
			buildHashConstTag,
			len(pts)-1,
		)
	}
	filenameRe, err := regexp.Compile(strings.Join([]string{
		// match part 0
		"(", regexp.QuoteMeta(pts[0]), ")",
		// match build id
		"([a-zA-Z0-9]*)",
		// match part 1
		"(", regexp.QuoteMeta(pts[1]), ")",
	}, ""))
	if err != nil {
		return err
	}
	le.
		WithField("filename-match-re", filenameRe.String()).
		Debug("compilation complete, cleaning up old versions")
	// list files in target dir
	dirContents, err := os.ReadDir(pluginOutputDirectory)
	if err != nil {
		return err
	}
	for _, df := range dirContents {
		if df.IsDir() || !df.Type().IsRegular() {
			continue
		}
		dfName := df.Name()
		if !filenameRe.MatchString(dfName) {
			continue
		}
		dfBase := path.Base(dfName)
		if err := os.RemoveAll(path.Join(pluginOutputDirectory, dfBase)); err != nil {
			le.WithError(err).Warn("could not delete old version")
			continue
		} else {
			le.Debugf("deleted old version: %s", dfBase)
		}
	}
	return nil
}

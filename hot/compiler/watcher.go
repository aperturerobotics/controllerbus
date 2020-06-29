package hot_compiler

import (
	"context"
	gast "go/ast"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	debounce_fswatcher "github.com/aperturerobotics/controllerbus/util/debounce-fswatcher"
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
	analyzeEveryTime  bool
}

// NewWatcher constructs a new watcher.
func NewWatcher(le *logrus.Entry, packageLookupPath string, packagePaths []string, analyzeEveryTime bool) *Watcher {
	return &Watcher{
		le:                le,
		packagePaths:      packagePaths,
		packageLookupPath: packageLookupPath,
		analyzeEveryTime:  analyzeEveryTime,
	}
}

// CompilePlugin analyzes and compiles a plugin.
//
// Recognizes and replaces {buildHash} in the output filename.
func CompilePlugin(
	ctx context.Context,
	le *logrus.Entry,
	an *Analysis,
	pluginCodegenPath string,
	pluginOutputPath string,
	pluginBinaryID string,
	pluginBinaryVersion string,
	preWriteOutFileHook func(nextOutFilePath, nextOutFileContentsPath string) error,
) (*gast.File, error) {
	le.Debug("compiling plugin")
	wr, err := GeneratePluginWrapper(
		ctx,
		le,
		an,
		pluginBinaryID,
		pluginBinaryVersion,
	)
	if err != nil {
		return nil, err
	}
	err = CompilePluginFromFile(
		le,
		wr,
		pluginCodegenPath,
		pluginOutputPath,
		preWriteOutFileHook,
	)
	if err != nil {
		return nil, err
	}
	return wr, err
}

// WatchCompilePlugin watches and compiles package.
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
		WithField("plugin-output-filename", path.Base(pluginOutputPath)).
		Debugf("analyzing packages: %v", w.packagePaths)
	an, err := AnalyzePackages(ctx, w.le, w.packageLookupPath, w.packagePaths)
	if err != nil {
		return err
	}

	le.
		WithField("codegen-path", pluginCodegenPath).
		Info("hot: starting to build/watch plugin")
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	compilePlugin := func() error {
		_, err := CompilePlugin(
			ctx,
			le,
			an,
			pluginCodegenPath,
			pluginOutputPath,
			pluginBinaryID,
			pluginBinaryVersion,
			func(nextOutFilePath, nextOutFileContentsPath string) error {
				if err := CleanupOldVersions(le, pluginOutputPath); err != nil {
					return err
				}
				return nil
			},
		)
		if err == nil && compiledCb != nil {
			err = compiledCb(w.packagePaths, pluginOutputPath)
		}
		return err
	}

	watchedFiles := make(map[string]struct{})
	for {
		err = compilePlugin()
		if err != nil {
			return err
		}

		// build file watchlist
		codefileMap := an.GetProgramCodeFiles(w.packagePaths, "")
		nextWatchedFiles := make(map[string]struct{})
		for _, filePaths := range codefileMap {
			for _, filePath := range filePaths {
				nextWatchedFiles[filePath] = struct{}{}
			}
		}
		for filePath := range watchedFiles {
			if _, ok := nextWatchedFiles[filePath]; ok {
				delete(nextWatchedFiles, filePath)
				continue
			}
			le.Debugf("removing watcher for file: %s", filePath)
			if err := watcher.Remove(filePath); err != nil {
				return err
			}
		}
		for filePath := range nextWatchedFiles {
			le.Debugf("adding watcher for file: %s", filePath)
			watchedFiles[filePath] = struct{}{}
			if err := watcher.Add(filePath); err != nil {
				return err
			}
		}

		le.Debugf(
			"hot: watching %d packages with %d files",
			len(w.packagePaths),
			len(watchedFiles),
		)

		// wait for a file change
		for {
			happened, err := debounce_fswatcher.DebounceFSWatcherEvents(
				ctx,
				watcher,
				time.Second,
			)
			if err != nil {
				return err
			}
			le.Infof("re-analyzing packages after %d filesystem events", len(happened))
			if w.analyzeEveryTime {
				break
			}
			// re-sync without re-analyzing
			if err := compilePlugin(); err != nil {
				return err
			}
		}
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
	dirContents, err := ioutil.ReadDir(pluginOutputDirectory)
	if err != nil {
		return err
	}
	for _, df := range dirContents {
		if df.IsDir() || !df.Mode().IsRegular() {
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

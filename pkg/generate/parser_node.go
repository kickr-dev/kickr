package generate

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	engine "github.com/kickr-dev/engine/pkg"
	"github.com/kickr-dev/engine/pkg/files"
	"github.com/kickr-dev/engine/pkg/parser"

	"github.com/kickr-dev/kickr/pkg/generate/types"
)

// ParserNode detects the presence of a ParserNode.js project by looking for a package.json file.
//
// In case of success, the function will set the language to "node"
// and the worker to "main" if the main property is present in the package.json file.
func ParserNode(ctx context.Context, destdir string, config *types.KickrWrapper) error {
	// try to parse nodejs language at base directory
	err := parserNode(ctx, destdir, config)
	if err == nil {
		engine.GetLogger().Infof("node detected, a '%s' is present and valid", parser.FilePackageJSON)
		return nil // parsing was successful with base directory
	}
	if !errors.Is(err, fs.ErrNotExist) {
		return err // already wrapped
	}

	if config.CI == nil || config.CI.Website == nil || config.CI.Website.Directory == "" {
		// nothing to do since there's no website configured or the website directory is the base directory (must be linked to another language)
		return nil
	}

	// try to parse nodejs inside website directory
	if err := parserNode(ctx, filepath.Join(destdir, config.CI.Website.Directory), config); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return err // already wrapped
	}
	engine.GetLogger().Infof("node detected in '%s', a '%s' is present and valid", config.CI.Website.Directory, parser.FilePackageJSON)
	return nil
}

var _ engine.Parser[types.KickrWrapper] = ParserNode // ensure interface is implemented

func parserNode(_ context.Context, destdir string, config *types.KickrWrapper) error { //nolint:revive
	var jsonfile parser.PackageJSON
	if err := files.ReadJSON(filepath.Join(destdir, parser.FilePackageJSON), &jsonfile, os.ReadFile); err != nil {
		return fmt.Errorf("read json: %w", err)
	}

	if err := jsonfile.Validate(); err != nil {
		return fmt.Errorf("validate '%s': %w", parser.FilePackageJSON, err)
	}

	config.SetLanguage("node", jsonfile)
	if jsonfile.Main != nil {
		config.AddWorker("main")
	}
	return nil
}

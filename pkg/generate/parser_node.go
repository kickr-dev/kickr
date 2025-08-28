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
func ParserNode(_ context.Context, destdir string, config *types.KickrGen) error {
	var jsonfile parser.PackageJSON
	jsonpath := filepath.Join(destdir, parser.FilePackageJSON)
	if err := files.ReadJSON(jsonpath, &jsonfile, os.ReadFile); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("read json: %w", err)
	}
	engine.GetLogger().Infof("node detected, a '%s' is present and valid", parser.FilePackageJSON)

	if err := jsonfile.Validate(); err != nil {
		return fmt.Errorf("validate '%s': %w", parser.FilePackageJSON, err)
	}

	config.SetLanguage("node", jsonfile)
	if jsonfile.Main != nil {
		config.Executables.AddWorker("main")
	}
	return nil
}

var _ engine.Parser[types.KickrGen] = ParserNode // ensure interface is implemented

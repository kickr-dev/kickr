package generate

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"

	engine "github.com/kickr-dev/engine/pkg"
	"github.com/kickr-dev/engine/pkg/parser"

	"github.com/kickr-dev/kickr/pkg/generate/types"
)

// ParserGolang detects the presence of a go.mod file
// and adds go.mod parsed configuration as 'go' in languages.
//
// It also detects the presence of main.go files in cmd folder
// and adds them to executables composition.
//
// If a hugo config or theme file is present, it will be detected
// and 'hugo' will be set as the language ('go' will not in that case).
func ParserGolang(_ context.Context, destdir string, config *types.KickrWrapper) error {
	var hasHugo bool
	if hugoc, ok := parser.Hugo(destdir); ok {
		engine.GetLogger().Infof("hugo detected, theme or hugo files are present")
		config.SetLanguage("hugo", hugoc)
		hasHugo = true
	}
	if config.CI != nil && config.CI.Website != nil && config.CI.Website.Directory != "" {
		if hugoc, ok := parser.Hugo(filepath.Join(destdir, config.CI.Website.Directory)); ok {
			engine.GetLogger().Infof("hugo detected in '%s', theme or hugo files are present", config.CI.Website.Directory)

			// note: may override configuration from base directory (to see later if it's problematic)
			// it shouldn't because with hugo we are deploying a website, as such by providing the website directory
			// the user understands that it's this path to use
			config.SetLanguage("hugo", hugoc)

			// no affect of hasHugo here since this could be a website for a documentation in a more global repository with golang
		}
	}
	// in case the base directory is hugo based, we don't continue
	if hasHugo {
		return nil
	}

	gomod, err := parser.ReadGomod(destdir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("read '%s': %w", parser.FileGomod, err)
	}

	engine.GetLogger().Infof("golang detected, file '%s' is present and valid", parser.FileGomod)
	config.SetLanguage("go", gomod)

	executables, err := parser.ReadGoCmd(destdir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("read '%s': %w", parser.FolderCMD, err)
	}

	config.Executables = executables
	return nil
}

var _ engine.Parser[types.KickrWrapper] = ParserGolang // ensure interface is implemented

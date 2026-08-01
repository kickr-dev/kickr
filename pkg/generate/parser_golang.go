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
func ParserGolang(ctx context.Context, destdir string, repo *types.Repository) error {
	root, err := parserHugo(ctx, destdir, repo)
	if err != nil {
		return fmt.Errorf("parse hugo: %w", err)
	}
	if root {
		// there's a hugo module at destdir base path, we should skip checking go.work or go.mod
		// this could evolve in the future depending on end user needs
		return nil
	}

	// read go.work first
	gowork, err := parser.ReadGowork(destdir)
	if err == nil {
		engine.GetLogger().Infof("golang detected, file '%s' is present and valid", parser.FileGowork)
		repo.Module(types.RootModule).SetLanguage(types.LanguageGo, gowork)

		// each 'use' directive declares a go module, it's the only way for kickr to know about them
		for _, use := range gowork.Uses {
			executables, err := parser.ReadGoCmd(filepath.Join(destdir, use.Use))
			if err != nil && !errors.Is(err, fs.ErrNotExist) {
				return fmt.Errorf("read '%s' in '%s': %w", parser.FolderCMD, use.Use, err)
			}
			repo.Module(use.Use).SetLanguage(types.LanguageGo, use.Gomod).SetExecutables(executables)
		}
	} else if !errors.Is(err, parser.ErrNoGowork) {
		return fmt.Errorf("read '%s': %w", parser.FileGowork, err)
	}

	// still, try to read a go.mod (it will override go.work data but it's fine since only Go and Toolchain are used)
	gomod, err := parser.ReadGomod(destdir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("read '%s': %w", parser.FileGomod, err)
	}
	engine.GetLogger().Infof("golang detected, file '%s' is present and valid", parser.FileGomod)
	repo.Module(types.RootModule).SetLanguage(types.LanguageGo, gomod)

	// parse cmd directory only if there's a go.mod for base directory
	executables, err := parser.ReadGoCmd(destdir)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("read '%s': %w", parser.FolderCMD, err)
	}
	repo.Module(types.RootModule).SetExecutables(executables)
	return nil
}

var _ engine.Parser[types.Repository] = ParserGolang // ensure interface is implemented

func parserHugo(_ context.Context, destdir string, repo *types.Repository) (bool, error) {
	var root bool

	// try to parse destdir for hugo
	hugoc, err := parser.ReadHugo(destdir)
	if err == nil {
		engine.GetLogger().Infof("hugo detected, theme or hugo files are present")
		root = true
		repo.Module(types.RootModule).SetLanguage(types.LanguageHugo, hugoc)
	} else if !errors.Is(err, parser.ErrNoHugo) {
		return false, fmt.Errorf("read hugo: %w", err)
	}

	// try to parse website directory for hugo
	if repo.Website != nil && repo.Website.Directory != "" {
		hugoc, err := parser.ReadHugo(filepath.Join(destdir, repo.Website.Directory))
		if err == nil {
			engine.GetLogger().Infof("hugo detected in '%s', theme or hugo files are present", repo.Website.Directory)
			repo.Module(repo.Website.Directory).SetLanguage(types.LanguageHugo, hugoc)
		} else if !errors.Is(err, parser.ErrNoHugo) {
			return false, fmt.Errorf("read hugo in '%s': %w", repo.Website.Directory, err)
		}
	}

	return root, nil
}

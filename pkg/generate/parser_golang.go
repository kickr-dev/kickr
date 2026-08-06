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
	"github.com/kickr-dev/kickr/pkg/kickr/v1"
)

// ParserHugo detects Hugo sites in repository modules and sets their language accordingly.
func ParserHugo(_ context.Context, destdir string, repo *types.Repository) error {
	errs := make([]error, 0, len(repo.Modules))
	for i, module := range repo.Modules {
		hugo, err := parser.ReadHugo(filepath.Join(destdir, module.Dir()))
		if err != nil {
			if !errors.Is(err, parser.ErrNoHugo) {
				errs = append(errs, fmt.Errorf("read hugo in '%s': %w", module.Dir(), err))
			}
			continue
		}

		engine.GetLogger().Infof("hugo detected in '%s', theme or hugo files are present", module.Dir())
		repo.Modules[i].SetLanguage(types.LanguageHugo, hugo)
	}
	return errors.Join(errs...) // already wrapped
}

// ParserGolang detects Golang modules in the repository via go.work and go.mod files.
//
// In case an Hugo configuration exists in the root module,
// Golang parsing is skipped.
func ParserGolang(ctx context.Context, destdir string, repo *types.Repository) error {
	ri := repo.ModuleIndexOf(types.RootModule)
	if ri >= 0 {
		if _, ok := repo.Modules[ri].Languages[types.LanguageHugo]; ok {
			return nil // root module has hugo language, skip Golang parsing
		}
	}
	if ri < 0 {
		return nil // Golang parsing goes exclusively through root, either by go.work indicating all modules or go.mod
	}

	// read go.work first
	if err := gowork(ctx, destdir, repo); err != nil {
		return err // already wrapped
	}
	// still, try to read a go.mod (it will override go.work data but it's fine since only Go and Toolchain are used)
	if err := gomod(ctx, destdir, repo); err != nil {
		return err // already wrapped
	}
	return nil
}

var _ engine.Parser[types.Repository] = ParserGolang // ensure interface is implemented

// gowork reads destdir go.work (if it exists) and its 'uses' go.mod.
func gowork(_ context.Context, destdir string, repo *types.Repository) error {
	ri := repo.ModuleIndexOf(types.RootModule)

	work, err := parser.ReadGowork(destdir)
	if err != nil {
		if !errors.Is(err, parser.ErrNoGowork) {
			return fmt.Errorf("read '%s': %w", parser.FileGowork, err)
		}
		return nil
	}
	engine.GetLogger().Infof("golang detected, file '%s' is present and valid", parser.FileGowork)
	repo.Modules[ri].SetLanguage(types.LanguageGo, work)

	// each 'use' directive declares a go module, it's the only way for kickr to know about them
	errs := make([]error, 0, len(work.Uses))
	for _, use := range work.Uses {
		i := repo.ModuleIndexOf(use.Use)
		if i < 0 {
			repo.Modules = append(repo.Modules, types.Module{
				Config:    kickr.Module{},
				Directory: filepath.Clean(use.Use),
				Parent:    repo,
			})
			i = len(repo.Modules) - 1
		}
		repo.Modules[i].SetLanguage(types.LanguageGo, use.Gomod)

		executables, err := parser.ReadGoCmd(filepath.Join(destdir, use.Use))
		if err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				errs = append(errs, fmt.Errorf("read '%s' in '%s': %w", parser.FolderCMD, use.Use, err))
			}
			continue
		}
		repo.Modules[i].SetExecutables(executables)
	}
	return errors.Join(errs...) // already wrapped
}

// gomod reads destdir go.mod (if it exists) and its cmd directory.
func gomod(_ context.Context, destdir string, repo *types.Repository) error {
	ri := repo.ModuleIndexOf(types.RootModule)

	mod, err := parser.ReadGomod(destdir)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("read '%s': %w", parser.FileGomod, err)
		}
		return nil
	}
	engine.GetLogger().Infof("golang detected, file '%s' is present and valid", parser.FileGomod)
	repo.Modules[ri].SetLanguage(types.LanguageGo, mod)

	// parse cmd directory only if there's a go.mod for base directory
	executables, err := parser.ReadGoCmd(destdir)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("read '%s': %w", parser.FolderCMD, err)
		}
		return nil
	}
	repo.Modules[ri].SetExecutables(executables)
	return nil
}

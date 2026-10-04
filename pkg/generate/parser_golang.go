package generate

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"path/filepath"

	engine "github.com/kickr-dev/engine/pkg"
	"github.com/kickr-dev/engine/pkg/generator"
	"github.com/kickr-dev/engine/pkg/parser"

	"github.com/kickr-dev/kickr/pkg/generate/types"
	"github.com/kickr-dev/kickr/pkg/kickr/v1"
)

// GoImage is the container image repository of Go, whose version digest is pinned by ParserGolang.
const GoImage = "docker.io/library/golang"

// ParserGolang detects Golang modules in the repository via go.work and go.mod files.
//
// In case an Hugo configuration exists in the root module,
// Golang parsing is skipped.
//
// Each module image is fetched from its toolchain (or go) version and pinned by digest.
func ParserGolang(httpClient *http.Client) func(ctx context.Context, destdir string, repo *types.Repository) error {
	if httpClient == nil {
		httpClient = http.DefaultClient //nolint:revive
	}
	return func(ctx context.Context, destdir string, repo *types.Repository) error {
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
		if err := gowork(ctx, httpClient, destdir, repo); err != nil {
			return err // already wrapped
		}
		// still, try to read a go.mod (it will override go.work data but it's fine since only Go and Toolchain are used)
		if err := gomod(ctx, httpClient, destdir, repo); err != nil {
			return err // already wrapped
		}
		return nil
	}
}

var _ engine.Parser[types.Repository] = ParserGolang(nil) // ensure interface is implemented

// gowork reads destdir go.work (if it exists) and its 'uses' go.mod.
func gowork(ctx context.Context, httpClient *http.Client, destdir string, repo *types.Repository) error {
	ri := repo.ModuleIndexOf(types.RootModule)

	work, err := parser.ReadGowork(destdir)
	if err != nil {
		if !errors.Is(err, parser.ErrNoGowork) {
			return fmt.Errorf("read '%s': %w", parser.FileGowork, err)
		}
		return nil
	}
	engine.GetLogger().Infof("golang detected, file '%s' is present and valid", parser.FileGowork)

	image, err := goImage(ctx, httpClient, work.Go, work.Toolchain)
	if err != nil {
		engine.GetLogger().Warnf("failed to fetch golang image for '%s', using default one: %v", parser.FileGowork, err)
	}
	work.ContainerImage = image
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
func gomod(ctx context.Context, httpClient *http.Client, destdir string, repo *types.Repository) error {
	ri := repo.ModuleIndexOf(types.RootModule)

	mod, err := parser.ReadGomod(destdir)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("read '%s': %w", parser.FileGomod, err)
		}
		return nil
	}
	engine.GetLogger().Infof("golang detected, file '%s' is present and valid", parser.FileGomod)

	image, err := goImage(ctx, httpClient, mod.Go, mod.Toolchain)
	if err != nil {
		engine.GetLogger().Warnf("failed to fetch golang image for '%s', using default one: %v", parser.FileGomod, err)
	}
	mod.ContainerImage = image
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

// goImage fetches the golang image matching toolchain (or goversion when toolchain isn't set) and returns it pinned by digest.
func goImage(ctx context.Context, httpClient *http.Client, goversion, toolchain string) (string, error) {
	version := goversion
	if toolchain != "" && toolchain != "default" {
		version = toolchain
	}
	image, err := generator.FetchContainerImage(ctx, httpClient, GoImage, version+"-trixie")
	if err != nil {
		return "", fmt.Errorf("fetch image: %w", err)
	}
	return image, nil
}

package generate

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"

	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	engine "github.com/kickr-dev/engine/pkg"

	"github.com/kickr-dev/kickr/pkg/generate/types"
)

// TerraformModule represents a terraform module.
//
// Parsing mainly comes from terraform-config-inspect library.
type TerraformModule struct {
	*tfconfig.Module

	Backend string
}

var backends = []string{"http", "s3"}

// var moduleRegexp = regexp.MustCompile(`terraform-\S+-\S+`)

// ParserTerraform detects the presence of a terraform module with its 'main.tf' at destdir base directory
// or within specified modules (by config.Terraform.Modules) with their own 'main.tf'.
//
// In case a module is specified by the configuration but doesn't contain a 'main.tf' then it will return an error.
//
// All successful module parses are added into config languages.
func ParserTerraform(_ context.Context, destdir string, config *types.Repository) error {
	// terraform module at root
	if tfconfig.IsModuleDir(destdir) {
		tfmodule, dErr := tfconfig.LoadModule(destdir)
		if dErr != nil {
			return fmt.Errorf("load module: %w", dErr)
		}

		backend, err := terraformBackend(destdir)
		if err != nil {
			engine.GetLogger().Warnf("failed to read backend type: %s", err.Error())
		}
		if backend != "" && !slices.Contains(backends, backend) {
			engine.GetLogger().Warnf("backend '%s' doesn't have an associated behavior", backend)
		}

		config.SetLanguage("terraform", []types.Mono[TerraformModule]{{
			Directory: ".",
			Specifics: TerraformModule{Module: tfmodule, Backend: backend},
		}})
	}

	// no terraform modules specified
	if config.Terraform == nil || len(config.Terraform.Modules) == 0 {
		return nil
	}

	var (
		errs    = make([]error, 0, len(config.Terraform.Modules))
		modules = make([]types.Mono[TerraformModule], 0, len(config.Terraform.Modules))
	)
	for _, module := range config.Terraform.Modules {
		if !tfconfig.IsModuleDir(filepath.Join(destdir, module)) {
			errs = append(errs, fmt.Errorf("module '%s' isn't a terraform module", module))
			continue
		}

		tfmodule, dErr := tfconfig.LoadModule(filepath.Join(destdir, module))
		if dErr != nil {
			errs = append(errs, fmt.Errorf("load module '%s': %w", module, dErr))
			continue
		}

		backend, err := terraformBackend(filepath.Join(destdir, module))
		if err != nil {
			engine.GetLogger().Warnf("failed to read backend type: %s", err.Error())
		}
		if backend != "" && !slices.Contains(backends, backend) {
			engine.GetLogger().Warnf("backend '%s' doesn't have an associated behavior", backend)
		}

		modules = append(modules, types.Mono[TerraformModule]{
			Directory: module,
			Specifics: TerraformModule{Module: tfmodule, Backend: backend},
		})
	}
	if err := errors.Join(errs...); err != nil {
		return err // already wrapped
	}

	if len(modules) > 0 {
		config.SetLanguage("terraform", modules)
	}
	return nil
}

var _ engine.Parser[types.Repository] = ParserTerraform // ensure interface is implemented

var backendRegexp = regexp.MustCompile(`backend "(\S+)" {`)

func terraformBackend(destdir string) (string, error) {
	for _, file := range []string{"backend.tf", "state.tf"} {
		bytes, err := os.ReadFile(filepath.Join(destdir, file))
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return "", fmt.Errorf("read file: %w", err)
		}

		if matches := backendRegexp.FindSubmatch(bytes); len(matches) > 1 {
			return string(matches[1]), nil
		}
	}

	entries, err := os.ReadDir(destdir)
	if err != nil {
		return "", fmt.Errorf("read dir: %w", err)
	}

	errs := make([]error, 0, len(entries))
	for _, entry := range entries {
		bytes, err := os.ReadFile(filepath.Join(destdir, entry.Name()))
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			errs = append(errs, fmt.Errorf("read file: %w", err))
			continue
		}

		if matches := backendRegexp.FindSubmatch(bytes); len(matches) > 1 {
			return string(matches[1]), nil
		}
	}
	return "", errors.Join(errs...)
}

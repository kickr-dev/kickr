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

// ParserTerraform detects the presence of a terraform module with its 'main.tf' at destdir base directory
// or within specified modules (by config.Terraform.Modules) with their own 'main.tf'.
//
// In case a module is specified by the configuration but doesn't contain a 'main.tf' then it will return an error.
//
// All successful module parses are added into config modules.
func ParserTerraform(_ context.Context, destdir string, repo *types.Repository) error {
	if tfconfig.IsModuleDir(destdir) {
		tfmodule, diags := tfconfig.LoadModule(destdir)
		if diags.HasErrors() {
			return fmt.Errorf("load module '%s': %w", filepath.Base(destdir), diags)
		}

		backend, err := terraformBackend(destdir)
		if err != nil {
			engine.GetLogger().Warnf("failed to read backend type: %s", err.Error())
		}
		if backend != "" && !slices.Contains(backends, backend) {
			engine.GetLogger().Warnf("backend '%s' doesn't have an associated behavior", backend)
		}
		repo.Module(types.RootModule).SetLanguage(types.LanguageTerraform, TerraformModule{Module: tfmodule, Backend: backend})
	}

	// no terraform modules specified
	if repo.Terraform == nil || len(repo.Terraform.Modules) == 0 {
		return nil
	}

	errs := make([]error, 0, len(repo.Terraform.Modules))
	for _, directory := range repo.Terraform.Modules {
		moduledir := filepath.Join(destdir, directory)
		if !tfconfig.IsModuleDir(moduledir) {
			errs = append(errs, fmt.Errorf("module '%s' isn't a terraform module", directory))
			continue
		}

		tfmodule, diags := tfconfig.LoadModule(moduledir)
		if diags.HasErrors() {
			errs = append(errs, fmt.Errorf("load module '%s': %w", directory, diags))
			continue
		}

		backend, err := terraformBackend(moduledir)
		if err != nil {
			engine.GetLogger().Warnf("failed to read backend type: %s", err.Error())
		}
		if backend != "" && !slices.Contains(backends, backend) {
			engine.GetLogger().Warnf("backend '%s' doesn't have an associated behavior", backend)
		}

		repo.Module(directory).SetLanguage(types.LanguageTerraform, TerraformModule{Module: tfmodule, Backend: backend})
	}
	return errors.Join(errs...) // already wrapped
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

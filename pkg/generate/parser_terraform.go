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

// ParserTerraform detects the presence of a terraform module with its 'main.tf' in every repository module.
//
// A module declaring 'terraform:' but whose directory isn't a terraform module returns an error.
// A module declaring neither is scanned silently: present means terraform, absent means not.
//
// All successful module parses are added into config modules.
func ParserTerraform(_ context.Context, destdir string, repo *types.Repository) error {
	errs := make([]error, 0, len(repo.Modules))
	for i, module := range repo.Modules {
		if module.Config.Terraform == nil {
			continue
		}

		moduledir := filepath.Join(destdir, module.Dir())
		if !tfconfig.IsModuleDir(moduledir) {
			errs = append(errs, fmt.Errorf("module '%s' isn't a terraform module", module.Dir()))
			continue
		}

		tfmodule, diags := tfconfig.LoadModule(moduledir)
		if diags.HasErrors() {
			errs = append(errs, fmt.Errorf("load module '%s': %w", module.Dir(), diags))
			continue
		}

		backend, err := terraformBackend(moduledir)
		if err != nil {
			engine.GetLogger().Warnf("failed to read backend type: %s", err.Error())
		}
		if backend != "" && !slices.Contains(backends, backend) {
			engine.GetLogger().Warnf("backend '%s' doesn't have an associated behavior", backend)
		}

		repo.Modules[i].SetLanguage(types.LanguageTerraform, TerraformModule{Module: tfmodule, Backend: backend})
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

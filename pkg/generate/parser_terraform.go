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
	"strings"

	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	engine "github.com/kickr-dev/engine/pkg"

	"github.com/kickr-dev/kickr/pkg/generate/types"
)

// TerraformModule represents a terraform module.
//
// Parsing mainly comes from terraform-config-inspect library.
type TerraformModule struct {
	*tfconfig.Module

	PublishName     string
	PublishProvider string

	StateBackend string
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

		// publish name and provider are derived from the repository name, naming convention is 'terraform-<provider>-<name>',
		// where 'provider' is the terraform provider on which the module applies
		// and 'name' is the module name.
		//
		// since a terraform module can include and publish nested modules under 'modules' directory,
		// such modules are suffixed with their directory name (e.g. 'terraform-aws-vpc-subnet' for a module in 'modules/subnet').
		//
		// see https://developer.hashicorp.com/terraform/registry/modules/publish#requirements
		// see https://developer.hashicorp.com/terraform/language/modules/develop/structure
		var name, provider string
		if module.Config.Terraform.Publish != "" {
			provider, name = "local", repo.VCS.ProjectName
			if rest, ok := strings.CutPrefix(repo.VCS.ProjectName, "terraform-"); ok {
				provider, name, _ = strings.Cut(rest, "-")
			}
			if module.Dir() != types.RootModule {
				name += "-" + engine.ToSlug(strings.TrimPrefix(module.Dir(), "modules/"))
			}
		}

		// backend detection only applies to module with an apply step
		var backend string
		if module.Config.HasTerraformApply() {
			var err error
			backend, err = terraformBackend(moduledir)
			if err != nil {
				engine.GetLogger().Warnf("failed to read backend type: %s", err.Error())
			}
			if backend != "" && !slices.Contains(backends, backend) {
				engine.GetLogger().Warnf("backend '%s' doesn't have an associated behavior", backend)
			}
		}

		repo.Modules[i].SetLanguage(types.LanguageTerraform, TerraformModule{
			Module:          tfmodule,
			StateBackend:    backend,
			PublishName:     name,
			PublishProvider: provider,
		})
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

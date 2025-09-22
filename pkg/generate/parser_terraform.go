package generate

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	engine "github.com/kickr-dev/engine/pkg"

	"github.com/kickr-dev/kickr/pkg/generate/types"
)

// var moduleRegexp = regexp.MustCompile(`terraform-\S+-\S+`)

// ParserTerraform detects the presence of a terraform module with its 'main.tf' at destdir base directory
// or within specified modules (by config.Terraform.Modules) with their own 'main.tf'.
//
// In case a module is specified by the configuration but doesn't contain a 'main.tf' then it will return an error.
//
// All successful module parses are added into config languages.
func ParserTerraform(_ context.Context, destdir string, config *types.KickrWrapper) error {
	// terraform module at root, nothing more to do
	if tfconfig.IsModuleDir(destdir) {
		tfmodule, err := tfconfig.LoadModule(destdir)
		if err != nil {
			return fmt.Errorf("load module: %w", err)
		}

		config.SetLanguage("terraform", []types.Mono[*tfconfig.Module]{{
			Directory: ".",
			Specifics: tfmodule,
		}})
	}

	// no terraform modules specified
	if config.Terraform == nil || len(config.Terraform.Modules) == 0 {
		return nil
	}

	var (
		errs    = make([]error, 0, len(config.Terraform.Modules))
		modules = make([]types.Mono[*tfconfig.Module], 0, len(config.Terraform.Modules))
	)
	for _, module := range config.Terraform.Modules {
		if !tfconfig.IsModuleDir(filepath.Join(destdir, module)) {
			errs = append(errs, fmt.Errorf("module '%s' isn't a terraform module", module))
			continue
		}

		tfmodule, err := tfconfig.LoadModule(filepath.Join(destdir, module))
		if err != nil {
			errs = append(errs, fmt.Errorf("load module '%s': %w", module, err))
			continue
		}
		modules = append(modules, types.Mono[*tfconfig.Module]{Directory: module, Specifics: tfmodule})
	}
	if err := errors.Join(errs...); err != nil {
		return err // already wrapped
	}

	if len(modules) > 0 {
		config.SetLanguage("terraform", modules)
	}
	return nil
}

var _ engine.Parser[types.KickrWrapper] = ParserTerraform // ensure interface is implemented

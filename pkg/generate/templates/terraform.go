package templates

import (
	engine "github.com/kickr-dev/engine/pkg"

	"github.com/kickr-dev/kickr/pkg/generate/types"
)

// Terraform returns the slice of templates related to Terraform / OpenTofu generation (terraform-docs, tflint).
func Terraform() []engine.Template[types.Repository] {
	// Terraform wasn't parsed during parsers processing
	noTerraform := func(repo types.Repository) bool {
		return len(repo.ModulesWith(types.LanguageTerraform)) == 0
	}

	return []engine.Template[types.Repository]{
		{
			Delimiters: engine.DelimitersChevron(),
			Globs:      []string{".terraform-docs.yml" + engine.TmplExtension},
			Out:        ".terraform-docs.yml",
			Remove: func(repo types.Repository) bool {
				return len(repo.ModulesWithTerraformPublish()) == 0
			},
		},
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      []string{".tflint.hcl" + engine.TmplExtension},
			Out:        ".tflint.hcl",
			Remove:     noTerraform,
		},
	}
}

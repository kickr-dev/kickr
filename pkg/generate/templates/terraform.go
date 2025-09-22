package templates

import (
	engine "github.com/kickr-dev/engine/pkg"

	"github.com/kickr-dev/kickr/pkg/generate/types"
)

// Terraform returns the slice of templates related to Terraform / OpenTofu generation (tflint).
func Terraform() []engine.Template[types.KickrWrapper] {
	// Terraform wasn't parsed during parsers processing
	noTerraform := func(config types.KickrWrapper) bool {
		_, ok := config.Languages["terraform"]
		return !ok
	}

	return []engine.Template[types.KickrWrapper]{
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      []string{".tflint.hcl" + engine.TmplExtension},
			Out:        ".tflint.hcl",
			Remove:     noTerraform,
		},
	}
}

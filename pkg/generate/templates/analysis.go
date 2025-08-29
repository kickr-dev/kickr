package templates

import (
	"slices"

	engine "github.com/kickr-dev/engine/pkg"
	"github.com/kickr-dev/engine/pkg/parser"

	"github.com/kickr-dev/kickr/pkg/generate/types"
	kickr "github.com/kickr-dev/kickr/pkg/kickr/v1"
)

// CodeCov returns the slice of templates related to codecov configuration.
func CodeCov() []engine.Template[types.KickrWrapper] {
	name := ".codecov.yml"
	return []engine.Template[types.KickrWrapper]{
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      []string{name + engine.TmplExtension},
			Out:        name,
			Remove: func(config types.KickrWrapper) bool {
				return !config.IsCI(parser.GitHub) || !slices.Contains(config.CI.Options, kickr.OptionCodeCov)
			},
		},
	}
}

// Sonar returns the slice of templates related to SonarCloud / SonarQube configuration.
func Sonar() []engine.Template[types.KickrWrapper] {
	name := "sonar.properties"
	return []engine.Template[types.KickrWrapper]{
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      []string{name + engine.TmplExtension},
			Out:        name,
			Remove: func(config types.KickrWrapper) bool {
				return config.CI == nil || !slices.Contains(config.CI.Options, kickr.OptionSonarQube)
			},
		},
	}
}

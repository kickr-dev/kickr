package templates

import (
	"slices"

	engine "github.com/kickr-dev/engine/pkg"
	"github.com/kickr-dev/engine/pkg/parser"

	kickr "github.com/kickr-dev/kickr/pkg/configuration"
)

// CodeCov returns the slice of templates related to codecov configuration.
func CodeCov() []engine.Template[kickr.Config] {
	name := ".codecov.yml"
	return []engine.Template[kickr.Config]{
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      []string{name + engine.TmplExtension},
			Out:        name,
			Remove: func(config kickr.Config) bool {
				return !config.IsCI(parser.GitHub) || !slices.Contains(config.CI.Options, kickr.CodeCov)
			},
		},
	}
}

// Sonar returns the slice of templates related to SonarCloud / SonarQube configuration.
func Sonar() []engine.Template[kickr.Config] {
	name := "sonar.properties"
	return []engine.Template[kickr.Config]{
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      []string{name + engine.TmplExtension},
			Out:        name,
			Remove: func(config kickr.Config) bool {
				return config.CI == nil || !slices.Contains(config.CI.Options, kickr.Sonar)
			},
		},
	}
}

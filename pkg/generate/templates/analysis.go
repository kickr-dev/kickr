package templates

import (
	"slices"

	engine "github.com/kickr-dev/engine/pkg"

	"github.com/kickr-dev/kickr/pkg/generate/types"
	kickr "github.com/kickr-dev/kickr/pkg/kickr/v1"
)

// CodeCov returns the slice of templates related to codecov configuration.
func CodeCov() []engine.Template[types.Repository] {
	name := ".codecov.yml"
	return []engine.Template[types.Repository]{
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      []string{name + engine.TmplExtension},
			Out:        name,
			Remove: func(repo types.Repository) bool {
				return repo.Config.GitHub == nil || !slices.Contains(repo.Config.GitHub.Options, kickr.GitHubOptionsCodecov)
			},
		},
	}
}

// Sonar returns the slice of templates related to SonarCloud / SonarQube configuration.
func Sonar() []engine.Template[types.Repository] {
	name := "sonar.properties"
	return []engine.Template[types.Repository]{
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      []string{name + engine.TmplExtension},
			Out:        name,
			Remove: func(repo types.Repository) bool {
				return !repo.Config.HasSonarQube()
			},
		},
	}
}

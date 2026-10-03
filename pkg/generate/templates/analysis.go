package templates

import (
	"path"
	"slices"

	engine "github.com/kickr-dev/engine/pkg"

	"github.com/kickr-dev/kickr/pkg/generate/types"
	kickr "github.com/kickr-dev/kickr/pkg/kickr/v1"
)

// CodeCov returns the slice of templates related to codecov configuration.
func CodeCov() []engine.Template[types.Repository] {
	return []engine.Template[types.Repository]{
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      []string{".codecov.yml" + engine.TmplExtension},
			Out:        ".codecov.yml",
			Remove: func(repo types.Repository) bool {
				return repo.Config.GitHub == nil || !slices.Contains(repo.Config.GitHub.Options, kickr.GitHubOptionsCodecov)
			},
		},
	}
}

// Plumber returns the slice of templates related to Plumber configuration.
func Plumber() []engine.Template[types.Repository] {
	return []engine.Template[types.Repository]{
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      []string{path.Join(".github", "plumber.yml"+engine.TmplExtension)},
			Out:        path.Join(".github", "plumber.yml"),
			Remove: func(repo types.Repository) bool {
				return repo.Config.GitHub == nil || slices.Contains(repo.Config.GitHub.Exclude, kickr.GitHubExcludePlumber)
			},
		},
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      []string{path.Join(".gitlab", "plumber.yml"+engine.TmplExtension)},
			Out:        path.Join(".gitlab", "plumber.yml"),
			Remove: func(repo types.Repository) bool {
				return repo.Config.GitLab == nil || slices.Contains(repo.Config.GitLab.Exclude, kickr.GitLabExcludePlumber)
			},
		},
	}
}

// Sonar returns the slice of templates related to SonarCloud / SonarQube configuration.
func Sonar() []engine.Template[types.Repository] {
	return []engine.Template[types.Repository]{
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      []string{"sonar.properties" + engine.TmplExtension},
			Out:        "sonar.properties",
			Remove: func(repo types.Repository) bool {
				return !repo.Config.HasSonarQube()
			},
		},
	}
}

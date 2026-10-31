package templates

import (
	"path"
	"slices"

	engine "gitlab.com/kickr-dev/engine/pkg"

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
			GeneratePolicy: engine.RemoveIf(func(repo types.Repository) bool {
				return repo.Config.GitHub == nil || !slices.Contains(repo.Config.GitHub.Options, kickr.GitHubOptionsCodecov)
			}, engine.PolicyDefault),
		},
	}
}

// Plumber returns the slice of templates related to Plumber configuration and CI/CD.
func Plumber() []engine.Template[types.Repository] {
	return slices.Concat(githubPlumber(), gitlabPlumber())
}

func githubPlumber() []engine.Template[types.Repository] {
	return []engine.Template[types.Repository]{
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      []string{path.Join(".github", "plumber.yml"+engine.TmplExtension)},
			Out:        path.Join(".github", "plumber.yml"),
			GeneratePolicy: engine.RemoveIf(func(repo types.Repository) bool {
				return repo.Config.GitHub == nil || slices.Contains(repo.Config.GitHub.Exclude, kickr.GitHubExcludePlumber)
			}, engine.PolicyDefault),
		},
		{
			Delimiters: engine.DelimitersChevron(),
			Globs:      []string{path.Join(".github", "workflows", "plumber.yml"+engine.TmplExtension)},
			Out:        path.Join(".github", "workflows", "plumber.yml"),
			GeneratePolicy: engine.RemoveIf(func(repo types.Repository) bool {
				return repo.Config.GitHub == nil || slices.Contains(repo.Config.GitHub.Exclude, kickr.GitHubExcludePlumber)
			}, engine.PolicyDefault),
		},
	}
}

func gitlabPlumber() []engine.Template[types.Repository] {
	return []engine.Template[types.Repository]{
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      []string{path.Join(".gitlab", "pipelines", "plumber.yml"+engine.TmplExtension)},
			Out:        path.Join(".gitlab", "pipelines", "plumber.yml"),
			GeneratePolicy: engine.RemoveIf(func(repo types.Repository) bool {
				return repo.Config.GitLab == nil || slices.Contains(repo.Config.GitLab.Exclude, kickr.GitLabExcludePlumber)
			}, engine.PolicyDefault),
		},
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      []string{path.Join(".gitlab", "plumber.yml"+engine.TmplExtension)},
			Out:        path.Join(".gitlab", "plumber.yml"),
			GeneratePolicy: engine.RemoveIf(func(repo types.Repository) bool {
				return repo.Config.GitLab == nil || slices.Contains(repo.Config.GitLab.Exclude, kickr.GitLabExcludePlumber)
			}, engine.PolicyDefault),
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
			GeneratePolicy: engine.RemoveIf(func(repo types.Repository) bool {
				return !repo.Config.HasSonarQube()
			}, engine.PolicyDefault),
		},
	}
}

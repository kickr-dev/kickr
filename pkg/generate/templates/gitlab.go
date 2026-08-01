package templates

import (
	"path"
	"slices"

	engine "github.com/kickr-dev/engine/pkg"

	"github.com/kickr-dev/kickr/pkg/generate/types"
	"github.com/kickr-dev/kickr/pkg/kickr/v1"
)

// GitLab returns the slice of templates related to GitLab configuration.
func GitLab() (templates []engine.Template[types.Repository]) {
	gitlabci := path.Join(".gitlab-ci.yml")
	templates = append(templates, engine.Template[types.Repository]{
		Delimiters: engine.DelimitersBracket(),
		Globs:      []string{gitlabci + engine.TmplExtension},
		Out:        gitlabci,
		Remove:     func(repo types.Repository) bool { return repo.GitLab == nil },
	})

	semrel := path.Join(".gitlab", "pipelines", "semantic-release.yml")
	templates = append(templates, engine.Template[types.Repository]{
		Delimiters: engine.DelimitersBracket(),
		Globs:      []string{semrel + engine.TmplExtension},
		Out:        semrel,
		Remove:     func(repo types.Repository) bool { return repo.GitLab == nil },
	})

	deployment := path.Join(".gitlab", "pipelines", "deployment.yml")
	templates = append(templates, engine.Template[types.Repository]{
		Delimiters: engine.DelimitersBracket(),
		Globs:      engine.GlobsWithPart(deployment),
		Out:        deployment,
		Remove: func(repo types.Repository) bool {
			if repo.GitLab == nil {
				return true
			}
			return !repo.HasDocker() && repo.Helm == nil && !repo.HasLanguage(types.LanguageTerraform) && repo.Website == nil //nolint:revive
		},
	})

	deploymentOverrides := path.Join(".gitlab", "deployment.overrides.yml")
	templates = append(templates, engine.Template[types.Repository]{
		Delimiters: engine.DelimitersBracket(),
		Globs:      []string{deploymentOverrides + engine.TmplExtension},
		Out:        deploymentOverrides,
		Remove: func(repo types.Repository) bool {
			if repo.GitLab == nil || !slices.Contains(repo.GitLab.Options, kickr.GitLabOptionsOverridesDeployment) {
				return true
			}
			return !repo.HasDocker() && repo.Helm == nil && !repo.HasLanguage(types.LanguageTerraform) && repo.Website == nil //nolint:revive
		},
	})

	integration := path.Join(".gitlab", "pipelines", "integration.yml")
	templates = append(templates, engine.Template[types.Repository]{
		Delimiters: engine.DelimitersBracket(),
		Globs:      []string{integration + engine.TmplExtension},
		Out:        integration,
		Remove:     func(repo types.Repository) bool { return repo.GitLab == nil },
	})

	integrationOverrides := path.Join(".gitlab", "integration.overrides.yml")
	templates = append(templates, engine.Template[types.Repository]{
		Delimiters: engine.DelimitersBracket(),
		Globs:      []string{integrationOverrides + engine.TmplExtension},
		Out:        integrationOverrides,
		Remove: func(repo types.Repository) bool {
			return repo.GitLab == nil || !slices.Contains(repo.GitLab.Options, kickr.GitLabOptionsOverridesIntegration)
		},
	})

	kickrp := path.Join(".gitlab", "pipelines", "kickr.yml")
	templates = append(templates, engine.Template[types.Repository]{
		Delimiters: engine.DelimitersBracket(),
		Globs:      []string{kickrp + engine.TmplExtension},
		Out:        kickrp,
		Remove: func(repo types.Repository) bool {
			return repo.GitLab == nil || !slices.Contains(repo.GitLab.Options, kickr.GitLabOptionsKickr)
		},
	})

	variables := path.Join(".gitlab", "variables.yml")
	templates = append(templates, engine.Template[types.Repository]{
		Delimiters: engine.DelimitersBracket(),
		Globs:      []string{variables + engine.TmplExtension},
		Out:        variables,
		Remove:     func(repo types.Repository) bool { return repo.GitLab == nil },
	})

	return templates
}

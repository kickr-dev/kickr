package templates

import (
	"path"
	"slices"

	engine "gitlab.com/kickr-dev/engine/pkg"

	"github.com/kickr-dev/kickr/pkg/generate/types"
	"github.com/kickr-dev/kickr/pkg/kickr/v1"
)

// GitLab returns the slice of templates related to GitLab configuration.
func GitLab() []engine.Template[types.Repository] {
	return slices.Concat(gitlabPipelines(), gitlabConfig())
}

func gitlabPipelines() (templates []engine.Template[types.Repository]) {
	hasDeployment := func(repo types.Repository) bool {
		return repo.Config.Helm != nil || repo.HasDocker() || len(repo.ModulesWith(types.LanguageTerraform)) > 0 ||
			len(repo.ModulesWithDeployment(kickr.DeploymentTargetNetlify)) > 0 || len(repo.ModulesWithDeployment(kickr.DeploymentTargetPages)) > 0
	}

	semrel := path.Join(".gitlab", "pipelines", "semantic-release.yml")
	templates = append(templates, engine.Template[types.Repository]{
		Delimiters:     engine.DelimitersBracket(),
		Globs:          []string{semrel + engine.TmplExtension},
		Out:            semrel,
		GeneratePolicy: engine.RemoveIf(func(repo types.Repository) bool { return repo.Config.GitLab == nil }, engine.PolicyDefault),
	})

	deployment := path.Join(".gitlab", "pipelines", "deployment.yml")
	templates = append(templates, engine.Template[types.Repository]{
		Delimiters: engine.DelimitersBracket(),
		Globs:      engine.GlobsWithPart(deployment),
		Out:        deployment,
		GeneratePolicy: engine.RemoveIf(func(repo types.Repository) bool {
			return repo.Config.GitLab == nil || !hasDeployment(repo)
		}, engine.PolicyDefault),
	})

	deploymentOverrides := path.Join(".gitlab", "deployment.overrides.yml")
	templates = append(templates, engine.Template[types.Repository]{
		Delimiters: engine.DelimitersBracket(),
		Globs:      []string{deploymentOverrides + engine.TmplExtension},
		Out:        deploymentOverrides,
		GeneratePolicy: func(out string, repo types.Repository) engine.GeneratePolicy {
			if repo.Config.GitLab == nil || !slices.Contains(repo.Config.GitLab.Options, kickr.GitLabOptionsOverridesDeployment) || !hasDeployment(repo) {
				return engine.PolicyRemove
			}
			return engine.SkipIfExists(out, engine.PolicyDefault)
		},
	})

	integration := path.Join(".gitlab", "pipelines", "integration.yml")
	templates = append(templates, engine.Template[types.Repository]{
		Delimiters:     engine.DelimitersBracket(),
		Globs:          []string{integration + engine.TmplExtension},
		Out:            integration,
		GeneratePolicy: engine.RemoveIf(func(repo types.Repository) bool { return repo.Config.GitLab == nil }, engine.PolicyDefault),
	})

	integrationOverrides := path.Join(".gitlab", "integration.overrides.yml")
	templates = append(templates, engine.Template[types.Repository]{
		Delimiters: engine.DelimitersBracket(),
		Globs:      []string{integrationOverrides + engine.TmplExtension},
		Out:        integrationOverrides,
		GeneratePolicy: func(out string, repo types.Repository) engine.GeneratePolicy {
			if repo.Config.GitLab == nil || !slices.Contains(repo.Config.GitLab.Options, kickr.GitLabOptionsOverridesIntegration) {
				return engine.PolicyRemove
			}
			return engine.SkipIfExists(out, engine.PolicyDefault)
		},
	})

	kickrp := path.Join(".gitlab", "pipelines", "kickr.yml")
	templates = append(templates, engine.Template[types.Repository]{
		Delimiters: engine.DelimitersBracket(),
		Globs:      []string{kickrp + engine.TmplExtension},
		Out:        kickrp,
		GeneratePolicy: engine.RemoveIf(func(repo types.Repository) bool {
			return repo.Config.GitLab == nil || !slices.Contains(repo.Config.GitLab.Options, kickr.GitLabOptionsKickr)
		}, engine.PolicyDefault),
	})

	return templates
}

func gitlabConfig() (templates []engine.Template[types.Repository]) {
	gitlabci := path.Join(".gitlab-ci.yml")
	templates = append(templates, engine.Template[types.Repository]{
		Delimiters:     engine.DelimitersBracket(),
		Globs:          []string{gitlabci + engine.TmplExtension},
		Out:            gitlabci,
		GeneratePolicy: engine.RemoveIf(func(repo types.Repository) bool { return repo.Config.GitLab == nil }, engine.PolicyDefault),
	})

	variables := path.Join(".gitlab", "variables.yml")
	templates = append(templates, engine.Template[types.Repository]{
		Delimiters:     engine.DelimitersBracket(),
		Globs:          []string{variables + engine.TmplExtension},
		Out:            variables,
		GeneratePolicy: engine.RemoveIf(func(repo types.Repository) bool { return repo.Config.GitLab == nil }, engine.PolicyDefault),
	})

	return templates
}

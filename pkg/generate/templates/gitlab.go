package templates

import (
	"path"
	"slices"

	engine "github.com/kickr-dev/engine/pkg"
	"github.com/kickr-dev/engine/pkg/parser"

	"github.com/kickr-dev/kickr/pkg/generate/types"
	"github.com/kickr-dev/kickr/pkg/kickr/v1"
)

// GitLab returns the slice of templates related to GitLab configuration.
func GitLab() []engine.Template[types.Repository] {
	var templates []engine.Template[types.Repository]

	gitlabci := path.Join(".gitlab-ci.yml")
	templates = append(templates, engine.Template[types.Repository]{
		Delimiters: engine.DelimitersBracket(),
		Globs:      []string{gitlabci + engine.TmplExtension},
		Out:        gitlabci,
		Remove:     func(config types.Repository) bool { return !config.IsCI(parser.GitLab) },
	})

	semrel := path.Join(".gitlab", "pipelines", "semantic-release.yml")
	templates = append(templates, engine.Template[types.Repository]{
		Delimiters: engine.DelimitersBracket(),
		Globs:      []string{semrel + engine.TmplExtension},
		Out:        semrel,
		Remove:     func(config types.Repository) bool { return !config.IsCI(parser.GitLab) },
	})

	deployment := path.Join(".gitlab", "pipelines", "deployment.yml")
	templates = append(templates, engine.Template[types.Repository]{
		Delimiters: engine.DelimitersBracket(),
		Globs:      engine.GlobsWithPart(deployment),
		Out:        deployment,
		Remove: func(config types.Repository) bool {
			return !config.IsCI(parser.GitLab) || !config.HasDeployment()
		},
	})

	integration := path.Join(".gitlab", "pipelines", "integration.yml")
	templates = append(templates, engine.Template[types.Repository]{
		Delimiters: engine.DelimitersBracket(),
		Globs:      []string{integration + engine.TmplExtension},
		Out:        integration,
		Remove:     func(config types.Repository) bool { return !config.IsCI(parser.GitLab) },
	})

	dependencies := path.Join(".gitlab", "pipelines", "dependencies.yml")
	templates = append(templates, engine.Template[types.Repository]{
		Delimiters: engine.DelimitersBracket(),
		Globs:      []string{dependencies + engine.TmplExtension},
		Out:        dependencies,
		Remove: func(config types.Repository) bool {
			return !config.IsCI(parser.GitLab) || config.Dependencies == nil || //nolint:revive,nolintlint
				(config.Dependencies.Manager == kickr.ManagerRenovate && !slices.Contains(config.CI.Options, kickr.OptionRenovate))
		},
	})

	kickrp := path.Join(".gitlab", "pipelines", "kickr.yml")
	templates = append(templates, engine.Template[types.Repository]{
		Delimiters: engine.DelimitersBracket(),
		Globs:      []string{kickrp + engine.TmplExtension},
		Out:        kickrp,
		Remove: func(config types.Repository) bool {
			return !config.IsCI(parser.GitLab) || !slices.Contains(config.CI.Options, kickr.OptionKickr)
		},
	})

	return templates
}

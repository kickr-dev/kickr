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
		Remove:     func(config types.Repository) bool { return config.GitLab == nil },
	})

	semrel := path.Join(".gitlab", "pipelines", "semantic-release.yml")
	templates = append(templates, engine.Template[types.Repository]{
		Delimiters: engine.DelimitersBracket(),
		Globs:      []string{semrel + engine.TmplExtension},
		Out:        semrel,
		Remove:     func(config types.Repository) bool { return config.GitLab == nil },
	})

	deployment := path.Join(".gitlab", "pipelines", "deployment.yml")
	templates = append(templates, engine.Template[types.Repository]{
		Delimiters: engine.DelimitersBracket(),
		Globs:      engine.GlobsWithPart(deployment),
		Out:        deployment,
		Remove: func(config types.Repository) bool {
			return config.GitLab == nil || (config.Docker == nil && config.Helm == nil && config.Terraform == nil && config.Website == nil)
		},
	})

	integration := path.Join(".gitlab", "pipelines", "integration.yml")
	templates = append(templates, engine.Template[types.Repository]{
		Delimiters: engine.DelimitersBracket(),
		Globs:      []string{integration + engine.TmplExtension},
		Out:        integration,
		Remove:     func(config types.Repository) bool { return config.GitLab == nil },
	})

	kickrp := path.Join(".gitlab", "pipelines", "kickr.yml")
	templates = append(templates, engine.Template[types.Repository]{
		Delimiters: engine.DelimitersBracket(),
		Globs:      []string{kickrp + engine.TmplExtension},
		Out:        kickrp,
		Remove: func(config types.Repository) bool {
			return config.GitLab == nil || !slices.Contains(config.GitLab.Options, kickr.GitLabOptionsKickr)
		},
	})

	return templates
}

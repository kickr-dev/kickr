package templates

import (
	"path"
	"slices"
	"strings"

	engine "github.com/kickr-dev/engine/pkg"
	"github.com/kickr-dev/engine/pkg/parser"

	"github.com/kickr-dev/kickr/pkg/generate/types"
	kickr "github.com/kickr-dev/kickr/pkg/kickr/v1"
)

// Dependabot returns the slice of templates related to dependabot configuration.
func Dependabot() []engine.Template[types.Repository] {
	return []engine.Template[types.Repository]{
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      []string{path.Join(".github", "dependabot.yml"+engine.TmplExtension)},
			Out:        path.Join(".github", "dependabot.yml"),
			Remove: func(config types.Repository) bool {
				return config.Dependencies == nil || config.Dependencies.Manager != kickr.ManagerDependabot || config.Platform != parser.GitHub
			},
		},
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      []string{path.Join(".gitlab", "dependabot.yml"+engine.TmplExtension)},
			Out:        path.Join(".gitlab", "dependabot.yml"),
			Remove: func(config types.Repository) bool {
				return config.Dependencies == nil || config.Dependencies.Manager != kickr.ManagerDependabot || !config.IsCI(parser.GitLab)
			},
		},
	}
}

// Renovate returns the slice of templates related to renovate configuration.
func Renovate() []engine.Template[types.Repository] {
	return []engine.Template[types.Repository]{
		{
			Delimiters: engine.DelimitersChevron(),
			Globs:      []string{path.Join(".github", "workflows", "renovate.yml"+engine.TmplExtension)},
			Out:        path.Join(".github", "workflows", "renovate.yml"),
			Remove: func(config types.Repository) bool {
				manager := config.Dependencies != nil && config.Dependencies.Manager == kickr.ManagerRenovate
				ci := config.CI != nil && config.CI.Provider == parser.GitHub && slices.ContainsFunc(config.CI.Options, func(v string) bool { return strings.HasPrefix(v, "renovate:") })
				return !manager || !ci
			},
		},
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      []string{"renovate.json" + engine.TmplExtension},
			Out:        "renovate.json",
			Remove: func(config types.Repository) bool {
				return config.Dependencies == nil || config.Dependencies.Manager != kickr.ManagerRenovate
			},
		},
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      []string{path.Join(".ci", "renovate.json"+engine.TmplExtension)},
			Out:        path.Join(".github", "renovate.json"),
			Remove: func(config types.Repository) bool {
				manager := config.Dependencies != nil && config.Dependencies.Manager == kickr.ManagerRenovate
				ci := config.CI != nil && config.CI.Provider == parser.GitHub && slices.ContainsFunc(config.CI.Options, func(v string) bool { return strings.HasPrefix(v, "renovate:") })
				return !manager || !ci
			},
		},
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      []string{path.Join(".ci", "renovate.json"+engine.TmplExtension)},
			Out:        path.Join(".gitlab", "renovate.json"),
			Remove: func(config types.Repository) bool {
				manager := config.Dependencies != nil && config.Dependencies.Manager == kickr.ManagerRenovate
				ci := config.CI != nil && config.CI.Provider == parser.GitLab && slices.Contains(config.CI.Options, "renovate")
				return !manager || !ci
			},
		},
	}
}

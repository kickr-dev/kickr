package templates

import (
	"path"

	engine "github.com/kickr-dev/engine/pkg"
	"github.com/kickr-dev/engine/pkg/parser"

	"github.com/kickr-dev/kickr/pkg/generate/types"
	kickr "github.com/kickr-dev/kickr/pkg/kickr/v1"
)

// Dependabot returns the slice of templates related to dependabot configuration.
func Dependabot() []engine.Template[types.KickrGen] {
	return []engine.Template[types.KickrGen]{
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      []string{path.Join(".github", "dependabot.yml"+engine.TmplExtension)},
			Out:        path.Join(".github", "dependabot.yml"),
			Remove: func(config types.KickrGen) bool {
				return config.Dependencies == nil || config.Dependencies.Manager != kickr.ManagerDependabot || config.Platform != parser.GitHub
			},
		},
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      []string{path.Join(".gitlab", "dependabot.yml"+engine.TmplExtension)},
			Out:        path.Join(".gitlab", "dependabot.yml"),
			Remove: func(config types.KickrGen) bool {
				return config.Dependencies == nil || config.Dependencies.Manager != kickr.ManagerDependabot || !config.IsCI(parser.GitLab)
			},
		},
	}
}

// Renovate returns the slice of templates related to renovate configuration.
func Renovate() []engine.Template[types.KickrGen] {
	return []engine.Template[types.KickrGen]{
		{
			Delimiters: engine.DelimitersChevron(),
			Globs:      []string{path.Join(".github", "workflows", "renovate.yml"+engine.TmplExtension)},
			Out:        path.Join(".github", "workflows", "renovate.yml"),
			Remove: func(config types.KickrGen) bool {
				manager := config.Dependencies != nil && config.Dependencies.Manager == kickr.ManagerRenovate
				ci := config.CI != nil && config.CI.Provider == parser.GitHub && config.CI.Renovate != nil
				return !manager || !ci
			},
		},
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      []string{"renovate.json5" + engine.TmplExtension},
			Out:        "renovate.json5",
			Remove: func(config types.KickrGen) bool {
				return config.Dependencies == nil || config.Dependencies.Manager != kickr.ManagerRenovate
			},
		},
	}
}

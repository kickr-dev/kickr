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

// Renovate returns the slice of templates related to renovate configuration.
func Renovate() []engine.Template[types.Repository] {
	return []engine.Template[types.Repository]{
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      []string{"renovate.json" + engine.TmplExtension},
			Out:        "renovate.json",
			Remove: func(config types.Repository) bool {
				return slices.Contains(config.Exclude, kickr.ExcludeRenovate)
			},
		},
		{
			Delimiters: engine.DelimitersChevron(),
			Globs:      []string{path.Join(".github", "workflows", "renovate.yml"+engine.TmplExtension)},
			Out:        path.Join(".github", "workflows", "renovate.yml"),
			Remove: func(config types.Repository) bool {
				return !config.IsCI(parser.GitHub) || !slices.ContainsFunc(config.CI.Options, func(v string) bool { return strings.HasPrefix(v, kickr.OptionRenovate+":") })
			},
		},
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      []string{path.Join(".ci", "renovate.json"+engine.TmplExtension)},
			Out:        path.Join(".github", "renovate.json"),
			Remove: func(config types.Repository) bool {
				return !config.IsCI(parser.GitHub) || !slices.ContainsFunc(config.CI.Options, func(v string) bool { return strings.HasPrefix(v, kickr.OptionRenovate+":") })
			},
		},
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      []string{path.Join(".gitlab", "pipelines", "renovate.yml"+engine.TmplExtension)},
			Out:        path.Join(".gitlab", "pipelines", "renovate.yml"),
			Remove: func(config types.Repository) bool {
				return !config.IsCI(parser.GitLab) || !slices.Contains(config.CI.Options, kickr.OptionRenovate)
			},
		},
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      []string{path.Join(".ci", "renovate.json"+engine.TmplExtension)},
			Out:        path.Join(".gitlab", "renovate.json"),
			Remove: func(config types.Repository) bool {
				return !config.IsCI(parser.GitLab) || !slices.Contains(config.CI.Options, kickr.OptionRenovate)
			},
		},
	}
}

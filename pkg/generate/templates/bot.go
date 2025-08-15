package templates

import (
	"path"
	"slices"

	engine "github.com/kickr-dev/engine/pkg"
	"github.com/kickr-dev/engine/pkg/parser"

	kickr "github.com/kickr-dev/kickr/pkg/configuration"
)

// Dependabot returns the slice of templates related to dependabot configuration.
func Dependabot() []engine.Template[kickr.Config] {
	return []engine.Template[kickr.Config]{
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      []string{path.Join(".github", "dependabot.yml"+engine.TmplExtension)},
			Out:        path.Join(".github", "dependabot.yml"),
			Remove: func(config kickr.Config) bool {
				return config.Bot != kickr.Dependabot || config.Platform != parser.GitHub
			},
		},
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      []string{path.Join(".gitlab", "dependabot.yml"+engine.TmplExtension)},
			Out:        path.Join(".gitlab", "dependabot.yml"),
			Remove: func(config kickr.Config) bool {
				return config.Bot != kickr.Dependabot || !config.IsCI(parser.GitLab)
			},
		},
	}
}

// Renovate returns the slice of templates related to renovate configuration.
func Renovate() []engine.Template[kickr.Config] {
	return []engine.Template[kickr.Config]{
		{
			Delimiters: engine.DelimitersChevron(),
			Globs:      []string{path.Join(".github", "workflows", "renovate.yml"+engine.TmplExtension)},
			Out:        path.Join(".github", "workflows", "renovate.yml"),
			Remove: func(config kickr.Config) bool {
				return config.Bot != kickr.Renovate || !config.IsCI(parser.GitHub)
			},
		},
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      []string{path.Join("scripts", "sh", "renovate.sh"+engine.TmplExtension)},
			Out:        path.Join("scripts", "sh", "renovate.sh"),
			Remove: func(config kickr.Config) bool {
				return config.Bot != kickr.Renovate || !slices.Contains(config.Include, kickr.RenovatePostUpgrade)
			},
		},
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      []string{"renovate.json5" + engine.TmplExtension},
			Out:        "renovate.json5",
			Remove:     func(config kickr.Config) bool { return config.Bot != kickr.Renovate },
		},
	}
}

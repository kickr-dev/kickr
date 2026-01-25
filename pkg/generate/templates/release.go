package templates

import (
	"path"

	engine "github.com/kickr-dev/engine/pkg"
	"github.com/kickr-dev/engine/pkg/parser"

	"github.com/kickr-dev/kickr/pkg/generate/types"
)

// SemanticRelease returns the slice of templates related to semantic-release configuration.
func SemanticRelease() []engine.Template[types.Repository] {
	return []engine.Template[types.Repository]{
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      []string{".releaserc.yml" + engine.TmplExtension},
			Out:        ".releaserc.yml",
			Remove: func(config types.Repository) bool {
				return config.CI == nil || config.CI.Release == nil
			},
		},
		{
			Delimiters:     engine.DelimitersBracket(),
			Globs:          []string{path.Join(".gitlab", "semrel-plugins.txt"+engine.TmplExtension)},
			Out:            path.Join(".gitlab", "semrel-plugins.txt"),
			GeneratePolicy: engine.PolicyAlways, // always generate semrel-plugins.txt
			Remove: func(config types.Repository) bool {
				return !config.IsCI(parser.GitLab) || config.CI.Release == nil //nolint:revive
			},
		},
	}
}

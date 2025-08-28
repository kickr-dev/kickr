package templates

import (
	"path"

	engine "github.com/kickr-dev/engine/pkg"
	"github.com/kickr-dev/engine/pkg/parser"

	"github.com/kickr-dev/kickr/pkg/generate/types"
)

// SemanticRelease returns the slice of templates related to semantic-release configuration.
func SemanticRelease() []engine.Template[types.KickrGen] {
	return []engine.Template[types.KickrGen]{
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      []string{".releaserc.yml" + engine.TmplExtension},
			Out:        ".releaserc.yml",
			Remove:     func(config types.KickrGen) bool { return !config.HasRelease() },
		},
		{
			Delimiters:     engine.DelimitersBracket(),
			Globs:          []string{path.Join(".gitlab", "semrel-plugins.txt"+engine.TmplExtension)},
			Out:            path.Join(".gitlab", "semrel-plugins.txt"),
			GeneratePolicy: engine.PolicyAlways, // always generate semrel-plugins.txt
			Remove: func(config types.KickrGen) bool {
				return !config.HasRelease() || !config.IsCI(parser.GitLab)
			},
		},
	}
}

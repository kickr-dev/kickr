package templates

import (
	"path"

	engine "github.com/kickr-dev/engine/pkg"
	"github.com/kickr-dev/engine/pkg/parser"

	"github.com/kickr-dev/kickr/pkg/generate/types"
)

// GitLab returns the slice of templates related to GitLab configuration.
func GitLab() []engine.Template[types.KickrGen] {
	srcs := []string{".gitlab-ci.yml", path.Join(".gitlab", "workflows", ".gitlab-ci.yml")}

	templates := make([]engine.Template[types.KickrGen], 0, len(srcs))
	for _, src := range srcs {
		templates = append(templates, engine.Template[types.KickrGen]{
			Delimiters: engine.DelimitersBracket(),
			Globs:      []string{src + engine.TmplExtension},
			Out:        src,
			Remove:     func(config types.KickrGen) bool { return !config.IsCI(parser.GitLab) },
		})
	}
	return templates
}

package templates

import (
	"path"

	engine "github.com/kickr-dev/engine/pkg"
	"github.com/kickr-dev/engine/pkg/parser"

	"github.com/kickr-dev/kickr/pkg/generate/types"
)

// GitLab returns the slice of templates related to GitLab configuration.
func GitLab() []engine.Template[types.Repository] {
	srcs := []string{".gitlab-ci.yml", path.Join(".gitlab", "workflows", ".gitlab-ci.yml")}

	templates := make([]engine.Template[types.Repository], 0, len(srcs))
	for _, src := range srcs {
		templates = append(templates, engine.Template[types.Repository]{
			Delimiters: engine.DelimitersBracket(),
			Globs:      []string{src + engine.TmplExtension},
			Out:        src,
			Remove:     func(config types.Repository) bool { return !config.IsCI(parser.GitLab) },
		})
	}
	return templates
}

package templates

import (
	"path"

	engine "github.com/kickr-dev/engine/pkg"
	"github.com/kickr-dev/engine/pkg/parser"

	kickr "github.com/kickr-dev/kickr/pkg/configuration"
)

// GitLab returns the slice of templates related to GitLab configuration.
func GitLab() []engine.Template[kickr.Config] {
	srcs := []string{".gitlab-ci.yml", path.Join(".gitlab", "workflows", ".gitlab-ci.yml")}

	templates := make([]engine.Template[kickr.Config], 0, len(srcs))
	for _, src := range srcs {
		templates = append(templates, engine.Template[kickr.Config]{
			Delimiters: engine.DelimitersBracket(),
			Globs:      []string{src + engine.TmplExtension},
			Out:        src,
			Remove:     func(config kickr.Config) bool { return !config.IsCI(parser.GitLab) },
		})
	}
	return templates
}

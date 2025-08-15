package templates

import (
	"path"
	"slices"

	engine "github.com/kickr-dev/engine/pkg"

	kickr "github.com/kickr-dev/kickr/pkg/configuration"
)

// Makefile returns the slice of templates related to make configuration (build, test, docker make tasks).
func Makefile() []engine.Template[kickr.Config] {
	return []engine.Template[kickr.Config]{
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      []string{"Makefile" + engine.TmplExtension},
			Out:        "Makefile",
			Remove: func(config kickr.Config) bool {
				_, ok := config.Languages["node"] // don't generate makefiles with node
				return ok || slices.Contains(config.Exclude, kickr.Makefile)
			},
		},
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      engine.GlobsWithPart(path.Join("scripts", "mk", "build.mk")),
			Out:        path.Join("scripts", "mk", "build.mk"),
			Remove: func(config kickr.Config) bool {
				_, ok := config.Languages["node"] // don't generate makefiles with node
				return ok || slices.Contains(config.Exclude, kickr.Makefile)
			},
		},
	}
}

package templates

import (
	"path"
	"slices"

	engine "github.com/kickr-dev/engine/pkg"

	"github.com/kickr-dev/kickr/pkg/generate/types"
	kickr "github.com/kickr-dev/kickr/pkg/kickr/v1"
)

// Makefile returns the slice of templates related to make configuration (build, test, docker make tasks).
func Makefile() []engine.Template[types.KickrWrapper] {
	return []engine.Template[types.KickrWrapper]{
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      []string{"Makefile" + engine.TmplExtension},
			Out:        "Makefile",
			Remove: func(config types.KickrWrapper) bool {
				_, ok := config.Languages["node"] // don't generate makefiles with node
				return ok || slices.Contains(config.Exclude, kickr.ExcludeMakefile)
			},
		},
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      engine.GlobsWithPart(path.Join("scripts", "mk", "build.mk")),
			Out:        path.Join("scripts", "mk", "build.mk"),
			Remove: func(config types.KickrWrapper) bool {
				_, ok := config.Languages["node"] // don't generate makefiles with node
				return ok || slices.Contains(config.Exclude, kickr.ExcludeMakefile)
			},
		},
	}
}

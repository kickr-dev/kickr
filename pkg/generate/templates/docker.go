package templates

import (
	engine "github.com/kickr-dev/engine/pkg"

	"github.com/kickr-dev/kickr/pkg/generate/types"
)

// Docker returns the slice of templates related to Docker generation (Dockerfile, .dockerignore, etc.).
func Docker() []engine.Template[types.Module] {
	return []engine.Template[types.Module]{
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      engine.GlobsWithPart("Dockerfile"),
			Out:        "Dockerfile",
			Remove: func(module types.Module) bool {
				return !module.HasDocker()
			},
		},
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      []string{".dockerignore" + engine.TmplExtension},
			Out:        ".dockerignore",
			Remove: func(module types.Module) bool {
				return !module.HasDocker()
			},
		},
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      []string{"launcher.sh" + engine.TmplExtension},
			Out:        "launcher.sh",
			// launcher.sh is a specific thing to golang being able to have multiple binaries inside a simple project (cmd folder)
			// however, it may change in the future with python (or rust or others ?) depending on flexibility in repositories layout
			Remove: func(module types.Module) bool {
				_, ok := module.Languages[types.LanguageGo]
				return !ok || !module.HasDocker() || module.Binaries() <= 1
			},
		},
	}
}

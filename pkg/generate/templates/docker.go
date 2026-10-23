package templates

import (
	engine "github.com/kickr-dev/engine/pkg"
	"github.com/kickr-dev/engine/pkg/files"

	"github.com/kickr-dev/kickr/pkg/generate/types"
)

// Docker returns the slice of templates related to Docker generation (Dockerfile, .dockerignore, etc.).
func Docker() []engine.Template[types.Module] {
	return []engine.Template[types.Module]{
		{
			Delimiters:  engine.DelimitersBracket(),
			EmptyPolicy: engine.PolicyKeepEmpty,
			Globs:       engine.GlobsWithPart("Dockerfile"),
			Out:         "Dockerfile",
			GeneratePolicy: engine.RemoveIf(func(module types.Module) bool {
				return !module.HasDocker()
			}, engine.PolicyDefault),
		},
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      []string{".dockerignore" + engine.TmplExtension},
			Out:        ".dockerignore",
			GeneratePolicy: engine.RemoveIf(func(module types.Module) bool {
				return !module.HasDocker()
			}, engine.PolicyDefault),
		},
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      []string{"launcher.sh" + engine.TmplExtension},
			Mode:       files.RwxRxRxRx,
			Out:        "launcher.sh",
			// launcher.sh is a specific thing to golang being able to have multiple binaries inside a simple project (cmd folder)
			// however, it may change in the future with python (or rust or others ?) depending on flexibility in repositories layout
			GeneratePolicy: engine.RemoveIf(func(module types.Module) bool {
				_, ok := module.Languages[types.LanguageGo]
				return !ok || !module.HasDocker() || module.Binaries() <= 1
			}, engine.PolicyDefault),
		},
	}
}

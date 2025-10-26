package templates

import (
	engine "github.com/kickr-dev/engine/pkg"

	"github.com/kickr-dev/kickr/pkg/generate/types"
)

// Docker returns the slice of templates related to Docker generation (Dockerfile, .dockerignore, etc.).
func Docker() []engine.Template[types.Repository] {
	return []engine.Template[types.Repository]{
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      engine.GlobsWithPart("Dockerfile"),
			Out:        "Dockerfile",
			Remove: func(config types.Repository) bool {
				return config.CI == nil || config.CI.Docker == nil || config.Binaries() == 0
			},
		},
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      []string{".dockerignore" + engine.TmplExtension},
			Out:        ".dockerignore",
			Remove: func(config types.Repository) bool {
				return config.CI == nil || config.CI.Docker == nil || config.Binaries() == 0
			},
		},
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      []string{"launcher.sh" + engine.TmplExtension},
			Out:        "launcher.sh",
			// launcher.sh is a specific thing to golang being able to have multiple binaries inside a simple project (cmd folder)
			// however, it may change in the future with python (or rust or others ?) depending on flexibility in repositories layout
			Remove: func(config types.Repository) bool {
				_, ok := config.Languages["go"]
				return !ok || config.CI == nil || config.CI.Docker == nil || config.Binaries() <= 1
			},
		},
	}
}

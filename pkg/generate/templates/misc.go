package templates

import (
	"slices"

	engine "github.com/kickr-dev/engine/pkg"

	kickr "github.com/kickr-dev/kickr/pkg/configuration"
)

// Misc returns the slice of templates globally related to a code repository (README.md, CODEOWNERS, etc.).
func Misc() []engine.Template[kickr.Config] {
	return []engine.Template[kickr.Config]{
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      []string{"CODEOWNERS" + engine.TmplExtension},
			Out:        "CODEOWNERS",
		},
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      []string{"README.md" + engine.TmplExtension},
			Out:        "README.md",
		},
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      []string{".pre-commit-config.yaml" + engine.TmplExtension},
			Out:        ".pre-commit-config.yaml",
			Remove:     func(config kickr.Config) bool { return slices.Contains(config.Exclude, kickr.PreCommit) },
		},
	}
}

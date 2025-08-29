package templates

import (
	"slices"

	engine "github.com/kickr-dev/engine/pkg"

	"github.com/kickr-dev/kickr/pkg/generate/types"
	kickr "github.com/kickr-dev/kickr/pkg/kickr/v1"
)

// Misc returns the slice of templates globally related to a code repository (README.md, CODEOWNERS, etc.).
func Misc() []engine.Template[types.KickrWrapper] {
	return []engine.Template[types.KickrWrapper]{
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
			Remove:     func(config types.KickrWrapper) bool { return slices.Contains(config.Exclude, kickr.ExcludePreCommit) },
		},
	}
}

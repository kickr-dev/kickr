package templates

import (
	"path"
	"slices"

	engine "github.com/kickr-dev/engine/pkg"

	"github.com/kickr-dev/kickr/pkg/generate/types"
	kickr "github.com/kickr-dev/kickr/pkg/kickr/v1"
)

// Misc returns the slice of templates globally related to a code repository (README.md, CODEOWNERS, etc.).
func Misc() []engine.Template[types.Repository] {
	return []engine.Template[types.Repository]{
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
			Globs:      []string{"SECURITY.md" + engine.TmplExtension},
			Out:        "SECURITY.md",
		},
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      []string{".pre-commit-config.yaml" + engine.TmplExtension},
			Out:        ".pre-commit-config.yaml",
			Remove:     func(config types.Repository) bool { return slices.Contains(config.Exclude, kickr.ExcludePreCommit) },
		},
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      []string{path.Join("scripts", "sh", "conventionalcommits-branch.sh"+engine.TmplExtension)},
			Out:        path.Join("scripts", "sh", "conventionalcommits-branch.sh"),
			Remove: func(config types.Repository) bool {
				return slices.Contains(config.Exclude, kickr.ExcludePreCommit) || !slices.Contains(config.PreCommit, kickr.PreCommitConventionalCommits)
			},
		},
	}
}

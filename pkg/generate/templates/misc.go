package templates

import (
	"path"
	"slices"

	engine "github.com/kickr-dev/engine/pkg"
	"github.com/kickr-dev/engine/pkg/files"

	"github.com/kickr-dev/kickr/pkg/generate/types"
	kickr "github.com/kickr-dev/kickr/pkg/kickr/v1"
)

// Misc returns the slice of templates globally related to a code repository (CODEOWNERS, README.md, SECURITY.md, etc.).
//
// Existing README.md and CONTRIBUTING.md files are kept as is, except for their kickr sections (badges and guidelines), which are regenerated.
func Misc() []engine.Template[types.Repository] {
	return []engine.Template[types.Repository]{
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      []string{"CODEOWNERS" + engine.TmplExtension},
			Out:        "CODEOWNERS",
		},
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      engine.GlobsWithPart("CONTRIBUTING.md"),
			Out:        "CONTRIBUTING.md",
			Remove: func(repo types.Repository) bool {
				return slices.Contains(repo.Config.Exclude, kickr.ExcludeContributing)
			},
			Sections: []engine.Section{
				{Begin: "<!-- BEGIN_KICKR_GUIDELINES -->", End: "<!-- END_KICKR_GUIDELINES -->"},
			},
		},
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      engine.GlobsWithPart("README.md"),
			Out:        "README.md",
			Sections: []engine.Section{
				{Begin: "<!-- BEGIN_KICKR_BADGES -->", End: "<!-- END_KICKR_BADGES -->"},
			},
		},
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      []string{"SECURITY.md" + engine.TmplExtension},
			Out:        "SECURITY.md",
			Remove:     func(repo types.Repository) bool { return slices.Contains(repo.Config.Exclude, kickr.ExcludeSecurity) },
		},
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      []string{".pre-commit-config.yaml" + engine.TmplExtension},
			Out:        ".pre-commit-config.yaml",
			Remove:     func(repo types.Repository) bool { return slices.Contains(repo.Config.Exclude, kickr.ExcludePreCommit) },
		},
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      []string{path.Join("scripts", "sh", "conventionalcommits-branch.sh"+engine.TmplExtension)},
			Mode:       files.RwxRxRxRx,
			Out:        path.Join("scripts", "sh", "conventionalcommits-branch.sh"),
			Remove: func(repo types.Repository) bool {
				return slices.Contains(repo.Config.Exclude, kickr.ExcludePreCommit) || !slices.Contains(repo.Config.PreCommit, kickr.PreCommitConventionalCommits)
			},
		},
	}
}

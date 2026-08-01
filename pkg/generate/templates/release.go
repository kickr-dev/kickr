package templates

import (
	"path"

	engine "github.com/kickr-dev/engine/pkg"

	"github.com/kickr-dev/kickr/pkg/generate/types"
)

// SemanticRelease returns the slice of templates related to semantic-release configuration.
func SemanticRelease() []engine.Template[types.Repository] {
	return []engine.Template[types.Repository]{
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      []string{".releaserc.yml" + engine.TmplExtension},
			Out:        ".releaserc.yml",
			Remove: func(repo types.Repository) bool {
				gitlab := repo.GitLab != nil && repo.GitLab.Release != nil
				github := repo.GitHub != nil && repo.GitHub.Release != nil
				return !github && !gitlab
			},
		},
		{
			Delimiters:     engine.DelimitersBracket(),
			Globs:          []string{path.Join(".github", "semrel-plugins.txt"+engine.TmplExtension)},
			Out:            path.Join(".github", "semrel-plugins.txt"),
			GeneratePolicy: engine.PolicyAlways, // always generate semrel-plugins.txt
			Remove: func(repo types.Repository) bool {
				return repo.GitHub == nil || repo.GitHub.Release == nil
			},
		},
		{
			Delimiters:     engine.DelimitersBracket(),
			Globs:          []string{path.Join(".gitlab", "semrel-plugins.txt"+engine.TmplExtension)},
			Out:            path.Join(".gitlab", "semrel-plugins.txt"),
			GeneratePolicy: engine.PolicyAlways, // always generate semrel-plugins.txt
			Remove: func(repo types.Repository) bool {
				return repo.GitLab == nil || repo.GitLab.Release == nil
			},
		},
	}
}

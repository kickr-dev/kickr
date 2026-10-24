package templates

import (
	"path"

	engine "gitlab.com/kickr-dev/engine/pkg"

	"github.com/kickr-dev/kickr/pkg/generate/types"
)

// Golang returns the slice of templates related to Golang generation for given a module.
func Golang() []engine.Template[types.Module] {
	return []engine.Template[types.Module]{
		{
			Delimiters: engine.DelimitersChevron(),
			Globs:      []string{path.Join("internal", "build", "build.go"+engine.TmplExtension)},
			Out:        path.Join("internal", "build", "build.go"),
			GeneratePolicy: engine.RemoveIf(func(module types.Module) bool {
				_, ok := module.Languages[types.LanguageGo]
				return !ok || module.Binaries() == 0
			}, engine.PolicyDefault),
		},
	}
}

// RepositoryGolang returns the slice of templates related to Golang generation for the repository's root (golangci-lint, goreleaser, etc.).
func RepositoryGolang() []engine.Template[types.Repository] {
	return []engine.Template[types.Repository]{
		{
			Delimiters: engine.DelimitersChevron(),
			Globs:      []string{".golangci.yml" + engine.TmplExtension},
			Out:        ".golangci.yml",
			GeneratePolicy: engine.RemoveIf(func(repo types.Repository) bool {
				return len(repo.ModulesWith(types.LanguageGo)) == 0
			}, engine.PolicyDefault),
		},
		{
			Delimiters: engine.DelimitersChevron(),
			Globs:      []string{".goreleaser.yml" + engine.TmplExtension},
			Out:        ".goreleaser.yml",
			GeneratePolicy: engine.RemoveIf(func(repo types.Repository) bool {
				for _, module := range repo.ModulesWith(types.LanguageGo) {
					if len(module.Clis) > 0 {
						return false
					}
				}
				return true
			}, engine.PolicyDefault),
		},
	}
}

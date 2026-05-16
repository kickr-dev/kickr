package generate

import (
	"context"
	"path/filepath"

	engine "github.com/kickr-dev/engine/pkg"
	"github.com/kickr-dev/engine/pkg/files"

	"github.com/kickr-dev/kickr/pkg/generate/types"
)

// ParserGlob is a simple parser checking whether specific globs exist within the destdir project.
//
// It adds any matches into config Globs property.
func ParserGlob(_ context.Context, destdir string, config *types.Repository) error {
	checks := []struct {
		Glob string
		Name string
	}{
		{Glob: ".gitmodules", Name: "gitmodules"},
		{Glob: "*.*sh", Name: "shell"},
		{Glob: "*.tmpl", Name: "tmpl"},
		{Glob: "go.mod", Name: "gomod"},
	}
	for _, check := range checks {
		matches := files.Glob(destdir, check.Glob,
			files.GlobExcludedDirectories("node_modules", "testdata"),
			files.GlobExcludedFiles("conventionalcommits-branch.sh"))
		if len(matches) == 0 {
			continue
		}

		paths := make([]string, 0, len(matches))
		for _, match := range matches {
			path, err := filepath.Rel(destdir, match)
			if err != nil {
				engine.GetLogger().Warnf("failed to get relative path file: %v", err)
				continue
			}
			paths = append(paths, filepath.ToSlash(path))
		}
		config.SetGlob(check.Name, paths)
	}
	return nil
}

var _ engine.Parser[types.Repository] = ParserGlob

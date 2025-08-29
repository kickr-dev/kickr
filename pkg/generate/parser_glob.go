package generate

import (
	"context"

	engine "github.com/kickr-dev/engine/pkg"
	"github.com/kickr-dev/engine/pkg/files"
	"github.com/kickr-dev/kickr/pkg/generate/types"
)

// ParserGlob is a simple parser checking whether specific globs exist within the destdir project.
//
// It adds any matches into config Globs property.
func ParserGlob(ctx context.Context, destdir string, config *types.KickrGen) error {
	checks := []struct {
		Glob string
		Name string
	}{
		{Glob: ".gitmodules", Name: "gitmodules"},
		{Glob: "*.*sh", Name: "shell"},
		{Glob: "*.tmpl", Name: "tmpl"},
	}
	for _, check := range checks {
		if matches := files.Glob(destdir, check.Glob); len(matches) > 0 {
			config.SetGlob(check.Name)
		}
	}
	return nil
}

var _ engine.Parser[types.KickrGen] = ParserGlob

package generate

import (
	"context"
	"errors"

	"github.com/go-git/go-git/v5"
	engine "github.com/kickr-dev/engine/pkg"
	"github.com/kickr-dev/engine/pkg/parser"

	"github.com/kickr-dev/kickr/pkg/generate/types"
)

// ParserGit adds git configuration (if the current repository is a git repository)
// to the configuration.
func ParserGit(_ context.Context, destdir string, repo *types.Repository) error {
	vcs, err := parser.Git(destdir)
	if err != nil {
		for _, is := range []error{git.ErrRepositoryNotExists, git.ErrRemoteNotFound} {
			if errors.Is(err, is) {
				engine.GetLogger().Warnf("failed to retrieve git vcs configuration: %v", err)
				return nil
			}
		}
		return err // not really an added value to wrap here
	}
	engine.GetLogger().Infof("git repository detected")

	repo.VCS = vcs
	if repo.Config.Platform != "" {
		repo.VCS.Platform = repo.Config.Platform
	} else {
		repo.Config.Platform = repo.VCS.Platform
	}
	return nil
}

var _ engine.Parser[types.Repository] = ParserGit // ensure interface is implemented

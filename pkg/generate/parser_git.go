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
func ParserGit(_ context.Context, destdir string, config *types.KickrGen) error {
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

	config.VCS = vcs
	if config.Platform != "" {
		config.VCS.Platform = config.Platform
	} else {
		config.Platform = config.VCS.Platform
	}
	return nil
}

var _ engine.Parser[types.KickrGen] = ParserGit // ensure interface is implemented

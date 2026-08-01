package generate_test

import (
	"path/filepath"
	"testing"

	"github.com/kickr-dev/engine/pkg/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kickr-dev/kickr/pkg/generate"
	"github.com/kickr-dev/kickr/pkg/generate/types"
	"github.com/kickr-dev/kickr/pkg/kickr/v1"
	"github.com/kickr-dev/kickr/testutils"
)

func TestParserGit(t *testing.T) {
	ctx := t.Context()

	t.Run("success_no_vcs", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()
		repo := types.Repository{}

		// Act
		err := generate.ParserGit(ctx, destdir, &repo)

		// Assert
		require.NoError(t, err)
		assert.Zero(t, repo)
	})

	t.Run("success_vcs", func(t *testing.T) {
		// Arrange
		expected := types.Repository{
			Kickr: kickr.Kickr{Platform: parser.GitLab},
			VCS: parser.VCS{
				Platform:    parser.GitLab,
				ProjectHost: "gitlab.com",
				ProjectName: "kickr",
				ProjectPath: "kickr-dev/kickr",
			},
		}
		repo := types.Repository{}

		// Act
		err := generate.ParserGit(ctx, filepath.Join(testutils.Testdata(t), ".."), &repo)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expected, types.Repository{
			Kickr: repo.Kickr,
			VCS: parser.VCS{
				Platform:    repo.VCS.Platform,
				ProjectHost: repo.VCS.ProjectHost,
				ProjectName: repo.VCS.ProjectName,
				ProjectPath: repo.VCS.ProjectPath,
				// ignore tags
			},
		})
	})

	t.Run("success_platform_already_present", func(t *testing.T) {
		// Arrange
		expected := types.Repository{
			Kickr: kickr.Kickr{Platform: parser.GitHub},
			VCS: parser.VCS{
				Platform:    parser.GitHub,
				ProjectHost: "gitlab.com",
				ProjectName: "kickr",
				ProjectPath: "kickr-dev/kickr",
			},
		}
		repo := types.Repository{
			Kickr: kickr.Kickr{Platform: parser.GitHub},
		}

		// Act
		err := generate.ParserGit(ctx, filepath.Join(testutils.Testdata(t), ".."), &repo)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expected, types.Repository{
			Kickr: repo.Kickr,
			VCS: parser.VCS{
				Platform:    repo.VCS.Platform,
				ProjectHost: repo.VCS.ProjectHost,
				ProjectName: repo.VCS.ProjectName,
				ProjectPath: repo.VCS.ProjectPath,
				// ignore tags
			},
		})
	})
}

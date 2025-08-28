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
		config := types.KickrGen{}

		// Act
		err := generate.ParserGit(ctx, destdir, &config)

		// Assert
		require.NoError(t, err)
		assert.Zero(t, config)
	})

	t.Run("success_vcs", func(t *testing.T) {
		// Arrange
		expected := types.KickrGen{
			Kickr: kickr.Kickr{Platform: parser.GitHub},
			VCS: parser.VCS{
				Platform:    parser.GitHub,
				ProjectHost: "github.com",
				ProjectName: "kickr",
				ProjectPath: "kickr-dev/kickr",
			},
		}
		config := types.KickrGen{}

		// Act
		err := generate.ParserGit(ctx, filepath.Join(testutils.Testdata(t), ".."), &config)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expected, config)
	})

	t.Run("success_platform_already_present", func(t *testing.T) {
		// Arrange
		expected := types.KickrGen{
			Kickr: kickr.Kickr{Platform: parser.GitLab},
			VCS: parser.VCS{
				Platform:    parser.GitLab,
				ProjectHost: "github.com",
				ProjectName: "kickr",
				ProjectPath: "kickr-dev/kickr",
			},
		}
		config := types.KickrGen{
			Kickr: kickr.Kickr{Platform: parser.GitLab},
		}

		// Act
		err := generate.ParserGit(ctx, filepath.Join(testutils.Testdata(t), ".."), &config)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expected, config)
	})
}

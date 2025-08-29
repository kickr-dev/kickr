package generate_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kickr-dev/engine/pkg/files"
	"github.com/kickr-dev/engine/pkg/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kickr-dev/kickr/pkg/generate"
	"github.com/kickr-dev/kickr/pkg/generate/types"
)

func TestParserNode(t *testing.T) {
	ctx := t.Context()

	t.Run("error_read_packagejson", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()
		require.NoError(t, os.Mkdir(filepath.Join(destdir, parser.FilePackageJSON), files.RwxRxRxRx))

		// Act
		err := generate.ParserNode(ctx, destdir, &types.KickrWrapper{})

		// Assert
		assert.ErrorContains(t, err, "read json")
	})

	t.Run("error_validate_packagejson", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(destdir, parser.FilePackageJSON), []byte("{}"), files.RwRR))

		// Act
		err := generate.ParserNode(ctx, destdir, &types.KickrWrapper{})

		// Assert
		assert.ErrorIs(t, err, parser.ErrMissingPackageName)
		assert.ErrorIs(t, err, parser.ErrInvalidPackageManager)
	})

	t.Run("success_no_main", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()
		require.NoError(t, os.WriteFile(
			filepath.Join(destdir, parser.FilePackageJSON),
			[]byte(`{ "name": "kickr", "packageManager": "bun@1.1.6" }`), files.RwRR))

		expected := types.KickrWrapper{
			Languages: map[string]any{
				"node": parser.PackageJSON{
					Name:           "kickr",
					PackageManager: "bun@1.1.6",
				},
			},
		}
		config := types.KickrWrapper{}

		// Act
		err := generate.ParserNode(ctx, destdir, &config)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expected, config)
	})

	t.Run("success_main", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()
		require.NoError(t, os.WriteFile(
			filepath.Join(destdir, parser.FilePackageJSON),
			[]byte(`{ "name": "kickr", "packageManager": "bun@1.1.6", "main": "index.js" }`), files.RwRR))

		expected := types.KickrWrapper{
			Executables: parser.Executables{
				Workers: map[string]struct{}{"main": {}},
			},
			Languages: map[string]any{
				"node": parser.PackageJSON{
					Main:           func() *string { v := "index.js"; return &v }(),
					Name:           "kickr",
					PackageManager: "bun@1.1.6",
				},
			},
		}
		config := types.KickrWrapper{}

		// Act
		err := generate.ParserNode(ctx, destdir, &config)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expected, config)
	})
}

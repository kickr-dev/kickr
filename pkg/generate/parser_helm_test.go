package generate_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kickr-dev/engine/pkg/files"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	kickr "github.com/kickr-dev/kickr/pkg/configuration"
	"github.com/kickr-dev/kickr/pkg/generate"
)

func TestParserChart(t *testing.T) {
	ctx := t.Context()

	t.Run("error_merge_values", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()
		kickrfile := filepath.Join(destdir, "chart", kickr.File)
		require.NoError(t, os.MkdirAll(kickrfile, files.RwxRxRxRx))

		// Act
		err := generate.ParserHelm(ctx, destdir, &kickr.Config{CI: &kickr.CI{Helm: &kickr.Helm{}}})

		// Assert
		assert.ErrorContains(t, err, "read yaml")
	})

	t.Run("success_merge_values", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()
		chartdir := filepath.Join(destdir, "chart")
		require.NoError(t, os.Mkdir(chartdir, files.RwxRxRxRx))
		require.NoError(t, os.WriteFile(
			filepath.Join(chartdir, kickr.File),
			[]byte("description: a description"), files.RwRR))

		expected := kickr.Config{
			CI: &kickr.CI{Helm: &kickr.Helm{}},
			Languages: map[string]any{
				"helm": map[string]any{
					"ci":          map[string]any{},
					"description": "a description",
				},
			},
		}
		config := kickr.Config{CI: &kickr.CI{Helm: &kickr.Helm{}}}

		// Act
		err := generate.ParserHelm(ctx, destdir, &config)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expected, config)
	})
}

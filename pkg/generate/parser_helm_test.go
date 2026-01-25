package generate_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kickr-dev/engine/pkg/files"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kickr-dev/kickr/pkg/generate"
	"github.com/kickr-dev/kickr/pkg/generate/types"
	kickr "github.com/kickr-dev/kickr/pkg/kickr/v1"
)

func TestParserChart(t *testing.T) {
	ctx := t.Context()

	t.Run("error_merge_values", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()
		kickrfile := filepath.Join(destdir, "chart", kickr.CustomValues)
		require.NoError(t, os.MkdirAll(kickrfile, files.RwxRxRxRx))

		// Act
		err := generate.ParserHelm(ctx, destdir, &types.Repository{Kickr: kickr.Kickr{Helm: &kickr.Helm{}}})

		// Assert
		assert.ErrorContains(t, err, "read yaml")
	})

	t.Run("success_merge_values", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()
		chartdir := filepath.Join(destdir, "chart")
		require.NoError(t, os.MkdirAll(chartdir, files.RwxRxRxRx))
		require.NoError(t, os.WriteFile(
			filepath.Join(chartdir, kickr.CustomValues),
			[]byte("description: a description"), files.RwRR))

		expected := types.Repository{
			Kickr: kickr.Kickr{Helm: &kickr.Helm{}},
			Languages: map[string]any{
				"helm": map[string]any{
					"description": "a description",
					"docker":      map[string]any{},

					"clis":    nil,
					"crons":   nil,
					"jobs":    nil,
					"workers": nil,

					"maintainers": nil,
					"projectName": "",
					"projectPath": "",
				},
			},
		}
		config := types.Repository{Kickr: kickr.Kickr{Helm: &kickr.Helm{}}}

		// Act
		err := generate.ParserHelm(ctx, destdir, &config)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expected, config)
	})
}

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

		expected := types.Repository{Kickr: kickr.Kickr{Helm: &kickr.Helm{}}}
		expected.Module(types.RootModule).SetLanguage(types.LanguageHelm, map[string]any{
			"description": "a description",
			"docker":      map[string]any{},

			"clis":    map[string]any{},
			"crons":   map[string]any{},
			"jobs":    map[string]any{},
			"workers": map[string]any{},

			"maintainers": nil,
			"projectName": "",
			"projectPath": "",
		})
		repo := types.Repository{Kickr: kickr.Kickr{Helm: &kickr.Helm{}}}

		// Act
		err := generate.ParserHelm(ctx, destdir, &repo)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expected, repo)
	})

	t.Run("success_module_image_repository", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()
		chartdir := filepath.Join(destdir, "chart")
		require.NoError(t, os.MkdirAll(chartdir, files.RwxRxRxRx))
		require.NoError(t, os.WriteFile(
			filepath.Join(chartdir, kickr.CustomValues),
			[]byte("workers:\n  api:\n    replicaCount: 2"), files.RwRR))

		conf := kickr.Kickr{Helm: &kickr.Helm{}}
		vcs := parser.VCS{ProjectPath: "org/repo"}

		expected := types.Repository{Kickr: conf, VCS: vcs}
		expected.Module(types.RootModule).AddWorker("main")
		expected.Module(types.RootModule).SetLanguage(types.LanguageHelm, map[string]any{
			"description": "",
			"docker":      map[string]any{},

			"clis":  map[string]any{},
			"crons": map[string]any{},
			"jobs":  map[string]any{},
			"workers": map[string]any{
				"main": map[string]any{"image": map[string]any{"repository": "org/repo"}},
				"api": map[string]any{
					"image":        map[string]any{"repository": "org/repo/services-api"},
					"replicaCount": uint64(2),
				},
			},

			"maintainers": nil,
			"projectName": "",
			"projectPath": "org/repo",
		})
		expected.Module("services/api").AddWorker("api")

		repo := types.Repository{Kickr: conf, VCS: vcs}
		repo.Module(types.RootModule).AddWorker("main")
		repo.Module("services/api").AddWorker("api")

		// Act
		err := generate.ParserHelm(ctx, destdir, &repo)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expected, repo)
	})
}

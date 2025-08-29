package generate_test

import (
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kickr-dev/engine/pkg/files"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kickr-dev/kickr/pkg/generate"
	"github.com/kickr-dev/kickr/pkg/generate/types"
)

func TestParserGlob(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		type testcase struct {
			Files    []string
			GlobName string
		}

		cases := []testcase{
			{Files: []string{".gitmodules", path.Join("subdir", ".gitmodules")}, GlobName: "gitmodules"},
			{Files: []string{"script.bash", path.Join("subdir", "script.bash")}, GlobName: "shell"},
			{Files: []string{"script.ksh", path.Join("subdir", "scripts.ksh")}, GlobName: "shell"},
			{Files: []string{"script.sh", path.Join("subdir", "script.sh")}, GlobName: "shell"},
			{Files: []string{"script.zsh", path.Join("subdir", "script.zsh")}, GlobName: "shell"},
			{Files: []string{"template.tmpl", path.Join("subdir", "template.tmpl")}, GlobName: "tmpl"},
		}
		for _, tc := range cases {
			t.Run(strings.Join(tc.Files, "_"), func(t *testing.T) {
				// Arrange
				destdir := t.TempDir()
				for _, file := range tc.Files {
					target, err := filepath.Localize(file)
					require.NoError(t, err)

					dir := filepath.Join(destdir, filepath.Dir(target))
					require.NoError(t, os.MkdirAll(dir, files.RwxRxRxRx))

					file, err := os.Create(filepath.Join(destdir, target))
					require.NoError(t, err)
					require.NoError(t, file.Close())
				}

				expected := map[string]any{tc.GlobName: tc.Files}
				config := types.KickrWrapper{}

				// Act
				err := generate.ParserGlob(t.Context(), destdir, &config)

				// Assert
				require.NoError(t, err)
				assert.Equal(t, expected, config.Globs)
			})
		}
	})
}

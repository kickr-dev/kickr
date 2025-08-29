package generate_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kickr-dev/kickr/pkg/generate"
	"github.com/kickr-dev/kickr/pkg/generate/types"
)

func TestParserGlob(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		type testcase struct {
			File     string
			GlobName string
		}
		cases := []testcase{
			{File: ".gitmodules", GlobName: "gitmodules"},
			{File: "script.bash", GlobName: "shell"},
			{File: "script.ksh", GlobName: "shell"},
			{File: "script.sh", GlobName: "shell"},
			{File: "script.zsh", GlobName: "shell"},
			{File: "template.tmpl", GlobName: "tmpl"},
		}

		for _, tc := range cases {
			t.Run(tc.File, func(t *testing.T) {
				// Arrange
				destdir := t.TempDir()
				file, err := os.Create(filepath.Join(destdir, tc.File))
				require.NoError(t, err)
				require.NoError(t, file.Close())
				config := types.KickrWrapper{}

				// Act
				err = generate.ParserGlob(t.Context(), destdir, &config)

				// Assert
				require.NoError(t, err)
				assert.Contains(t, config.Globs, tc.GlobName)
			})
		}
	})
}

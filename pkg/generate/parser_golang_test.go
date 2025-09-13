package generate_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/kickr-dev/engine/pkg/files"
	"github.com/kickr-dev/engine/pkg/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kickr-dev/kickr/pkg/generate"
	"github.com/kickr-dev/kickr/pkg/generate/types"
	"github.com/kickr-dev/kickr/pkg/kickr/v1"
)

func TestParserGolang(t *testing.T) {
	ctx := t.Context()

	t.Run("error_read_gomod", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(destdir, parser.FileGomod), files.RwxRxRxRx))

		// Act
		err := generate.ParserGolang(ctx, destdir, &types.KickrWrapper{})

		// Assert
		assert.ErrorContains(t, err, fmt.Sprintf("read '%s'", parser.FileGomod))
	})

	t.Run("error_parse_hugo", func(t *testing.T) {
		for _, dir := range []string{"", "docs"} {
			t.Run(dir, func(t *testing.T) {
				// Arrange
				destdir := t.TempDir()

				require.NoError(t, os.MkdirAll(filepath.Join(destdir, dir), files.RwxRxRxRx))
				err := os.WriteFile(filepath.Join(destdir, dir, "hugo.toml"), []byte("{ invalid toml }"), files.RwRR)
				require.NoError(t, err)

				config := types.KickrWrapper{
					Kickr: kickr.Kickr{
						CI: &kickr.CI{Website: &kickr.Website{Directory: dir}},
					},
				}

				// Act
				err = generate.ParserGolang(ctx, destdir, &config)

				// Assert
				require.NotErrorIs(t, err, parser.ErrNoHugo)
				assert.ErrorContains(t, err, "parse hugo")
			})
		}
	})

	t.Run("error_no_use_gomod", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()

		err := os.WriteFile(filepath.Join(destdir, parser.FileGowork), []byte(
			`go 1.22

			use (
				./lib1
			)`,
		), files.RwRR)
		require.NoError(t, err)

		config := types.KickrWrapper{}

		// Act
		err = generate.ParserGolang(ctx, destdir, &config)

		// Assert
		require.NotErrorIs(t, err, parser.ErrNoGowork)
		require.ErrorContains(t, err, "read 'go.work'")
	})

	t.Run("success_no_gowork_gomod", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()
		config := types.KickrWrapper{}

		// Act
		err := generate.ParserGolang(ctx, destdir, &config)

		// Assert
		require.NoError(t, err)
		assert.Zero(t, config)
	})

	t.Run("success_hugo", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()

		hugoconfig, err := os.Create(filepath.Join(destdir, "hugo.toml"))
		require.NoError(t, err)
		require.NoError(t, hugoconfig.Close())

		expected := types.KickrWrapper{
			Languages: map[string]any{
				"hugo": []types.Mono[parser.HugoCompose]{
					{Directory: ".", Specifics: parser.HugoCompose{HugoConfig: &parser.HugoConfig{}}},
				},
			},
		}
		config := types.KickrWrapper{}

		// Act
		err = generate.ParserGolang(ctx, destdir, &config)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expected, config)
	})

	t.Run("success_gomod_hugo_doc", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()

		err := os.WriteFile(filepath.Join(destdir, parser.FileGomod), []byte(
			`module github.com/kickr-dev/kickr

			go 1.22`,
		), files.RwRR)
		require.NoError(t, err)

		require.NoError(t, os.MkdirAll(filepath.Join(destdir, "docs"), files.RwxRxRxRx))
		hugoconfig, err := os.Create(filepath.Join(destdir, "docs", "hugo.toml"))
		require.NoError(t, err)
		require.NoError(t, hugoconfig.Close())

		expected := types.KickrWrapper{
			Kickr: kickr.Kickr{
				CI: &kickr.CI{Website: &kickr.Website{Directory: "docs"}},
			},
			Languages: map[string]any{
				"go": parser.Gomod{
					Module: "github.com/kickr-dev/kickr",
					Go:     "1.22",
					Tools:  []string{},
				},
				"hugo": []types.Mono[parser.HugoCompose]{
					{Directory: "docs", Specifics: parser.HugoCompose{HugoConfig: &parser.HugoConfig{}}},
				},
			},
		}
		config := types.KickrWrapper{
			Kickr: kickr.Kickr{
				CI: &kickr.CI{Website: &kickr.Website{Directory: "docs"}},
			},
		}

		// Act
		err = generate.ParserGolang(ctx, destdir, &config)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expected, config)
	})

	t.Run("success_gomod", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()

		err := os.WriteFile(filepath.Join(destdir, parser.FileGomod), []byte(
			`module github.com/kickr-dev/kickr

			go 1.22

			tool (
				example.com/tool-example
			)`,
		), files.RwRR)
		require.NoError(t, err)

		expected := types.KickrWrapper{
			Languages: map[string]any{
				"go": parser.Gomod{
					Module: "github.com/kickr-dev/kickr",
					Go:     "1.22",
					Tools:  []string{"example.com/tool-example"},
				},
			},
		}
		config := types.KickrWrapper{}

		// Act
		err = generate.ParserGolang(ctx, destdir, &config)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expected, config)
	})

	t.Run("success_gowork", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()

		err := os.WriteFile(filepath.Join(destdir, parser.FileGowork), []byte(
			`go 1.22

			use (
				./lib1
			)`,
		), files.RwRR)
		require.NoError(t, err)

		require.NoError(t, os.MkdirAll(filepath.Join(destdir, "lib1"), files.RwxRxRxRx))
		err = os.WriteFile(filepath.Join(destdir, "lib1", parser.FileGomod), []byte("module github.com/kickr-dev/kickr\ngo 1.22"), files.RwRR)
		require.NoError(t, err)

		expected := types.KickrWrapper{
			Languages: map[string]any{
				"go": parser.Gowork{
					Go: "1.22",
					Uses: []parser.GoworkUse{
						{
							Gomod: parser.Gomod{
								Go:     "1.22",
								Module: "github.com/kickr-dev/kickr",
								Tools:  []string{},
							},
							Use: "./lib1",
						},
					},
				},
			},
		}
		config := types.KickrWrapper{}

		// Act
		err = generate.ParserGolang(ctx, destdir, &config)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expected, config)
	})

	t.Run("success_gomod_cmd", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()

		err := os.WriteFile(filepath.Join(destdir, parser.FileGomod), []byte(
			`module github.com/kickr-dev/kickr

			go 1.22`,
		), files.RwRR)
		require.NoError(t, err)

		cmd := filepath.Join(destdir, parser.FolderCMD)
		require.NoError(t, os.MkdirAll(cmd, files.RwxRxRxRx))
		cli := filepath.Join(cmd, "name")
		require.NoError(t, os.MkdirAll(cli, files.RwxRxRxRx))
		main, err := os.Create(filepath.Join(cli, parser.FileMain))
		require.NoError(t, err)
		require.NoError(t, main.Close())

		expected := types.KickrWrapper{
			Executables: parser.Executables{
				Clis: map[string]any{"name": struct{}{}},
			},
			Languages: map[string]any{
				"go": parser.Gomod{
					Module: "github.com/kickr-dev/kickr",
					Go:     "1.22",
					Tools:  []string{},
				},
			},
		}
		config := types.KickrWrapper{}

		// Act
		err = generate.ParserGolang(ctx, destdir, &config)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expected, config)
	})
}

package generate_test

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/kickr-dev/engine/pkg/files"
	"github.com/kickr-dev/engine/pkg/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kickr-dev/kickr/pkg/generate"
	"github.com/kickr-dev/kickr/pkg/generate/types"
)

func TestParserHugo(t *testing.T) {
	ctx := t.Context()

	httpmock.Activate()
	t.Cleanup(httpmock.DeactivateAndReset)

	t.Run("error_parse_hugo", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()
		require.NoError(t, os.MkdirAll(destdir, files.RwxRxRxRx))
		require.NoError(t, os.WriteFile(filepath.Join(destdir, "hugo.toml"), []byte("{ invalid toml }"), files.RwRR))

		repo := types.Repository{
			Modules: []types.Module{
				{Directory: types.RootModule},
			},
		}

		// Act
		err := generate.ParserHugo(http.DefaultClient)(ctx, destdir, &repo)

		// Assert
		require.NotErrorIs(t, err, parser.ErrNoHugo)
		assert.ErrorContains(t, err, "read hugo in '.'")
	})

	t.Run("success_no_hugo", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()
		expected := types.Repository{
			Modules: []types.Module{
				{Directory: types.RootModule},
			},
		}
		repo := types.Repository{
			Modules: []types.Module{
				{Directory: types.RootModule},
			},
		}

		// Act
		err := generate.ParserHugo(http.DefaultClient)(ctx, destdir, &repo)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expected, repo)
	})

	t.Run("success_root", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()
		hugoconfig, err := os.Create(filepath.Join(destdir, "hugo.toml"))
		require.NoError(t, err)
		require.NoError(t, hugoconfig.Close())

		expected := types.Repository{
			Modules: []types.Module{
				{
					Directory: types.RootModule,
					Languages: map[string]any{types.LanguageHugo: parser.HugoCompose{HugoConfig: &parser.HugoConfig{}}},
				},
			},
		}
		repo := types.Repository{
			Modules: []types.Module{
				{Directory: types.RootModule},
			},
		}

		// Act
		err = generate.ParserHugo(http.DefaultClient)(ctx, destdir, &repo)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expected, repo)
	})

	t.Run("success_docs", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(destdir, "docs"), files.RwxRxRxRx))
		hugoconfig, err := os.Create(filepath.Join(destdir, "docs", "hugo.toml"))
		require.NoError(t, err)
		require.NoError(t, hugoconfig.Close())

		expected := types.Repository{
			Modules: []types.Module{
				{Directory: types.RootModule},
				{
					Directory: "docs",
					Languages: map[string]any{types.LanguageHugo: parser.HugoCompose{HugoConfig: &parser.HugoConfig{}}},
				},
			},
		}
		repo := types.Repository{
			Modules: []types.Module{
				{Directory: types.RootModule},
				{Directory: "docs"},
			},
		}

		// Act
		err = generate.ParserHugo(http.DefaultClient)(ctx, destdir, &repo)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expected, repo)
	})

	t.Run("success_container_image_max", func(t *testing.T) {
		// Arrange
		t.Setenv("DOCKER_CONFIG", t.TempDir())

		manifest := "https://ghcr.io/v2/gohugoio/hugo/manifests/v0.160.0"
		digest := "sha256:71d7c66aa2305f7d239c35d3c04b688c709ee0243d792afb868a3731494f08bb"
		httpmock.RegisterResponder(http.MethodHead, manifest,
			httpmock.NewStringResponder(http.StatusOK, "{}").
				HeaderSet(http.Header{"Content-Type": {"application/vnd.oci.image.index.v1+json"}, "Docker-Content-Digest": {digest}}).
				SetContentLength())

		destdir := t.TempDir()
		content := "[module.hugoVersion]\nmax = '0.160.0'\nmin = '0.150.0'"
		require.NoError(t, os.WriteFile(filepath.Join(destdir, "hugo.toml"), []byte(content), files.RwRR))

		expected := types.Repository{
			Modules: []types.Module{
				{
					Directory: types.RootModule,
					Languages: map[string]any{
						types.LanguageHugo: parser.HugoCompose{
							ContainerImage: "ghcr.io/gohugoio/hugo:v0.160.0@" + digest,
							HugoConfig: &parser.HugoConfig{
								Module: parser.HugoModule{HugoVersion: parser.HugoVersion{Max: "0.160.0", Min: "0.150.0"}},
							},
						},
					},
				},
			},
		}
		repo := types.Repository{
			Modules: []types.Module{
				{Directory: types.RootModule},
			},
		}

		// Act
		err := generate.ParserHugo(http.DefaultClient)(ctx, destdir, &repo)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expected, repo)
	})

	t.Run("success_container_image_min", func(t *testing.T) {
		// Arrange
		t.Setenv("DOCKER_CONFIG", t.TempDir())

		manifest := "https://ghcr.io/v2/gohugoio/hugo/manifests/v0.150.0"
		digest := "sha256:71d7c66aa2305f7d239c35d3c04b688c709ee0243d792afb868a3731494f08bb"
		httpmock.RegisterResponder(http.MethodHead, manifest,
			httpmock.NewStringResponder(http.StatusOK, "{}").
				HeaderSet(http.Header{"Content-Type": {"application/vnd.oci.image.index.v1+json"}, "Docker-Content-Digest": {digest}}).
				SetContentLength())

		destdir := t.TempDir()
		content := "[module.hugoVersion]\nmin = 'v0.150.0'"
		require.NoError(t, os.WriteFile(filepath.Join(destdir, "hugo.toml"), []byte(content), files.RwRR))

		expected := types.Repository{
			Modules: []types.Module{
				{
					Directory: types.RootModule,
					Languages: map[string]any{
						types.LanguageHugo: parser.HugoCompose{
							ContainerImage: "ghcr.io/gohugoio/hugo:v0.150.0@" + digest,
							HugoConfig: &parser.HugoConfig{
								Module: parser.HugoModule{HugoVersion: parser.HugoVersion{Min: "v0.150.0"}},
							},
						},
					},
				},
			},
		}
		repo := types.Repository{
			Modules: []types.Module{
				{Directory: types.RootModule},
			},
		}

		// Act
		err := generate.ParserHugo(http.DefaultClient)(ctx, destdir, &repo)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expected, repo)
	})
}

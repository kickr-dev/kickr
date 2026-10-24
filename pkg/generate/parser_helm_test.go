package generate_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/kickr-dev/engine/pkg/parser"
	"gitlab.com/kickr-dev/workload"

	"github.com/kickr-dev/kickr/pkg/generate"
	"github.com/kickr-dev/kickr/pkg/generate/types"
	kickr "github.com/kickr-dev/kickr/pkg/kickr/v1"
)

func TestParserChart(t *testing.T) {
	ctx := t.Context()

	t.Run("success_description", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()

		expected := types.Repository{
			Config: kickr.Kickr{Description: "a description", Helm: &kickr.Helm{}},
			Modules: []types.Module{
				{
					Directory: types.RootModule,
					Languages: map[string]any{types.LanguageHelm: map[string]any{
						"description":     "a description",
						"docker":          map[string]any{},
						"workloadVersion": workload.Version(),

						"clis":    map[string]any{},
						"crons":   map[string]any{},
						"jobs":    map[string]any{},
						"workers": map[string]any{},

						"maintainers": nil,
						"projectName": "",
						"projectPath": "",
					}},
				},
			},
		}
		repo := types.Repository{
			Config: kickr.Kickr{Description: "a description", Helm: &kickr.Helm{}},
			Modules: []types.Module{
				{Directory: types.RootModule},
			},
		}

		// Act
		err := generate.ParserHelm(ctx, destdir, &repo)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expected, repo)
	})

	t.Run("success_module_image_repository", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()

		expected := types.Repository{
			Config: kickr.Kickr{Helm: &kickr.Helm{}},
			VCS:    parser.VCS{ProjectPath: "org/repo"},
			Modules: []types.Module{
				{
					Directory:   types.RootModule,
					Config:      kickr.Module{Path: types.RootModule, Deployment: &kickr.Deployment{Target: kickr.DeploymentTargetHelm}},
					Executables: parser.Executables{Workers: map[string]any{"main": struct{}{}}},
					Languages: map[string]any{types.LanguageHelm: map[string]any{
						"description":     "",
						"docker":          map[string]any{},
						"workloadVersion": workload.Version(),

						"clis":  map[string]any{},
						"crons": map[string]any{},
						"jobs":  map[string]any{},
						"workers": map[string]any{
							"main": map[string]any{"image": map[string]any{"repository": "org/repo"}},
							"api":  map[string]any{"image": map[string]any{"repository": "org/repo/services-api"}},
						},

						"maintainers": nil,
						"projectName": "",
						"projectPath": "org/repo",
					}},
				},
				{
					Directory:   "services/api",
					Config:      kickr.Module{Path: "services/api", Deployment: &kickr.Deployment{Target: kickr.DeploymentTargetHelm}},
					Executables: parser.Executables{Workers: map[string]any{"api": struct{}{}}},
				},
			},
		}
		repo := types.Repository{
			Config: kickr.Kickr{Helm: &kickr.Helm{}},
			VCS:    parser.VCS{ProjectPath: "org/repo"},
			Modules: []types.Module{
				{
					Directory:   types.RootModule,
					Config:      kickr.Module{Path: types.RootModule, Deployment: &kickr.Deployment{Target: kickr.DeploymentTargetHelm}},
					Executables: parser.Executables{Workers: map[string]any{"main": struct{}{}}},
				},
				{
					Directory:   "services/api",
					Config:      kickr.Module{Path: "services/api", Deployment: &kickr.Deployment{Target: kickr.DeploymentTargetHelm}},
					Executables: parser.Executables{Workers: map[string]any{"api": struct{}{}}},
				},
			},
		}

		// Act
		err := generate.ParserHelm(ctx, destdir, &repo)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expected, repo)
	})
}

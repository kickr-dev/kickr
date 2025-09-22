package generate_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/kickr-dev/engine/pkg/files"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kickr-dev/kickr/pkg/generate"
	"github.com/kickr-dev/kickr/pkg/generate/types"
	"github.com/kickr-dev/kickr/pkg/kickr/v1"
)

func TestParserTerraform(t *testing.T) {
	ctx := t.Context()

	t.Run("error_invalid_module", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()
		config := types.Repository{
			Kickr: kickr.Kickr{
				Terraform: &kickr.Terraform{
					Modules: []string{"path"},
				},
			},
		}

		// Act
		err := generate.ParserTerraform(ctx, destdir, &config)

		// Assert
		assert.ErrorContains(t, err, "module 'path' isn't a terraform module")
	})

	t.Run("error_invalid_module_file", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(destdir, "main.tf"), []byte("invalid file"), files.RwRR))

		// Act
		err := generate.ParserTerraform(ctx, destdir, &types.Repository{})

		// Assert
		assert.ErrorContains(t, err, "load module:")
	})

	t.Run("error_invalid_submodules_file", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()

		require.NoError(t, os.MkdirAll(filepath.Join(destdir, "module"), files.RwxRxRxRx))
		require.NoError(t, os.WriteFile(filepath.Join(destdir, "module", "main.tf"), []byte("invalid file"), files.RwRR))

		require.NoError(t, os.MkdirAll(filepath.Join(destdir, "another_module"), files.RwxRxRxRx))
		require.NoError(t, os.WriteFile(filepath.Join(destdir, "another_module", "main.tf"), []byte("invalid file"), files.RwRR))

		config := types.Repository{
			Kickr: kickr.Kickr{
				Terraform: &kickr.Terraform{Modules: []string{"module", "another_module"}},
			},
		}

		// Act
		err := generate.ParserTerraform(ctx, destdir, &config)

		// Assert
		assert.ErrorContains(t, err, "load module 'module':")
		assert.ErrorContains(t, err, "load module 'another_module':")
	})

	t.Run("success_no_terraform", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()
		config := types.Repository{}

		// Act
		err := generate.ParserTerraform(ctx, destdir, &config)

		// Assert
		require.NoError(t, err)
		assert.Zero(t, config)
	})

	t.Run("success_root", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(destdir, "main.tf"), []byte(`variable "my_var" {}`), files.RwRR))
		require.NoError(t, os.WriteFile(filepath.Join(destdir, "state.tf"), []byte(`terraform { backend "local" {} }`), files.RwRR))

		config := types.Repository{}
		expected := types.Repository{
			Languages: map[string]any{
				"terraform": []types.Mono[generate.TerraformModule]{
					{
						Directory: ".",
						Specifics: generate.TerraformModule{
							Backend: "local",
							Module: &tfconfig.Module{
								Path: destdir,
								Variables: map[string]*tfconfig.Variable{
									"my_var": {
										Name:     "my_var",
										Required: true,
										Pos: tfconfig.SourcePos{
											Filename: filepath.Join(destdir, "main.tf"),
											Line:     1,
										},
									},
								},
								Outputs:           map[string]*tfconfig.Output{},
								RequiredProviders: map[string]*tfconfig.ProviderRequirement{},
								ProviderConfigs:   map[string]*tfconfig.ProviderConfig{},
								ManagedResources:  map[string]*tfconfig.Resource{},
								DataResources:     map[string]*tfconfig.Resource{},
								ModuleCalls:       map[string]*tfconfig.ModuleCall{},
							},
						},
					},
				},
			},
		}

		// Act
		err := generate.ParserTerraform(ctx, destdir, &config)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expected, config)
	})

	t.Run("success_subdirectories", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()

		require.NoError(t, os.MkdirAll(filepath.Join(destdir, "module"), files.RwxRxRxRx))
		require.NoError(t, os.WriteFile(filepath.Join(destdir, "module", "versions.tf"), []byte(`terraform { backend "s3" {} }`), files.RwRR))
		require.NoError(t, os.WriteFile(filepath.Join(destdir, "module", "main.tf"), []byte(`variable "module_var" {}`), files.RwRR))

		require.NoError(t, os.MkdirAll(filepath.Join(destdir, "another_module"), files.RwxRxRxRx))
		require.NoError(t, os.WriteFile(filepath.Join(destdir, "another_module", "main.tf"), []byte(`variable "another_module_var" {}`), files.RwRR))
		require.NoError(t, os.WriteFile(filepath.Join(destdir, "another_module", "backend.tf"), []byte(`terraform { backend "http" {} }`), files.RwRR))

		config := types.Repository{
			Kickr: kickr.Kickr{
				Terraform: &kickr.Terraform{Modules: []string{"module", "another_module"}},
			},
		}
		expected := types.Repository{
			Kickr: kickr.Kickr{
				Terraform: &kickr.Terraform{Modules: []string{"module", "another_module"}},
			},
			Languages: map[string]any{
				"terraform": []types.Mono[generate.TerraformModule]{
					{
						Directory: "module",
						Specifics: generate.TerraformModule{
							Backend: "s3",
							Module: &tfconfig.Module{
								Path: filepath.Join(destdir, "module"),
								Variables: map[string]*tfconfig.Variable{
									"module_var": {
										Name:     "module_var",
										Required: true,
										Pos: tfconfig.SourcePos{
											Filename: filepath.Join(destdir, "module", "main.tf"),
											Line:     1,
										},
									},
								},
								Outputs:           map[string]*tfconfig.Output{},
								RequiredProviders: map[string]*tfconfig.ProviderRequirement{},
								ProviderConfigs:   map[string]*tfconfig.ProviderConfig{},
								ManagedResources:  map[string]*tfconfig.Resource{},
								DataResources:     map[string]*tfconfig.Resource{},
								ModuleCalls:       map[string]*tfconfig.ModuleCall{},
							},
						},
					},
					{
						Directory: "another_module",
						Specifics: generate.TerraformModule{
							Backend: "http",
							Module: &tfconfig.Module{
								Path: filepath.Join(destdir, "another_module"),
								Variables: map[string]*tfconfig.Variable{
									"another_module_var": {
										Name:     "another_module_var",
										Required: true,
										Pos: tfconfig.SourcePos{
											Filename: filepath.Join(destdir, "another_module", "main.tf"),
											Line:     1,
										},
									},
								},
								Outputs:           map[string]*tfconfig.Output{},
								RequiredProviders: map[string]*tfconfig.ProviderRequirement{},
								ProviderConfigs:   map[string]*tfconfig.ProviderConfig{},
								ManagedResources:  map[string]*tfconfig.Resource{},
								DataResources:     map[string]*tfconfig.Resource{},
								ModuleCalls:       map[string]*tfconfig.ModuleCall{},
							},
						},
					},
				},
			},
		}

		// Act
		err := generate.ParserTerraform(ctx, destdir, &config)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expected, config)
	})
}

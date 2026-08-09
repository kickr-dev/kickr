package kickr_test

import (
	"os"
	"path/filepath"
	"testing"

	"dario.cat/mergo"
	"github.com/kickr-dev/engine/pkg/files"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	schemas "github.com/kickr-dev/kickr/.schemas"
	kickr "github.com/kickr-dev/kickr/pkg/kickr/v1"
)

type testcase struct {
	kickr.Kickr

	Name string
}

// merge combines the two input kickr configurations.
func merge(t testing.TB, base, complement kickr.Kickr) kickr.Kickr {
	t.Helper()

	require.NoError(t, mergo.Merge(&base, complement, mergo.WithAppendSlice))
	return base
}

// validate runs the same validation than the generate command does on a '.kickr.yml' file built from conf.
func validate(t *testing.T, conf kickr.Kickr) error {
	t.Helper()
	destdir := t.TempDir()

	require.NoError(t, files.WriteYAML(filepath.Join(destdir, ".kickr.yml"), conf, kickr.EncodeOpts()...))
	return files.Validate(files.ReadYAMLFunc(schemas.FS(), kickr.Schema), files.ReadYAMLFunc(os.DirFS(destdir), ".kickr.yml"))
}

func TestSchemaModules_Errors(t *testing.T) {
	base := kickr.Kickr{Version: 1, Maintainers: []*kickr.Maintainer{{Name: "maintainer"}}}

	t.Run("github_multiple_pages", func(t *testing.T) {
		// Arrange
		conf := merge(t, base, kickr.Kickr{
			GitHub: &kickr.GitHub{},
			Modules: []kickr.Module{
				{Path: "docs", Deployment: &kickr.Deployment{Target: kickr.DeploymentTargetPages}},
				{Path: "blog", Deployment: &kickr.Deployment{Target: kickr.DeploymentTargetPages}},
			},
		})

		// Act
		err := validate(t, conf)

		// Assert
		assert.ErrorContains(t, err, "at '/modules'")
		assert.ErrorContains(t, err, "max 1 items required to match contains schema")
	})

	t.Run("gitlab_multiple_pages", func(t *testing.T) {
		// Arrange
		conf := merge(t, base, kickr.Kickr{
			GitLab: &kickr.GitLab{},
			Modules: []kickr.Module{
				{Path: "docs", Deployment: &kickr.Deployment{Target: kickr.DeploymentTargetPages}},
				{Path: "blog", Deployment: &kickr.Deployment{Target: kickr.DeploymentTargetPages}},
			},
		})

		// Act
		err := validate(t, conf)

		// Assert
		assert.ErrorContains(t, err, "at '/modules'")
		assert.ErrorContains(t, err, "max 1 items required to match contains schema")
	})

	t.Run("github_mixed_engines", func(t *testing.T) {
		// Arrange
		conf := merge(t, base, kickr.Kickr{
			GitHub: &kickr.GitHub{},
			Modules: []kickr.Module{
				{Path: "infra/prod", Terraform: &kickr.Terraform{Engine: kickr.TerraformEngineTerraform}},
				{Path: "infra/staging", Terraform: &kickr.Terraform{Engine: kickr.TerraformEngineTofu}},
			},
		})

		// Act
		err := validate(t, conf)

		// Assert
		assert.ErrorContains(t, err, "at '/modules'")
		assert.ErrorContains(t, err, "not")
	})

	t.Run("gitlab_mixed_engines", func(t *testing.T) {
		// Arrange
		conf := merge(t, base, kickr.Kickr{
			GitLab: &kickr.GitLab{},
			Modules: []kickr.Module{
				{Path: "infra/prod", Terraform: &kickr.Terraform{Engine: kickr.TerraformEngineTerraform}},
				{Path: "infra/staging", Terraform: &kickr.Terraform{Engine: kickr.TerraformEngineTofu}},
			},
		})

		// Act
		err := validate(t, conf)

		// Assert
		assert.ErrorContains(t, err, "at '/modules'")
		assert.ErrorContains(t, err, "not")
	})

	t.Run("github_missing_path", func(t *testing.T) {
		// Arrange
		conf := merge(t, base, kickr.Kickr{
			GitHub: &kickr.GitHub{},
			Modules: []kickr.Module{
				{Deployment: &kickr.Deployment{Target: kickr.DeploymentTargetPages}},
			},
		})

		// Act
		err := validate(t, conf)

		// Assert
		assert.ErrorContains(t, err, "at '/modules/0'")
		assert.ErrorContains(t, err, "'path'")
	})

	t.Run("gitlab_missing_path", func(t *testing.T) {
		// Arrange
		conf := merge(t, base, kickr.Kickr{
			GitLab: &kickr.GitLab{},
			Modules: []kickr.Module{
				{Deployment: &kickr.Deployment{Target: kickr.DeploymentTargetPages}},
			},
		})

		// Act
		err := validate(t, conf)

		// Assert
		assert.ErrorContains(t, err, "at '/modules/0'")
		assert.ErrorContains(t, err, "'path'")
	})

	t.Run("github_missing_deployment_target", func(t *testing.T) {
		// Arrange
		conf := merge(t, base, kickr.Kickr{GitHub: &kickr.GitHub{}, Modules: []kickr.Module{{Path: "docs", Deployment: &kickr.Deployment{Auto: true}}}})

		// Act
		err := validate(t, conf)

		// Assert
		assert.ErrorContains(t, err, "at '/modules/0/deployment'")
		assert.ErrorContains(t, err, "'target'")
	})

	t.Run("gitlab_missing_deployment_target", func(t *testing.T) {
		// Arrange
		conf := merge(t, base, kickr.Kickr{GitLab: &kickr.GitLab{}, Modules: []kickr.Module{{Path: "docs", Deployment: &kickr.Deployment{Auto: true}}}})

		// Act
		err := validate(t, conf)

		// Assert
		assert.ErrorContains(t, err, "at '/modules/0/deployment'")
		assert.ErrorContains(t, err, "'target'")
	})

	t.Run("github_missing_engine", func(t *testing.T) {
		// Arrange
		conf := merge(t, base, kickr.Kickr{GitHub: &kickr.GitHub{}, Modules: []kickr.Module{
			{Path: "infra", Terraform: &kickr.Terraform{Apply: kickr.TerraformApplyManual, Environments: []string{kickr.EnvironmentProduction}}},
		}})

		// Act
		err := validate(t, conf)

		// Assert
		assert.ErrorContains(t, err, "at '/modules/0/terraform'")
		assert.ErrorContains(t, err, "'engine'")
	})

	t.Run("gitlab_missing_engine", func(t *testing.T) {
		// Arrange
		conf := merge(t, base, kickr.Kickr{GitLab: &kickr.GitLab{}, Modules: []kickr.Module{
			{Path: "infra", Terraform: &kickr.Terraform{Apply: kickr.TerraformApplyManual, Environments: []string{kickr.EnvironmentProduction}}},
		}})

		// Act
		err := validate(t, conf)

		// Assert
		assert.ErrorContains(t, err, "at '/modules/0/terraform'")
		assert.ErrorContains(t, err, "'engine'")
	})

	t.Run("github_apply_without_environments", func(t *testing.T) {
		// Arrange
		conf := merge(t, base, kickr.Kickr{GitHub: &kickr.GitHub{}, Modules: []kickr.Module{
			{Path: "infra", Terraform: &kickr.Terraform{Engine: kickr.TerraformEngineTofu, Apply: kickr.TerraformApplyManual}},
		}})

		// Act
		err := validate(t, conf)

		// Assert
		assert.ErrorContains(t, err, "at '/modules/0/terraform'")
		assert.ErrorContains(t, err, "'environments'")
	})

	t.Run("gitlab_apply_without_environments", func(t *testing.T) {
		// Arrange
		conf := merge(t, base, kickr.Kickr{GitLab: &kickr.GitLab{}, Modules: []kickr.Module{
			{Path: "infra", Terraform: &kickr.Terraform{Engine: kickr.TerraformEngineTofu, Apply: kickr.TerraformApplyManual}},
		}})

		// Act
		err := validate(t, conf)

		// Assert
		assert.ErrorContains(t, err, "at '/modules/0/terraform'")
		assert.ErrorContains(t, err, "'environments'")
	})

	t.Run("github_environments_without_apply", func(t *testing.T) {
		// Arrange
		conf := merge(t, base, kickr.Kickr{GitHub: &kickr.GitHub{}, Modules: []kickr.Module{
			{Path: "infra", Terraform: &kickr.Terraform{Engine: kickr.TerraformEngineTofu, Environments: []string{kickr.EnvironmentProduction}}},
		}})

		// Act
		err := validate(t, conf)

		// Assert
		assert.ErrorContains(t, err, "at '/modules/0/terraform'")
		assert.ErrorContains(t, err, "'apply'")
	})

	t.Run("gitlab_environments_without_apply", func(t *testing.T) {
		// Arrange
		conf := merge(t, base, kickr.Kickr{GitLab: &kickr.GitLab{}, Modules: []kickr.Module{
			{Path: "infra", Terraform: &kickr.Terraform{Engine: kickr.TerraformEngineTofu, Environments: []string{kickr.EnvironmentProduction}}},
		}})

		// Act
		err := validate(t, conf)

		// Assert
		assert.ErrorContains(t, err, "at '/modules/0/terraform'")
		assert.ErrorContains(t, err, "'apply'")
	})

	t.Run("no_platform_deployment_without_platform", func(t *testing.T) {
		// Arrange
		conf := merge(t, base, kickr.Kickr{Modules: []kickr.Module{{Path: "docs", Deployment: &kickr.Deployment{Target: kickr.DeploymentTargetNetlify}}}})

		// Act
		err := validate(t, conf)

		// Assert
		assert.ErrorContains(t, err, "at '/modules/0/deployment': must not be provided")
	})
}

func TestSchemaModules(t *testing.T) {
	base := kickr.Kickr{Version: 1, Maintainers: []*kickr.Maintainer{{Name: "maintainer"}}}

	cases := []testcase{
		{Name: "github_no_modules", Kickr: kickr.Kickr{GitHub: &kickr.GitHub{}}},
		{Name: "gitlab_no_modules", Kickr: kickr.Kickr{GitLab: &kickr.GitLab{}}},
		{
			Name: "github_one_netlify_and_one_pages",
			Kickr: kickr.Kickr{
				GitHub: &kickr.GitHub{},
				Modules: []kickr.Module{
					{Path: "."},
					{Path: "docs", Deployment: &kickr.Deployment{Target: kickr.DeploymentTargetNetlify, Auto: true}},
					{Path: "blog", Deployment: &kickr.Deployment{Target: kickr.DeploymentTargetPages}},
				},
			},
		},
		{
			Name: "gitlab_one_netlify_and_one_pages",
			Kickr: kickr.Kickr{
				GitLab: &kickr.GitLab{},
				Modules: []kickr.Module{
					{Path: "."},
					{Path: "docs", Deployment: &kickr.Deployment{Target: kickr.DeploymentTargetNetlify, Auto: true}},
					{Path: "blog", Deployment: &kickr.Deployment{Target: kickr.DeploymentTargetPages}},
				},
			},
		},
		{
			Name: "github_multiple_netlify",
			Kickr: kickr.Kickr{
				GitHub: &kickr.GitHub{},
				Modules: []kickr.Module{
					{Path: "docs", Deployment: &kickr.Deployment{Target: kickr.DeploymentTargetNetlify}},
					{Path: "blog", Deployment: &kickr.Deployment{Target: kickr.DeploymentTargetNetlify}},
				},
			},
		},
		{
			Name: "gitlab_multiple_netlify",
			Kickr: kickr.Kickr{
				GitLab: &kickr.GitLab{},
				Modules: []kickr.Module{
					{Path: "docs", Deployment: &kickr.Deployment{Target: kickr.DeploymentTargetNetlify}},
					{Path: "blog", Deployment: &kickr.Deployment{Target: kickr.DeploymentTargetNetlify}},
				},
			},
		},
		{
			Name: "github_multiple_tofu",
			Kickr: kickr.Kickr{
				GitHub: &kickr.GitHub{},
				Modules: []kickr.Module{
					{
						Path: "infra/prod",
						Terraform: &kickr.Terraform{
							Engine:       kickr.TerraformEngineTofu,
							Apply:        kickr.TerraformApplyManual,
							Environments: []string{kickr.EnvironmentProduction},
						},
					},
					{
						Path: "infra/staging",
						Terraform: &kickr.Terraform{
							Engine:       kickr.TerraformEngineTofu,
							Apply:        kickr.TerraformApplyAuto,
							Environments: []string{kickr.EnvironmentStaging},
						},
					},
				},
			},
		},
		{
			Name: "gitlab_multiple_tofu",
			Kickr: kickr.Kickr{
				GitLab: &kickr.GitLab{},
				Modules: []kickr.Module{
					{
						Path: "infra/prod",
						Terraform: &kickr.Terraform{
							Engine:       kickr.TerraformEngineTofu,
							Apply:        kickr.TerraformApplyManual,
							Environments: []string{kickr.EnvironmentProduction},
						},
					},
					{
						Path: "infra/staging",
						Terraform: &kickr.Terraform{
							Engine:       kickr.TerraformEngineTofu,
							Apply:        kickr.TerraformApplyAuto,
							Environments: []string{kickr.EnvironmentStaging},
						},
					},
				},
			},
		},
		{
			Name: "github_multiple_terraform",
			Kickr: kickr.Kickr{
				GitHub: &kickr.GitHub{},
				Modules: []kickr.Module{
					{
						Path: "infra/prod",
						Terraform: &kickr.Terraform{
							Engine:       kickr.TerraformEngineTerraform,
							Apply:        kickr.TerraformApplyManual,
							Environments: []string{kickr.EnvironmentProduction},
						},
					},
					{
						Path: "infra/staging",
						Terraform: &kickr.Terraform{
							Engine:       kickr.TerraformEngineTerraform,
							Apply:        kickr.TerraformApplyAuto,
							Environments: []string{kickr.EnvironmentStaging},
						},
					},
				},
			},
		},
		{
			Name: "gitlab_multiple_terraform",
			Kickr: kickr.Kickr{
				GitLab: &kickr.GitLab{},
				Modules: []kickr.Module{
					{
						Path: "infra/prod",
						Terraform: &kickr.Terraform{
							Engine:       kickr.TerraformEngineTerraform,
							Apply:        kickr.TerraformApplyManual,
							Environments: []string{kickr.EnvironmentProduction},
						},
					},
					{
						Path: "infra/staging",
						Terraform: &kickr.Terraform{
							Engine:       kickr.TerraformEngineTerraform,
							Apply:        kickr.TerraformApplyAuto,
							Environments: []string{kickr.EnvironmentStaging},
						},
					},
				},
			},
		},
		{
			Name: "github_deployment_helm",
			Kickr: kickr.Kickr{
				GitHub: &kickr.GitHub{},
				Modules: []kickr.Module{
					{Path: "apps/api", Deployment: &kickr.Deployment{Target: kickr.DeploymentTargetHelm}},
					{Path: "apps/worker", Deployment: &kickr.Deployment{Target: kickr.DeploymentTargetHelm}},
				},
			},
		},
		{
			Name: "gitlab_deployment_helm",
			Kickr: kickr.Kickr{
				GitLab: &kickr.GitLab{},
				Modules: []kickr.Module{
					{Path: "apps/api", Deployment: &kickr.Deployment{Target: kickr.DeploymentTargetHelm}},
					{Path: "apps/worker", Deployment: &kickr.Deployment{Target: kickr.DeploymentTargetHelm}},
				},
			},
		},
		{
			Name: "github_exclusions",
			Kickr: kickr.Kickr{
				GitHub: &kickr.GitHub{},
				Modules: []kickr.Module{
					{Path: "docs", Exclude: []string{kickr.ModuleExcludeDocker, kickr.ModuleExcludeMakefile}},
				},
			},
		},
		{
			Name: "gitlab_exclusions",
			Kickr: kickr.Kickr{
				GitLab: &kickr.GitLab{},
				Modules: []kickr.Module{
					{Path: "docs", Exclude: []string{kickr.ModuleExcludeDocker, kickr.ModuleExcludeMakefile}},
				},
			},
		},
		{
			Name: "no_platform_path_and_exclude",
			Kickr: kickr.Kickr{
				Modules: []kickr.Module{
					{Path: "."},
					{Path: "docs", Exclude: []string{kickr.ModuleExcludeDocker}},
				},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			// Arrange
			conf := merge(t, base, tc.Kickr)

			// Act
			err := validate(t, conf)

			// Assert
			assert.NoError(t, err)
		})
	}
}

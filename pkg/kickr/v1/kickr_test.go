package kickr_test

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"dario.cat/mergo"
	"github.com/kickr-dev/engine/pkg/files"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	schemas "github.com/kickr-dev/kickr/.schemas"
	"github.com/kickr-dev/kickr/pkg/generate/types"
	"github.com/kickr-dev/kickr/pkg/kickr/v1"
)

// merge combines the input kickr configuration on top of the schema's minimal valid configuration.
func merge(t testing.TB, config kickr.Kickr) kickr.Kickr {
	t.Helper()

	result := kickr.Kickr{Version: 1, Maintainers: []*kickr.Maintainer{{Name: "maintainer"}}}
	require.NoError(t, mergo.Merge(&result, config, mergo.WithAppendSlice))
	return result
}

// validate runs the same validation than the generate command does on a '.kickr.yml' file built from config.
func validate(t *testing.T, config kickr.Kickr) error {
	t.Helper()
	destdir := t.TempDir()

	require.NoError(t, files.WriteYAML(filepath.Join(destdir, ".kickr.yml"), config, kickr.EncodeOpts()...))
	return files.Validate(files.ReadYAMLFunc(schemas.FS(), kickr.Schema), files.ReadYAMLFunc(os.DirFS(destdir), ".kickr.yml"))
}

func TestValidate_Errors(t *testing.T) {
	type testcase struct {
		Config      kickr.Kickr
		ErrContains []string
		Name        string
	}

	cases := []testcase{
		// required
		{
			Name:        "missing_version",
			Config:      kickr.Kickr{Maintainers: []*kickr.Maintainer{{Name: "maintainer"}}},
			ErrContains: []string{"at '/': missing property 'version'"},
		},
		{
			Name:        "missing_maintainers",
			Config:      kickr.Kickr{Version: 1},
			ErrContains: []string{"at '/': missing property 'maintainers'"},
		},
		{
			Name:        "maintainer_missing_name",
			Config:      merge(t, kickr.Kickr{Maintainers: []*kickr.Maintainer{{Email: "maintainer@example.com"}}}),
			ErrContains: []string{"at '/maintainers/1': missing property 'name'"},
		},
		{
			Name:        "module_missing_path",
			Config:      merge(t, kickr.Kickr{Modules: []kickr.Module{{}}}),
			ErrContains: []string{"at '/modules/0': missing property 'path'"},
		},
		{
			Name:        "terraform_missing_engine",
			Config:      merge(t, kickr.Kickr{Modules: []kickr.Module{{Path: types.RootModule, Terraform: &kickr.Terraform{}}}}),
			ErrContains: []string{"at '/modules/0/terraform': missing property 'engine'"},
		},
		{
			Name: "module_deployment_missing_target",
			Config: merge(t, kickr.Kickr{
				GitHub:  &kickr.GitHub{},
				Modules: []kickr.Module{{Path: types.RootModule, Deployment: &kickr.Deployment{Auto: true}}},
			}),
			ErrContains: []string{"at '/modules/0/deployment': missing property 'target'"},
		},

		// CI-gating: docker/helm/module.deployment/module.terraform ci fields forbidden without github/gitlab
		{
			Name:        "no_ci_docker_auto",
			Config:      merge(t, kickr.Kickr{Docker: &kickr.Docker{Auto: true}}),
			ErrContains: []string{"at '/docker/auto': must not be provided"},
		},
		{
			Name:        "no_ci_docker_path",
			Config:      merge(t, kickr.Kickr{Docker: &kickr.Docker{Path: "kickr-dev/kickr"}}),
			ErrContains: []string{"at '/docker/path': must not be provided"},
		},
		{
			Name:        "no_ci_docker_registry",
			Config:      merge(t, kickr.Kickr{Docker: &kickr.Docker{Registry: "ghcr.io"}}),
			ErrContains: []string{"at '/docker/registry': must not be provided"},
		},
		{
			Name:        "no_ci_helm_deploy",
			Config:      merge(t, kickr.Kickr{Helm: &kickr.Helm{Deploy: kickr.HelmDeployAuto}}),
			ErrContains: []string{"at '/helm/deploy': must not be provided"},
		},
		{
			Name:        "no_ci_helm_environments",
			Config:      merge(t, kickr.Kickr{Helm: &kickr.Helm{Environments: []string{kickr.EnvironmentStaging}}}),
			ErrContains: []string{"at '/helm/environments': must not be provided"},
		},
		{
			Name:        "no_ci_helm_publish",
			Config:      merge(t, kickr.Kickr{Helm: &kickr.Helm{Publish: kickr.HelmPublishAuto}}),
			ErrContains: []string{"at '/helm/publish': must not be provided"},
		},
		{
			Name:        "no_ci_helm_registry",
			Config:      merge(t, kickr.Kickr{Helm: &kickr.Helm{Registry: "ghcr.io"}}),
			ErrContains: []string{"at '/helm/registry': must not be provided"},
		},
		{
			Name:        "no_ci_helm_path",
			Config:      merge(t, kickr.Kickr{Helm: &kickr.Helm{Path: "kickr-dev/kickr"}}),
			ErrContains: []string{"at '/helm/path': must not be provided"},
		},
		{
			Name: "no_ci_module_deployment",
			Config: merge(t, kickr.Kickr{
				Modules: []kickr.Module{{Path: types.RootModule, Deployment: &kickr.Deployment{Target: kickr.DeploymentTargetPages}}},
			}),
			ErrContains: []string{"at '/modules/0/deployment': must not be provided"},
		},
		{
			Name: "no_ci_terraform_apply",
			Config: merge(t, kickr.Kickr{
				Modules: []kickr.Module{
					{Path: types.RootModule, Terraform: &kickr.Terraform{Engine: kickr.TerraformEngineTofu, Apply: kickr.TerraformApplyAuto}},
				},
			}),
			ErrContains: []string{"at '/modules/0/terraform/apply': must not be provided"},
		},
		{
			Name: "no_ci_terraform_environments",
			Config: merge(t, kickr.Kickr{
				Modules: []kickr.Module{
					{
						Path:      types.RootModule,
						Terraform: &kickr.Terraform{Engine: kickr.TerraformEngineTofu, Environments: []string{kickr.EnvironmentStaging}},
					},
				},
			}),
			ErrContains: []string{"at '/modules/0/terraform/environments': must not be provided"},
		},

		// dependentRequired: ci/helm and ci/terraform
		{
			Name:        "helm_environments_without_deploy",
			Config:      merge(t, kickr.Kickr{GitHub: &kickr.GitHub{}, Helm: &kickr.Helm{Environments: []string{kickr.EnvironmentStaging}}}),
			ErrContains: []string{"at '/helm': properties 'deploy' required, if 'environments' exists"},
		},
		{
			Name:        "helm_deploy_without_environments",
			Config:      merge(t, kickr.Kickr{GitHub: &kickr.GitHub{}, Helm: &kickr.Helm{Deploy: kickr.HelmDeployAuto}}),
			ErrContains: []string{"at '/helm': properties 'environments' required, if 'deploy' exists"},
		},
		{
			Name:        "helm_publish_without_path_registry",
			Config:      merge(t, kickr.Kickr{GitHub: &kickr.GitHub{}, Helm: &kickr.Helm{Publish: kickr.HelmPublishAuto}}),
			ErrContains: []string{"at '/helm': properties 'path', 'registry' required, if 'publish' exists"},
		},
		{
			Name: "helm_publish_without_registry",
			Config: merge(t, kickr.Kickr{
				GitHub: &kickr.GitHub{},
				Helm:   &kickr.Helm{Publish: kickr.HelmPublishAuto, Path: "kickr-dev/kickr"},
			}),
			ErrContains: []string{"at '/helm': properties 'registry' required, if 'publish' exists"},
		},
		{
			Name:        "helm_path_without_registry",
			Config:      merge(t, kickr.Kickr{GitHub: &kickr.GitHub{}, Helm: &kickr.Helm{Path: "kickr-dev/kickr"}}),
			ErrContains: []string{"at '/helm': properties 'registry' required, if 'path' exists"},
		},
		{
			Name:        "helm_registry_without_path",
			Config:      merge(t, kickr.Kickr{GitHub: &kickr.GitHub{}, Helm: &kickr.Helm{Registry: "ghcr.io"}}),
			ErrContains: []string{"at '/helm': properties 'path' required, if 'registry' exists"},
		},
		{
			Name: "terraform_environments_without_apply",
			Config: merge(t, kickr.Kickr{
				GitHub: &kickr.GitHub{},
				Modules: []kickr.Module{
					{
						Path:      types.RootModule,
						Terraform: &kickr.Terraform{Engine: kickr.TerraformEngineTofu, Environments: []string{kickr.EnvironmentStaging}},
					},
				},
			}),
			ErrContains: []string{"at '/modules/0/terraform': properties 'apply' required, if 'environments' exists"},
		},
		{
			Name: "terraform_apply_without_environments",
			Config: merge(t, kickr.Kickr{
				GitHub: &kickr.GitHub{},
				Modules: []kickr.Module{
					{Path: types.RootModule, Terraform: &kickr.Terraform{Engine: kickr.TerraformEngineTofu, Apply: kickr.TerraformApplyAuto}},
				},
			}),
			ErrContains: []string{"at '/modules/0/terraform': properties 'environments' required, if 'apply' exists"},
		},

		// structural allOf / contains
		{
			Name: "modules_two_pages_targets",
			Config: merge(t, kickr.Kickr{
				GitHub: &kickr.GitHub{},
				Modules: []kickr.Module{
					{Path: types.RootModule, Deployment: &kickr.Deployment{Target: kickr.DeploymentTargetPages}},
					{Path: "docs", Deployment: &kickr.Deployment{Target: kickr.DeploymentTargetPages}},
				},
			}),
			ErrContains: []string{"at '/modules': max 1 items required to match contains schema"},
		},
		{
			Name: "modules_mixed_terraform_engine_no_ci",
			Config: merge(t, kickr.Kickr{
				Modules: []kickr.Module{
					{Path: types.RootModule, Terraform: &kickr.Terraform{Engine: kickr.TerraformEngineTofu}},
					{Path: "docs", Terraform: &kickr.Terraform{Engine: kickr.TerraformEngineTerraform}},
				},
			}),
			ErrContains: []string{"at '/modules': 'not' failed"},
		},
		{
			Name: "modules_mixed_terraform_engine_with_ci",
			Config: merge(t, kickr.Kickr{
				GitHub: &kickr.GitHub{},
				Modules: []kickr.Module{
					{Path: types.RootModule, Terraform: &kickr.Terraform{Engine: kickr.TerraformEngineTofu}},
					{Path: "docs", Terraform: &kickr.Terraform{Engine: kickr.TerraformEngineTerraform}},
				},
			}),
			ErrContains: []string{"at '/modules': 'not' failed"},
		},
		{
			Name:        "gitlab_release_auth_forbidden",
			Config:      merge(t, kickr.Kickr{GitLab: &kickr.GitLab{Release: &kickr.Release{Auth: kickr.ReleaseAuthGitHubToken}}}),
			ErrContains: []string{"at '/gitlab/release/auth': must not be provided"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			// Act
			err := validate(t, tc.Config)

			// Assert
			require.Error(t, err)
			for _, contains := range tc.ErrContains {
				assert.ErrorContains(t, err, contains)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	type testcase struct {
		Config kickr.Kickr
		Name   string
	}

	platforms := func(t testing.TB, name string, config kickr.Kickr) []testcase {
		t.Helper()
		github, gitlab := config, config
		github.GitHub = &kickr.GitHub{}
		gitlab.GitLab = &kickr.GitLab{}
		return []testcase{
			{Name: "github_" + name, Config: merge(t, github)},
			{Name: "gitlab_" + name, Config: merge(t, gitlab)},
		}
	}

	cases := slices.Concat(
		[]testcase{
			{Name: "minimal_no_ci", Config: merge(t, kickr.Kickr{})},
			{
				Name: "modules_same_terraform_engine_no_ci",
				Config: merge(t, kickr.Kickr{
					Modules: []kickr.Module{
						{Path: types.RootModule, Terraform: &kickr.Terraform{Engine: kickr.TerraformEngineTofu}},
						{Path: "docs", Terraform: &kickr.Terraform{Engine: kickr.TerraformEngineTofu}},
					},
				}),
			},
			{
				Name: "modules_mixed_terraform_and_plain_no_ci",
				Config: merge(t, kickr.Kickr{
					Modules: []kickr.Module{
						{Path: types.RootModule, Terraform: &kickr.Terraform{Engine: kickr.TerraformEngineTofu}},
						{Path: "docs"},
					},
				}),
			},
		},
		platforms(t, "docker_ci_fields", kickr.Kickr{
			Docker: &kickr.Docker{Auto: true, Path: "kickr-dev/kickr", Registry: "ghcr.io", Port: 3000},
		}),
		platforms(t, "helm_deploy_environments", kickr.Kickr{
			Helm: &kickr.Helm{Deploy: kickr.HelmDeployAuto, Environments: []string{kickr.EnvironmentStaging}},
		}),
		platforms(t, "helm_publish_path_registry", kickr.Kickr{
			Helm: &kickr.Helm{Publish: kickr.HelmPublishAuto, Path: "kickr-dev/kickr", Registry: "ghcr.io"},
		}),
		platforms(t, "helm_full", kickr.Kickr{
			Helm: &kickr.Helm{
				Deploy:       kickr.HelmDeployAuto,
				Environments: []string{kickr.EnvironmentStaging},
				Publish:      kickr.HelmPublishAuto,
				Path:         "kickr-dev/kickr",
				Registry:     "ghcr.io",
			},
		}),
		platforms(t, "module_deployment_pages", kickr.Kickr{
			Modules: []kickr.Module{{Path: types.RootModule, Deployment: &kickr.Deployment{Target: kickr.DeploymentTargetPages, Auto: true}}},
		}),
		platforms(t, "module_deployment_helm", kickr.Kickr{
			Modules: []kickr.Module{{Path: types.RootModule, Deployment: &kickr.Deployment{Target: kickr.DeploymentTargetHelm}}},
		}),
		platforms(t, "module_terraform_apply_environments", kickr.Kickr{
			Modules: []kickr.Module{
				{
					Path: types.RootModule,
					Terraform: &kickr.Terraform{
						Engine:       kickr.TerraformEngineTofu,
						Apply:        kickr.TerraformApplyAuto,
						Environments: []string{kickr.EnvironmentStaging},
					},
				},
			},
		}),
		platforms(t, "modules_one_pages_among_many", kickr.Kickr{
			Modules: []kickr.Module{
				{Path: types.RootModule, Deployment: &kickr.Deployment{Target: kickr.DeploymentTargetPages}},
				{Path: "docs", Deployment: &kickr.Deployment{Target: kickr.DeploymentTargetHelm}},
				{Path: "apps/api"},
			},
		}),
		[]testcase{
			{
				Name: "github_release_auth",
				Config: merge(t, kickr.Kickr{
					GitHub: &kickr.GitHub{
						Release: &kickr.Release{
							Auth:    kickr.ReleaseAuthGitHubToken,
							Auto:    true,
							Options: []string{kickr.ReleaseOptionsBackmerge},
						},
					},
				}),
			},
			{
				Name: "gitlab_release_no_auth",
				Config: merge(t, kickr.Kickr{
					GitLab: &kickr.GitLab{
						Release: &kickr.Release{Auto: true, Options: []string{kickr.ReleaseOptionsBackmerge}},
					},
				}),
			},
			{
				Name: "both_platforms_ci_fields",
				Config: merge(t, kickr.Kickr{
					GitHub: &kickr.GitHub{},
					GitLab: &kickr.GitLab{},
					Docker: &kickr.Docker{Auto: true},
					Helm:   &kickr.Helm{Deploy: kickr.HelmDeployAuto, Environments: []string{kickr.EnvironmentStaging}},
				}),
			},
			{
				Name: "kitchen_sink",
				Config: merge(t, kickr.Kickr{
					GitHub: &kickr.GitHub{
						Release: &kickr.Release{
							Auth:    kickr.ReleaseAuthGitHubToken,
							Auto:    true,
							Options: []string{kickr.ReleaseOptionsBackmerge},
						},
					},
					Docker: &kickr.Docker{Auto: true, Path: "kickr-dev/kickr", Registry: "ghcr.io", Port: 3000},
					Helm: &kickr.Helm{
						Deploy:       kickr.HelmDeployAuto,
						Environments: []string{kickr.EnvironmentStaging},
						Publish:      kickr.HelmPublishAuto,
						Path:         "kickr-dev/kickr",
						Registry:     "ghcr.io",
					},
					Modules: []kickr.Module{
						{
							Path:       types.RootModule,
							Deployment: &kickr.Deployment{Target: kickr.DeploymentTargetPages, Auto: true},
							Terraform: &kickr.Terraform{
								Engine:       kickr.TerraformEngineTofu,
								Apply:        kickr.TerraformApplyAuto,
								Environments: []string{kickr.EnvironmentStaging},
							},
						},
						{Path: "docs"},
					},
				}),
			},
		},
	)
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			// Act
			err := validate(t, tc.Config)

			// Assert
			assert.NoError(t, err)
		})
	}
}

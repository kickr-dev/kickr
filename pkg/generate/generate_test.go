package generate_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	engine "github.com/kickr-dev/engine/pkg"
	"github.com/kickr-dev/engine/pkg/files"
	"github.com/kickr-dev/engine/pkg/parser"
	compare "github.com/kilianpaquier/compare/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kickr-dev/kickr/pkg/generate"
	"github.com/kickr-dev/kickr/pkg/generate/templates"
	"github.com/kickr-dev/kickr/pkg/generate/types"
	kickr "github.com/kickr-dev/kickr/pkg/kickr/v1"
	"github.com/kickr-dev/kickr/testutils"
)

func TestGenerate_NoLang(t *testing.T) {
	ctx := t.Context()

	t.Run("success_chart", func(t *testing.T) {
		type testcase struct {
			Helm     *kickr.Helm
			Provider string
		}

		cases := []testcase{
			{Provider: parser.GitHub, Helm: &kickr.Helm{}},
			{Provider: parser.GitHub, Helm: &kickr.Helm{Publish: kickr.HelmPublishAuto}},
			{Provider: parser.GitHub, Helm: &kickr.Helm{Path: "path/to/kickr", Publish: kickr.HelmPublishManual, Registry: "chartmuseum.example.com"}},

			{Provider: parser.GitLab, Helm: &kickr.Helm{}},
			{Provider: parser.GitLab, Helm: &kickr.Helm{Publish: kickr.HelmPublishAuto}},
			{Provider: parser.GitLab, Helm: &kickr.Helm{Path: "path/to/kickr", Publish: kickr.HelmPublishManual, Registry: "chartmuseum.example.com"}},
		}
		for _, tc := range cases {
			publish := "nil"
			if tc.Helm.Publish != "" {
				publish = tc.Helm.Publish
			}
			name := fmt.Sprint(tc.Provider, "_publish_", publish)
			t.Run(name, func(t *testing.T) {
				// Arrange
				config := types.Repository{
					Kickr: kickr.Kickr{
						CI:       &kickr.CI{Provider: tc.Provider},
						Exclude:  []string{kickr.ExcludeMakefile, kickr.ExcludeRenovate},
						Helm:     tc.Helm,
						Platform: tc.Provider,
					},
				}

				// Act & Assert
				test(ctx, t, config)
			})
		}
	})

	t.Run("success_kickr", func(t *testing.T) {
		type testcase struct {
			Option   string
			Provider string
		}

		cases := []testcase{
			{Option: kickr.OptionsKickrGitHubApp, Provider: parser.GitHub},
			{Option: kickr.OptionsKickrPersonalToken, Provider: parser.GitHub},
			{Option: kickr.OptionsKickr, Provider: parser.GitLab},
		}
		for _, tc := range cases {
			name := tc.Provider
			if option := strings.Split(tc.Option, ":"); len(option) > 1 {
				name += "_" + option[1]
			}
			t.Run(name, func(t *testing.T) {
				// Arrange
				config := types.Repository{
					Kickr: kickr.Kickr{
						CI:       &kickr.CI{Provider: tc.Provider, Options: []string{tc.Option}},
						Exclude:  []string{kickr.ExcludeMakefile, kickr.ExcludeShell},
						Platform: tc.Provider,
					},
				}

				// Act & Assert
				test(ctx, t, config)
			})
		}
	})

	t.Run("success_renovate", func(t *testing.T) {
		t.Run("github", func(t *testing.T) {
			for _, auth := range []string{kickr.OptionsRenovateGitHubApp, kickr.OptionsRenovatePersonalToken} {
				t.Run(auth, func(t *testing.T) {
					// Arrange
					config := types.Repository{
						Kickr: kickr.Kickr{
							CI:       &kickr.CI{Provider: parser.GitHub, Options: []string{auth}},
							Exclude:  []string{kickr.ExcludeMakefile, kickr.ExcludeShell},
							Platform: parser.GitHub,
						},
					}

					// Act & Assert
					test(ctx, t, config)
				})
			}
		})

		t.Run("gitlab", func(t *testing.T) {
			// Arrange
			config := types.Repository{
				Kickr: kickr.Kickr{
					CI:       &kickr.CI{Provider: parser.GitLab, Options: []string{kickr.OptionsRenovate}},
					Exclude:  []string{kickr.ExcludeMakefile, kickr.ExcludeShell},
					Platform: parser.GitLab,
				},
			}

			// Act & Assert
			test(ctx, t, config)
		})

		t.Run("templates", func(t *testing.T) {
			// Arrange
			tmpl := func(_ context.Context, destdir string, _ *types.Repository) error {
				file, err := os.Create(filepath.Join(destdir, "template.tmpl"))
				if err != nil {
					return fmt.Errorf("create: %w", err)
				}
				return file.Close()
			}

			config := types.Repository{
				Kickr: kickr.Kickr{Exclude: []string{kickr.ExcludeMakefile}},
			}

			// Act & Assert
			test(ctx, t, config, tmpl)
		})
	})

	t.Run("success_precommit", func(t *testing.T) {
		t.Run("enabled", func(t *testing.T) {
			for _, provider := range []string{parser.GitHub, parser.GitLab} {
				t.Run(provider, func(t *testing.T) {
					// Arrange
					config := types.Repository{
						Kickr: kickr.Kickr{
							CI:        &kickr.CI{Provider: provider},
							Exclude:   []string{kickr.ExcludeMakefile, kickr.ExcludeRenovate},
							PreCommit: []string{kickr.PreCommitAutoCommit, kickr.PreCommitGitflowBranches, kickr.PreCommitConventionalCommits},
						},
					}

					// Act & Assert
					test(ctx, t, config)
				})
			}
		})

		t.Run("disabled", func(t *testing.T) {
			for _, provider := range []string{parser.GitHub, parser.GitLab} {
				t.Run(provider, func(t *testing.T) {
					// Arrange
					config := types.Repository{
						Kickr: kickr.Kickr{
							CI:        &kickr.CI{Provider: provider},
							Exclude:   []string{kickr.ExcludeMakefile, kickr.ExcludePreCommit, kickr.ExcludeRenovate},
							PreCommit: []string{kickr.PreCommitAutoCommit, kickr.PreCommitGitflowBranches, kickr.PreCommitConventionalCommits},
						},
					}

					// Act & Assert
					test(ctx, t, config)
				})
			}
		})
	})

	t.Run("success_release", func(t *testing.T) {
		type testcase struct {
			Auth     string
			Auto     bool
			Provider string
		}

		cases := []testcase{
			{Provider: parser.GitHub},
			{Provider: parser.GitHub, Auto: true},

			{Provider: parser.GitHub, Auth: kickr.ReleaseAuthGitHubApp},
			{Provider: parser.GitHub, Auth: kickr.ReleaseAuthGitHubToken},
			{Provider: parser.GitHub, Auth: kickr.ReleaseAuthPersonalToken},

			{Provider: parser.GitLab},
			{Provider: parser.GitLab, Auto: true},
		}
		for _, tc := range cases {
			name := tc.Provider
			if tc.Auto {
				name += "_auto"
			}
			if tc.Auth != "" {
				name += "_auth_" + tc.Auth
			}

			t.Run(name, func(t *testing.T) {
				// Arrange
				config := types.Repository{
					Kickr: kickr.Kickr{
						CI: &kickr.CI{
							Provider: tc.Provider,
							Release:  &kickr.Release{Auto: tc.Auto, Auth: tc.Auth},
						},
						Exclude:  []string{kickr.ExcludeMakefile, kickr.ExcludeRenovate},
						Platform: tc.Provider,
					},
				}

				// Act & Assert
				test(ctx, t, config)
			})
		}
	})
}

func TestGenerate_Shell(t *testing.T) {
	ctx := t.Context()

	shell := func(_ context.Context, destdir string, _ *types.Repository) error {
		return os.WriteFile(filepath.Join(destdir, "script.sh"), []byte("#!/bin/sh\n"), files.RwxRxRxRx)
	}

	t.Run("success_ci", func(t *testing.T) {
		for _, provider := range []string{parser.GitLab, parser.GitHub} {
			t.Run(provider, func(t *testing.T) {
				// Arrange
				config := types.Repository{
					Kickr: kickr.Kickr{
						CI:      &kickr.CI{Provider: provider},
						Exclude: []string{kickr.ExcludeMakefile, kickr.ExcludeRenovate},
					},
				}

				// Act & Assert
				test(ctx, t, config, shell)
			})
		}
	})

	t.Run("success_precommit", func(t *testing.T) {
		for _, precommit := range []bool{true, false} {
			t.Run(strconv.FormatBool(precommit), func(t *testing.T) {
				// Arrange
				config := types.Repository{Kickr: kickr.Kickr{Exclude: []string{kickr.ExcludeMakefile, kickr.ExcludeRenovate}}}
				if !precommit {
					config.Exclude = append(config.Exclude, kickr.ExcludePreCommit)
				}

				// Act & Assert
				test(ctx, t, config, shell)
			})
		}
	})
}

func TestGenerate_Golang(t *testing.T) {
	ctx := t.Context()

	t.Run("success_cli", func(t *testing.T) {
		// Arrange
		golang := func(provider string) func(ctx context.Context, destdir string, config *types.Repository) error {
			return func(_ context.Context, destdir string, _ *types.Repository) error {
				gomod := fmt.Appendf(nil, "module %s.com/kickr-dev/kickr\n\ngo 1.23\n", provider)
				if err := os.WriteFile(filepath.Join(destdir, parser.FileGomod), gomod, files.RwRR); err != nil {
					return fmt.Errorf("write file: %w", err)
				}

				cmd := filepath.Join(destdir, parser.FolderCMD)
				if err := os.MkdirAll(cmd, files.RwxRxRxRx); err != nil {
					return fmt.Errorf("mkdir all: %w", err)
				}
				for _, bin := range []string{"name"} {
					if err := os.MkdirAll(filepath.Join(cmd, bin), files.RwxRxRxRx); err != nil {
						return fmt.Errorf("mkdir all: %w", err)
					}
					file, err := os.Create(filepath.Join(cmd, bin, parser.FileMain))
					if err != nil {
						return fmt.Errorf("create: %w", err)
					}
					if err := file.Close(); err != nil {
						return fmt.Errorf("close: %w", err)
					}
				}

				return nil
			}
		}

		for _, provider := range []string{parser.GitLab, parser.GitHub} {
			t.Run(provider, func(t *testing.T) {
				// Arrange
				config := types.Repository{
					Kickr: kickr.Kickr{
						CI:       &kickr.CI{Provider: provider, Release: &kickr.Release{}},
						Platform: provider,
					},
				}

				// Act & Assert
				test(ctx, t, config, golang(provider))
			})
		}
	})

	t.Run("success_library", func(t *testing.T) {
		// Arrange
		golang := func(platform string) func(ctx context.Context, destdir string, config *types.Repository) error {
			return func(_ context.Context, destdir string, _ *types.Repository) error {
				gomod := fmt.Appendf(nil, "module %s.com/kickr-dev/kickr\n\ngo 1.23\n", platform)
				if err := os.WriteFile(filepath.Join(destdir, parser.FileGomod), gomod, files.RwRR); err != nil {
					return fmt.Errorf("write file: %w", err)
				}
				return nil
			}
		}

		for _, platform := range []string{parser.GitLab, parser.GitHub} {
			t.Run(platform, func(t *testing.T) {
				// Arrange
				config := types.Repository{
					Kickr: kickr.Kickr{
						Exclude:   []string{kickr.ExcludeMakefile, kickr.ExcludeRenovate},
						PreCommit: []string{kickr.PreCommitGomodTidy},
						Platform:  platform,
					},
				}

				// Act & Assert
				test(ctx, t, config, golang(platform))
			})
		}
	})

	t.Run("success_multiple_libraries", func(t *testing.T) {
		// Arrange
		golang := func(platform string) func(ctx context.Context, destdir string, config *types.Repository) error {
			return func(_ context.Context, destdir string, _ *types.Repository) error {
				if err := os.WriteFile(filepath.Join(destdir, parser.FileGowork), []byte("go 1.23\n\nuse (\n\t./kickr\n\t./engine\n)\n"), files.RwRR); err != nil {
					return fmt.Errorf("write file: %w", err)
				}

				for _, dir := range []string{"kickr", "engine"} {
					if err := os.MkdirAll(filepath.Join(destdir, dir), files.RwxRxRxRx); err != nil {
						return fmt.Errorf("mkdir all: %w", err)
					}
					gomod := fmt.Appendf(nil, "module %s.com/kickr-dev/%s\n\ngo 1.23\n", platform, dir)
					if err := os.WriteFile(filepath.Join(destdir, dir, parser.FileGomod), gomod, files.RwRR); err != nil {
						return fmt.Errorf("write file: %w", err)
					}
				}
				return nil
			}
		}

		for _, platform := range []string{parser.GitLab, parser.GitHub} {
			t.Run(platform, func(t *testing.T) {
				// Arrange
				config := types.Repository{
					Kickr: kickr.Kickr{
						Exclude:   []string{kickr.ExcludeMakefile, kickr.ExcludeRenovate},
						PreCommit: []string{kickr.PreCommitGomodTidy},
						Platform:  platform,
					},
				}

				// Act & Assert
				test(ctx, t, config, golang(platform))
			})
		}
	})

	t.Run("success_multiple_bin_helm", func(t *testing.T) {
		// Arrange
		golang := func(provider string) func(ctx context.Context, destdir string, config *types.Repository) error {
			return func(_ context.Context, destdir string, _ *types.Repository) error {
				gomod := fmt.Appendf(nil, "module %s.com/kickr-dev/kickr\n\ngo 1.23\n", provider)
				if err := os.WriteFile(filepath.Join(destdir, parser.FileGomod), gomod, files.RwRR); err != nil {
					return fmt.Errorf("write file: %w", err)
				}

				cmd := filepath.Join(destdir, parser.FolderCMD)
				if err := os.MkdirAll(cmd, files.RwxRxRxRx); err != nil {
					return fmt.Errorf("mkdir all: %w", err)
				}
				for _, bin := range []string{"cron-name", "job-name", "worker-name"} {
					if err := os.MkdirAll(filepath.Join(cmd, bin), files.RwxRxRxRx); err != nil {
						return fmt.Errorf("mkdir all: %w", err)
					}
					file, err := os.Create(filepath.Join(cmd, bin, parser.FileMain))
					if err != nil {
						return fmt.Errorf("create: %w", err)
					}
					if err := file.Close(); err != nil {
						return fmt.Errorf("close: %w", err)
					}
				}

				return nil
			}
		}

		for _, provider := range []string{parser.GitLab, parser.GitHub} {
			t.Run(provider, func(t *testing.T) {
				// Arrange
				config := types.Repository{
					Kickr: kickr.Kickr{
						CI: &kickr.CI{
							Provider: provider,
							Options: []string{
								kickr.OptionsCodecov,
								kickr.OptionsCodeQL,
								kickr.OptionsHardenRunner,
								kickr.OptionsLabeler,
								kickr.OptionsOSSFScorecard,
								kickr.OptionsSonarQube,
								kickr.OptionsStepSecurityActions,
							},
							Release: &kickr.Release{},
						},
						Docker: &kickr.Docker{Path: "path/to/registry", Registry: "registry.example.com"},
						Helm: &kickr.Helm{
							Deploy:       kickr.HelmDeployManual,
							Environments: []string{kickr.EnvironmentStaging, kickr.EnvironmentProduction},
							Path:         "path/to/repository",
							Publish:      kickr.HelmPublishManual,
							Registry:     "chartmuseum.example.com",
						},
						Description: "A useful project description",
						Exclude:     []string{kickr.ExcludeRenovate, kickr.ExcludeShell},
						Platform:    provider,
						PreCommit:   []string{kickr.PreCommitGolangciLint},
					},
				}

				// Act & Assert
				test(ctx, t, config, golang(provider))
			})
		}
	})
}

func TestGenerate_Hugo(t *testing.T) {
	ctx := t.Context()

	hugo := func(_ context.Context, destdir string, _ *types.Repository) error {
		file, err := os.Create(filepath.Join(destdir, "hugo.toml"))
		if err != nil {
			return fmt.Errorf("create: %w", err)
		}
		return file.Close()
	}

	t.Run("success_no_website", func(t *testing.T) {
		for _, provider := range []string{parser.GitHub, parser.GitLab} {
			t.Run(provider, func(t *testing.T) {
				// Arrange
				config := types.Repository{
					Kickr: kickr.Kickr{
						CI:       &kickr.CI{Provider: provider},
						Platform: provider,
					},
				}

				// Act & Assert
				test(ctx, t, config, hugo)
			})
		}
	})

	t.Run("success_hosting", func(t *testing.T) {
		type testcase struct {
			Provider string
			Website  *kickr.Website
		}

		cases := []testcase{
			{Provider: parser.GitHub, Website: &kickr.Website{Hosting: kickr.WebsiteHostingNetlify, Auto: true}},
			{Provider: parser.GitHub, Website: &kickr.Website{Hosting: kickr.WebsiteHostingNetlify}},

			{Provider: parser.GitLab, Website: &kickr.Website{Hosting: kickr.WebsiteHostingNetlify, Auto: true}},
			{Provider: parser.GitLab, Website: &kickr.Website{Hosting: kickr.WebsiteHostingNetlify}},

			{Provider: parser.GitHub, Website: &kickr.Website{Hosting: kickr.WebsiteHostingPages, Auto: true}},
			{Provider: parser.GitHub, Website: &kickr.Website{Hosting: kickr.WebsiteHostingPages}},

			{Provider: parser.GitLab, Website: &kickr.Website{Hosting: kickr.WebsiteHostingPages, Auto: true}},
			{Provider: parser.GitLab, Website: &kickr.Website{Hosting: kickr.WebsiteHostingPages}},
		}
		for _, tc := range cases {
			name := fmt.Sprint(tc.Provider, "_", tc.Website.Hosting, "_auto_", tc.Website.Auto)
			t.Run(name, func(t *testing.T) {
				// Arrange
				config := types.Repository{
					Kickr: kickr.Kickr{
						CI:       &kickr.CI{Provider: tc.Provider},
						Platform: tc.Provider,
						Website:  tc.Website,
					},
				}

				// Act & Assert
				test(ctx, t, config, hugo)
			})
		}
	})
}

func TestGenerate_Node(t *testing.T) {
	ctx := t.Context()

	t.Run("success_package_managers", func(t *testing.T) {
		// Arrange
		node := func(tc string) func(ctx context.Context, destdir string, config *types.Repository) error {
			return func(_ context.Context, destdir string, _ *types.Repository) error {
				content := fmt.Appendf(nil, `{ "name": "kickr", "packageManager": "%s" }`+"\n", tc)
				return os.WriteFile(filepath.Join(destdir, parser.FilePackageJSON), content, files.RwRR)
			}
		}

		for _, tc := range []string{"bun@1.1.6", "npm@7.0.0", "pnpm@9.0.0", "yarn@1.22.10"} {
			t.Run(tc, func(t *testing.T) {
				// Arrange
				config := types.Repository{
					Kickr: kickr.Kickr{
						CI:       &kickr.CI{Provider: parser.GitHub},
						Platform: parser.GitHub,
					},
				}

				// Act & Assert
				test(ctx, t, config, node(tc))
			})
		}
	})

	t.Run("success_library", func(t *testing.T) {
		type testcase struct {
			Options        []string
			PackageManager string
			Provider       string
		}

		// Arrange
		node := func(tc testcase) func(ctx context.Context, destdir string, config *types.Repository) error {
			return func(_ context.Context, destdir string, _ *types.Repository) error {
				content := fmt.Appendf(nil, `{ "name": "kickr", "packageManager": "%s" }`+"\n", tc.PackageManager)
				return os.WriteFile(filepath.Join(destdir, parser.FilePackageJSON), content, files.RwRR)
			}
		}

		cases := []testcase{
			{Provider: parser.GitHub, Options: []string{kickr.OptionsRenovatePersonalToken}, PackageManager: "bun@1.1.6"},
			{Provider: parser.GitLab, Options: []string{kickr.OptionsRenovate}, PackageManager: "bun@1.1.6"},
		}
		for _, tc := range cases {
			name := fmt.Sprint(tc.Provider, "_", tc.PackageManager)
			t.Run(name, func(t *testing.T) {
				// Arrange
				config := types.Repository{
					Kickr: kickr.Kickr{
						CI: &kickr.CI{
							Options:  tc.Options,
							Provider: tc.Provider,
							Release:  &kickr.Release{Options: []string{kickr.ReleaseOptionsBackmerge}},
						},
						Platform: tc.Provider,
					},
				}

				// Act & Assert
				test(ctx, t, config, node(tc))
			})
		}
	})

	t.Run("success_hosting", func(t *testing.T) {
		type testcase struct {
			Provider string
			Website  *kickr.Website
		}

		// Arrange
		node := func(_ context.Context, destdir string, _ *types.Repository) error {
			return os.WriteFile(filepath.Join(destdir, parser.FilePackageJSON),
				[]byte(`{ "name": "kickr", "packageManager": "bun@1.1.6", "main": "index.js" }`+"\n"), files.RwRR)
		}

		cases := []testcase{
			{Provider: parser.GitHub, Website: &kickr.Website{Hosting: kickr.WebsiteHostingNetlify, Auto: true}},
			{Provider: parser.GitHub, Website: &kickr.Website{Hosting: kickr.WebsiteHostingNetlify}},

			{Provider: parser.GitLab, Website: &kickr.Website{Hosting: kickr.WebsiteHostingNetlify, Auto: true}},
			{Provider: parser.GitLab, Website: &kickr.Website{Hosting: kickr.WebsiteHostingNetlify}},

			{Provider: parser.GitHub, Website: &kickr.Website{Hosting: kickr.WebsiteHostingPages, Auto: true}},
			{Provider: parser.GitHub, Website: &kickr.Website{Hosting: kickr.WebsiteHostingPages}},

			{Provider: parser.GitLab, Website: &kickr.Website{Hosting: kickr.WebsiteHostingPages, Auto: true}},
			{Provider: parser.GitLab, Website: &kickr.Website{Hosting: kickr.WebsiteHostingPages}},
		}
		for _, tc := range cases {
			name := fmt.Sprint(tc.Provider, "_", tc.Website.Hosting, "_auto_", tc.Website.Auto)
			t.Run(name, func(t *testing.T) {
				// Arrange
				config := types.Repository{
					Kickr: kickr.Kickr{
						CI:       &kickr.CI{Provider: tc.Provider},
						Exclude:  []string{kickr.ExcludeMakefile, kickr.ExcludeRenovate},
						Platform: tc.Provider,
						Website:  tc.Website,
					},
				}

				// Act & Assert
				test(ctx, t, config, node)
			})
		}
	})

	t.Run("success_helm", func(t *testing.T) {
		type testcase struct {
			Helm     *kickr.Helm
			Provider string
		}

		// Arrange
		node := func(_ context.Context, destdir string, _ *types.Repository) error {
			return os.WriteFile(filepath.Join(destdir, parser.FilePackageJSON),
				[]byte(`{ "name": "kickr", "packageManager": "bun@1.1.6", "main": "index.js" }`+"\n"), files.RwRR)
		}

		cases := []testcase{
			{Provider: parser.GitHub, Helm: &kickr.Helm{}},
			{Provider: parser.GitHub, Helm: &kickr.Helm{Deploy: kickr.HelmDeployAuto, Environments: []string{kickr.EnvironmentReview}}},
			{Provider: parser.GitHub, Helm: &kickr.Helm{Deploy: kickr.HelmDeployManual, Environments: []string{kickr.EnvironmentIntegration}}},
			{Provider: parser.GitHub, Helm: &kickr.Helm{Publish: kickr.HelmPublishAuto}},
			{Provider: parser.GitHub, Helm: &kickr.Helm{Publish: kickr.HelmPublishManual}},

			{Provider: parser.GitLab, Helm: &kickr.Helm{}},
			{Provider: parser.GitLab, Helm: &kickr.Helm{Deploy: kickr.HelmDeployAuto, Environments: []string{kickr.EnvironmentReview}}},
			{Provider: parser.GitLab, Helm: &kickr.Helm{Deploy: kickr.HelmDeployManual, Environments: []string{kickr.EnvironmentIntegration}}},
			{Provider: parser.GitLab, Helm: &kickr.Helm{Publish: kickr.HelmPublishAuto}},
			{Provider: parser.GitLab, Helm: &kickr.Helm{Publish: kickr.HelmPublishManual}},
		}
		for _, tc := range cases {
			publish := "none"
			if tc.Helm.Publish != "" {
				publish = tc.Helm.Publish
			}
			deploy := "none"
			if tc.Helm.Deploy != "" {
				deploy = tc.Helm.Deploy
			}

			name := fmt.Sprint(tc.Provider, "_deploy_", deploy, "_publish_", publish)
			t.Run(name, func(t *testing.T) {
				// Arrange
				config := types.Repository{
					Kickr: kickr.Kickr{
						CI:       &kickr.CI{Provider: tc.Provider},
						Exclude:  []string{kickr.ExcludeMakefile, kickr.ExcludeRenovate},
						Helm:     tc.Helm,
						Platform: tc.Provider,
					},
				}

				// Act & Assert
				test(ctx, t, config, node)
			})
		}
	})
}

func TestGenerate_Terraform(t *testing.T) {
	ctx := t.Context()

	t.Run("success_multiple_modules", func(t *testing.T) {
		type testcase struct {
			Apply    string
			Engine   string
			Provider string
		}

		// Arrange
		terraform := func(subdir string) func(ctx context.Context, destdir string, config *types.Repository) error {
			return func(_ context.Context, destdir string, _ *types.Repository) error {
				if err := os.MkdirAll(filepath.Join(destdir, subdir), files.RwxRxRxRx); err != nil {
					return fmt.Errorf("mkdir all: %w", err)
				}
				return os.WriteFile(filepath.Join(destdir, subdir, "main.tf"), []byte(
					`terraform { backend "http" {} }`+"\n"+
						`variable "my_secret" { sensitive = true }`+"\n"+
						`variable "github_var" {}`+"\n"+
						`variable "my_var" {}`+"\n"), files.RwRR)
			}
		}

		cases := []testcase{
			{Apply: kickr.TerraformApplyManual, Engine: kickr.TerraformEngineOpenTofu, Provider: parser.GitHub},
			{Apply: kickr.TerraformApplyAuto, Engine: kickr.TerraformEngineTerraform, Provider: parser.GitHub},

			{Apply: kickr.TerraformApplyManual, Engine: kickr.TerraformEngineOpenTofu, Provider: parser.GitLab},
			{Apply: kickr.TerraformApplyAuto, Engine: kickr.TerraformEngineTerraform, Provider: parser.GitLab},
		}
		for _, tc := range cases {
			t.Run(tc.Provider+"_"+tc.Engine, func(t *testing.T) {
				// Arrange
				config := types.Repository{
					Kickr: kickr.Kickr{
						CI: &kickr.CI{
							Provider: tc.Provider,
							Release:  &kickr.Release{},
						},
						Platform: tc.Provider,
						Terraform: &kickr.Terraform{
							Apply:        tc.Apply,
							Engine:       tc.Engine,
							Environments: []string{kickr.EnvironmentProduction},
							Modules:      []string{"modules/one", "modules/two"},
						},
					},
				}

				// Act & Assert
				test(ctx, t, config, terraform(filepath.Join("modules", "one")), terraform(filepath.Join("modules", "two")))
			})
		}
	})

	t.Run("root_module", func(t *testing.T) {
		// Arrange
		terraform := func(_ context.Context, destdir string, _ *types.Repository) error {
			return os.WriteFile(filepath.Join(destdir, "main.tf"), []byte(`terraform { backend "s3" {} }`+"\n"), files.RwRR)
		}

		for _, provider := range []string{parser.GitHub, parser.GitLab} {
			t.Run(provider, func(t *testing.T) {
				// Arrange
				config := types.Repository{
					Kickr: kickr.Kickr{
						CI: &kickr.CI{
							Provider: provider,
						},
						PreCommit: []string{kickr.PreCommitTerraform},
						Platform:  provider,
						Terraform: &kickr.Terraform{
							Apply:        kickr.TerraformApplyAuto,
							Environments: []string{kickr.EnvironmentProduction},
						},
					},
				}

				// Act & Assert
				test(ctx, t, config, terraform)
			})
		}
	})
}

func TestGenerate_MonoRepo(t *testing.T) {
	ctx := t.Context()

	type testcase struct {
		Provider string
		Hosting  string
	}

	golang := func(tc testcase) func(ctx context.Context, destdir string, config *types.Repository) error {
		return func(_ context.Context, destdir string, _ *types.Repository) error {
			gomod := fmt.Appendf(nil, "module %s.com/kickr-dev/kickr\n\ngo 1.23\n", tc.Provider)
			if err := os.WriteFile(filepath.Join(destdir, parser.FileGomod), gomod, files.RwRR); err != nil {
				return fmt.Errorf("write file: %w", err)
			}
			return nil
		}
	}

	hugo := func(_ context.Context, destdir string, _ *types.Repository) error {
		if err := os.MkdirAll(filepath.Join(destdir, "docs"), files.RwxRxRxRx); err != nil {
			return fmt.Errorf("mkdir all: %w", err)
		}
		file, err := os.Create(filepath.Join(destdir, "docs", "hugo.toml"))
		if err != nil {
			return fmt.Errorf("create: %w", err)
		}
		return file.Close()
	}

	node := func(subdir string) engine.Parser[types.Repository] {
		return func(_ context.Context, destdir string, _ *types.Repository) error {
			if err := os.MkdirAll(filepath.Join(destdir, subdir), files.RwxRxRxRx); err != nil {
				return fmt.Errorf("mkdir all: %w", err)
			}
			return os.WriteFile(filepath.Join(destdir, subdir, parser.FilePackageJSON),
				[]byte(`{ "name": "kickr", "packageManager": "bun@1.1.6", "main": "index.js", "private": true }`+"\n"), files.RwRR)
		}
	}

	terraform := func(subdir string) func(ctx context.Context, destdir string, config *types.Repository) error {
		return func(_ context.Context, destdir string, _ *types.Repository) error {
			if err := os.MkdirAll(filepath.Join(destdir, subdir), files.RwxRxRxRx); err != nil {
				return fmt.Errorf("mkdir all: %w", err)
			}
			return os.WriteFile(filepath.Join(destdir, subdir, "main.tf"), []byte(`variable "my_var" {}`+"\n"), files.RwRR)
		}
	}

	t.Run("success_node_hugo_doc", func(t *testing.T) {
		cases := []testcase{
			{Provider: parser.GitLab, Hosting: kickr.WebsiteHostingPages},
			{Provider: parser.GitLab, Hosting: kickr.WebsiteHostingNetlify},
			{Provider: parser.GitHub, Hosting: kickr.WebsiteHostingPages},
			{Provider: parser.GitHub, Hosting: kickr.WebsiteHostingNetlify},
		}
		for _, tc := range cases {
			name := fmt.Sprint(tc.Provider, "_", tc.Hosting)
			t.Run(name, func(t *testing.T) {
				// Arrange
				config := types.Repository{
					Kickr: kickr.Kickr{
						CI: &kickr.CI{
							Provider: tc.Provider,
						},
						Exclude:  []string{kickr.ExcludePreCommit, kickr.ExcludeRenovate},
						Platform: tc.Provider,
						Website:  &kickr.Website{Hosting: tc.Hosting, Directory: "docs"},
					},
				}

				// Act & Assert
				test(ctx, t, config, node(""), hugo)
			})
		}
	})

	t.Run("success_go_hugo_doc", func(t *testing.T) {
		cases := []testcase{
			{Provider: parser.GitLab, Hosting: kickr.WebsiteHostingPages},
			{Provider: parser.GitLab, Hosting: kickr.WebsiteHostingNetlify},
			{Provider: parser.GitHub, Hosting: kickr.WebsiteHostingPages},
			{Provider: parser.GitHub, Hosting: kickr.WebsiteHostingNetlify},
		}
		for _, tc := range cases {
			name := fmt.Sprint(tc.Provider, "_", tc.Hosting)
			t.Run(name, func(t *testing.T) {
				// Arrange
				config := types.Repository{
					Kickr: kickr.Kickr{
						CI: &kickr.CI{
							Provider: tc.Provider,
						},
						Exclude:  []string{kickr.ExcludePreCommit, kickr.ExcludeRenovate},
						Platform: tc.Provider,
						Website:  &kickr.Website{Hosting: tc.Hosting, Directory: "docs"},
					},
				}

				// Act & Assert
				test(ctx, t, config, golang(tc), hugo)
			})
		}
	})

	t.Run("success_go_node_doc", func(t *testing.T) {
		cases := []testcase{
			{Provider: parser.GitLab, Hosting: kickr.WebsiteHostingPages},
			{Provider: parser.GitLab, Hosting: kickr.WebsiteHostingNetlify},
			{Provider: parser.GitHub, Hosting: kickr.WebsiteHostingPages},
			{Provider: parser.GitHub, Hosting: kickr.WebsiteHostingNetlify},
		}
		for _, tc := range cases {
			name := fmt.Sprint(tc.Provider, "_", tc.Hosting)
			t.Run(name, func(t *testing.T) {
				// Arrange
				config := types.Repository{
					Kickr: kickr.Kickr{
						CI: &kickr.CI{
							Provider: tc.Provider,
						},
						Exclude:  []string{kickr.ExcludeMakefile, kickr.ExcludePreCommit, kickr.ExcludeRenovate},
						Platform: tc.Provider,
						Website:  &kickr.Website{Hosting: tc.Hosting, Directory: "docs"},
					},
				}

				// Act & Assert
				test(ctx, t, config, golang(tc), node("docs"))
			})
		}
	})

	t.Run("success_go_self_terraform", func(t *testing.T) {
		cases := []testcase{
			{Provider: parser.GitLab},
			{Provider: parser.GitHub},
		}
		for _, tc := range cases {
			t.Run(tc.Provider, func(t *testing.T) {
				// Arrange
				config := types.Repository{
					Kickr: kickr.Kickr{
						CI:        &kickr.CI{Provider: tc.Provider},
						Exclude:   []string{kickr.ExcludeMakefile, kickr.ExcludePreCommit},
						Platform:  tc.Provider,
						Terraform: &kickr.Terraform{Engine: kickr.TerraformEngineOpenTofu, Modules: []string{".terraform"}},
					},
				}

				// Act & Assert
				test(ctx, t, config, golang(tc), terraform(".terraform"))
			})
		}
	})
}

func ParserInfo(_ context.Context, _ string, config *types.Repository) error {
	config.VCS = parser.VCS{
		Platform:    config.Platform,
		ProjectHost: config.Platform + ".com",
		ProjectName: "kickr",
		ProjectPath: "kickr-dev/kickr",
	}
	return nil
}

// test verifies every generation with provided config, parser and t.Name folder expected results.
func test(ctx context.Context, t *testing.T, config types.Repository, parsers ...engine.Parser[types.Repository]) {
	t.Helper()

	// Arrange
	config.Maintainers = append(config.Maintainers, &kickr.Maintainer{Name: "kilianpaquier"})
	assertdir := filepath.Join(testutils.Testdata(t), t.Name())
	require.NoError(t, os.MkdirAll(assertdir, files.RwxRxRxRx))

	destdir := t.TempDir()
	if ok, _ := strconv.ParseBool(os.Getenv("TESTDATA")); ok {
		destdir = assertdir
	}

	// Act
	_, err := engine.Generate(ctx, destdir, config,
		slices.Concat(parsers, []engine.Parser[types.Repository]{
			// must be kept first since it parses Git informations (useful for next parsers)
			// generate.ParserGit,
			ParserInfo,

			generate.ParserGlob,
			generate.ParserGolang,
			generate.ParserNode,
			generate.ParserTerraform,

			// must be kept last since it marshals config and merges it with chart overrides
			generate.ParserHelm,
		}),
		[]engine.Generator[types.Repository]{
			engine.GeneratorTemplates(templates.FS(), slices.Concat(templates.CodeCov(), templates.Sonar())),                              // coverage
			engine.GeneratorTemplates(templates.FS(), slices.Concat(templates.GitHub(), templates.GitLab(), templates.SemanticRelease())), // ci
			engine.GeneratorTemplates(templates.FS(), templates.Chart()),                                                                  // chart
			engine.GeneratorTemplates(templates.FS(), templates.Docker()),                                                                 // docker
			engine.GeneratorTemplates(templates.FS(), templates.Golang()),                                                                 // golang
			engine.GeneratorTemplates(templates.FS(), templates.Makefile()),                                                               // makefile
			engine.GeneratorTemplates(templates.FS(), templates.Misc()),                                                                   // misc
			engine.GeneratorTemplates(templates.FS(), templates.Renovate()),                                                               // renovate
			engine.GeneratorTemplates(templates.FS(), templates.Terraform()),                                                              // terraform
		})

	// Assert
	require.NoError(t, err)
	assert.NoError(t, compare.Dirs(assertdir, destdir))
}

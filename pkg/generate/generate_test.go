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
		// Arrange
		cases := []kickr.CI{
			{Provider: parser.GitHub, Helm: &kickr.Helm{}},
			{Provider: parser.GitHub, Helm: &kickr.Helm{Publish: kickr.HelmAuto}},
			{Provider: parser.GitHub, Helm: &kickr.Helm{Path: "path/to/kickr", Publish: kickr.HelmManual, Registry: "chartmuseum.example.com"}},

			{Provider: parser.GitLab, Helm: &kickr.Helm{}},
			{Provider: parser.GitLab, Helm: &kickr.Helm{Publish: kickr.HelmAuto}},
			{Provider: parser.GitLab, Helm: &kickr.Helm{Path: "path/to/kickr", Publish: kickr.HelmManual, Registry: "chartmuseum.example.com"}},
		}
		for _, ci := range cases {
			publish := "nil"
			if ci.Helm != nil && ci.Helm.Publish != "" {
				publish = ci.Helm.Publish
			}
			name := fmt.Sprint(ci.Provider, "_publish_", publish)
			t.Run(name, func(t *testing.T) {
				// Arrange
				config := types.Repository{
					Kickr: kickr.Kickr{
						CI:       &ci,
						Exclude:  []string{kickr.ExcludeMakefile},
						Platform: ci.Provider,
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
			{Option: kickr.OptionKickr + ":github-app", Provider: parser.GitHub},
			{Option: kickr.OptionKickr + ":personal-token", Provider: parser.GitHub},
			{Option: kickr.OptionKickr, Provider: parser.GitLab},
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
			for _, auth := range []string{kickr.AuthGitHubApp, kickr.AuthPersonalToken} {
				t.Run(auth, func(t *testing.T) {
					// Arrange
					config := types.Repository{
						Kickr: kickr.Kickr{
							CI:       &kickr.CI{Provider: parser.GitHub, Options: []string{"renovate:" + auth}},
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
					CI:       &kickr.CI{Provider: parser.GitLab, Options: []string{kickr.OptionRenovate}},
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
		type testcase struct {
			CI        string
			PreCommit bool
		}

		cases := []testcase{
			{CI: parser.GitHub},
			{CI: parser.GitHub, PreCommit: true},
			{CI: parser.GitLab},
			{CI: parser.GitLab, PreCommit: true},
		}
		for _, tc := range cases {
			name := tc.CI + "_" + strconv.FormatBool(tc.PreCommit)
			t.Run(name, func(t *testing.T) {
				// Arrange
				config := types.Repository{
					Kickr: kickr.Kickr{
						CI:      &kickr.CI{Provider: tc.CI},
						Exclude: []string{kickr.ExcludeMakefile},
					},
				}
				if !tc.PreCommit {
					config.Exclude = append(config.Exclude, kickr.ExcludePreCommit)
				} else {
					config.PreCommit = append(config.PreCommit, kickr.PreCommitAutoCommit, kickr.PreCommitGitflowBranches, kickr.PreCommitConventionalCommits)
				}

				// Act & Assert
				test(ctx, t, config)
			})
		}
	})

	t.Run("success_release", func(t *testing.T) {
		type testcase struct {
			Auth string
			Auto bool
			CI   string
		}

		cases := []testcase{
			{CI: parser.GitHub},
			{CI: parser.GitHub, Auto: true},

			{CI: parser.GitHub, Auth: kickr.AuthGitHubApp},
			{CI: parser.GitHub, Auth: kickr.AuthGitHubToken},
			{CI: parser.GitHub, Auth: kickr.AuthPersonalToken},

			{CI: parser.GitLab},
			{CI: parser.GitLab, Auto: true},
		}
		for _, tc := range cases {
			name := tc.CI
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
							Provider: tc.CI,
							Release:  &kickr.Release{Auto: tc.Auto, Auth: tc.Auth},
						},
						Exclude:  []string{kickr.ExcludeMakefile},
						Platform: tc.CI,
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
		for _, ci := range []string{parser.GitLab, parser.GitHub} {
			t.Run(ci, func(t *testing.T) {
				// Arrange
				config := types.Repository{
					Kickr: kickr.Kickr{
						CI:      &kickr.CI{Provider: ci},
						Exclude: []string{kickr.ExcludeMakefile},
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
				config := types.Repository{Kickr: kickr.Kickr{Exclude: []string{kickr.ExcludeMakefile}}}
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
		golang := func(ci string) func(ctx context.Context, destdir string, config *types.Repository) error {
			return func(_ context.Context, destdir string, _ *types.Repository) error {
				gomod := fmt.Appendf(nil, "module %s.com/kickr-dev/kickr\n\ngo 1.23\n", ci)
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

		for _, ci := range []string{parser.GitLab, parser.GitHub} {
			t.Run(ci, func(t *testing.T) {
				// Arrange
				config := types.Repository{
					Kickr: kickr.Kickr{
						CI:       &kickr.CI{Provider: ci, Release: &kickr.Release{}},
						Platform: ci,
					},
				}

				// Act & Assert
				test(ctx, t, config, golang(ci))
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
						Exclude:   []string{kickr.ExcludeMakefile},
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
						Exclude:   []string{kickr.ExcludeMakefile},
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
		golang := func(ci string) func(ctx context.Context, destdir string, config *types.Repository) error {
			return func(_ context.Context, destdir string, _ *types.Repository) error {
				gomod := fmt.Appendf(nil, "module %s.com/kickr-dev/kickr\n\ngo 1.23\n", ci)
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

		for _, ci := range []string{parser.GitLab, parser.GitHub} {
			t.Run(ci, func(t *testing.T) {
				// Arrange
				config := types.Repository{
					Kickr: kickr.Kickr{
						CI: &kickr.CI{
							Provider: ci,
							Docker:   &kickr.Docker{Path: "path/to/registry", Registry: "registry.example.com"},
							Helm: &kickr.Helm{
								Deploy:       kickr.HelmManual,
								Environments: []string{"staging", "production"},
								Path:         "path/to/repository",
								Publish:      kickr.HelmManual,
								Registry:     "chartmuseum.example.com",
							},
							Options: []string{
								kickr.OptionCodeCov,
								kickr.OptionCodeQL,
								kickr.OptionHardenRunner,
								kickr.OptionLabeler,
								kickr.OptionScoreCardOSSF,
								kickr.OptionSonarQube,
								kickr.OptionStepSecurityActions,
							},
							Release: &kickr.Release{},
						},
						Description: "A useful project description",
						Exclude:     []string{kickr.ExcludeShell},
						Platform:    ci,
						PreCommit:   []string{kickr.PreCommitGolangciLint},
					},
				}

				// Act & Assert
				test(ctx, t, config, golang(ci))
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
		for _, ci := range []string{parser.GitHub, parser.GitLab} {
			t.Run(ci, func(t *testing.T) {
				// Arrange
				config := types.Repository{
					Kickr: kickr.Kickr{CI: &kickr.CI{Provider: ci}, Platform: ci},
				}

				// Act & Assert
				test(ctx, t, config, hugo)
			})
		}
	})

	t.Run("success_netlify", func(t *testing.T) {
		cases := []kickr.CI{
			{Provider: parser.GitHub, Website: &kickr.Website{Hosting: kickr.HostingNetlify, Auto: true}},
			{Provider: parser.GitHub, Website: &kickr.Website{Hosting: kickr.HostingNetlify}},

			{Provider: parser.GitLab, Website: &kickr.Website{Hosting: kickr.HostingNetlify, Auto: true}},
			{Provider: parser.GitLab, Website: &kickr.Website{Hosting: kickr.HostingNetlify}},
		}
		for _, ci := range cases {
			name := fmt.Sprint(ci.Provider, "_auto_", ci.Website.Auto)
			t.Run(name, func(t *testing.T) {
				// Arrange
				config := types.Repository{
					Kickr: kickr.Kickr{CI: &ci, Platform: ci.Provider},
				}

				// Act & Assert
				test(ctx, t, config, hugo)
			})
		}
	})

	t.Run("success_pages", func(t *testing.T) {
		cases := []kickr.CI{
			{Provider: parser.GitHub, Website: &kickr.Website{Hosting: kickr.HostingPages, Auto: true}},
			{Provider: parser.GitHub, Website: &kickr.Website{Hosting: kickr.HostingPages}},

			{Provider: parser.GitLab, Website: &kickr.Website{Hosting: kickr.HostingPages, Auto: true}},
			{Provider: parser.GitLab, Website: &kickr.Website{Hosting: kickr.HostingPages}},
		}
		for _, ci := range cases {
			name := fmt.Sprint(ci.Provider, "_auto_", ci.Website.Auto)
			t.Run(name, func(t *testing.T) {
				// Arrange
				config := types.Repository{
					Kickr: kickr.Kickr{CI: &ci, Platform: ci.Provider},
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
			CI             string
			Options        []string
			PackageManager string
		}

		// Arrange
		node := func(tc testcase) func(ctx context.Context, destdir string, config *types.Repository) error {
			return func(_ context.Context, destdir string, _ *types.Repository) error {
				content := fmt.Appendf(nil, `{ "name": "kickr", "packageManager": "%s" }`+"\n", tc.PackageManager)
				return os.WriteFile(filepath.Join(destdir, parser.FilePackageJSON), content, files.RwRR)
			}
		}

		cases := []testcase{
			{CI: parser.GitHub, Options: []string{"renovate:personal-token"}, PackageManager: "bun@1.1.6"},
			{CI: parser.GitLab, Options: []string{kickr.OptionRenovate}, PackageManager: "bun@1.1.6"},
		}
		for _, tc := range cases {
			name := fmt.Sprint(tc.CI, "_", tc.PackageManager)
			t.Run(name, func(t *testing.T) {
				// Arrange
				config := types.Repository{
					Kickr: kickr.Kickr{
						CI: &kickr.CI{
							Options:  tc.Options,
							Provider: tc.CI,
							Release:  &kickr.Release{Options: []string{kickr.OptionBackmerge}},
						},
						Platform: tc.CI,
					},
				}

				// Act & Assert
				test(ctx, t, config, node(tc))
			})
		}
	})

	t.Run("success_netlify", func(t *testing.T) {
		// Arrange
		node := func(_ context.Context, destdir string, _ *types.Repository) error {
			return os.WriteFile(filepath.Join(destdir, parser.FilePackageJSON),
				[]byte(`{ "name": "kickr", "packageManager": "bun@1.1.6", "main": "index.js" }`+"\n"), files.RwRR)
		}

		cases := []kickr.CI{
			{Provider: parser.GitHub, Website: &kickr.Website{Hosting: kickr.HostingNetlify, Auto: true}},
			{Provider: parser.GitHub, Website: &kickr.Website{Hosting: kickr.HostingNetlify}},

			{Provider: parser.GitLab, Website: &kickr.Website{Hosting: kickr.HostingNetlify, Auto: true}},
			{Provider: parser.GitLab, Website: &kickr.Website{Hosting: kickr.HostingNetlify}},
		}
		for _, ci := range cases {
			name := fmt.Sprint(ci.Provider, "_auto_", ci.Website.Auto)
			t.Run(name, func(t *testing.T) {
				// Arrange
				config := types.Repository{
					Kickr: kickr.Kickr{
						CI:       &ci,
						Exclude:  []string{kickr.ExcludeMakefile},
						Platform: ci.Provider,
					},
				}

				// Act & Assert
				test(ctx, t, config, node)
			})
		}
	})

	t.Run("success_pages", func(t *testing.T) {
		// Arrange
		node := func(_ context.Context, destdir string, _ *types.Repository) error {
			return os.WriteFile(filepath.Join(destdir, parser.FilePackageJSON),
				[]byte(`{ "name": "kickr", "packageManager": "bun@1.1.6", "main": "index.js" }`+"\n"), files.RwRR)
		}

		cases := []kickr.CI{
			{Provider: parser.GitHub, Website: &kickr.Website{Hosting: kickr.HostingPages, Auto: true}},
			{Provider: parser.GitHub, Website: &kickr.Website{Hosting: kickr.HostingPages}},

			{Provider: parser.GitLab, Website: &kickr.Website{Hosting: kickr.HostingPages, Auto: true}},
			{Provider: parser.GitLab, Website: &kickr.Website{Hosting: kickr.HostingPages}},
		}
		for _, ci := range cases {
			name := fmt.Sprint(ci.Provider, "_auto_", ci.Website.Auto)
			t.Run(name, func(t *testing.T) {
				// Arrange
				config := types.Repository{
					Kickr: kickr.Kickr{
						CI:       &ci,
						Exclude:  []string{kickr.ExcludeMakefile},
						Platform: ci.Provider,
					},
				}

				// Act & Assert
				test(ctx, t, config, node)
			})
		}
	})

	t.Run("success_helm", func(t *testing.T) {
		// Arrange
		node := func(_ context.Context, destdir string, _ *types.Repository) error {
			return os.WriteFile(filepath.Join(destdir, parser.FilePackageJSON),
				[]byte(`{ "name": "kickr", "packageManager": "bun@1.1.6", "main": "index.js" }`+"\n"), files.RwRR)
		}

		cases := []kickr.CI{
			{Provider: parser.GitHub, Helm: &kickr.Helm{}},
			{Provider: parser.GitHub, Helm: &kickr.Helm{Deploy: kickr.HelmAuto, Environments: []string{"review"}}},
			{Provider: parser.GitHub, Helm: &kickr.Helm{Deploy: kickr.HelmManual, Environments: []string{"integration"}}},
			{Provider: parser.GitHub, Helm: &kickr.Helm{Publish: kickr.HelmAuto}},
			{Provider: parser.GitHub, Helm: &kickr.Helm{Publish: kickr.HelmManual}},

			{Provider: parser.GitLab, Helm: &kickr.Helm{}},
			{Provider: parser.GitLab, Helm: &kickr.Helm{Deploy: kickr.HelmAuto, Environments: []string{"review"}}},
			{Provider: parser.GitLab, Helm: &kickr.Helm{Deploy: kickr.HelmManual, Environments: []string{"integration"}}},
			{Provider: parser.GitLab, Helm: &kickr.Helm{Publish: kickr.HelmAuto}},
			{Provider: parser.GitLab, Helm: &kickr.Helm{Publish: kickr.HelmManual}},
		}
		for _, ci := range cases {
			publish := "none"
			if ci.Helm.Publish != "" {
				publish = ci.Helm.Publish
			}
			deploy := "none"
			if ci.Helm.Deploy != "" {
				deploy = ci.Helm.Deploy
			}

			name := fmt.Sprint(ci.Provider, "_deploy_", deploy, "_publish_", publish)
			t.Run(name, func(t *testing.T) {
				// Arrange
				config := types.Repository{
					Kickr: kickr.Kickr{
						CI:       &ci,
						Exclude:  []string{kickr.ExcludeMakefile},
						Platform: ci.Provider,
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
			{Apply: kickr.TerraformManual, Engine: kickr.EngineOpentofu, Provider: parser.GitHub},
			{Apply: kickr.TerraformAuto, Engine: kickr.EngineTerraform, Provider: parser.GitHub},

			{Apply: kickr.TerraformManual, Engine: kickr.EngineOpentofu, Provider: parser.GitLab},
			{Apply: kickr.TerraformAuto, Engine: kickr.EngineTerraform, Provider: parser.GitLab},
		}
		for _, tc := range cases {
			t.Run(tc.Provider+"_"+tc.Engine, func(t *testing.T) {
				// Arrange
				config := types.Repository{
					Kickr: kickr.Kickr{
						CI: &kickr.CI{
							Provider:  tc.Provider,
							Release:   &kickr.Release{},
							Terraform: &kickr.TerraformCI{Apply: tc.Apply, Environments: []string{"production"}},
						},
						Platform: tc.Provider,
						Terraform: &kickr.Terraform{
							Engine:  tc.Engine,
							Modules: []string{"modules/one", "modules/two"},
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
							Terraform: &kickr.TerraformCI{
								Apply:        kickr.TerraformAuto,
								Environments: []string{"production"},
							},
						},
						PreCommit: []string{kickr.PreCommitTerraform},
						Platform:  provider,
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
				[]byte(`{ "name": "kickr", "packageManager": "bun@1.1.6", "main": "index.js" }`+"\n"), files.RwRR)
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
			{Provider: parser.GitLab, Hosting: kickr.HostingPages},
			{Provider: parser.GitLab, Hosting: kickr.HostingNetlify},
			{Provider: parser.GitHub, Hosting: kickr.HostingPages},
			{Provider: parser.GitHub, Hosting: kickr.HostingNetlify},
		}
		for _, tc := range cases {
			name := fmt.Sprint(tc.Provider, "_", tc.Hosting)
			t.Run(name, func(t *testing.T) {
				// Arrange
				config := types.Repository{
					Kickr: kickr.Kickr{
						CI: &kickr.CI{
							Provider: tc.Provider,
							Website:  &kickr.Website{Hosting: tc.Hosting, Directory: "docs"},
						},
						Exclude:  []string{kickr.ExcludePreCommit},
						Platform: tc.Provider,
					},
				}

				// Act & Assert
				test(ctx, t, config, node(""), hugo)
			})
		}
	})

	t.Run("success_go_hugo_doc", func(t *testing.T) {
		cases := []testcase{
			{Provider: parser.GitLab, Hosting: kickr.HostingPages},
			{Provider: parser.GitLab, Hosting: kickr.HostingNetlify},
			{Provider: parser.GitHub, Hosting: kickr.HostingPages},
			{Provider: parser.GitHub, Hosting: kickr.HostingNetlify},
		}
		for _, tc := range cases {
			name := fmt.Sprint(tc.Provider, "_", tc.Hosting)
			t.Run(name, func(t *testing.T) {
				// Arrange
				config := types.Repository{
					Kickr: kickr.Kickr{
						CI: &kickr.CI{
							Provider: tc.Provider,
							Website:  &kickr.Website{Hosting: tc.Hosting, Directory: "docs"},
						},
						Exclude:  []string{kickr.ExcludePreCommit},
						Platform: tc.Provider,
					},
				}

				// Act & Assert
				test(ctx, t, config, golang(tc), hugo)
			})
		}
	})

	t.Run("success_go_node_doc", func(t *testing.T) {
		cases := []testcase{
			{Provider: parser.GitLab, Hosting: kickr.HostingPages},
			{Provider: parser.GitLab, Hosting: kickr.HostingNetlify},
			{Provider: parser.GitHub, Hosting: kickr.HostingPages},
			{Provider: parser.GitHub, Hosting: kickr.HostingNetlify},
		}
		for _, tc := range cases {
			name := fmt.Sprint(tc.Provider, "_", tc.Hosting)
			t.Run(name, func(t *testing.T) {
				// Arrange
				config := types.Repository{
					Kickr: kickr.Kickr{
						CI: &kickr.CI{
							Provider: tc.Provider,
							Website:  &kickr.Website{Hosting: tc.Hosting, Directory: "docs"},
						},
						Exclude:  []string{kickr.ExcludeMakefile, kickr.ExcludePreCommit},
						Platform: tc.Provider,
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
						Terraform: &kickr.Terraform{Engine: kickr.EngineOpentofu, Modules: []string{".terraform"}},
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

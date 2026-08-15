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

	"dario.cat/mergo"
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

type testcase struct {
	Config kickr.Kickr
	Name   string
}

func TestGenerate_NoLang(t *testing.T) {
	ctx := t.Context()

	t.Run("chart", func(t *testing.T) {
		cases := []testcase{
			{
				Name:   "github",
				Config: kickr.Kickr{GitHub: &kickr.GitHub{}, Helm: &kickr.Helm{}},
			},
			{
				Name:   "github_publish_auto",
				Config: kickr.Kickr{GitHub: &kickr.GitHub{}, Helm: &kickr.Helm{Publish: kickr.HelmPublishAuto}},
			},
			{
				Name: "github_publish_manual",
				Config: kickr.Kickr{
					GitHub: &kickr.GitHub{},
					Helm:   &kickr.Helm{Path: "path/to/kickr", Publish: kickr.HelmPublishManual, Registry: "chartmuseum.example.com"},
				},
			},
			{
				Name:   "gitlab",
				Config: kickr.Kickr{GitLab: &kickr.GitLab{}, Helm: &kickr.Helm{}},
			},
			{
				Name:   "gitlab_publish_auto",
				Config: kickr.Kickr{GitLab: &kickr.GitLab{}, Helm: &kickr.Helm{Publish: kickr.HelmPublishAuto}},
			},
			{
				Name: "gitlab_publish_manual",
				Config: kickr.Kickr{
					GitLab: &kickr.GitLab{},
					Helm:   &kickr.Helm{Path: "path/to/kickr", Publish: kickr.HelmPublishManual, Registry: "chartmuseum.example.com"},
				},
			},
		}
		for _, tc := range cases {
			t.Run(tc.Name, func(t *testing.T) {
				// Arrange
				repo := types.Repository{
					Config: merge(t, kickr.Kickr{
						Exclude: []string{kickr.ExcludeMakefile, kickr.ExcludeRenovate},
					}, tc.Config),
				}

				// Act & Assert
				test(ctx, t, repo)
			})
		}
	})

	t.Run("kickr", func(t *testing.T) {
		cases := []testcase{
			{
				Name:   "github_github_app",
				Config: kickr.Kickr{GitHub: &kickr.GitHub{Options: []string{kickr.GitHubOptionsKickrGitHubApp}}},
			},
			{
				Name:   "github_personal_token",
				Config: kickr.Kickr{GitHub: &kickr.GitHub{Options: []string{kickr.GitHubOptionsKickrPersonalToken}}},
			},
			{
				Name:   "gitlab",
				Config: kickr.Kickr{GitLab: &kickr.GitLab{Options: []string{kickr.GitLabOptionsKickr}}},
			},
		}
		for _, tc := range cases {
			t.Run(tc.Name, func(t *testing.T) {
				// Arrange
				repo := types.Repository{
					Config: merge(t, kickr.Kickr{
						Exclude: []string{kickr.ExcludeMakefile, kickr.ExcludeShell},
					}, tc.Config),
				}

				// Act & Assert
				test(ctx, t, repo)
			})
		}
	})

	t.Run("renovate", func(t *testing.T) {
		cases := []testcase{
			{
				Name:   "github_github_app",
				Config: kickr.Kickr{GitHub: &kickr.GitHub{Options: []string{kickr.GitHubOptionsRenovateGitHubApp}}},
			},
			{
				Name:   "github_personal_token",
				Config: kickr.Kickr{GitHub: &kickr.GitHub{Options: []string{kickr.GitHubOptionsRenovatePersonalToken}}},
			},
			{
				Name:   "gitlab",
				Config: kickr.Kickr{GitLab: &kickr.GitLab{Options: []string{kickr.GitLabOptionsRenovate}}},
			},
		}
		for _, tc := range cases {
			t.Run(tc.Name, func(t *testing.T) {
				// Arrange
				repo := types.Repository{
					Config: merge(t, kickr.Kickr{
						Exclude: []string{kickr.ExcludeMakefile, kickr.ExcludeShell},
					}, tc.Config),
				}

				// Act & Assert
				test(ctx, t, repo)
			})
		}

		t.Run("templates", func(t *testing.T) {
			// Arrange
			tmpl := func(_ context.Context, destdir string, _ *types.Repository) error {
				file, err := os.Create(filepath.Join(destdir, "template.tmpl"))
				if err != nil {
					return fmt.Errorf("create: %w", err)
				}
				return file.Close()
			}

			repo := types.Repository{
				Config: kickr.Kickr{Exclude: []string{kickr.ExcludeMakefile}},
			}

			// Act & Assert
			test(ctx, t, repo, tmpl)
		})
	})

	t.Run("precommit", func(t *testing.T) {
		cases := []testcase{
			{Name: "enabled_github", Config: kickr.Kickr{GitHub: &kickr.GitHub{}}},
			{Name: "enabled_gitlab", Config: kickr.Kickr{GitLab: &kickr.GitLab{}}},
			{
				Name:   "disabled_github",
				Config: kickr.Kickr{GitHub: &kickr.GitHub{}, Exclude: []string{kickr.ExcludePreCommit}},
			},
			{
				Name:   "disabled_gitlab",
				Config: kickr.Kickr{GitLab: &kickr.GitLab{}, Exclude: []string{kickr.ExcludePreCommit}},
			},
		}
		for _, tc := range cases {
			t.Run(tc.Name, func(t *testing.T) {
				// Arrange
				repo := types.Repository{
					Config: merge(t, kickr.Kickr{
						Exclude:   []string{kickr.ExcludeMakefile, kickr.ExcludeRenovate},
						PreCommit: []string{kickr.PreCommitAutoCommit, kickr.PreCommitGitflowBranches, kickr.PreCommitConventionalCommits},
					}, tc.Config),
				}

				// Act & Assert
				test(ctx, t, repo)
			})
		}
	})

	t.Run("release", func(t *testing.T) {
		cases := []testcase{
			{
				Name:   "github",
				Config: kickr.Kickr{GitHub: &kickr.GitHub{Release: &kickr.Release{Options: []string{kickr.ReleaseOptionsBackmerge}}}},
			},
			{
				Name:   "github_auto",
				Config: kickr.Kickr{GitHub: &kickr.GitHub{Release: &kickr.Release{Auto: true}}},
			},
			{
				Name:   "github_auth_github_app",
				Config: kickr.Kickr{GitHub: &kickr.GitHub{Release: &kickr.Release{Auth: kickr.ReleaseAuthGitHubApp}}},
			},
			{
				Name:   "github_auth_github_token",
				Config: kickr.Kickr{GitHub: &kickr.GitHub{Release: &kickr.Release{Auth: kickr.ReleaseAuthGitHubToken}}},
			},
			{
				Name:   "github_auth_personal_token",
				Config: kickr.Kickr{GitHub: &kickr.GitHub{Release: &kickr.Release{Auth: kickr.ReleaseAuthPersonalToken}}},
			},
			{
				Name:   "gitlab",
				Config: kickr.Kickr{GitLab: &kickr.GitLab{Release: &kickr.Release{Options: []string{kickr.ReleaseOptionsBackmerge}}}},
			},
			{
				Name:   "gitlab_auto",
				Config: kickr.Kickr{GitLab: &kickr.GitLab{Release: &kickr.Release{Auto: true}}},
			},
		}
		for _, tc := range cases {
			t.Run(tc.Name, func(t *testing.T) {
				// Arrange
				repo := types.Repository{
					Config: merge(t, kickr.Kickr{
						Exclude: []string{kickr.ExcludeMakefile, kickr.ExcludeRenovate},
					}, tc.Config),
				}

				// Act & Assert
				test(ctx, t, repo)
			})
		}
	})
}

func TestGenerate_Shell(t *testing.T) {
	ctx := t.Context()

	shell := func(_ context.Context, destdir string, _ *types.Repository) error {
		return os.WriteFile(filepath.Join(destdir, "script.sh"), []byte("#!/bin/sh\n"), files.RwxRxRxRx)
	}

	cases := []testcase{
		{Name: "github", Config: kickr.Kickr{GitHub: &kickr.GitHub{}}},
		{Name: "gitlab", Config: kickr.Kickr{GitLab: &kickr.GitLab{}}},
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			// Arrange
			repo := types.Repository{
				Config: merge(t, kickr.Kickr{Exclude: []string{kickr.ExcludeMakefile, kickr.ExcludeRenovate}}, tc.Config),
			}

			// Act & Assert
			test(ctx, t, repo, shell)
		})
	}

	t.Run("precommit", func(t *testing.T) {
		cases := []testcase{
			{Name: "disabled", Config: kickr.Kickr{Exclude: []string{kickr.ExcludePreCommit}}},
			{Name: "enabled", Config: kickr.Kickr{}},
		}
		for _, tc := range cases {
			t.Run(tc.Name, func(t *testing.T) {
				// Arrange
				repo := types.Repository{
					Config: merge(t, kickr.Kickr{Exclude: []string{kickr.ExcludeMakefile, kickr.ExcludeRenovate}}, tc.Config),
				}

				// Act & Assert
				test(ctx, t, repo, shell)
			})
		}
	})
}

func TestGenerate_Golang(t *testing.T) {
	ctx := t.Context()

	t.Run("cli", func(t *testing.T) {
		// Arrange
		golang := func(provider string) func(ctx context.Context, destdir string, repo *types.Repository) error {
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

		cases := []testcase{
			{Name: "github", Config: kickr.Kickr{GitHub: &kickr.GitHub{Release: &kickr.Release{}}}},
			{Name: "gitlab", Config: kickr.Kickr{GitLab: &kickr.GitLab{Release: &kickr.Release{}}}},
		}
		for _, tc := range cases {
			t.Run(tc.Name, func(t *testing.T) {
				// Arrange
				repo := types.Repository{Config: merge(t, kickr.Kickr{}, tc.Config)}

				// Act & Assert
				test(ctx, t, repo, golang(tc.Name))
			})
		}
	})

	t.Run("library", func(t *testing.T) {
		// Arrange
		golang := func(platform string) func(ctx context.Context, destdir string, repo *types.Repository) error {
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
				repo := types.Repository{
					Config: kickr.Kickr{
						Exclude:   []string{kickr.ExcludeMakefile, kickr.ExcludeRenovate},
						PreCommit: []string{kickr.PreCommitGomodTidy},
						Platform:  platform,
					},
				}

				// Act & Assert
				test(ctx, t, repo, golang(platform))
			})
		}
	})

	t.Run("multiple_bin_helm", func(t *testing.T) {
		// Arrange
		golang := func(provider string) func(ctx context.Context, destdir string, repo *types.Repository) error {
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

		cases := []testcase{
			{
				Name: "github",
				Config: kickr.Kickr{
					GitHub: &kickr.GitHub{
						Options: []string{
							kickr.GitHubOptionsCodecov,
							kickr.GitHubOptionsCodeQL,
							kickr.GitHubOptionsHardenRunner,
							kickr.GitHubOptionsLabeler,
							kickr.GitHubOptionsOSSFScorecard,
							kickr.GitHubOptionsSonarQube,
							kickr.GitHubOptionsStepSecurityActions,
						},
						Release: &kickr.Release{},
					},
				},
			},
			{
				Name: "gitlab",
				Config: kickr.Kickr{
					GitLab: &kickr.GitLab{
						Options: []string{
							kickr.GitLabOptionsSonarQube,
							kickr.GitLabOptionsOverridesIntegration,
							kickr.GitLabOptionsOverridesDeployment,
						},
						Release: &kickr.Release{},
					},
				},
			},
		}
		for _, tc := range cases {
			t.Run(tc.Name, func(t *testing.T) {
				// Arrange
				repo := types.Repository{
					Config: merge(t, kickr.Kickr{
						Docker: &kickr.Docker{Path: "path/to/image", Registry: "registry.example.com"},
						Helm: &kickr.Helm{
							Deploy:       kickr.HelmDeployManual,
							Environments: []string{kickr.EnvironmentStaging, kickr.EnvironmentProduction},
							Path:         "path/to/repository",
							Publish:      kickr.HelmPublishManual,
							Registry:     "chartmuseum.example.com",
						},
						Description: "A useful project description",
						Exclude:     []string{kickr.ExcludeRenovate, kickr.ExcludeShell},
						Modules:     []kickr.Module{{Path: types.RootModule, Deployment: &kickr.Deployment{Target: kickr.DeploymentTargetHelm}}},
						PreCommit:   []string{kickr.PreCommitGolangciLint},
					}, tc.Config),
				}

				// Act & Assert
				test(ctx, t, repo, golang(tc.Name))
			})
		}
	})

	t.Run("multiple_libraries", func(t *testing.T) {
		// Arrange
		golang := func(platform string) func(ctx context.Context, destdir string, repo *types.Repository) error {
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
				repo := types.Repository{
					Config: kickr.Kickr{
						Exclude:   []string{kickr.ExcludeMakefile, kickr.ExcludeRenovate},
						PreCommit: []string{kickr.PreCommitGomodTidy},
						Platform:  platform,
					},
				}

				// Act & Assert
				test(ctx, t, repo, golang(platform))
			})
		}
	})
}

func TestGenerate_Hugo(t *testing.T) {
	ctx := t.Context()

	hugo := func(hugodir string) engine.Parser[types.Repository] {
		return func(_ context.Context, destdir string, _ *types.Repository) error {
			if err := os.MkdirAll(filepath.Join(destdir, hugodir), files.RwxRxRxRx); err != nil {
				return fmt.Errorf("mkdir all: %w", err)
			}
			file, err := os.Create(filepath.Join(destdir, hugodir, "hugo.toml"))
			if err != nil {
				return fmt.Errorf("create: %w", err)
			}
			return file.Close()
		}
	}

	t.Run("no_website", func(t *testing.T) {
		cases := []testcase{
			{Name: "github", Config: kickr.Kickr{GitHub: &kickr.GitHub{}}},
			{Name: "gitlab", Config: kickr.Kickr{GitLab: &kickr.GitLab{}}},
		}
		for _, tc := range cases {
			t.Run(tc.Name, func(t *testing.T) {
				// Arrange
				repo := types.Repository{Config: merge(t, kickr.Kickr{}, tc.Config)}

				// Act & Assert
				test(ctx, t, repo, hugo(""))
			})
		}
	})

	t.Run("hosting", func(t *testing.T) {
		cases := []testcase{
			{
				Name: "github_netlify_auto",
				Config: kickr.Kickr{
					GitHub: &kickr.GitHub{},
					Modules: []kickr.Module{
						{Path: types.RootModule, Deployment: &kickr.Deployment{Target: kickr.DeploymentTargetNetlify, Auto: true}},
					},
				},
			},
			{
				Name: "github_netlify",
				Config: kickr.Kickr{
					GitHub: &kickr.GitHub{},
					Modules: []kickr.Module{
						{Path: types.RootModule, Deployment: &kickr.Deployment{Target: kickr.DeploymentTargetNetlify}},
					},
				},
			},
			{
				Name: "gitlab_netlify_auto",
				Config: kickr.Kickr{
					GitLab: &kickr.GitLab{},
					Modules: []kickr.Module{
						{Path: types.RootModule, Deployment: &kickr.Deployment{Target: kickr.DeploymentTargetNetlify, Auto: true}},
					},
				},
			},
			{
				Name: "gitlab_netlify",
				Config: kickr.Kickr{
					GitLab: &kickr.GitLab{},
					Modules: []kickr.Module{
						{Path: types.RootModule, Deployment: &kickr.Deployment{Target: kickr.DeploymentTargetNetlify}},
					},
				},
			},
			{
				Name: "github_pages_auto",
				Config: kickr.Kickr{
					GitHub: &kickr.GitHub{},
					Modules: []kickr.Module{
						{Path: types.RootModule, Deployment: &kickr.Deployment{Target: kickr.DeploymentTargetPages, Auto: true}},
					},
				},
			},
			{
				Name: "github_pages",
				Config: kickr.Kickr{
					GitHub: &kickr.GitHub{},
					Modules: []kickr.Module{
						{Path: types.RootModule, Deployment: &kickr.Deployment{Target: kickr.DeploymentTargetPages}},
					},
				},
			},
			{
				Name: "gitlab_pages_auto",
				Config: kickr.Kickr{
					GitLab: &kickr.GitLab{},
					Modules: []kickr.Module{
						{Path: types.RootModule, Deployment: &kickr.Deployment{Target: kickr.DeploymentTargetPages, Auto: true}},
					},
				},
			},
			{
				Name: "gitlab_pages",
				Config: kickr.Kickr{
					GitLab: &kickr.GitLab{},
					Modules: []kickr.Module{
						{Path: types.RootModule, Deployment: &kickr.Deployment{Target: kickr.DeploymentTargetPages}},
					},
				},
			},
		}
		for _, tc := range cases {
			t.Run(tc.Name, func(t *testing.T) {
				// Arrange
				repo := types.Repository{
					Config: merge(t, kickr.Kickr{}, tc.Config),
				}

				// Act & Assert
				test(ctx, t, repo, hugo(""))
			})
		}
	})

	t.Run("docker", func(t *testing.T) {
		cases := []testcase{
			{Name: "github", Config: kickr.Kickr{GitHub: &kickr.GitHub{}}},
			{Name: "gitlab", Config: kickr.Kickr{GitLab: &kickr.GitLab{}}},
		}
		for _, tc := range cases {
			t.Run(tc.Name, func(t *testing.T) {
				// Arrange
				repo := types.Repository{
					Config: merge(t, kickr.Kickr{Docker: &kickr.Docker{}}, tc.Config),
				}

				// Act & Assert
				test(ctx, t, repo, hugo(""))
			})
		}
	})

	t.Run("netlify_and_pages", func(t *testing.T) {
		cases := []testcase{
			{Name: "github", Config: kickr.Kickr{GitHub: &kickr.GitHub{}}},
			{Name: "gitlab", Config: kickr.Kickr{GitLab: &kickr.GitLab{}}},
		}
		for _, tc := range cases {
			t.Run(tc.Name, func(t *testing.T) {
				// Arrange
				repo := types.Repository{
					Config: merge(t, kickr.Kickr{
						Exclude: []string{kickr.ExcludePreCommit, kickr.ExcludeRenovate},
						Modules: []kickr.Module{
							{Deployment: &kickr.Deployment{Target: kickr.DeploymentTargetPages}, Path: "docs"},
							{Deployment: &kickr.Deployment{Target: kickr.DeploymentTargetNetlify}, Path: "website"},
						},
					}, tc.Config),
				}

				// Act & Assert
				test(ctx, t, repo, hugo("docs"), hugo("website"))
			})
		}
	})
}

func TestGenerate_Node(t *testing.T) {
	ctx := t.Context()

	t.Run("package_managers", func(t *testing.T) {
		// Arrange
		node := func(tc string) func(ctx context.Context, destdir string, repo *types.Repository) error {
			return func(_ context.Context, destdir string, _ *types.Repository) error {
				content := fmt.Appendf(nil, `{ "name": "kickr", "packageManager": "%s" }`+"\n", tc)
				return os.WriteFile(filepath.Join(destdir, parser.FilePackageJSON), content, files.RwRR)
			}
		}

		for _, tc := range []string{"bun@1.1.6", "npm@7.0.0", "pnpm@9.0.0", "yarn@1.22.10"} {
			t.Run(tc, func(t *testing.T) {
				// Arrange
				repo := types.Repository{
					Config: kickr.Kickr{
						GitHub:   &kickr.GitHub{},
						Platform: parser.GitHub,
					},
				}

				// Act & Assert
				test(ctx, t, repo, node(tc))
			})
		}
	})

	t.Run("library", func(t *testing.T) {
		// Arrange
		node := func(_ context.Context, destdir string, _ *types.Repository) error {
			return os.WriteFile(filepath.Join(destdir, parser.FilePackageJSON),
				[]byte(`{ "name": "kickr", "packageManager": "bun@1.1.6", "main": "index.js" }`+"\n"), files.RwRR)
		}

		cases := []testcase{
			{Name: "github_bun", Config: kickr.Kickr{GitHub: &kickr.GitHub{Release: &kickr.Release{}}}},
			{Name: "gitlab_bun", Config: kickr.Kickr{GitLab: &kickr.GitLab{Release: &kickr.Release{}}}},
		}
		for _, tc := range cases {
			t.Run(tc.Name, func(t *testing.T) {
				// Arrange
				repo := types.Repository{Config: merge(t, kickr.Kickr{}, tc.Config)}

				// Act & Assert
				test(ctx, t, repo, node)
			})
		}
	})

	t.Run("hosting", func(t *testing.T) {
		// Arrange
		node := func(_ context.Context, destdir string, _ *types.Repository) error {
			return os.WriteFile(filepath.Join(destdir, parser.FilePackageJSON),
				[]byte(`{ "name": "kickr", "packageManager": "bun@1.1.6", "main": "index.js" }`+"\n"), files.RwRR)
		}

		cases := []testcase{
			{
				Name: "github_netlify_auto",
				Config: kickr.Kickr{
					GitHub: &kickr.GitHub{},
					Modules: []kickr.Module{
						{Path: types.RootModule, Deployment: &kickr.Deployment{Target: kickr.DeploymentTargetNetlify, Auto: true}},
					},
				},
			},
			{
				Name: "github_netlify",
				Config: kickr.Kickr{
					GitHub: &kickr.GitHub{},
					Modules: []kickr.Module{
						{Path: types.RootModule, Deployment: &kickr.Deployment{Target: kickr.DeploymentTargetNetlify}},
					},
				},
			},
			{
				Name: "gitlab_netlify_auto",
				Config: kickr.Kickr{
					GitLab: &kickr.GitLab{},
					Modules: []kickr.Module{
						{Path: types.RootModule, Deployment: &kickr.Deployment{Target: kickr.DeploymentTargetNetlify, Auto: true}},
					},
				},
			},
			{
				Name: "gitlab_netlify",
				Config: kickr.Kickr{
					GitLab: &kickr.GitLab{},
					Modules: []kickr.Module{
						{Path: types.RootModule, Deployment: &kickr.Deployment{Target: kickr.DeploymentTargetNetlify}},
					},
				},
			},
			{
				Name: "github_pages_auto",
				Config: kickr.Kickr{
					GitHub: &kickr.GitHub{},
					Modules: []kickr.Module{
						{Path: types.RootModule, Deployment: &kickr.Deployment{Target: kickr.DeploymentTargetPages, Auto: true}},
					},
				},
			},
			{
				Name: "github_pages",
				Config: kickr.Kickr{
					GitHub: &kickr.GitHub{},
					Modules: []kickr.Module{
						{Path: types.RootModule, Deployment: &kickr.Deployment{Target: kickr.DeploymentTargetPages}},
					},
				},
			},
			{
				Name: "gitlab_pages_auto",
				Config: kickr.Kickr{
					GitLab: &kickr.GitLab{},
					Modules: []kickr.Module{
						{Path: types.RootModule, Deployment: &kickr.Deployment{Target: kickr.DeploymentTargetPages, Auto: true}},
					},
				},
			},
			{
				Name: "gitlab_pages",
				Config: kickr.Kickr{
					GitLab: &kickr.GitLab{},
					Modules: []kickr.Module{
						{Path: types.RootModule, Deployment: &kickr.Deployment{Target: kickr.DeploymentTargetPages}},
					},
				},
			},
		}
		for _, tc := range cases {
			t.Run(tc.Name, func(t *testing.T) {
				// Arrange
				repo := types.Repository{
					Config: merge(t, kickr.Kickr{
						Exclude: []string{kickr.ExcludeMakefile, kickr.ExcludeRenovate},
					}, tc.Config),
				}

				// Act & Assert
				test(ctx, t, repo, node)
			})
		}
	})

	t.Run("helm", func(t *testing.T) {
		// Arrange
		node := func(_ context.Context, destdir string, _ *types.Repository) error {
			return os.WriteFile(filepath.Join(destdir, parser.FilePackageJSON),
				[]byte(`{ "name": "kickr", "packageManager": "bun@1.1.6", "main": "index.js" }`+"\n"), files.RwRR)
		}

		cases := []testcase{
			{
				Name: "github",
				Config: kickr.Kickr{
					GitHub: &kickr.GitHub{},
					Helm:   &kickr.Helm{},
					Modules: []kickr.Module{
						{Path: types.RootModule, Deployment: &kickr.Deployment{Target: kickr.DeploymentTargetHelm}},
					},
				},
			},
			{
				Name: "github_deploy_auto",
				Config: kickr.Kickr{
					GitHub: &kickr.GitHub{},
					Helm:   &kickr.Helm{Deploy: kickr.HelmDeployAuto, Environments: []string{kickr.EnvironmentReview}},
					Modules: []kickr.Module{
						{Path: types.RootModule, Deployment: &kickr.Deployment{Target: kickr.DeploymentTargetHelm}},
					},
				},
			},
			{
				Name: "github_deploy_manual",
				Config: kickr.Kickr{
					GitHub: &kickr.GitHub{},
					Helm:   &kickr.Helm{Deploy: kickr.HelmDeployManual, Environments: []string{kickr.EnvironmentIntegration}},
					Modules: []kickr.Module{
						{Path: types.RootModule, Deployment: &kickr.Deployment{Target: kickr.DeploymentTargetHelm}},
					},
				},
			},
			{
				Name: "github_publish_auto",
				Config: kickr.Kickr{
					GitHub: &kickr.GitHub{},
					Helm:   &kickr.Helm{Publish: kickr.HelmPublishAuto},
					Modules: []kickr.Module{
						{Path: types.RootModule, Deployment: &kickr.Deployment{Target: kickr.DeploymentTargetHelm}},
					},
				},
			},
			{
				Name: "github_publish_manual",
				Config: kickr.Kickr{
					GitHub: &kickr.GitHub{},
					Helm:   &kickr.Helm{Publish: kickr.HelmPublishManual},
					Modules: []kickr.Module{
						{Path: types.RootModule, Deployment: &kickr.Deployment{Target: kickr.DeploymentTargetHelm}},
					},
				},
			},
			{
				Name: "gitlab",
				Config: kickr.Kickr{
					GitLab: &kickr.GitLab{},
					Helm:   &kickr.Helm{},
					Modules: []kickr.Module{
						{Path: types.RootModule, Deployment: &kickr.Deployment{Target: kickr.DeploymentTargetHelm}},
					},
				},
			},
			{
				Name: "gitlab_deploy_auto",
				Config: kickr.Kickr{
					GitLab: &kickr.GitLab{},
					Helm:   &kickr.Helm{Deploy: kickr.HelmDeployAuto, Environments: []string{kickr.EnvironmentReview}},
					Modules: []kickr.Module{
						{Path: types.RootModule, Deployment: &kickr.Deployment{Target: kickr.DeploymentTargetHelm}},
					},
				},
			},
			{
				Name: "gitlab_deploy_manual",
				Config: kickr.Kickr{
					GitLab: &kickr.GitLab{},
					Helm:   &kickr.Helm{Deploy: kickr.HelmDeployManual, Environments: []string{kickr.EnvironmentIntegration}},
					Modules: []kickr.Module{
						{Path: types.RootModule, Deployment: &kickr.Deployment{Target: kickr.DeploymentTargetHelm}},
					},
				},
			},
			{
				Name: "gitlab_publish_auto",
				Config: kickr.Kickr{
					GitLab: &kickr.GitLab{},
					Helm:   &kickr.Helm{Publish: kickr.HelmPublishAuto},
					Modules: []kickr.Module{
						{Path: types.RootModule, Deployment: &kickr.Deployment{Target: kickr.DeploymentTargetHelm}},
					},
				},
			},
			{
				Name: "gitlab_publish_manual",
				Config: kickr.Kickr{
					GitLab: &kickr.GitLab{},
					Helm:   &kickr.Helm{Publish: kickr.HelmPublishManual},
					Modules: []kickr.Module{
						{Path: types.RootModule, Deployment: &kickr.Deployment{Target: kickr.DeploymentTargetHelm}},
					},
				},
			},
		}
		for _, tc := range cases {
			t.Run(tc.Name, func(t *testing.T) {
				// Arrange
				repo := types.Repository{
					Config: merge(t, kickr.Kickr{
						Exclude: []string{kickr.ExcludeMakefile, kickr.ExcludeRenovate},
					}, tc.Config),
				}

				// Act & Assert
				test(ctx, t, repo, node)
			})
		}
	})

	t.Run("docker", func(t *testing.T) {
		// Arrange
		node := func(_ context.Context, destdir string, _ *types.Repository) error {
			return os.WriteFile(filepath.Join(destdir, parser.FilePackageJSON),
				[]byte(`{ "name": "kickr", "packageManager": "bun@1.1.6", "main": "index.js" }`+"\n"), files.RwRR)
		}

		cases := []testcase{
			{Name: "github", Config: kickr.Kickr{GitHub: &kickr.GitHub{}}},
			{Name: "gitlab", Config: kickr.Kickr{GitLab: &kickr.GitLab{}}},
		}
		for _, tc := range cases {
			t.Run(tc.Name, func(t *testing.T) {
				// Arrange
				repo := types.Repository{
					Config: merge(t, kickr.Kickr{Docker: &kickr.Docker{}}, tc.Config),
				}

				// Act & Assert
				test(ctx, t, repo, node)
			})
		}
	})

	t.Run("docker_and_publish", func(t *testing.T) {
		// Arrange
		node := func(_ context.Context, destdir string, _ *types.Repository) error {
			return os.WriteFile(filepath.Join(destdir, parser.FilePackageJSON),
				[]byte(`{
					"name": "kickr",
					"packageManager": "bun@1.1.6",
					"main": "index.js",
					"publishConfig": { "registry": "https://registry.npmjs.org" }
				}`+"\n"), files.RwRR)
		}

		cases := []testcase{
			{Name: "github", Config: kickr.Kickr{GitHub: &kickr.GitHub{Release: &kickr.Release{}}}},
			{Name: "gitlab", Config: kickr.Kickr{GitLab: &kickr.GitLab{Release: &kickr.Release{}}}},
		}
		for _, tc := range cases {
			t.Run(tc.Name, func(t *testing.T) {
				// Arrange
				repo := types.Repository{
					Config: merge(t, kickr.Kickr{Docker: &kickr.Docker{}}, tc.Config),
				}

				// Act & Assert
				test(ctx, t, repo, node)
			})
		}
	})
}

func TestGenerate_Terraform(t *testing.T) {
	ctx := t.Context()

	t.Run("multiple_modules", func(t *testing.T) {
		// Arrange
		terraform := func(subdir string) func(ctx context.Context, destdir string, repo *types.Repository) error {
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
			{
				Name: "github_tofu_apply_manual",
				Config: kickr.Kickr{
					GitHub: &kickr.GitHub{Release: &kickr.Release{}},
					Modules: []kickr.Module{
						{Path: types.RootModule},
						{
							Path: "modules/one",
							Terraform: &kickr.Terraform{
								Apply:        kickr.TerraformApplyManual,
								Engine:       kickr.TerraformEngineTofu,
								Environments: []string{kickr.EnvironmentProduction},
							},
						},
						{
							Path: "modules/two",
							Terraform: &kickr.Terraform{
								Apply:        kickr.TerraformApplyManual,
								Engine:       kickr.TerraformEngineTofu,
								Environments: []string{kickr.EnvironmentProduction},
							},
						},
					},
				},
			},
			{
				Name: "github_terraform_apply_auto",
				Config: kickr.Kickr{
					GitHub: &kickr.GitHub{Release: &kickr.Release{}},
					Modules: []kickr.Module{
						{Path: types.RootModule},
						{
							Path: "modules/one",
							Terraform: &kickr.Terraform{
								Apply:        kickr.TerraformApplyAuto,
								Engine:       kickr.TerraformEngineTerraform,
								Environments: []string{kickr.EnvironmentProduction},
							},
						},
						{
							Path: "modules/two",
							Terraform: &kickr.Terraform{
								Apply:        kickr.TerraformApplyAuto,
								Engine:       kickr.TerraformEngineTerraform,
								Environments: []string{kickr.EnvironmentProduction},
							},
						},
					},
				},
			},
			{
				Name: "gitlab_tofu_apply_manual",
				Config: kickr.Kickr{
					GitLab: &kickr.GitLab{Release: &kickr.Release{}},
					Modules: []kickr.Module{
						{Path: types.RootModule},
						{
							Path: "modules/one",
							Terraform: &kickr.Terraform{
								Apply:        kickr.TerraformApplyManual,
								Engine:       kickr.TerraformEngineTofu,
								Environments: []string{kickr.EnvironmentProduction},
							},
						},
						{
							Path: "modules/two",
							Terraform: &kickr.Terraform{
								Apply:        kickr.TerraformApplyManual,
								Engine:       kickr.TerraformEngineTofu,
								Environments: []string{kickr.EnvironmentProduction},
							},
						},
					},
				},
			},
			{
				Name: "gitlab_terraform_apply_auto",
				Config: kickr.Kickr{
					GitLab: &kickr.GitLab{Release: &kickr.Release{}},
					Modules: []kickr.Module{
						{Path: types.RootModule},
						{
							Path: "modules/one",
							Terraform: &kickr.Terraform{
								Apply:        kickr.TerraformApplyAuto,
								Engine:       kickr.TerraformEngineTerraform,
								Environments: []string{kickr.EnvironmentProduction},
							},
						},
						{
							Path: "modules/two",
							Terraform: &kickr.Terraform{
								Apply:        kickr.TerraformApplyAuto,
								Engine:       kickr.TerraformEngineTerraform,
								Environments: []string{kickr.EnvironmentProduction},
							},
						},
					},
				},
			},
		}
		for _, tc := range cases {
			t.Run(tc.Name, func(t *testing.T) {
				// Arrange
				repo := types.Repository{
					Config: merge(t, kickr.Kickr{}, tc.Config),
				}

				// Act & Assert
				test(ctx, t, repo, terraform(filepath.Join("modules", "one")), terraform(filepath.Join("modules", "two")))
			})
		}
	})

	t.Run("root_module", func(t *testing.T) {
		// Arrange
		terraform := func(_ context.Context, destdir string, _ *types.Repository) error {
			return os.WriteFile(filepath.Join(destdir, "main.tf"), []byte(`terraform { backend "s3" {} }`+"\n"), files.RwRR)
		}

		cases := []testcase{
			{Name: "github", Config: kickr.Kickr{GitHub: &kickr.GitHub{}}},
			{Name: "gitlab", Config: kickr.Kickr{GitLab: &kickr.GitLab{}}},
		}
		for _, tc := range cases {
			t.Run(tc.Name, func(t *testing.T) {
				// Arrange
				repo := types.Repository{
					Config: merge(t, kickr.Kickr{
						PreCommit: []string{kickr.PreCommitTerraform},
						Modules: []kickr.Module{
							{
								Path: types.RootModule,
								Terraform: &kickr.Terraform{
									Engine:       kickr.TerraformEngineTofu,
									Environments: []string{kickr.EnvironmentProduction},
								},
							},
						},
					}, tc.Config),
				}

				// Act & Assert
				test(ctx, t, repo, terraform)
			})
		}
	})
}

func TestGenerate_MonoRepo(t *testing.T) {
	ctx := t.Context()

	golang := func(godir, cmd, provider string) engine.Parser[types.Repository] {
		return func(_ context.Context, destdir string, _ *types.Repository) error {
			if err := os.MkdirAll(filepath.Join(destdir, godir), files.RwxRxRxRx); err != nil {
				return fmt.Errorf("mkdir all: %w", err)
			}

			gomod := fmt.Appendf(nil, "module %s.com/kickr-dev/kickr\n\ngo 1.23\n", provider)
			if err := os.WriteFile(filepath.Join(destdir, godir, parser.FileGomod), gomod, files.RwRR); err != nil {
				return fmt.Errorf("write file: %w", err)
			}

			cmd := filepath.Join(destdir, godir, parser.FolderCMD, cmd)
			if err := os.MkdirAll(cmd, files.RwxRxRxRx); err != nil {
				return fmt.Errorf("mkdir all: %w", err)
			}
			file, err := os.Create(filepath.Join(cmd, parser.FileMain))
			if err != nil {
				return fmt.Errorf("create: %w", err)
			}
			return file.Close()
		}
	}

	hugo := func(hugodir string) engine.Parser[types.Repository] {
		return func(_ context.Context, destdir string, _ *types.Repository) error {
			if err := os.MkdirAll(filepath.Join(destdir, hugodir), files.RwxRxRxRx); err != nil {
				return fmt.Errorf("mkdir all: %w", err)
			}
			file, err := os.Create(filepath.Join(destdir, hugodir, "hugo.toml"))
			if err != nil {
				return fmt.Errorf("create: %w", err)
			}
			return file.Close()
		}
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

	terraform := func(subdir string) func(ctx context.Context, destdir string, repo *types.Repository) error {
		return func(_ context.Context, destdir string, _ *types.Repository) error {
			if err := os.MkdirAll(filepath.Join(destdir, subdir), files.RwxRxRxRx); err != nil {
				return fmt.Errorf("mkdir all: %w", err)
			}
			return os.WriteFile(filepath.Join(destdir, subdir, "main.tf"), []byte(`variable "my_var" {}`+"\n"), files.RwRR)
		}
	}

	gowork := func(uses ...string) engine.Parser[types.Repository] {
		return func(_ context.Context, destdir string, _ *types.Repository) error {
			content := fmt.Sprintf("go 1.23\nuse (\n\t%s\n)\n", strings.Join(uses, "\n\t"))
			return os.WriteFile(filepath.Join(destdir, parser.FileGowork), []byte(content), files.RwRR)
		}
	}

	t.Run("go_node_doc", func(t *testing.T) {
		cases := []testcase{
			{Name: "github_netlify", Config: kickr.Kickr{GitHub: &kickr.GitHub{}}},
			{Name: "github_pages", Config: kickr.Kickr{GitHub: &kickr.GitHub{}}},
			{Name: "gitlab_netlify", Config: kickr.Kickr{GitLab: &kickr.GitLab{}}},
			{Name: "gitlab_pages", Config: kickr.Kickr{GitLab: &kickr.GitLab{}}},
		}
		for _, tc := range cases {
			t.Run(tc.Name, func(t *testing.T) {
				// Arrange
				provider, target, _ := strings.Cut(tc.Name, "_")
				repo := types.Repository{
					Config: merge(t, kickr.Kickr{
						Docker:  &kickr.Docker{},
						Exclude: []string{kickr.ExcludeMakefile, kickr.ExcludePreCommit, kickr.ExcludeRenovate},
						Modules: []kickr.Module{
							{Deployment: &kickr.Deployment{Target: target}, Exclude: []string{kickr.ModuleExcludeDocker}, Path: "docs"},
						},
					}, tc.Config),
				}

				// Act & Assert
				test(ctx, t, repo, golang(types.RootModule, "api", provider), node("docs"))
			})
		}
	})

	t.Run("go_hugo_doc", func(t *testing.T) {
		cases := []testcase{
			{Name: "github_netlify", Config: kickr.Kickr{GitHub: &kickr.GitHub{}}},
			{Name: "github_pages", Config: kickr.Kickr{GitHub: &kickr.GitHub{}}},
			{Name: "gitlab_netlify", Config: kickr.Kickr{GitLab: &kickr.GitLab{}}},
			{Name: "gitlab_pages", Config: kickr.Kickr{GitLab: &kickr.GitLab{}}},
		}
		for _, tc := range cases {
			t.Run(tc.Name, func(t *testing.T) {
				// Arrange
				provider, target, _ := strings.Cut(tc.Name, "_")
				repo := types.Repository{
					Config: merge(t, kickr.Kickr{
						Docker:  &kickr.Docker{},
						Exclude: []string{kickr.ExcludePreCommit, kickr.ExcludeRenovate},
						Modules: []kickr.Module{
							{Deployment: &kickr.Deployment{Target: target}, Exclude: []string{kickr.ModuleExcludeDocker}, Path: "docs"},
						},
					}, tc.Config),
				}

				// Act & Assert
				test(ctx, t, repo, golang(types.RootModule, "api", provider), hugo("docs"))
			})
		}
	})

	t.Run("node_hugo_doc", func(t *testing.T) {
		cases := []testcase{
			{Name: "github_netlify", Config: kickr.Kickr{GitHub: &kickr.GitHub{}}},
			{Name: "github_pages", Config: kickr.Kickr{GitHub: &kickr.GitHub{}}},
			{Name: "gitlab_netlify", Config: kickr.Kickr{GitLab: &kickr.GitLab{}}},
			{Name: "gitlab_pages", Config: kickr.Kickr{GitLab: &kickr.GitLab{}}},
		}
		for _, tc := range cases {
			t.Run(tc.Name, func(t *testing.T) {
				// Arrange
				_, target, _ := strings.Cut(tc.Name, "_")
				repo := types.Repository{
					Config: merge(t, kickr.Kickr{
						Exclude: []string{kickr.ExcludePreCommit, kickr.ExcludeRenovate},
						Modules: []kickr.Module{
							{Deployment: &kickr.Deployment{Target: target}, Path: "docs"},
							{Deployment: &kickr.Deployment{Target: target}, Path: "website"},
						},
					}, tc.Config),
				}

				// Act & Assert
				test(ctx, t, repo, node(types.RootModule), hugo("docs"), hugo("website"))
			})
		}
	})

	t.Run("frontend_backend_docker", func(t *testing.T) {
		cases := []testcase{
			{Name: "github", Config: kickr.Kickr{GitHub: &kickr.GitHub{}}},
			{Name: "gitlab", Config: kickr.Kickr{GitLab: &kickr.GitLab{}}},
		}
		for _, tc := range cases {
			t.Run(tc.Name, func(t *testing.T) {
				// Arrange
				repo := types.Repository{
					Config: merge(t, kickr.Kickr{
						Docker:  &kickr.Docker{},
						Exclude: []string{kickr.ExcludePreCommit, kickr.ExcludeRenovate},
						Modules: []kickr.Module{{Path: "backend"}, {Path: "frontend"}},
					}, tc.Config),
				}

				// Act & Assert
				test(ctx, t, repo, gowork("./backend"), golang("backend", "api", tc.Name), node("frontend"))
			})
		}
	})

	t.Run("go_self_terraform", func(t *testing.T) {
		cases := []testcase{
			{Name: "github", Config: kickr.Kickr{GitHub: &kickr.GitHub{}}},
			{Name: "gitlab", Config: kickr.Kickr{GitLab: &kickr.GitLab{}}},
		}
		for _, tc := range cases {
			t.Run(tc.Name, func(t *testing.T) {
				// Arrange
				repo := types.Repository{
					Config: merge(t, kickr.Kickr{
						Exclude: []string{kickr.ExcludeMakefile, kickr.ExcludePreCommit},
						Modules: []kickr.Module{
							{Path: ".terraform", Terraform: &kickr.Terraform{Engine: kickr.TerraformEngineTofu}},
						},
					}, tc.Config),
				}

				// Act & Assert
				test(ctx, t, repo, golang(types.RootModule, "api", tc.Name), terraform(".terraform"))
			})
		}
	})
}

func TestGenerate_MultiPlatforms(t *testing.T) {
	ctx := t.Context()

	t.Run("multiple_ci", func(t *testing.T) {
		cases := []testcase{
			{
				Name: "github_primary",
				Config: kickr.Kickr{
					GitLab: &kickr.GitLab{Exclude: []string{kickr.GitLabExcludePreCommit}},
					GitHub: &kickr.GitHub{Release: &kickr.Release{}},
				},
			},
			{
				Name: "gitlab_primary",
				Config: kickr.Kickr{
					GitLab: &kickr.GitLab{Release: &kickr.Release{}},
					GitHub: &kickr.GitHub{Exclude: []string{kickr.GitHubExcludePreCommit}},
				},
			},
		}
		for _, tc := range cases {
			t.Run(tc.Name, func(t *testing.T) {
				// Arrange
				repo := types.Repository{
					Config: merge(t, kickr.Kickr{
						GitHub: &kickr.GitHub{},
						GitLab: &kickr.GitLab{},
						PreCommit: []string{
							kickr.PreCommitAutoCommit,
							kickr.PreCommitConventionalCommits,
							kickr.PreCommitGitflowBranches,
							kickr.PreCommitGolangciLint,
							kickr.PreCommitGomodTidy,
							kickr.PreCommitTerraform,
						},
					}, tc.Config),
				}

				// Act & Assert
				test(ctx, t, repo)
			})
		}
	})
}

func ParserInfo(_ context.Context, _ string, repo *types.Repository) error {
	repo.VCS = parser.VCS{
		Platform:    repo.Config.Platform,
		ProjectHost: repo.Config.Platform + ".com",
		ProjectName: "kickr",
		ProjectPath: "kickr-dev/kickr",
	}
	return nil
}

// merge combines the two input kickr configurations.
//
// Should only be used to avoid conditions between CICD provides
// and not as a general use case.
func merge(t testing.TB, base, complement kickr.Kickr) kickr.Kickr {
	t.Helper()

	require.NoError(t, mergo.Merge(&base, complement, mergo.WithAppendSlice))
	if base.Platform == "" {
		if base.GitHub != nil {
			base.Platform = parser.GitHub
		}
		if base.GitLab != nil {
			base.Platform = parser.GitLab
		}
	}
	return base
}

// test verifies every generation with provided repo, parser and t.Name folder expected results.
func test(ctx context.Context, t *testing.T, repo types.Repository, parsers ...engine.Parser[types.Repository]) {
	t.Helper()

	// Arrange
	repo.Config.Maintainers = append(repo.Config.Maintainers, &kickr.Maintainer{Name: "kilianpaquier"})
	assertdir := filepath.Join(testutils.Testdata(t), t.Name())
	require.NoError(t, os.MkdirAll(assertdir, files.RwxRxRxRx))

	destdir := t.TempDir()
	if ok, _ := strconv.ParseBool(os.Getenv("TESTDATA")); ok {
		destdir = assertdir
	}

	// Act
	err := engine.Generate(ctx, destdir, repo,
		slices.Concat(parsers, []engine.Parser[types.Repository]{
			// must be kept first since it parses Git informations (useful for next parsers)
			// generate.ParserGit,
			ParserInfo,
			generate.ParserModules,

			generate.ParserGlob,
			generate.ParserHugo,
			generate.ParserGolang,
			generate.ParserNode,
			generate.ParserTerraform,

			// must be kept last since it marshals repo and merges it with chart overrides
			generate.ParserHelm,
		}),
		[]engine.Generator[types.Repository]{
			engine.GeneratorModules(templates.FS(), types.RepositoryModules, templates.Docker()),   // module docker
			engine.GeneratorModules(templates.FS(), types.RepositoryModules, templates.Golang()),   // module golang
			engine.GeneratorModules(templates.FS(), types.RepositoryModules, templates.Makefile()), // module makefile

			engine.GeneratorTemplates(templates.FS(), slices.Concat(templates.CodeCov(), templates.Sonar())),                              // coverage
			engine.GeneratorTemplates(templates.FS(), slices.Concat(templates.GitHub(), templates.GitLab(), templates.SemanticRelease())), // ci
			engine.GeneratorTemplates(templates.FS(), templates.Chart()),                                                                  // chart
			engine.GeneratorTemplates(templates.FS(), templates.RepositoryGolang()),                                                       // golang
			engine.GeneratorTemplates(templates.FS(), templates.RepositoryMakefile()),                                                     // makefile
			engine.GeneratorTemplates(templates.FS(), templates.Misc()),                                                                   // misc
			engine.GeneratorTemplates(templates.FS(), templates.Renovate()),                                                               // renovate
			engine.GeneratorTemplates(templates.FS(), templates.Terraform()),                                                              // terraform
		})

	// Assert
	require.NoError(t, err)
	assert.NoError(t, compare.Dirs(assertdir, destdir))
}

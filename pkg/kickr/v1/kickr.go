//nolint:revive // should be generated
package kickr

type Kickr struct {
	Version int `json:"version,omitempty" yaml:"version,omitempty"`

	// Description is a free-form text that may be used in various cases like in the Helm chart, the Dockerfile, etc.
	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	CI *CI `json:"ci,omitempty" yaml:"ci,omitempty"`

	Exclude []string `json:"exclude,omitempty" yaml:"exclude,omitempty"`

	// License can be one of the available options with GitLab license API.
	//
	// With this property, kickr will automatically download the appropriate license file
	// and save it in the project.
	//
	// Enums:
	// 	- agpl-3.0
	// 	- apache-2.0
	// 	- bsd-2-clause
	// 	- bsd-3-clause
	// 	- bsl-1.0
	// 	- cc0-1.0
	// 	- epl-2.0
	// 	- gpl-2.0
	// 	- gpl-3.0
	// 	- lgpl-2.1
	// 	- mit
	// 	- mpl-2.0
	// 	- unlicense
	License string `json:"license,omitempty" yaml:"license,omitempty"`

	// Maintainers defines the list of project maintainers.
	//
	// When using kickr, at least one maintainer must be given.
	//
	// Maintainers may be used in various places like in the Helm chart, Dockerfile, Goreleaser configuration, etc.
	Maintainers []*Maintainer `json:"maintainers,omitempty" yaml:"maintainers,omitempty"`

	// Platform defines the source code hosting platform (GitHub, GitLab, Gitea, Bitbucket).
	//
	// By default, it will be guessed based on Git remote URL.
	// However, in case of mirroring, or other specific cases (like custom hosting), it may be updated (and it won't be overridden) to provide a custom platform.
	//
	// The platform serves in various cases to generate the right files format (dependabot for instance) but also for badges in README.md.
	//
	// Enums:
	// 	- bitbucket
	// 	- gitea
	// 	- github
	// 	- gitlab
	Platform string `json:"platform,omitempty" yaml:"platform,omitempty"`

	// Pre-commit section can be provided to add additional, non-default, checks with pre-commit.
	//
	// All options available here are implemented in kickr templates, there's no possibly to directly give pre-commit hooks definition.
	//
	// In case a specific hook, not implemented, must be provided, then '.pre-commit-config.yaml' file can be modified
	// and the top comment indicating that it's generated can be removed to avoid being overridden.
	PreCommit []string `json:"pre-commit,omitempty" yaml:"pre-commit,omitempty"`

	// Terraform section can be provided to tune how terraform / opentofu templates (pre-commit, continuous integration) behaviors will be generated.
	//
	// By default, if a module is present at repository's root, then there's no need to provide the list of modules.
	// However, if current repository is composed of multiple root modules (in subdirectories),
	// then it's required to provide the full list of modules to generate templates correctly.
	//
	// The engine property can be provided to use terraform instead of opentofu (by default).
	Terraform *Terraform `json:"terraform,omitempty" yaml:"terraform,omitempty"`
}

type CI struct {
	// Provider.
	//
	// Enums:
	// 	- github
	// 	- gitlab
	Provider string `json:"provider,omitempty" yaml:"provider,omitempty"`

	// Options to add for CI/CD generation.
	//
	// Those are various and may differ depending on used CI/CD provider.
	Options []string `json:"options,omitempty" yaml:"options,omitempty"`

	Docker    *Docker      `json:"docker,omitempty"    yaml:"docker,omitempty"`
	Helm      *Helm        `json:"helm,omitempty"      yaml:"helm,omitempty"`
	Release   *Release     `json:"release,omitempty"   yaml:"release,omitempty"`
	Terraform *TerraformCI `json:"terraform,omitempty" yaml:"terraform,omitempty"`
	Website   *Website     `json:"website,omitempty"   yaml:"website,omitempty"`
}

type Renovate struct {
	// Auth corresponds to the authentication method used by Renovate when running in self-hosted mode.
	//
	// Since GitLab provides features around tokens, this property cannot be given when running with it (i.e. will always be 'personal-token').
	//
	// Enums:
	// 	- github-app
	// 	- github-token
	// 	- personal-token
	Auth string `json:"auth,omitempty" yaml:"auth,omitempty"`
}

type Maintainer struct {
	Name  string  `json:"name,omitempty"  yaml:"name,omitempty"`
	Email *string `json:"email,omitempty" yaml:"email,omitempty"`
	URL   *string `json:"url,omitempty"   yaml:"url,omitempty"`
}

// Terraform section can be provided to tune how terraform / opentofu templates (pre-commit, continuous integration) behaviors will be generated.
//
// By default, if a 'main.tf' is present at repository's root, then there's no need to provide the list of modules.
// However, if current repository is composed of multiple terraform root modules (in subdirectories),
// then it's required to provide the full list of modules to generate templates correctly.
//
// The engine property can be provided to use terraform instead of opentofu (by default).
type Terraform struct {
	// Engine.
	//
	// Enums:
	// 	- opentofu
	// 	- terraform
	Engine  string   `json:"engine,omitempty"  yaml:"engine,omitempty"`
	Modules []string `json:"modules,omitempty" yaml:"modules,omitempty"`
}

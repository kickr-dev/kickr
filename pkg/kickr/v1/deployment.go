//nolint:revive // should be generated
package kickr

type Docker struct {
	// Auto.
	//
	// Enums:
	// 	- true
	// 	- false
	Auto bool `json:"auto,omitempty" yaml:"auto,omitempty"`

	Path     string `json:"path,omitempty"     yaml:"path,omitempty"`
	Port     int    `json:"port,omitempty"     yaml:"port,omitempty"`
	Registry string `json:"registry,omitempty" yaml:"registry,omitempty"`
}

type Helm struct {
	// Deploy helm chart to kubernetes cluster(s).
	Deploy string `json:"deploy,omitempty" yaml:"deploy,omitempty"`

	// Environments to associate a specific Kubernetes context to a branch.
	//
	// Only those four environments can be provided, defining under the hood specific behaviors (for GitHub Actions):
	//  - 'integration' will only be run on protected
	//  - 'production' will only be run on the default branch
	//  - 'review' will only be run on non-protected branches
	//  - 'staging' will only be run on the default branch
	//
	// Concerning GitLab CI/CD, rules can be found at https://gitlab.com/to-be-continuous/helm#managed-deployment-environments.
	Environments []string `json:"environments,omitempty" yaml:"environments,omitempty"`

	// Path within the helm registry (default is the path on VCS remote URL).
	Path string `json:"path,omitempty" yaml:"path,omitempty"`

	// Publish helm chart to the registry.
	//
	// Enums:
	// 	- auto
	// 	- manual
	// 	- none
	Publish string `json:"publish,omitempty" yaml:"publish,omitempty"`

	// Registry URL (both OCI and non-OCI based registry are supported).
	//
	// Depending on the project registry (GitHub, Jfrog, Nexus, etc.), this property may need a more specific path like 'oci://jfrog.example.com/my/registry'.
	Registry string `json:"registry,omitempty" yaml:"registry,omitempty"`
}

type TerraformCI struct {
	// Apply strategy.
	Apply string `json:"apply,omitempty" yaml:"apply,omitempty"`

	// Environments to associate a specific terraform apply with.
	//
	// Only those four environments can be provided, defining under the hood specific behaviors (for GitHub Actions):
	//  - 'integration' will only be run on protected
	//  - 'production' will only be run on the default branch
	//  - 'review' will only be run on non-protected branches
	//  - 'staging' will only be run on the default branch
	//
	// Concerning GitLab CI/CD, rules can be found at https://gitlab.com/to-be-continuous/terraform#global-configuration.
	Environments []string `json:"environments,omitempty" yaml:"environments,omitempty"`
}

type Website struct {
	Auto bool `json:"auto,omitempty" yaml:"auto,omitempty"`

	Hosting string `json:"hosting,omitempty" yaml:"hosting,omitempty"`

	// Directory where is located the website to deploy (default is '.').
	Directory string `json:"directory,omitempty" yaml:"directory,omitempty"`
}

type Release struct {
	// Auth corresponds to the authentication method used by semantic-release when creating GitHub / GitLab releases.
	//
	// Since GitLab provides features around tokens, this property cannot be given when running with it (i.e. will always be 'personal-token').
	//
	// Enums:
	// 	- github-app
	// 	- github-token
	// 	- personal-token
	Auth string `json:"auth,omitempty" yaml:"auth,omitempty"`

	// Auto.
	//
	// Enums:
	// 	- true
	// 	- false
	Auto bool `json:"auto,omitempty" yaml:"auto,omitempty"`

	// Options to add in release job / release configuration.
	Options []string `json:"options,omitempty" yaml:"options,omitempty"`
}

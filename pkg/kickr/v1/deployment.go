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

type Netlify struct {
	// Type netlify.
	//
	// Since netlify can handle preview deployments, the production deployment is mapped to the default branch.
	// All other branches will be deployed with a subdomain corresponding to the branch sha8 (it will avoid creating many previews per branch).
	//
	// Build can be made with Nodejs or GoHugo projects.
	Auto bool `json:"auto,omitempty" yaml:"auto,omitempty"`
}

type Pages struct {
	// Type pages makes it so the project will be deployed on the 'ci.provider' Pages (GitHub, GitLab, etc.).
	//
	// Note that since pages deployment doesn't handle environments, the only deployment will be done on default branch.
	// All other branches will be ignored during Continuous Integration for that part.
	//
	// Build can be made with Nodejs or GoHugo projects.
	Auto bool `json:"auto,omitempty" yaml:"auto,omitempty"`
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

package kickr

// Just a convenient way to separate structs (target is to have them automatically generated from JSON Schem)
// from associated methods.

import (
	"cmp"
	"slices"
)

// IsHelmPublishAuto returns truthy in case the configuration has CI enabled, helm publish enabled and auto actived.
func (k Kickr) IsHelmPublishAuto() bool {
	return k.Helm != nil && k.Helm.Publish == HelmPublishAuto
}

// HasHelmPublish returns truthy in case the configuration has CI enabled, helm chart generation enabled
// and publication to a helm repository enabled.
func (k Kickr) HasHelmPublish() bool {
	return k.Helm != nil && k.Helm.Publish != ""
}

// HasHelmDeploy returns truthy in case the configuration has CI enabled, helm chart generation enabled
// and deployment to kubernetes cluster(s) enabled.
func (k Kickr) HasHelmDeploy() bool {
	return k.Helm != nil && k.Helm.Deploy != ""
}

// HasSonarQube returns truthy in case SonarQube analysis is enabled on either platform.
func (k Kickr) HasSonarQube() bool {
	if k.GitHub != nil && slices.Contains(k.GitHub.Options, GitHubOptionsSonarQube) {
		return true
	}
	if k.GitLab != nil && slices.Contains(k.GitLab.Options, GitLabOptionsSonarQube) {
		return true
	}
	return false
}

// TerraformEngine returns the terraform engine configured throughout the modules, if any.
func (k Kickr) TerraformEngine() string {
	for _, module := range k.Modules {
		if module.Terraform == nil {
			continue
		}
		if module.Terraform.Engine != "" {
			// engine is unique throughout all modules with terraform / opentofu
			// validation happens on schema end
			return module.Terraform.Engine
		}
		// engine is unique throughout all modules with terraform / opentofu
		// validation happens on schema end
		return TerraformEngineTofu
	}
	return ""
}

// HasTerraformApply returns truthy when at least one module has an apply strategy configured.
func (k Kickr) HasTerraformApply() bool {
	return slices.ContainsFunc(k.Modules, Module.HasTerraformApply)
}

// HasTerraformAutoApply returns truthy when at least one module has an apply strategy configured and set to auto.
func (k Kickr) HasTerraformAutoApply() bool {
	return slices.ContainsFunc(k.Modules, Module.HasTerraformAutoApply)
}

// HasDeploymentAuto returns truthy when at least one module has a deployment strategy configured and set to auto.
func (k Kickr) HasDeploymentAuto() bool {
	return slices.ContainsFunc(k.Modules, Module.HasDeploymentAuto)
}

// HasKickr returns truthy in case one option at least is provided for kickr auto-layout generation.
func (k Kickr) HasKickr() bool {
	if k.GitHub != nil && slices.ContainsFunc(k.GitHub.Options, func(o string) bool { return o == GitHubOptionsKickrGitHubApp || o == GitHubOptionsKickrPersonalToken }) {
		return true
	}
	if k.GitLab != nil && slices.ContainsFunc(k.GitLab.Options, func(o string) bool { return o == GitLabOptionsKickr }) {
		return true
	}
	return false
}

// EnsureDefaults migrates old properties into new fields
// and ensures default properties are always sets.
func (k *Kickr) EnsureDefaults() {
	slices.Sort(k.Exclude)
	slices.Sort(k.PreCommit)

	// sort maintainers per name
	slices.SortFunc(k.Maintainers, func(a, b *Maintainer) int {
		return cmp.Compare(a.Name, b.Name)
	})

	if k.GitHub != nil {
		slices.Sort(k.GitHub.Options)
		if k.GitHub.Release != nil {
			slices.Sort(k.GitHub.Release.Options)
		}
	}

	if k.GitLab != nil {
		slices.Sort(k.GitLab.Options)
		if k.GitLab.Release != nil {
			slices.Sort(k.GitLab.Release.Options)
		}
	}

	if k.Helm != nil {
		slices.Sort(k.Helm.Environments)
	}

	// sort modules per path
	slices.SortFunc(k.Modules, func(a, b Module) int {
		return cmp.Compare(a.Path, b.Path)
	})

	for _, module := range k.Modules {
		slices.Sort(module.Exclude)
		if module.Terraform != nil {
			slices.Sort(module.Terraform.Environments)
		}
	}
}

// HasTerraformApply returns truthy when the module has an apply strategy configured.
func (m Module) HasTerraformApply() bool {
	return m.Terraform != nil && m.Terraform.Apply != ""
}

// HasTerraformDocs returns truthy when the module needs terraform documentation.
func (m Module) HasTerraformDocs() bool {
	return m.Terraform != nil && m.Terraform.Apply == ""
}

// HasTerraformAutoApply returns truthy when the module has an apply strategy configured and set to auto.
func (m Module) HasTerraformAutoApply() bool {
	return m.Terraform != nil && m.Terraform.Apply == TerraformApplyAuto
}

// HasDeployment returns truthy when the module is deployed on the input target.
func (m Module) HasDeployment(target string) bool {
	return m.Deployment != nil && m.Deployment.Target == target
}

// HasDeploymentAuto returns truthy when the module has a deployment strategy configured and set to auto.
func (m Module) HasDeploymentAuto() bool {
	return m.Deployment != nil && m.Deployment.Auto
}

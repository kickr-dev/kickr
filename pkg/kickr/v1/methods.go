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

// IsHosting returns truthy in case the configuration has CI enabled, a website to deploy
// and the input hosting name matches the one configured.
func (k Kickr) IsHosting(hosting string) bool {
	return k.Website != nil && k.Website.Hosting == hosting
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

// HasTerraformApply returns truthy in case the configuration has CI enabled, terraform generation enabled
// and apply is enabled in terraform CI configuration.
func (k Kickr) HasTerraformApply() bool {
	return k.Terraform != nil && k.Terraform.Apply != ""
}

// HasKickr returns truthy in case one option at least is provided for kickr auto-layout generation.
func (k Kickr) HasKickr() bool {
	for _, cond := range []bool{
		k.GitLab != nil && slices.ContainsFunc(k.GitLab.Options, func(o string) bool {
			return o == GitLabOptionsKickr
		}),
		k.GitHub != nil && slices.ContainsFunc(k.GitHub.Options, func(o string) bool {
			return o == GitHubOptionsKickrGitHubApp || o == GitHubOptionsKickrPersonalToken
		}),
	} {
		if cond {
			return true
		}
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

	if k.Terraform != nil {
		slices.Sort(k.Terraform.Environments)
		slices.Sort(k.Terraform.Modules)
	}
}

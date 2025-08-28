package kickr

// Just a convenient way to separate structs (target is to have them automatically generated from JSON Schem)
// from associated methods.

import (
	"cmp"
	"slices"
)

// IsCI returns truthy in case the input provider is the one specified in configuration.
//
// It returns false if CI is disabled.
func (k Kickr) IsCI(provider string) bool {
	return k.CI != nil && k.CI.Provider == provider
}

// DependencyManager returns truthy in case the input manager matches the configuration dependencies manager.
func (k Kickr) DependencyManager(manager string) bool {
	return k.Dependencies != nil && k.Dependencies.Manager == manager
}

// HasRelease returns truthy in case the configuration has CI enabled and Release configuration.
func (k Kickr) HasRelease() bool {
	return k.CI != nil && k.CI.Release != nil
}

// IsReleaseAuto returns truthy in case the configuration has CI enabled, release enabled and auto actived.
func (k Kickr) IsReleaseAuto() bool {
	return k.CI != nil && k.CI.Release != nil && k.CI.Release.Auto
}

// IsHelmPublishAuto returns truthy in case the configuration has CI enabled, helm publish enabled and auto actived.
func (k Kickr) IsHelmPublishAuto() bool {
	return k.CI != nil && k.CI.Helm != nil && k.CI.Helm.Publish == HelmAuto
}

// HasHelmPublish returns truthy in case the configuration has CI enabled, helm chart generation enabled
// and publication to a helm repository enabled.
func (k Kickr) HasHelmPublish() bool {
	return k.CI != nil && k.CI.Helm != nil && slices.Contains([]string{HelmAuto, HelmManual}, k.CI.Helm.Publish)
}

// HasHelmDeploy returns truthy in case the configuration has CI enabled, helm chart generation enabled
// and deployment to kubernetes cluster(s) enabled.
func (k Kickr) HasHelmDeploy() bool {
	return k.CI != nil && k.CI.Helm != nil && slices.Contains([]string{HelmAuto, HelmManual}, k.CI.Helm.Deploy)
}

// HasAutoDeployment returns truthy in case the configuration has CI enabled,
// and at least one deployment section (docker, helm, netlify, pages, etc.) is in auto mode.
func (k Kickr) HasAutoDeployment() bool {
	if k.CI == nil {
		return false
	}

	docker := k.CI.Docker != nil && k.CI.Docker.Auto
	helm := k.CI.Helm != nil && (k.CI.Helm.Deploy == HelmAuto || k.CI.Helm.Publish == HelmAuto)
	netlify := k.CI.Netlify != nil && k.CI.Netlify.Auto
	pages := k.CI.Pages != nil && k.CI.Pages.Auto
	release := k.CI.Release != nil && k.CI.Release.Auto

	return docker || helm || netlify || pages || release
}

// HasDeployment returns truthy in case the configuration has CI enabled and Deployment configuration.
func (k Kickr) HasDeployment() bool {
	if k.CI == nil {
		return false
	}

	docker := k.CI.Docker != nil
	helm := k.CI.Helm != nil
	netlify := k.CI.Netlify != nil
	pages := k.CI.Pages != nil
	release := k.CI.Release != nil

	return docker || helm || netlify || pages || release
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

	if k.CI != nil {
		slices.Sort(k.CI.Options)

		if k.CI.Helm != nil {
			slices.Sort(k.CI.Helm.Environments)
		}
	}
}

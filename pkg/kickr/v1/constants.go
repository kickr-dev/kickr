//nolint:revive // should be generated
package kickr

const (
	File   = ".kickr"
	Schema = "kickr.v1.schema.json"
)

const (
	// ExcludeGoreleaser can be given in kickr exclusions ('exclude' key) to avoid generating a .goreleaser.yml file.
	//
	// By default, if a given project is a Go project,
	// and a cmd CLI is defined (cmd/<some useful CLI name>)
	// a .goreleaser.yml file is generated.
	//
	// As such, it's unnecessary to specify this property when your project isn't a Go one.
	ExcludeGoreleaser string = "goreleaser"

	// ExcludeMakefile can be given in kickr exclusions ('exclude' key) to avoid generating a ExcludeMakefile
	// and additional Makefiles in scripts/mk/*.mk.
	//
	// When generating a Node project, it's unnecessary to specify this property since no ExcludeMakefile will be generated anyway.
	// It's because Node projects contain all their scripts in package.json.
	ExcludeMakefile string = "makefile"

	// ExcludePreCommit can be given in kickr exclusions ('exclude' key) to avoid generating pre-commit files and associated Continuous Integration.
	ExcludePreCommit string = "pre-commit"

	// ExcludeShell can be given in kickr exclusions ('exclude' key)
	// to avoid generating shell (check / test / pre-commit) Continuous Integration.
	ExcludeShell string = "shell"
)

const (
	// OptionCodeCov is the codecov option for CI tuning.
	OptionCodeCov string = "codecov"
	// OptionSonarQube is the sonarqube option for CI tuning.
	OptionSonarQube string = "sonarqube"

	// OptionCodeQL is the codeql option for CI tuning.
	OptionCodeQL string = "codeql"
	// OptionHardenRunner is the CI option to ensure runners (with GitHub Actions) doesn't have too many open rights.
	OptionHardenRunner string = "harden-runner"
	// OptionLabeler is the auto labeling option for CI tuning.
	OptionLabeler string = "labeler"
	// OptionScoreCardOSSF is the CI option to add OSSF Scorecard badge and associated workflow (with GitHub Actions).
	OptionScoreCardOSSF string = "ossf-scorecard"

	// OptionBackmerge is the CI release option to backmerge stable branches between them.
	OptionBackmerge string = "backmerge"
)

const (
	// ManagerDependabot is the dependabot updater name for CI maintenance configuration.
	ManagerDependabot string = "dependabot"
	// ManagerRenovate is the renovate updater name for CI maintenance configuration.
	ManagerRenovate string = "renovate"
)

const (
	// AuthGitHubApp is the value for github release mode with a github app.
	AuthGitHubApp string = "github-app"
	// AuthGitHubToken is the value for github release mode with a github token.
	AuthGitHubToken string = "github-token"
	// AuthPersonalToken is the value for github release mode with a personal token (PAT).
	AuthPersonalToken string = "personal-token"
)

const (
	// HelmAuto is the constant indicating that Helm chart publication / deployment should be made automatically.
	HelmAuto string = "auto"
	// HelmManual is the constant indicating that Helm chart publication / deployment should be made manually.
	HelmManual string = "manual"
	// HelmNone is the constant indicating that Helm chart publication / deployment should not be made.
	HelmNone string = "none"
)

const (
	// PreCommitAutoCommit is an available pre-commit option to auto-commit issues identified by pre-commit.
	PreCommitAutoCommit = "auto-commit"

	// PreCommitGomodTidy is an available pre-commit option for pre-commit configuration file.
	//
	// It will ensure there's no lint issue before committing anything.
	PreCommitGolangciLint = "golangci-lint"

	// PreCommitGomodTidy is an available pre-commit option for pre-commit configuration file.
	//
	// It will ensure go.mod and go.sum are tidied before committing anything.
	PreCommitGomodTidy = "gomod-tidy"
)

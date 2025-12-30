package cobra

import (
	"net/http"
	"os"
	"path/filepath"
	"slices"

	engine "github.com/kickr-dev/engine/pkg"
	"github.com/kickr-dev/engine/pkg/files"
	"github.com/spf13/cobra"

	schemas "github.com/kickr-dev/kickr/.schemas"
	"github.com/kickr-dev/kickr/pkg/generate"
	"github.com/kickr-dev/kickr/pkg/generate/templates"
	"github.com/kickr-dev/kickr/pkg/generate/types"
	kickr "github.com/kickr-dev/kickr/pkg/kickr/v1"
)

var (
	force bool

	generateCmd = &cobra.Command{
		Use:     "generate",
		Aliases: []string{"g"},
		Short:   "Generate project layout",
		Run: gen(
			generate.GeneratorGitignore(http.DefaultClient), // gitignore
			generate.GeneratorLicense(http.DefaultClient),   // license

			engine.GeneratorTemplates(templates.FS(), slices.Concat(templates.CodeCov(), templates.Sonar())),                              // coverage
			engine.GeneratorTemplates(templates.FS(), slices.Concat(templates.Dependabot(), templates.Renovate())),                        // bot
			engine.GeneratorTemplates(templates.FS(), slices.Concat(templates.GitHub(), templates.GitLab(), templates.SemanticRelease())), // ci
			engine.GeneratorTemplates(templates.FS(), templates.Chart()),                                                                  // chart
			engine.GeneratorTemplates(templates.FS(), templates.Docker()),                                                                 // docker
			engine.GeneratorTemplates(templates.FS(), templates.Golang()),                                                                 // golang
			engine.GeneratorTemplates(templates.FS(), templates.Makefile()),                                                               // makefile
			engine.GeneratorTemplates(templates.FS(), templates.Misc()),                                                                   // misc
			engine.GeneratorTemplates(templates.FS(), templates.Terraform()),                                                              // terraform
		),
	}
)

func init() {
	rootCmd.AddCommand(generateCmd)

	generateCmd.PersistentFlags().BoolVarP(&force, "force", "f", false,
		"force generation of all files initially created by kickr (README.md, SECURITY.md, etc.) even if the initial generated notice has been removed")
}

func gen(generators ...engine.Generator[types.Repository]) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		dest := filepath.Join(wd, kickr.File)
		logger.Infof("generating layout in %s", wd)

		// initialize configuration if it does not exist
		if !files.Exists(dest) {
			initializeCmd.Run(cmd, args) // will fatal if initialization fails
		}

		// validate configuration
		if err := files.Validate(
			func(out any) error { return files.ReadYAML(kickr.Schema, out, schemas.ReadFile) }, // read schema
			func(out any) error { return files.ReadYAML(dest, out, os.ReadFile) },              // read configuration
		); err != nil {
			logger.Fatal(err)
		}

		// read configuration
		var config kickr.Kickr
		if err := files.ReadYAML(dest, &config, os.ReadFile); err != nil {
			logger.Fatal(err)
		}
		config.EnsureDefaults()

		// run generation
		engine.Configure(engine.WithForce(force), engine.WithLogger(logger))
		parsers := []engine.Parser[types.Repository]{
			// must be kept first since it parses Git informations (useful for next parsers)
			generate.ParserGit,

			generate.ParserGlob,
			generate.ParserGolang,
			generate.ParserNode,
			generate.ParserTerraform,

			// must be kept last since it marshals config and merges it with chart overrides
			generate.ParserHelm,
		}

		result, err := engine.Generate(ctx, wd, types.Repository{Kickr: config}, parsers, generators)
		if err != nil {
			logger.Fatal(err)
		}

		// save configuration again in case it was modified during generation
		if err := files.WriteYAML(dest, result.Kickr, kickr.EncodeOpts()...); err != nil {
			logger.Fatal(err)
		}
	}
}

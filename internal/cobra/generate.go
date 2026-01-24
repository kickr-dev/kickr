package cobra

import (
	"fmt"
	"net/http"
	"os"
	"slices"
	"strconv"

	engine "github.com/kickr-dev/engine/pkg"
	"github.com/kickr-dev/engine/pkg/files"
	"github.com/spf13/cobra"

	schemas "github.com/kickr-dev/kickr/.schemas"
	"github.com/kickr-dev/kickr/pkg/generate"
	"github.com/kickr-dev/kickr/pkg/generate/templates"
	"github.com/kickr-dev/kickr/pkg/generate/types"
	kickr "github.com/kickr-dev/kickr/pkg/kickr/v1"
)

const envPrefix = "kickr-"

const (
	flagForce      = "force"
	flagShortForce = "f"
)

func generators() []engine.Generator[types.Repository] {
	return []engine.Generator[types.Repository]{
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
	}
}

func generateCmd(wd *string, generators ...engine.Generator[types.Repository]) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate project layout",
		Args: func(cmd *cobra.Command, _ []string) error {
			// validate force environment variable
			if !cmd.Flags().Changed(flagForce) {
				if env := getenv(envPrefix + flagForce); env != "" {
					ff, err := strconv.ParseBool(env)
					if err != nil {
						return fmt.Errorf(`invalid argument %q for "--%s" flag: %w`, env, flagForce, err)
					}
					force = ff
				}
			}
			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			dest := kickr.File(*wd)
			if dest != "" {
				return nil
			}
			return initializeCmd(wd).RunE(cmd, args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			dest := kickr.File(*wd)
			logger.Infof("generating layout in %s", *wd)

			// validate configuration
			if err := files.Validate(
				func(out any) error { return files.ReadYAML(kickr.Schema, out, schemas.ReadFile) }, // read schema
				func(out any) error { return files.ReadYAML(dest, out, os.ReadFile) },              // read configuration
			); err != nil {
				return err
			}

			// read configuration
			var config kickr.Kickr
			if err := files.ReadYAML(dest, &config, os.ReadFile); err != nil {
				return err
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

			result, err := engine.Generate(cmd.Context(), *wd, types.Repository{Kickr: config}, parsers, generators)
			if err != nil {
				return err
			}

			// save configuration again in case it was modified during generation
			if err := files.WriteYAML(dest, result.Kickr, kickr.EncodeOpts()...); err != nil {
				return err
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, flagForce, flagShortForce, false,
		"force generation of all files initially created by kickr (README.md, SECURITY.md, etc.) even if the initial generated notice has been removed")

	return cmd
}

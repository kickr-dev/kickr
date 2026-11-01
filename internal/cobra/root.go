package cobra

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	engine "gitlab.com/kickr-dev/engine/pkg"

	"github.com/kickr-dev/kickr/pkg/generate"
	kickr "github.com/kickr-dev/kickr/pkg/kickr/v1"
)

const groupStandalone = "standalone"

const (
	flagDir        = "dir"
	flagShortDir   = "d"
	flagLogFormat  = "log-format"
	flagLogLevel   = "log-level"
	flagForce      = "force"
	flagShortForce = "f"
)

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	var wd string

	cmd := rootCmd(&wd)
	cmd.AddGroup(&cobra.Group{ID: groupStandalone, Title: "Standalone generation:"})
	cmd.AddCommand(chartCmd(&wd), dockerCmd(&wd), gitignoreCmd(&wd), gitlabCmd(&wd), golangciCmd(&wd), licenseCmd(&wd), plumberCmd(&wd))
	cmd.AddCommand(initializeCmd(&wd))
	cmd.AddCommand(version())
	cmd.AddCommand(generateCmd(&wd, generate.Generators(httpClient())...))

	if err := cmd.Execute(); err != nil {
		subcmd, _, _ := cmd.Find(os.Args[1:])
		usage(cmd, subcmd, err)
		logger.Fatal(err.Error())
	}
}

func rootCmd(wd *string) *cobra.Command {
	logFormat, logLevel := "text", "info"
	var force bool

	gen := generateCmd(wd, generate.Generators(httpClient())...)

	cmd := &cobra.Command{
		Use: filepath.Base(os.Args[0]),
		Long: `Kickr generates and maintains a repository layout (CI/CD, Helm chart, Docker, linters, community files, etc.)
from a kickr configuration and what it detects in the repository.

Run 'kickr' or 'kickr generate' to generate the whole layout, initializing the kickr configuration when missing.
Run a standalone generation command (e.g. 'kickr chart' or 'kickr gitlab') to generate only that part, with or without kickr configuration.
Without kickr configuration, standalone generations add no generated notice and regenerate their files on every run.`,
		SilenceErrors: true, // don't print errors with cobra, let logger.Fatal handle them
		SilenceUsage:  true, // don't print help on errors, let usage function handle printing depending on command error

		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			for _, err := range []error{setupLogger(logFormat, logLevel), setupForce(cmd, &force), setupWorkingDir(wd)} {
				if err != nil {
					return err
				}
			}
			// without kickr configuration, standalone generations don't own their files, so no notice and always regenerated
			unowned := cmd.GroupID == groupStandalone && kickr.File(*wd) == ""

			engine.SetLogger(logger)
			cmd.SetContext(engine.ContextWithOptions(cmd.Context(), engine.Options{Force: force || unowned, StripMarkers: unowned}))
			return nil
		},

		// defaulting command to generate
		Args:    gen.Args,
		PreRunE: gen.PreRunE,
		RunE:    gen.RunE,
	}

	cmd.PersistentFlags().StringVarP(wd, flagDir, flagShortDir, coalesce(getenv(envPrefix+"working-"+flagDir), getenv(envPrefix+flagDir)),
		"set directory where generation will be made (default is current directory)")
	cmd.PersistentFlags().StringVar(&logFormat, flagLogFormat, coalesce(getenv(flagLogFormat), logFormat), `set logging format (either "text" or "json")`)
	cmd.PersistentFlags().StringVar(&logLevel, flagLogLevel, coalesce(getenv(flagLogLevel), logLevel), "set logging level")
	cmd.PersistentFlags().BoolVarP(&force, flagForce, flagShortForce, false,
		"force generation of all files initially created by kickr (README.md, SECURITY.md, etc.) even if the initial generated notice has been removed")

	_ = setupLogger(logFormat, logLevel) // ensure logging is correctly configured with default values even when a bad input flag is given

	return cmd
}

func setupWorkingDir(wd *string) error {
	if wd == nil || *wd == "" {
		pwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get wd: %w", err)
		}
		*wd = pwd
		return nil
	}

	abs, err := filepath.Abs(*wd)
	if err != nil {
		return fmt.Errorf("absolute path: %w", err)
	}
	*wd = abs
	return nil
}

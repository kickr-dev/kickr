package cobra

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

const (
	flagDir       = "dir"
	flagShortDir  = "d"
	flagLogFormat = "log-format"
	flagLogLevel  = "log-level"
)

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	var wd string

	cmd := rootCmd(&wd)
	cmd.AddCommand(initializeCmd(&wd))
	cmd.AddCommand(version())
	cmd.AddCommand(generateCmd(&wd, generators()...))

	if err := cmd.Execute(); err != nil {
		subcmd, _, _ := cmd.Find(os.Args[1:])
		usage(cmd, subcmd, err)
		logger.Fatal(err.Error())
	}
}

func rootCmd(wd *string) *cobra.Command {
	logFormat, logLevel := "text", "info"

	generate := generateCmd(wd, generators()...)

	cmd := &cobra.Command{
		Use: "kickr",
		Long: `Kickr initializes or generates kickr projects. Kickr projects are only defined by a .kickr file
and multiple files automatically generated to avoid multiple hours to setup Continuous Integration, coverage, security analyzes, helm chart, etc.

Kickr generation can be done with 'kickr' command or 'kickr generate' command.`,
		SilenceErrors: true, // don't print errors with cobra, let logger.Fatal handle them
		SilenceUsage:  true, // don't print help on errors, let usage function handle printing depending on command error
		PersistentPreRunE: func(*cobra.Command, []string) error {
			if err := setupLogger(logFormat, logLevel); err != nil {
				return err
			}
			return setupWorkingDir(wd)
		},

		// defaulting command to generate
		Args:    generate.Args,
		PreRunE: generate.PreRunE,
		RunE:    generate.RunE,
	}

	cmd.Flags().AddFlagSet(generate.Flags())

	cmd.PersistentFlags().StringVarP(wd, flagDir, flagShortDir, coalesce(getenv(envPrefix+"working-"+flagDir), getenv(envPrefix+flagDir)),
		"set directory where generation will be made (default is current directory)")
	cmd.PersistentFlags().StringVar(&logFormat, flagLogFormat, coalesce(getenv(flagLogFormat), logFormat), `set logging format (either "text" or "json")`)
	cmd.PersistentFlags().StringVar(&logLevel, flagLogLevel, coalesce(getenv(flagLogLevel), logLevel), "set logging level")

	_ = cmd.PersistentPreRunE(nil, nil) // ensure logging is correctly configured with default values even when a bad input flag is given

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

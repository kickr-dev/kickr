package cobra

import (
	"github.com/spf13/cobra"

	"github.com/kickr-dev/kickr/internal/build"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show current kickr version",
	Run:   func(_ *cobra.Command, _ []string) { logger.Info(build.GetInfo()) },
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

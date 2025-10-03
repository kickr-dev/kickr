package cobra

import (
	"path/filepath"

	engine "github.com/kickr-dev/engine/pkg"
	"github.com/kickr-dev/engine/pkg/files"
	"github.com/spf13/cobra"

	"github.com/kickr-dev/kickr/pkg/initialize"
	kickr "github.com/kickr-dev/kickr/pkg/kickr/v1"
)

var initializeCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize new kickr project",
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		dest := filepath.Join(wd, kickr.File)

		if files.Exists(dest) {
			logger.Info("project already initialized")
			return
		}

		config, err := engine.Initialize(ctx, engine.WithFormGroups(initialize.Maintainer, initialize.License, initialize.Defaults))
		if err != nil {
			logger.Fatal(err)
		}

		if err := files.WriteYAML(dest, config, kickr.EncodeOpts()...); err != nil {
			logger.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(initializeCmd)
}

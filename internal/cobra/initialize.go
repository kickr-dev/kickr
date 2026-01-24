package cobra

import (
	engine "github.com/kickr-dev/engine/pkg"
	"github.com/kickr-dev/engine/pkg/files"
	"github.com/spf13/cobra"

	"github.com/kickr-dev/kickr/pkg/initialize"
	kickr "github.com/kickr-dev/kickr/pkg/kickr/v1"
)

func initializeCmd(wd *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize new kickr project",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if dest := kickr.File(*wd); dest != "" {
				logger.Info("project already initialized")
				return nil
			}
			dest := kickr.Files()[0]

			config, err := engine.Initialize(cmd.Context(), engine.WithFormGroups(initialize.Maintainer, initialize.License, initialize.Defaults))
			if err != nil {
				return err
			}

			if err := files.WriteYAML(dest, config, kickr.EncodeOpts()...); err != nil {
				return err
			}
			return nil
		},
	}
	return cmd
}

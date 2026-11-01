package cobra

import (
	"path/filepath"

	"github.com/spf13/cobra"
	engine "gitlab.com/kickr-dev/engine/pkg"
	"gitlab.com/kickr-dev/engine/pkg/files"

	"github.com/kickr-dev/kickr/pkg/generate"
	"github.com/kickr-dev/kickr/pkg/generate/types"
	kickr "github.com/kickr-dev/kickr/pkg/kickr/v1"
)

const envPrefix = "kickr-"

func generateCmd(wd *string, generators ...engine.Generator[types.Repository]) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate project layout",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			dest := kickr.File(*wd)
			if dest != "" {
				return nil
			}
			return initializeCmd(wd).RunE(cmd, args)
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			logger.Infof("generating layout in %s", *wd)

			config, err := kickr.Read(*wd)
			if err != nil {
				return err
			}
			if err := engine.Generate(cmd.Context(), *wd, types.Repository{Config: config}, generate.Parsers(httpClient()), generators); err != nil {
				return err
			}

			// save configuration again in case it was modified during generation
			return files.WriteYAML(filepath.Join(*wd, kickr.File(*wd)), config, kickr.EncodeOpts()...)
		},
	}
	return cmd
}

package cobra //nolint:testpackage

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootFlags(t *testing.T) {
	norun := func(cmd *cobra.Command) *cobra.Command {
		cmd.PreRunE = func(*cobra.Command, []string) error {
			return nil
		}
		cmd.RunE = func(*cobra.Command, []string) error {
			return nil
		}
		return cmd
	}

	t.Run("from_env", func(t *testing.T) {
		// Arrange
		t.Setenv("LOG_LEVEL", "debug")
		t.Setenv("LOG_FORMAT", "json")
		t.Setenv("KICKR_WORKING_DIR", "/working-dir")

		cmd := norun(rootCmd(new(string)))

		// Act
		err := cmd.ExecuteContext(t.Context())

		// Assert
		require.NoError(t, err)

		format, err := cmd.PersistentFlags().GetString(flagLogFormat)
		require.NoError(t, err)
		assert.Equal(t, "json", format)

		level, err := cmd.PersistentFlags().GetString(flagLogLevel)
		require.NoError(t, err)
		assert.Equal(t, "debug", level)

		dir, err := cmd.PersistentFlags().GetString(flagDir)
		require.NoError(t, err)
		assert.Equal(t, "/working-dir", dir)
	})

	t.Run("from_flags", func(t *testing.T) {
		// Arrange
		cmd := norun(rootCmd(new(string)))
		cmd.SetArgs([]string{"--" + flagLogFormat, "json", "--" + flagLogLevel, "debug", "--" + flagDir, "/working-dir"})

		// Act
		err := cmd.ExecuteContext(t.Context())

		// Assert
		require.NoError(t, err)

		format, err := cmd.PersistentFlags().GetString(flagLogFormat)
		require.NoError(t, err)
		assert.Equal(t, "json", format)

		level, err := cmd.PersistentFlags().GetString(flagLogLevel)
		require.NoError(t, err)
		assert.Equal(t, "debug", level)

		dir, err := cmd.PersistentFlags().GetString(flagDir)
		require.NoError(t, err)
		assert.Equal(t, "/working-dir", dir)
	})
}

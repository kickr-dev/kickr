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

	t.Run("log_format_invalid", func(t *testing.T) {
		// Arrange
		cmd := norun(rootCmd(new(string)))
		cmd.SetArgs([]string{"--" + flagLogFormat, "invalid"})

		// Act
		err := cmd.ExecuteContext(t.Context())

		// Assert
		assert.ErrorContains(t, err, `invalid argument "invalid" for "--log-format" flag`)
	})

	t.Run("log_level_invalid", func(t *testing.T) {
		// Arrange
		cmd := norun(rootCmd(new(string)))
		cmd.SetArgs([]string{"--" + flagLogLevel, "invalid"})

		// Act
		err := cmd.ExecuteContext(t.Context())

		// Assert
		assert.ErrorContains(t, err, `invalid argument "invalid" for "--log-level" flag`)
	})

	t.Run("force_invalid_env", func(t *testing.T) {
		// Arrange
		t.Setenv("KICKR_FORCE", "invalid")

		cmd := norun(rootCmd(new(string)))

		// Act
		err := cmd.ExecuteContext(t.Context())

		// Assert
		assert.ErrorContains(t, err, `invalid argument "invalid"`)
	})

	t.Run("force_from_env", func(t *testing.T) {
		// Arrange
		t.Setenv("KICKR_FORCE", "true")

		cmd := norun(rootCmd(new(string)))

		// Act
		err := cmd.ExecuteContext(t.Context())

		// Assert
		require.NoError(t, err)

		force, err := cmd.PersistentFlags().GetBool(flagForce)
		require.NoError(t, err)
		assert.True(t, force)
	})

	t.Run("force_flags_override_env", func(t *testing.T) {
		// Arrange
		t.Setenv("KICKR_FORCE", "false")

		cmd := norun(rootCmd(new(string)))
		cmd.SetArgs([]string{"--" + flagForce})

		// Act
		err := cmd.ExecuteContext(t.Context())

		// Assert
		require.NoError(t, err)

		force, err := cmd.PersistentFlags().GetBool(flagForce)
		require.NoError(t, err)
		assert.True(t, force)
	})
}

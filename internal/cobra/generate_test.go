package cobra //nolint:testpackage

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateFlags(t *testing.T) {
	norun := func(cmd *cobra.Command) *cobra.Command {
		cmd.PreRunE = func(*cobra.Command, []string) error {
			return nil
		}
		cmd.RunE = func(*cobra.Command, []string) error {
			return nil
		}
		return cmd
	}

	t.Run("invalid_env", func(t *testing.T) {
		// Arrange
		t.Setenv("KICKR_FORCE", "invalid")

		cmd := norun(generateCmd(new(string), generators()...))

		// Act
		err := cmd.ExecuteContext(t.Context())

		// Assert
		assert.ErrorContains(t, err, `invalid argument "invalid"`)
	})

	t.Run("from_env", func(t *testing.T) {
		// Arrange
		t.Setenv("KICKR_FORCE", "true")

		cmd := norun(generateCmd(new(string), generators()...))

		// Act
		err := cmd.ExecuteContext(t.Context())

		// Assert
		require.NoError(t, err)

		force, err := cmd.Flags().GetBool(flagForce)
		require.NoError(t, err)
		assert.True(t, force)
	})

	t.Run("flags_override_env", func(t *testing.T) {
		// Arrange
		t.Setenv("KICKR_FORCE", "false")

		cmd := norun(generateCmd(new(string), generators()...))
		cmd.SetArgs([]string{"--force"})

		// Act
		err := cmd.ExecuteContext(t.Context())

		// Assert
		require.NoError(t, err)

		force, err := cmd.Flags().GetBool(flagForce)
		require.NoError(t, err)
		assert.True(t, force)
	})
}

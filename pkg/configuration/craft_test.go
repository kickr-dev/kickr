package kickr_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	kickr "github.com/kickr-dev/kickr/pkg/configuration"
)

func TestEnsureDefaults(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Arrange
		config := kickr.Config{
			CI: &kickr.CI{
				Options: []string{"c", "b", "a"},
			},
		}

		// Act
		config.EnsureDefaults()

		// Assert
		assert.Equal(t, []string{"a", "b", "c"}, config.CI.Options)
	})
}

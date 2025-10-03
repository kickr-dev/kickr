package initialize_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kickr-dev/kickr/pkg/initialize"
	"github.com/kickr-dev/kickr/pkg/kickr/v1"
)

func TestDefaults(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Arrange
		config := kickr.Kickr{}

		// Act
		group := initialize.Defaults(&config)

		// Assert
		assert.Nil(t, group)
		assert.Equal(t, 1, config.Version)
	})
}

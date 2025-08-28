package initialize_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kickr-dev/kickr/pkg/initialize"
	kickr "github.com/kickr-dev/kickr/pkg/kickr/v1"
)

func TestLicense(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Act
		group := initialize.License(&kickr.Kickr{})

		// Assert
		content := group.Content()
		assert.Contains(t, content, "Would you like to specify a license (optional) ?")
		assert.Contains(t, content, "Which one ?")
	})
}

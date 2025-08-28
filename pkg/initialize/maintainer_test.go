package initialize_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kickr-dev/kickr/pkg/initialize"
	kickr "github.com/kickr-dev/kickr/pkg/kickr/v1"
)

func TestMaintainer(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Act
		group := initialize.Maintainer(&kickr.Kickr{})

		// Assert
		content := group.Content()
		assert.Contains(t, content, "What's the maintainer name (required) ?")
		assert.Contains(t, content, "What's the maintainer mail (optional) ?")
		assert.Contains(t, content, "What's the maintainer url (optional) ?")
	})
}

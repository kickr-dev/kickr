package initialize_test

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	engine "github.com/kickr-dev/engine/pkg"
	"github.com/stretchr/testify/require"

	"github.com/kickr-dev/kickr/pkg/initialize"
	kickr "github.com/kickr-dev/kickr/pkg/kickr/v1"
)

func TestLicense(t *testing.T) {
	ctx := t.Context()

	t.Run("success_none", func(t *testing.T) {
		// Arrange
		reader := NewPacedReader(selectSubmit, 2*time.Millisecond)

		// Act
		config, err := engine.Initialize(ctx,
			engine.WithFormGroups(initialize.License),
			engine.WithTeaOptions[kickr.Kickr](tea.WithInput(reader)))

		// Assert
		require.NoError(t, err)
		require.Empty(t, config.License)
	})

	t.Run("success_picked", func(t *testing.T) {
		// Arrange
		reader := NewPacedReader(arrowDown+selectSubmit, 2*time.Millisecond)

		// Act
		config, err := engine.Initialize(ctx,
			engine.WithFormGroups(initialize.License),
			engine.WithTeaOptions[kickr.Kickr](tea.WithInput(reader)))

		// Assert
		require.NoError(t, err)
		require.Equal(t, "agpl-3.0", config.License)
	})
}

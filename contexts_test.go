package process

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthWithAndFromContext(t *testing.T) {
	t.Run("from context without value", func(t *testing.T) {
		ctx := context.Background()
		assert.Nil(t, HealthFromContext(ctx))
	})
	t.Run("from context that has value", func(t *testing.T) {
		h := NewHealth()

		svcKey := "some-service"

		status, err := h.Register(svcKey)
		require.NoError(t, err)
		assert.NotNil(t, status)
		status.Update(true)

		ctx := context.Background()
		ctx = ContextWithHealth(ctx, h)

		ctxHealth := HealthFromContext(ctx)
		require.NotNil(t, ctxHealth)

		ctxStatus, ok := ctxHealth.Get(svcKey)
		require.True(t, ok)
		assert.True(t, ctxStatus.Healthy())
	})
}

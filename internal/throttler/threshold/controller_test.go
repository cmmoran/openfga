package threshold

import (
	"context"
	"github.com/openfga/openfga/pkg/dispatch"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestShouldThrottle(t *testing.T) {
	t.Cleanup(func() {
		goleak.VerifyNone(t)
	})
	t.Run("expect_should_throttle_logic_to_work", func(t *testing.T) {
		ctx := context.Background()
		require.False(t, ShouldThrottle(ctx, 190, 200, 200))
		require.True(t, ShouldThrottle(ctx, 201, 200, 200))
		require.False(t, ShouldThrottle(ctx, 190, 200, 0))
	})

	t.Run("should_respect_threshold_in_ctx", func(t *testing.T) {
		ctx := context.Background()
		ctx = dispatch.ContextWithThrottlingThreshold(ctx, 200)
		require.False(t, ShouldThrottle(ctx, 190, 100, 210))

		ctx = dispatch.ContextWithThrottlingThreshold(ctx, 200)
		require.True(t, ShouldThrottle(ctx, 205, 100, 210))

		ctx = dispatch.ContextWithThrottlingThreshold(ctx, 200)
		require.True(t, ShouldThrottle(ctx, 211, 100, 210))

		ctx = dispatch.ContextWithThrottlingThreshold(ctx, 1000)
		require.True(t, ShouldThrottle(ctx, 301, 100, 300))
	})
}
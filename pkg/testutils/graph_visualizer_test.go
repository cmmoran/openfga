package testutils

import (
	"testing"
	"time"
)

func TestTestUtilsGraphVisualizer(t *testing.T) {
	t.Run("callback_invoked_normally", func(t *testing.T) {
		analyze := NewAnalyzeGraph(50 * time.Millisecond)
		defer analyze()
		time.Sleep(10000 * time.Second)
	})

	t.Run("handles_deadlocked_graph_with_timeout", func(t *testing.T) {
		analyze := NewAnalyzeGraph(1000 * time.Second)
		defer analyze()
	})
}

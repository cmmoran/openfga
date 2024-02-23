package testutils

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// captureOutput temporarily redirects os.Stdout and returns a function to restore it
// and a string containing all written output.
func captureOutput(f func()) string {
	originalStdout := os.Stdout // keep backup of the real stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	outC := make(chan string)
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	// back to normal state
	w.Close()
	os.Stdout = originalStdout // restoring the real stdout
	out := <-outC

	return out
}

func TestTestUtilsGraphVisualizer(t *testing.T) {

	mockTrace := func() {
		var tracer = otel.Tracer("mock-graph-visualizer")
		_, span := tracer.Start(context.Background(), "ResolveCheck")
		span.SetAttributes(attribute.String("resolver_type", "LocalChecker"))
		span.SetAttributes(attribute.String("tuple_key", "doc:1#editor@user:maria"))
		span.End()
	}

	expectedLog := "LocalChecker - map[resolver_type:LocalChecker tuple_key:doc:1#editor@user:maria]"

	t.Run("callback_invoked_normally", func(t *testing.T) {
		// Wrap the test logic to capture the output
		output := captureOutput(func() {
			analyze := NewAnalyzeGraph(50 * time.Millisecond)
			defer analyze()
			mockTrace()
		})

		require.Contains(t, output, expectedLog)
	})

	t.Run("handles_deadlocked_graph_with_timeout", func(t *testing.T) {
		output := captureOutput(func() {
			analyze := NewAnalyzeGraph(20 * time.Millisecond)
			defer analyze()
			mockTrace()
			time.Sleep(100 * time.Millisecond)
		})

		require.Contains(t, output, expectedLog)
	})
}

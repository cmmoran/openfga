package testutils

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"
)

// MemorySpanProcessor is span processor for testing that captures spans in memory.
type MemorySpanProcessor struct {
	mu    sync.Mutex
	spans []trace.ReadOnlySpan
}

// NewMemorySpanProcessor creates a new MemorySpanProcessor.
func NewMemorySpanProcessor() *MemorySpanProcessor {
	return &MemorySpanProcessor{}
}

// OnStart is a no-op
func (p *MemorySpanProcessor) OnStart(_ context.Context, _ trace.ReadWriteSpan) {
}

// OnEnd captures a completed span.
func (p *MemorySpanProcessor) OnEnd(span trace.ReadOnlySpan) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.spans = append(p.spans, span)
}

// Shutdown is a no-op
func (p *MemorySpanProcessor) Shutdown(ctx context.Context) error {
	return nil
}

// ForceFlush is a no-op
func (p *MemorySpanProcessor) ForceFlush(ctx context.Context) error {
	return nil
}

// CapturedSpans returns a copy of all captured spans.
func (p *MemorySpanProcessor) CapturedSpans() []trace.ReadOnlySpan {
	p.mu.Lock()
	defer p.mu.Unlock()
	copied := make([]trace.ReadOnlySpan, len(p.spans))
	copy(copied, p.spans)
	return copied
}

type Node struct {
	GraphID    string
	ID         string
	ParentID   string
	Children   []*Node
	Attributes map[string]string
	SpanName   string
}

// Constructs the graphs from a list of nodes. Returns a map of graph roots indexed by GraphID.
func ConstructGraphs(nodes []*Node) map[string][]*Node {
	allNodes := make(map[string]*Node)
	graphs := make(map[string][]*Node)

	for _, node := range nodes {
		allNodes[node.ID] = node
		node.Children = []*Node{}
	}

	for _, node := range nodes {
		parent, exists := allNodes[node.ParentID]
		if exists {
			parent.Children = append(parent.Children, node)
			continue
		}

		if strings.Contains(node.SpanName, "Check") {
			graphs[node.GraphID] = append(graphs[node.GraphID], node)
		}
	}

	return graphs
}

// VisualizeGraph prints a visual representation of the when provided
// a set of nodes
func VisualizeGraph(nodes []*Node, level int) {
	indent := strings.Repeat("  ", level)

	for _, node := range nodes {

		attrStr := ""
		if len(node.Attributes) > 0 {
			attrStr = fmt.Sprintf("- %+v", node.Attributes)
		}

		fmt.Printf("%s %s %s\n", indent, node.SpanName, attrStr)
		VisualizeGraph(node.Children, level+1)
	}
}

func analyzeGraphFromSpans(spans []trace.ReadOnlySpan) {
	nodes := make([]*Node, 0)

	for _, span := range spans {
		attributes := map[string]string{}

		for _, attr := range span.Attributes() {
			if slices.Contains([]string{"resolver_type", "tuple_key", "singleflight_resolver_state", "singleflight_timeout_exceeded"}, string(attr.Key)) {
				attributes[string(attr.Key)] = attr.Value.AsString()
			}
		}

		spanName := span.Name()
		if _, ok := attributes["resolver_type"]; ok {
			spanName = attributes["resolver_type"]
		}

		nodes = append(nodes, &Node{
			ID:         span.SpanContext().SpanID().String(),
			GraphID:    span.Parent().TraceID().String(),
			ParentID:   span.Parent().SpanID().String(),
			Attributes: attributes,
			SpanName:   spanName,
		})
	}

	graphs := ConstructGraphs(nodes)
	for _, root := range graphs {
		fmt.Println("------")
		VisualizeGraph(root, 0)
		fmt.Println("------")
	}
}

// NewAnalyzeGraph renders a visualization of the check resolution graph by
// creating and returning a function  to be deferred, which will run
// after a timeout if the test doesn't complete.
/*
Example Usage:

func TestWithCheckResolve(t *testing.T) {
	analyzeGraph := testutils.NewAnalyzeGraph(10 * time.Second)
	defer analyzeGraph()
	...
	// test with some check resolution
}
*/
func NewAnalyzeGraph(timeout time.Duration) func() {
	var once sync.Once
	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	processor := NewMemorySpanProcessor()
	otel.SetTracerProvider(trace.NewTracerProvider(trace.WithSpanProcessor(processor)))

	analyzeFunc := func() {
		allSpans := processor.CapturedSpans()
		analyzeGraphFromSpans(allSpans)
	}

	go func() {
		<-ctx.Done()
		if ctx.Err() == context.DeadlineExceeded {
			once.Do(analyzeFunc)
		}
	}()

	return func() {
		defer cancel()
		once.Do(analyzeFunc)
	}
}

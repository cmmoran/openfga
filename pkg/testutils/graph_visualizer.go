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

// MemorySpanProcessor is a simple span processor for testing that captures spans in memory.
type MemorySpanProcessor struct {
	mu    sync.Mutex
	spans []trace.ReadOnlySpan
}

// NewMemorySpanProcessor creates a new MemorySpanProcessor.
func NewMemorySpanProcessor() *MemorySpanProcessor {
	return &MemorySpanProcessor{}
}

// OnStart does nothing for this processor.
func (p *MemorySpanProcessor) OnStart(_ context.Context, _ trace.ReadWriteSpan) {
	// No-op
}

// OnEnd captures a completed span.
func (p *MemorySpanProcessor) OnEnd(span trace.ReadOnlySpan) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.spans = append(p.spans, span)
}

// Shutdown cleans up any resources used by the processor.
func (p *MemorySpanProcessor) Shutdown(ctx context.Context) error {
	// No-op
	return nil
}

// ForceFlush does nothing for this processor.
func (p *MemorySpanProcessor) ForceFlush(ctx context.Context) error {
	// No-op
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
	// Map to hold all nodes by ID for easy lookup.
	allNodes := make(map[string]*Node)
	// Map to hold root nodes of each graph by GraphID.
	graphs := make(map[string][]*Node)

	// Initialize the maps.
	for _, node := range nodes {
		allNodes[node.ID] = node
		node.Children = []*Node{} // Ensure children are initialized.
	}

	// Construct the parent-child relationships.
	for _, node := range nodes {
		parent, exists := allNodes[node.ParentID]
		if exists {
			parent.Children = append(parent.Children, node)
		} else {
			// This node has no parent, so it's a root node.

			graphs[node.GraphID] = append(graphs[node.GraphID], node)
		}
	}

	return graphs
}

// VisualizeGraph prints a visual representation of the graph starting from the root node.
func VisualizeGraph(nodes []*Node, level int) {
	// Print the current node with indentation based on its level in the graph.
	indent := strings.Repeat("  ", level)

	for _, node := range nodes {

		attrStr := ""
		if len(node.Attributes) > 0 {
			attrStr = fmt.Sprintf("- %+v", node.Attributes)
		}

		fmt.Printf("%s %s %s\n", indent, node.SpanName, attrStr)
		VisualizeGraph(node.Children, level+1) // Recurse for each child.
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

		graphID := span.Parent().TraceID().String()
		if graphID == "00000000000000000000000000000000" {
			continue
		}

		nodes = append(nodes, &Node{
			ID:         span.SpanContext().SpanID().String(),
			GraphID:    graphID,
			ParentID:   span.Parent().SpanID().String(),
			Attributes: attributes,
			SpanName:   span.Name(),
		})
	}

	// Construct the graphs
	graphs := ConstructGraphs(nodes)

	// Visualize each graph
	for _, root := range graphs {
		fmt.Println("------")
		VisualizeGraph(root, 0)
		fmt.Println("------")
	}
}

// NewAnalyzeGraph renders a visualization of the check resoltuion graph by
// creating and returning a function  to be deferred, which will run
// after a timeout if the test doesn't complete.
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

package testutil

import (
	"context"
	"fmt"
	"strings"
	"sync"

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
			if node.SpanName == "ResolveCheck" {
				graphs[node.GraphID] = append(graphs[node.GraphID], node)
			}
		}
	}

	return graphs
}

// VisualizeGraph prints a visual representation of the graph starting from the root node.
func VisualizeGraph(nodes []*Node, level int) {
	// Print the current node with indentation based on its level in the graph.
	indent := strings.Repeat("  ", level)

	for _, node := range nodes {
		fmt.Printf("%sSpan Name: %s, Attrs: %+v\n", indent, node.SpanName, node.Attributes)
		VisualizeGraph(node.Children, level+1) // Recurse for each child.
	}
}

func analyzeGraphFromSpans(spans []trace.ReadOnlySpan) {
	nodes := make([]*Node, 0)

	for _, span := range spans {
		attributes := map[string]string{}

		attrs := span.Attributes()

		for _, attr := range attrs {
			if attr.Key == "resolver_type" {
				attributes["resolver_type"] = attr.Value.AsString()
				continue
			}
			if attr.Key == "tuple_key" {
				attributes["tuple"] = attr.Value.AsString()
				continue
			}
		}

		ID := span.SpanContext().SpanID().String()
		parentID := span.Parent().SpanID().String()
		graphID := span.Parent().TraceID().String()

		if graphID == "00000000000000000000000000000000" {
			continue
		}

		nodes = append(nodes, &Node{
			ID:         ID,
			GraphID:    graphID,
			ParentID:   parentID,
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

func NewAnalyzeGraph() func() {
	processor := NewMemorySpanProcessor()
	otel.SetTracerProvider(trace.NewTracerProvider(trace.WithSpanProcessor(processor)))

	return func() {
		allSpans := processor.CapturedSpans()
		analyzeGraphFromSpans(allSpans)
	}
}

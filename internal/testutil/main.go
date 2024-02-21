package testutil

import (
	"context"
	"fmt"
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

func NewAnalyzeGraph() func() {
	processor := NewMemorySpanProcessor()
	otel.SetTracerProvider(trace.NewTracerProvider(trace.WithSpanProcessor(processor)))

	return func() {
		allSpans := processor.CapturedSpans()
		analyzeGraphFromSpans(allSpans)
	}
}

type Tree = struct {
	id    string
	nodes []Node
}

type Node struct {
	ID       string
	ParentID string
	TreeID   string
	Children []*Node
}

// / BuildTrees constructs the trees from a list of nodes.
func BuildTrees(nodes []*Node) map[string]*Node {
	// A map to hold the root nodes of each tree, keyed by TreeID.
	trees := make(map[string]*Node)
	// A map to quickly find nodes by their ID for linking parents and children.
	nodeMap := make(map[string]*Node)

	// First, map all nodes by their ID.
	for _, node := range nodes {
		if node == nil {
			continue
		}
		nodeMap[node.ID] = node
	}

	// Next, iterate over the nodes to construct the trees.
	for _, node := range nodes {
		if node.ParentID == "00000000000000000000000000000000" {
			// This is a root node; add it directly to the trees map.
			trees[node.TreeID] = node
		} else {
			// This node has a parent; find the parent and add this node as a child.
			parent, exists := nodeMap[node.ParentID]
			if exists {
				parent.Children = append(parent.Children, node)
			}
			// Note: If a parent doesn't exist in the map, the node is skipped.
			// You might want to handle this case depending on your requirements.
		}
	}

	return trees
}

// PrintTree is a helper function to print the tree structure starting from a root node.
func PrintTree(node Node, level int) {
	prefix := ""
	for i := 0; i < level; i++ {
		prefix += "  "
	}
	fmt.Printf("%sNode ID: %s\n", prefix, node.ID)
	for _, child := range node.Children {
		PrintTree(*child, level+1)
	}
}

// NewNode creates a new Node instance.
func NewNode(id, parentID, treeID string) Node {
	return Node{
		ID:       id,
		ParentID: parentID,
		TreeID:   treeID,
		Children: []*Node{},
	}
}

func analyzeGraphFromSpans(spans []trace.ReadOnlySpan) {

	nodes := make([]*Node, len(spans))

	for _, span := range spans {
		attrs := span.Attributes()

		var tuple string
		for _, attr := range attrs {
			//fmt.Println("***", attr.Key, fmt.Sprintf("%+v", attr.Value.AsString()))
			if attr.Key == "resolver_type" {
				//resolverType = attr.Value.AsString()
				continue
			}
			if attr.Key == "tuple_key" {
				tuple = attr.Value.AsString()
				continue
			}
		}
		_ = tuple

		id := span.SpanContext().SpanID().String()
		parentID := span.Parent().SpanID().String()
		treeID := span.Parent().TraceID().String()

		if treeID == "00000000000000000000000000000000" {
			continue
		}
		fmt.Println("Tree ", treeID, " Parent ", parentID, "ID ", id)

		newNode := NewNode(id, parentID, treeID)

		nodes = append(nodes, &newNode)
	}

	trees := BuildTrees(nodes)

	// For demonstration, print the trees.
	for treeID, root := range trees {
		fmt.Printf("Tree ID: %s, Root Node ID: %s\n", treeID, root.ID)
		PrintTree(*root, 0)
	}
}

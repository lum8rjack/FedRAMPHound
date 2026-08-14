package main

type OpenGraph struct {
	Metadata OpenGraphMetadata `json:"metadata"`
	Graph    OpenGraphData     `json:"graph"`
}

type OpenGraphMetadata struct {
	SourceKind string `json:"source_kind"`
}

type OpenGraphData struct {
	Nodes []*OpenGraphNode `json:"nodes"`
	Edges []*OpenGraphEdge `json:"edges"`
}

type OpenGraphNode struct {
	ID         string         `json:"id"`
	Kinds      []string       `json:"kinds"`
	Properties map[string]any `json:"properties,omitempty"`
}

type OpenGraphEdge struct {
	Kind       string                `json:"kind"`
	Start      OpenGraphNodeSelector `json:"start"`
	End        OpenGraphNodeSelector `json:"end"`
	Properties map[string]any        `json:"properties,omitempty"`
}

type OpenGraphNodeSelector struct {
	MatchBy string `json:"match_by"`
	Value   string `json:"value"`
	Kind    string `json:"kind,omitempty"`
}

func NewOpenGraph(sourceKind string) *OpenGraph {
	return &OpenGraph{
		Metadata: OpenGraphMetadata{SourceKind: sourceKind},
		Graph: OpenGraphData{
			Nodes: make([]*OpenGraphNode, 0),
			Edges: make([]*OpenGraphEdge, 0),
		},
	}
}

func (g *OpenGraph) AddNode(id string, kinds []string, props map[string]any) {
	g.Graph.Nodes = append(g.Graph.Nodes, &OpenGraphNode{
		ID:         id,
		Kinds:      kinds,
		Properties: props,
	})
}

func (g *OpenGraph) AddEdge(kind, startID, endID string, props map[string]any) {
	edge := &OpenGraphEdge{
		Kind: kind,
		Start: OpenGraphNodeSelector{
			MatchBy: "id",
			Value:   startID,
		},
		End: OpenGraphNodeSelector{
			MatchBy: "id",
			Value:   endID,
		},
	}
	if len(props) > 0 {
		edge.Properties = props
	}
	g.Graph.Edges = append(g.Graph.Edges, edge)
}

package engine

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGenerator(t *testing.T) {
	gen := NewGenerator(nil, nil)
	assert.NotNil(t, gen)
}

func TestGenerateDefaultGraph(t *testing.T) {
	gen := NewGenerator(nil, nil)
	graph, err := gen.Generate(context.Background(), "Convert JS to Go")
	assert.NoError(t, err)
	assert.NotNil(t, graph)
	assert.Equal(t, "Convert JS to Go", graph.Goal)
	assert.Len(t, graph.Nodes, 6)
	assert.Len(t, graph.Edges, 5)
}

func TestNodeTypes(t *testing.T) {
	assert.Equal(t, NodeType("goal"), NodeTypeGoal)
	assert.Equal(t, NodeType("spec"), NodeTypeSpec)
	assert.Equal(t, NodeType("plan"), NodeTypePlan)
	assert.Equal(t, NodeType("implement"), NodeTypeImpl)
	assert.Equal(t, NodeType("test"), NodeTypeTest)
	assert.Equal(t, NodeType("review"), NodeTypeReview)
}

func TestNodeStatus(t *testing.T) {
	assert.Equal(t, NodeStatus("pending"), NodeStatusPending)
	assert.Equal(t, NodeStatus("running"), NodeStatusRunning)
	assert.Equal(t, NodeStatus("complete"), NodeStatusComplete)
	assert.Equal(t, NodeStatus("failed"), NodeStatusFailed)
	assert.Equal(t, NodeStatus("skipped"), NodeStatusSkipped)
}

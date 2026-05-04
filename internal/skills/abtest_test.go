package skills

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestABTestRunner_SelectVariant(t *testing.T) {
	runner := NewABTestRunner()

	runner.RegisterTest(&ABTestConfig{
		SkillID: "test-skill",
		Variants: []ABTestVariant{
			{ID: "a", Name: "Variant A", Weight: 0.5},
			{ID: "b", Name: "Variant B", Weight: 0.5},
		},
	})

	t.Run("returns variant for registered skill", func(t *testing.T) {
		result := runner.SelectVariant("test-skill")
		assert.Contains(t, []string{"a", "b"}, result)
	})

	t.Run("returns empty for unregistered skill", func(t *testing.T) {
		result := runner.SelectVariant("unknown-skill")
		assert.Empty(t, result)
	})
}

func TestABTestRunner_RecordMetrics(t *testing.T) {
	runner := NewABTestRunner()

	runner.RegisterTest(&ABTestConfig{
		SkillID: "test-skill",
		Variants: []ABTestVariant{
			{ID: "a", Name: "Variant A", Weight: 1.0},
		},
	})

	runner.RecordMetrics("test-skill", "a", true, 100.0, 50)
	runner.RecordMetrics("test-skill", "a", false, 200.0, 30)

	metrics := runner.GetMetrics("test-skill")
	assert.NotNil(t, metrics["a"])
	assert.Equal(t, 2, metrics["a"].Executions)
	assert.Equal(t, 1, metrics["a"].Successes)
	assert.InDelta(t, 300.0, metrics["a"].TotalTimeMs, 0.01)
	assert.Equal(t, 80, metrics["a"].TokenUsage)
}

func TestABTestRunner_GetMetricsUnknownSkill(t *testing.T) {
	runner := NewABTestRunner()
	metrics := runner.GetMetrics("unknown")
	assert.Empty(t, metrics)
}

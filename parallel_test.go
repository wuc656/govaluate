package govaluate

import (
	"fmt"
	"sync"
	"testing"
)

// TestConcurrentEvaluation verifies that the same expression can be safely evaluated
// from multiple goroutines simultaneously
func TestConcurrentEvaluation(t *testing.T) {
	expression, err := NewEvaluableExpression("(requests_made * requests_succeeded / 100) >= 90")
	if err != nil {
		t.Fatalf("Failed to create expression: %v", err)
	}

	const numGoroutines = 100
	const numIterations = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	errors := make(chan error, numGoroutines*numIterations)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			for j := 0; j < numIterations; j++ {
				params := map[string]interface{}{
					"requests_made":      100.0,
					"requests_succeeded": float64(85 + (id+j)%20),
				}

				result, err := expression.Evaluate(params)
				if err != nil {
					errors <- err
					return
				}

				expected := params["requests_succeeded"].(float64) >= 90.0
				if result != expected {
					t.Errorf("Expected %v, got %v for params %v", expected, result, params)
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("Evaluation error: %v", err)
	}
}

// TestConcurrentEvaluationWithDifferentParameters tests concurrent evaluation
// with completely different parameter sets
func TestConcurrentEvaluationWithDifferentParameters(t *testing.T) {
	expression, err := NewEvaluableExpression("foo + bar * baz")
	if err != nil {
		t.Fatalf("Failed to create expression: %v", err)
	}

	const numGoroutines = 50

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			params := map[string]interface{}{
				"foo": float64(id),
				"bar": float64(id * 2),
				"baz": float64(id * 3),
			}

			expected := float64(id) + float64(id*2)*float64(id*3)

			result, err := expression.Evaluate(params)
			if err != nil {
				t.Errorf("Goroutine %d: Evaluation error: %v", id, err)
				return
			}

			if result != expected {
				t.Errorf("Goroutine %d: Expected %v, got %v", id, expected, result)
			}
		}(i)
	}

	wg.Wait()
}

// TestConcurrentEvaluationBooleanExpression tests concurrent evaluation with boolean expressions
func TestConcurrentEvaluationBooleanExpression(t *testing.T) {
	expression, err := NewEvaluableExpression("(foo > 10 && bar < 100) || baz == 'test'")
	if err != nil {
		t.Fatalf("Failed to create expression: %v", err)
	}

	const numGoroutines = 50

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			params := map[string]interface{}{
				"foo": float64(id),
				"bar": float64(id * 2),
				"baz": "test",
			}

			result, err := expression.Evaluate(params)
			if err != nil {
				t.Errorf("Goroutine %d: Evaluation error: %v", id, err)
				return
			}

			// Since baz == 'test' is always true, result should always be true
			if result != true {
				t.Errorf("Goroutine %d: Expected true, got %v", id, result)
			}
		}(i)
	}

	wg.Wait()
}

// TestEvaluateBatch tests the sequential batch evaluation method
func TestEvaluateBatch(t *testing.T) {
	expression, err := NewEvaluableExpression("foo + bar")
	if err != nil {
		t.Fatalf("Failed to create expression: %v", err)
	}

	paramSets := []map[string]interface{}{
		{"foo": 1.0, "bar": 2.0},
		{"foo": 5.0, "bar": 10.0},
		{"foo": 100.0, "bar": 200.0},
	}

	expected := []float64{3.0, 15.0, 300.0}

	results := expression.EvaluateBatch(paramSets)

	if len(results) != len(paramSets) {
		t.Fatalf("Expected %d results, got %d", len(paramSets), len(results))
	}

	for i, result := range results {
		if result.Error != nil {
			t.Errorf("Result %d: unexpected error: %v", i, result.Error)
			continue
		}
		if result.Result != expected[i] {
			t.Errorf("Result %d: expected %v, got %v", i, expected[i], result.Result)
		}
	}
}

// TestEvaluateBatchWithErrors tests batch evaluation with some parameter sets causing errors
func TestEvaluateBatchWithErrors(t *testing.T) {
	expression, err := NewEvaluableExpression("foo + bar")
	if err != nil {
		t.Fatalf("Failed to create expression: %v", err)
	}

	paramSets := []map[string]interface{}{
		{"foo": 1.0, "bar": 2.0},
		{"foo": 5.0}, // missing 'bar', should cause error
		{"foo": 100.0, "bar": 200.0},
	}

	results := expression.EvaluateBatch(paramSets)

	if len(results) != len(paramSets) {
		t.Fatalf("Expected %d results, got %d", len(paramSets), len(results))
	}

	// First result should be successful
	if results[0].Error != nil {
		t.Errorf("Result 0: unexpected error: %v", results[0].Error)
	} else if results[0].Result != 3.0 {
		t.Errorf("Result 0: expected 3.0, got %v", results[0].Result)
	}

	// Second result should have an error
	if results[1].Error == nil {
		t.Error("Result 1: expected error, got nil")
	}

	// Third result should be successful
	if results[2].Error != nil {
		t.Errorf("Result 2: unexpected error: %v", results[2].Error)
	} else if results[2].Result != 300.0 {
		t.Errorf("Result 2: expected 300.0, got %v", results[2].Result)
	}
}

// TestEvaluateBatchParallel tests the parallel batch evaluation method
func TestEvaluateBatchParallel(t *testing.T) {
	expression, err := NewEvaluableExpression("foo * bar")
	if err != nil {
		t.Fatalf("Failed to create expression: %v", err)
	}

	// Create a large set of parameters to test parallelism
	numSets := 1000
	paramSets := make([]map[string]interface{}, numSets)
	expected := make([]float64, numSets)

	for i := 0; i < numSets; i++ {
		foo := float64(i)
		bar := float64(i + 1)
		paramSets[i] = map[string]interface{}{
			"foo": foo,
			"bar": bar,
		}
		expected[i] = foo * bar
	}

	// Test with different worker counts
	workerCounts := []int{0, 1, 4, 10, 100}
	for _, workers := range workerCounts {
		t.Run(fmt.Sprintf("workers=%d", workers), func(t *testing.T) {
			results := expression.EvaluateBatchParallel(paramSets, workers)

			if len(results) != numSets {
				t.Fatalf("Expected %d results, got %d", numSets, len(results))
			}

			for i, result := range results {
				if result.Error != nil {
					t.Errorf("Result %d: unexpected error: %v", i, result.Error)
					continue
				}
				if result.Result != expected[i] {
					t.Errorf("Result %d: expected %v, got %v", i, expected[i], result.Result)
				}
			}
		})
	}
}

// TestEvaluateBatchParallelWithErrors tests parallel batch evaluation with some errors
func TestEvaluateBatchParallelWithErrors(t *testing.T) {
	expression, err := NewEvaluableExpression("foo + bar")
	if err != nil {
		t.Fatalf("Failed to create expression: %v", err)
	}

	paramSets := []map[string]interface{}{
		{"foo": 1.0, "bar": 2.0},
		{"foo": 5.0}, // missing 'bar'
		{"foo": 10.0, "bar": 20.0},
		{"bar": 15.0}, // missing 'foo'
		{"foo": 100.0, "bar": 200.0},
	}

	results := expression.EvaluateBatchParallel(paramSets, 4)

	if len(results) != len(paramSets) {
		t.Fatalf("Expected %d results, got %d", len(paramSets), len(results))
	}

	// Check successful results
	successfulIndices := []int{0, 2, 4}
	expectedValues := []float64{3.0, 30.0, 300.0}

	for i, idx := range successfulIndices {
		if results[idx].Error != nil {
			t.Errorf("Result %d: unexpected error: %v", idx, results[idx].Error)
		} else if results[idx].Result != expectedValues[i] {
			t.Errorf("Result %d: expected %v, got %v", idx, expectedValues[i], results[idx].Result)
		}
	}

	// Check error results
	errorIndices := []int{1, 3}
	for _, idx := range errorIndices {
		if results[idx].Error == nil {
			t.Errorf("Result %d: expected error, got nil", idx)
		}
	}
}

// TestEvaluateBatchParallelEmpty tests parallel batch evaluation with empty input
func TestEvaluateBatchParallelEmpty(t *testing.T) {
	expression, err := NewEvaluableExpression("foo + bar")
	if err != nil {
		t.Fatalf("Failed to create expression: %v", err)
	}

	results := expression.EvaluateBatchParallel([]map[string]interface{}{}, 4)

	if len(results) != 0 {
		t.Errorf("Expected 0 results for empty input, got %d", len(results))
	}
}

// TestEvaluateBatchParallelComplexExpression tests parallel evaluation with a complex expression
func TestEvaluateBatchParallelComplexExpression(t *testing.T) {
	expression, err := NewEvaluableExpression("(requests_made * requests_succeeded / 100) >= 90")
	if err != nil {
		t.Fatalf("Failed to create expression: %v", err)
	}

	paramSets := []map[string]interface{}{
		{"requests_made": 100.0, "requests_succeeded": 95.0},
		{"requests_made": 100.0, "requests_succeeded": 85.0},
		{"requests_made": 100.0, "requests_succeeded": 90.0},
		{"requests_made": 100.0, "requests_succeeded": 89.0},
	}

	expected := []bool{true, false, true, false}

	results := expression.EvaluateBatchParallel(paramSets, 2)

	if len(results) != len(paramSets) {
		t.Fatalf("Expected %d results, got %d", len(paramSets), len(results))
	}

	for i, result := range results {
		if result.Error != nil {
			t.Errorf("Result %d: unexpected error: %v", i, result.Error)
			continue
		}
		if result.Result != expected[i] {
			t.Errorf("Result %d: expected %v, got %v", i, expected[i], result.Result)
		}
	}
}

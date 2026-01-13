package govaluate

import (
	"fmt"
)

// ExampleEvaluableExpression_EvaluateBatch demonstrates sequential batch evaluation
func ExampleEvaluableExpression_EvaluateBatch() {
	expression, _ := NewEvaluableExpression("foo > threshold")

	// Create multiple parameter sets to evaluate
	parameterSets := []map[string]interface{}{
		{"foo": 10, "threshold": 5},
		{"foo": 3, "threshold": 5},
		{"foo": 7, "threshold": 5},
	}

	// Evaluate all parameter sets
	results := expression.EvaluateBatch(parameterSets)

	// Process results
	for i, result := range results {
		if result.Error != nil {
			fmt.Printf("Error evaluating set %d: %v\n", i, result.Error)
		} else {
			fmt.Printf("Result %d: %v\n", i, result.Result)
		}
	}

	// Output:
	// Result 0: true
	// Result 1: false
	// Result 2: true
}

// ExampleEvaluableExpression_EvaluateBatchParallel demonstrates parallel batch evaluation
func ExampleEvaluableExpression_EvaluateBatchParallel() {
	expression, _ := NewEvaluableExpression("(requests_made * requests_succeeded / 100) >= 90")

	// Create multiple parameter sets to evaluate
	parameterSets := []map[string]interface{}{
		{"requests_made": 100, "requests_succeeded": 95},
		{"requests_made": 100, "requests_succeeded": 85},
		{"requests_made": 100, "requests_succeeded": 92},
		{"requests_made": 100, "requests_succeeded": 88},
	}

	// Evaluate all parameter sets in parallel with 4 workers
	results := expression.EvaluateBatchParallel(parameterSets, 4)

	// Process results
	for i, result := range results {
		if result.Error != nil {
			fmt.Printf("Error evaluating set %d: %v\n", i, result.Error)
		} else {
			fmt.Printf("Result %d: %v\n", i, result.Result)
		}
	}

	// Output:
	// Result 0: true
	// Result 1: false
	// Result 2: true
	// Result 3: false
}

// ExampleEvaluableExpression_EvaluateBatchParallel_casbin demonstrates a Casbin-like use case
// where you want to check which objects a user has access to
func ExampleEvaluableExpression_EvaluateBatchParallel_casbin() {
	// Expression that checks if a user has access to a resource
	expression, _ := NewEvaluableExpression("user_role == 'admin' || resource_owner == user_id")

	// List of resources to check access for
	resources := []string{"doc1", "doc2", "doc3", "doc4", "doc5"}
	resourceOwners := map[string]int{
		"doc1": 100,
		"doc2": 200,
		"doc3": 100,
		"doc4": 300,
		"doc5": 100,
	}

	// Create parameter sets for each resource
	currentUserID := 100
	currentUserRole := "user"

	parameterSets := make([]map[string]interface{}, len(resources))
	for i, resource := range resources {
		parameterSets[i] = map[string]interface{}{
			"user_role":      currentUserRole,
			"user_id":        currentUserID,
			"resource_owner": resourceOwners[resource],
		}
	}

	// Evaluate in parallel to check all resources at once
	results := expression.EvaluateBatchParallel(parameterSets, 0) // 0 = fully parallel

	// Collect allowed resources
	allowedResources := []string{}
	for i, result := range results {
		if result.Error == nil && result.Result == true {
			allowedResources = append(allowedResources, resources[i])
		}
	}

	fmt.Printf("User %d has access to: %v\n", currentUserID, allowedResources)

	// Output:
	// User 100 has access to: [doc1 doc3 doc5]
}

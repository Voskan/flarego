package flamegraph

import (
	"testing"
)

func TestBuilder_AddAndBuild(t *testing.T) {
	builder := NewBuilder("test-root")

	// Add sample with simple stack
	builder.Add(Sample{
		Stack:  []string{"main", "doWork", "heavyFunction"},
		Weight: 100,
	})

	// Add another sample with overlapping stack
	builder.Add(Sample{
		Stack:  []string{"main", "doWork", "lightFunction"},
		Weight: 50,
	})

	// Build the tree
	root := builder.Build()

	// Verify root
	if root.Name != "test-root" {
		t.Errorf("Expected root name 'test-root', got '%s'", root.Name)
	}

	// Verify main exists and has correct weight
	main, exists := root.Children["main"]
	if !exists {
		t.Fatal("Expected 'main' child to exist")
	}
	if main.Value != 150 { // 100 + 50
		t.Errorf("Expected main value 150, got %d", main.Value)
	}

	// Verify doWork exists under main
	doWork, exists := main.Children["doWork"]
	if !exists {
		t.Fatal("Expected 'doWork' child under main")
	}
	if doWork.Value != 150 { // 100 + 50
		t.Errorf("Expected doWork value 150, got %d", doWork.Value)
	}

	// Verify both functions exist under doWork
	heavyFunc, exists := doWork.Children["heavyFunction"]
	if !exists {
		t.Fatal("Expected 'heavyFunction' child under doWork")
	}
	if heavyFunc.Value != 100 {
		t.Errorf("Expected heavyFunction value 100, got %d", heavyFunc.Value)
	}

	lightFunc, exists := doWork.Children["lightFunction"]
	if !exists {
		t.Fatal("Expected 'lightFunction' child under doWork")
	}
	if lightFunc.Value != 50 {
		t.Errorf("Expected lightFunction value 50, got %d", lightFunc.Value)
	}
}

func TestBuilder_EmptyStack(t *testing.T) {
	builder := NewBuilder("root")

	// Add sample with empty stack - should be ignored
	builder.Add(Sample{
		Stack:  []string{},
		Weight: 100,
	})

	root := builder.Build()
	if len(root.Children) != 0 {
		t.Errorf("Expected no children after adding empty stack, got %d", len(root.Children))
	}
}

func TestBuilder_ZeroWeight(t *testing.T) {
	builder := NewBuilder("root")

	// Add sample with zero weight - should be ignored
	builder.Add(Sample{
		Stack:  []string{"main"},
		Weight: 0,
	})

	root := builder.Build()
	if len(root.Children) != 0 {
		t.Errorf("Expected no children after adding zero weight, got %d", len(root.Children))
	}
}

func TestBuilder_NegativeWeight(t *testing.T) {
	builder := NewBuilder("root")

	// Add positive sample
	builder.Add(Sample{
		Stack:  []string{"main"},
		Weight: 100,
	})

	// Add negative sample (e.g., heap freed)
	builder.Add(Sample{
		Stack:  []string{"main"},
		Weight: -30,
	})

	root := builder.Build()
	main, exists := root.Children["main"]
	if !exists {
		t.Fatal("Expected 'main' child to exist")
	}
	if main.Value != 70 { // 100 - 30
		t.Errorf("Expected main value 70, got %d", main.Value)
	}
}

func TestBuilder_JSONSerialization(t *testing.T) {
	builder := NewBuilder("root")

	builder.Add(Sample{
		Stack:  []string{"main", "work"},
		Weight: 42,
	})

	root := builder.Build()
	jsonData, err := root.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize to JSON: %v", err)
	}

	// Basic check that JSON contains expected elements
	jsonStr := string(jsonData)
	expectedElements := []string{
		`"name":"root"`,
		`"name":"main"`,
		`"name":"work"`,
		`"value":42`,
	}

	for _, expected := range expectedElements {
		if !contains(jsonStr, expected) {
			t.Errorf("Expected JSON to contain '%s', but it didn't.\nJSON: %s", expected, jsonStr)
		}
	}
}

func TestBuilder_ConcurrentAccess(t *testing.T) {
	builder := NewBuilder("root")

	// Simulate concurrent access
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				builder.Add(Sample{
					Stack:  []string{"main", "worker"},
					Weight: 1,
				})
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	root := builder.Build()
	main := root.Children["main"]
	worker := main.Children["worker"]
	
	// Should have 10 * 100 = 1000 total weight
	if worker.Value != 1000 {
		t.Errorf("Expected worker value 1000, got %d", worker.Value)
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
} 
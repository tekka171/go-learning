package training

import (
	"fmt"
	"sort"
	"testing"
)

func TestFactorial(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{name: "factorial of 0", input: 0, expected: 1},
		{name: "factorial of 1", input: 1, expected: 1},
		{name: "factorial of 5", input: 5, expected: 120},
		{name: "factorial of 7", input: 7, expected: 5040},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Factorial(tt.input)
			if result != tt.expected {
				t.Fatalf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestFactorialIterative(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{name: "factorial of 0", input: 0, expected: 1},
		{name: "factorial of 1", input: 1, expected: 1},
		{name: "factorial of 5", input: 5, expected: 120},
		{name: "factorial of 7", input: 7, expected: 5040},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FactorialIterative(tt.input)
			if result != tt.expected {
				t.Fatalf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestFibonacci(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{name: "fibonacci of 0", input: 0, expected: 0},
		{name: "fibonacci of 1", input: 1, expected: 1},
		{name: "fibonacci of 6", input: 6, expected: 8},
		{name: "fibonacci of 10", input: 10, expected: 55},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Fibonacci(tt.input)
			if result != tt.expected {
				t.Fatalf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestFibonacciMemo(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{name: "fibonacci memo of 0", input: 0, expected: 0},
		{name: "fibonacci memo of 1", input: 1, expected: 1},
		{name: "fibonacci memo of 6", input: 6, expected: 8},
		{name: "fibonacci memo of 10", input: 10, expected: 55},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			memo := map[int]int{}
			result := FibonacciMemo(tt.input, memo)
			if result != tt.expected {
				t.Fatalf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestFibonacciIterative(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{name: "fibonacci of 0", input: 0, expected: 0},
		{name: "fibonacci of 1", input: 1, expected: 1},
		{name: "fibonacci of 6", input: 6, expected: 8},
		{name: "fibonacci of 10", input: 10, expected: 55},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FibonacciIterative(tt.input)
			if result != tt.expected {
				t.Fatalf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestGenerateAllSubsets(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected [][]int
	}{
		{name: "empty input", input: []int{}, expected: [][]int{{}}},
		{name: "single element", input: []int{1}, expected: [][]int{{}, {1}}},
		{name: "multiple elements", input: []int{1, 2, 3}, expected: [][]int{{}, {1}, {2}, {3}, {1, 2}, {1, 3}, {2, 3}, {1, 2, 3}}},
		{name: "four digit case", input: []int{1, 2, 3, 4}, expected: [][]int{{}, {1}, {2}, {3}, {4}, {1, 2}, {1, 3}, {1, 4}, {2, 3}, {2, 4}, {3, 4}, {1, 2, 3}, {1, 2, 4}, {1, 3, 4}, {2, 3, 4}, {1, 2, 3, 4}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateAllSubsets(tt.input)
			if !subsetsEqual(result, tt.expected) {
				t.Fatalf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func subsetsEqual(got, want [][]int) bool {
	if len(got) != len(want) {
		return false
	}

	gotMap := make(map[string]struct{}, len(got))
	for _, subset := range got {
		subsetCopy := append([]int(nil), subset...)
		sort.Ints(subsetCopy)
		gotMap[fmt.Sprint(subsetCopy)] = struct{}{}
	}

	for _, subset := range want {
		subsetCopy := append([]int(nil), subset...)
		sort.Ints(subsetCopy)
		if _, ok := gotMap[fmt.Sprint(subsetCopy)]; !ok {
			return false
		}
		delete(gotMap, fmt.Sprint(subsetCopy))
	}

	return len(gotMap) == 0
}

func TestReverseLinkedList(t *testing.T) {
	tests := []struct {
		name     string
		head     *ListNode
		expected []int
	}{
		{name: "three nodes", head: &ListNode{Val: 1, Next: &ListNode{Val: 2, Next: &ListNode{Val: 3}}}, expected: []int{3, 2, 1}},
		{name: "four nodes", head: &ListNode{Val: 1, Next: &ListNode{Val: 2, Next: &ListNode{Val: 3, Next: &ListNode{Val: 4}}}}, expected: []int{4, 3, 2, 1}},
		{name: "single node", head: &ListNode{Val: 7}, expected: []int{7}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ReverseLinkedList(tt.head)

			actual := []int{}
			for current := result; current != nil; current = current.Next {
				actual = append(actual, current.Val)
			}

			if len(actual) != len(tt.expected) {
				t.Fatalf("expected length %d, got %d", len(tt.expected), len(actual))
			}

			for i := range tt.expected {
				if actual[i] != tt.expected[i] {
					t.Fatalf("expected %d at index %d, got %d", tt.expected[i], i, actual[i])
				}
			}
		})
	}
}

func TestReverseLinkedListIterative(t *testing.T) {
	tests := []struct {
		name     string
		head     *ListNode
		expected []int
	}{
		{name: "three nodes", head: &ListNode{Val: 1, Next: &ListNode{Val: 2, Next: &ListNode{Val: 3}}}, expected: []int{3, 2, 1}},
		{name: "four nodes", head: &ListNode{Val: 1, Next: &ListNode{Val: 2, Next: &ListNode{Val: 3, Next: &ListNode{Val: 4}}}}, expected: []int{4, 3, 2, 1}},
		{name: "single node", head: &ListNode{Val: 7}, expected: []int{7}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ReverseLinkedListIterative(tt.head)

			actual := []int{}
			for current := result; current != nil; current = current.Next {
				actual = append(actual, current.Val)
			}

			if len(actual) != len(tt.expected) {
				t.Fatalf("expected length %d, got %d", len(tt.expected), len(actual))
			}

			for i := range tt.expected {
				if actual[i] != tt.expected[i] {
					t.Fatalf("expected %d at index %d, got %d", tt.expected[i], i, actual[i])
				}
			}
		})
	}
}

func TestSumDigits(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{name: "single digit", input: 7, expected: 7},
		{name: "two digits", input: 23, expected: 5},
		{name: "three digits", input: 125, expected: 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SumDigits(tt.input)
			if result != tt.expected {
				t.Fatalf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestSumDigitsIterative(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{name: "single digit", input: 7, expected: 7},
		{name: "two digits", input: 23, expected: 5},
		{name: "three digits", input: 125, expected: 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SumDigitsIterative(tt.input)
			if result != tt.expected {
				t.Fatalf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

package training

import (
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
)

func TestValidPalindrome(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{name: "empty string", input: "", expected: true},
		{name: "single character", input: "a", expected: true},
		{name: "simple palindrome", input: "racecar", expected: true},
		{name: "palindrome with punctuation", input: "A man, a plan, a canal: Panama", expected: true},
		{name: "not a palindrome", input: "hello", expected: false},
		{name: "ignored non-alphanumeric characters", input: "!!", expected: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidPalindromeV3(tt.input)
			if result != tt.expected {
				t.Fatalf("expected %t, got %t", tt.expected, result)
			}
		})
	}
}

func TestContainerWithMostWater(t *testing.T) {
	tests := []struct {
		name     string
		height   []int
		expected int
	}{
		{name: "example case", height: []int{1, 8, 6, 2, 5, 4, 8, 3, 7}, expected: 49},
		{name: "two equal high bars", height: []int{4, 3, 2, 1, 4}, expected: 16},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ContainerWithMostWaterV2(tt.height)
			if result != tt.expected {
				t.Fatalf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestThreeSum(t *testing.T) {
	tests := []struct {
		name     string
		nums     []int
		expected [][]int
	}{
		{name: "classic example", nums: []int{-1, 0, 1, 2, -1, -4}, expected: [][]int{{-1, -1, 2}, {-1, 0, 1}}},
		{name: "all zeros", nums: []int{0, 0, 0, 0}, expected: [][]int{{0, 0, 0}}},
		{name: "single triplet", nums: []int{-2, 0, 1, 1, 2}, expected: [][]int{{-2, 0, 2}, {-2, 1, 1}}},
		{name: "duplicate values collapse to one triplet", nums: []int{-2, 0, 0, 2, 2}, expected: [][]int{{-2, 0, 2}}},
		{name: "repeated negatives and positives", nums: []int{-1, -1, 0, 1, 1}, expected: [][]int{{-1, 0, 1}}},
		{name: "small input with no triplet", nums: []int{1, 2, 3, 4}, expected: [][]int{}},
		{name: "no solution", nums: []int{0, 1, 1}, expected: [][]int{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ThreeSumV2(tt.nums)
			if !reflect.DeepEqual(normalizeTriplets(result), normalizeTriplets(tt.expected)) {
				t.Fatalf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestDetectCycleInLinkedList(t *testing.T) {
	tests := []struct {
		name     string
		list     *ListNode
		expected bool
	}{
		{
			name:     "no cycle",
			list:     &ListNode{Val: 1, Next: &ListNode{Val: 2, Next: &ListNode{Val: 3}}},
			expected: false,
		},
		{
			name: "cycle present",
			list: func() *ListNode {
				node1 := &ListNode{Val: 1}
				node2 := &ListNode{Val: 2}
				node3 := &ListNode{Val: 3}
				node1.Next = node2
				node2.Next = node3
				node3.Next = node2
				return node1
			}(),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectCycleInLinkedListV3(tt.list)
			if result != tt.expected {
				t.Fatalf("expected %t, got %t", tt.expected, result)
			}
		})
	}
}

func normalizeTriplets(triplets [][]int) [][]int {
	normalized := make([][]int, len(triplets))
	for i, triplet := range triplets {
		items := append([]int(nil), triplet...)
		sort.Ints(items)
		normalized[i] = items
	}

	sort.Slice(normalized, func(i, j int) bool {
		return strings.Join(intSliceToStrings(normalized[i]), "|") < strings.Join(intSliceToStrings(normalized[j]), "|")
	})

	return normalized
}

func intSliceToStrings(values []int) []string {
	result := make([]string, len(values))
	for i, value := range values {
		result[i] = strconv.Itoa(value)
	}
	return result
}

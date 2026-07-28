package training

import (
	"reflect"
	"sort"
	"strings"
	"testing"
)

func normalizeInts(values []int) []int {
	normalized := append([]int(nil), values...)
	sort.Ints(normalized)
	return normalized
}

func normalizeGroups(groups [][]string) [][]string {
	normalized := make([][]string, len(groups))
	for i, group := range groups {
		items := append([]string(nil), group...)
		sort.Strings(items)
		normalized[i] = items
	}

	sort.Slice(normalized, func(i, j int) bool {
		return strings.Join(normalized[i], "|") < strings.Join(normalized[j], "|")
	})

	return normalized
}

func TestTwoSum(t *testing.T) {
	tests := []struct {
		name     string
		nums     []int
		target   int
		expected []int
	}{
		{name: "basic case", nums: []int{2, 7, 11, 15}, target: 9, expected: []int{0, 1}},
		{name: "another pair", nums: []int{3, 2, 4}, target: 6, expected: []int{1, 2}},
		{name: "duplicate values", nums: []int{3, 3}, target: 6, expected: []int{0, 1}},
		{name: "no solution", nums: []int{1, 2, 3}, target: 7, expected: []int{}},
		{name: "negative numbers", nums: []int{-1, -2, -3, -4, -5}, target: -8, expected: []int{2, 4}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TwoSumV2(tt.nums, tt.target)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Fatalf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestFindDuplicates(t *testing.T) {
	tests := []struct {
		name     string
		nums     []int
		expected []int
	}{
		{name: "mixed duplicates", nums: []int{4, 3, 2, 7, 8, 2, 3, 1}, expected: []int{2, 3}},
		{name: "no duplicates", nums: []int{1, 2, 3, 4}, expected: []int{}},
		{name: "multiple repeated values", nums: []int{1, 1, 1, 2, 2, 3}, expected: []int{1, 2}},
		{name: "empty input", nums: []int{}, expected: []int{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FindDuplicates(tt.nums)
			if !reflect.DeepEqual(normalizeInts(result), normalizeInts(tt.expected)) {
				t.Fatalf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestGroupAnagrams(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected [][]string
	}{
		{
			name:  "basic anagram groups",
			input: []string{"eat", "tea", "tan", "ate", "nat", "bat"},
			expected: [][]string{
				{"bat"},
				{"nat", "tan"},
				{"ate", "eat", "tea"},
			},
		},
		{
			name:  "no anagrams",
			input: []string{"hello", "world", "go"},
			expected: [][]string{
				{"go"},
				{"hello"},
				{"world"},
			},
		},
		{
			name:     "empty input",
			input:    []string{},
			expected: [][]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GroupAnagramsV2(tt.input)
			if !reflect.DeepEqual(normalizeGroups(result), normalizeGroups(tt.expected)) {
				t.Fatalf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

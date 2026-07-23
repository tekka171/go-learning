package training

import (
	"reflect"
	"testing"
)

func TestReverseString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "basic case", input: "hello", expected: "olleh"},
		{name: "palindrome", input: "racecar", expected: "racecar"},
		{name: "single character", input: "a", expected: "a"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ReverseStringV3(tt.input)
			if result != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestIsAnagram(t *testing.T) {
	tests := []struct {
		name     string
		input1   string
		input2   string
		expected bool
	}{
		{name: "basic anagram", input1: "hello", input2: "olleh", expected: true},
		{name: "different length", input1: "apple", input2: "apples", expected: false},
		{name: "not anagram", input1: "hello", input2: "world", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsAnagramV3(tt.input1, tt.input2)
			if result != tt.expected {
				t.Fatalf("expected %t, got %t", tt.expected, result)
			}
		})
	}
}

func TestFirstNonRepeatingCharacter(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected rune
	}{
		{name: "simple case", input: "loveleetcode", expected: 'v'},
		{name: "another case", input: "aabbcc", expected: '\x00'},
		{name: "single character", input: "a", expected: 'a'},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FirstNonRepeatingCharacter(tt.input)
			if result != tt.expected {
				t.Fatalf("expected %c, got %c", tt.expected, result)
			}
		})
	}
}

func TestRemoveDuplicatesFromSortedArrayInPlace(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected int
	}{
		{name: "simple case", input: []int{1, 1, 2, 2, 3, 3}, expected: 3},
		{name: "simple case 2", input: []int{1, 1, 2, 2, 3, 3, 4}, expected: 4},
		{name: "empty case", input: []int{}, expected: 0},
		{name: "single character", input: []int{2}, expected: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RemoveDuplicatesFromSortedArrayInPlace(tt.input)
			if result != tt.expected {
				t.Fatalf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestMergeTwoSortedArraysInPlace(t *testing.T) {
	tests := []struct {
		name     string
		nums1    []int
		m        int
		nums2    []int
		n        int
		expected []int
	}{
		{name: "basic merge", nums1: []int{1, 2, 3, 0, 0, 0}, m: 3, nums2: []int{2, 5, 6}, n: 3, expected: []int{1, 2, 2, 3, 5, 6}},
		{name: "empty second array", nums1: []int{1}, m: 1, nums2: []int{}, n: 0, expected: []int{1}},
		{name: "negative numbers", nums1: []int{-1, 0, 0, 0}, m: 2, nums2: []int{-2, -1}, n: 2, expected: []int{-2, -1, -1, 0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			merged := append([]int(nil), tt.nums1...)
			MergeTwoSortedArraysInPlaceV2(merged, tt.m, tt.nums2, tt.n)

			if !reflect.DeepEqual(merged[:tt.m+tt.n], tt.expected) {
				t.Fatalf("expected %v, got %v", tt.expected, merged[:tt.m+tt.n])
			}
		})
	}
}

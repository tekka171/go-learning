package training

import (
	"slices"
	"unicode"
)

func ValidPalindrome(s string) bool { //brute force
	var cleaned []rune
	for _, char := range s {
		if unicode.IsLetter(char) || unicode.IsDigit(char) {
			cleaned = append(cleaned, unicode.ToLower(char))
		}
	}

	var reversed []rune
	// for i := range slices.Backward(cleaned) { //using slices.Backward
	for i := len(cleaned) - 1; i >= 0; i-- {
		reversed = append(reversed, cleaned[i])
	}

	return string(cleaned) == string(reversed)
}

func ValidPalindromeV2(s string) bool { //two pointer
	left := 0
	right := len(s) - 1

	for left < right {
		leftChar := rune(s[left])
		if !unicode.IsLetter(leftChar) && !unicode.IsDigit(leftChar) {
			left++
			continue
		}

		rightChar := rune(s[right])
		if !unicode.IsLetter(rightChar) && !unicode.IsDigit(rightChar) {
			right--
			continue
		}

		if unicode.ToLower(leftChar) != unicode.ToLower(rightChar) {
			return false
		}

		left++
		right--
	}

	return true
}

func ValidPalindromeV3(s string) bool { //two pointer, no rune convert, ASCII only
	left := 0
	right := len(s) - 1

	for left < right {
		if !isAlphanumericCharacter(s[left]) {
			left++
			continue
		}

		if !isAlphanumericCharacter(s[right]) {
			right--
			continue
		}

		if toLower(s[left]) != toLower(s[right]) {
			return false
		}

		left++
		right--
	}

	return true
}

func isAlphanumericCharacter(c byte) bool {
	if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
		return true
	}

	return false
}

func toLower(c byte) byte {
	if c >= 'A' && c <= 'Z' {
		return c + ('a' - 'A')
	}

	return c
}

func ContainerWithMostWater(heights []int) int { //Brute force
	var res int

	for i := range heights {
		for j := i; j < len(heights); j++ {
			width := j - i
			height := min(heights[i], heights[j])
			area := width * height
			if res < area {
				res = area
			}
		}
	}

	return res
}

func ContainerWithMostWaterV2(heights []int) int { //Two pointers
	var res int

	left := 0
	right := len(heights) - 1
	for left < right {
		width := right - left
		height := min(heights[left], heights[right])
		area := width * height
		if res < area {
			res = area
		}

		if heights[left] < heights[right] { //move the shorter wall in hope that the next one would be higher, so might get a higher min height
			left++
		} else {
			right--
		}
	}

	return res
}

// Given an array of integers, find all unique triplets [a, b, c] such that a + b + c == 0 (no duplicate triplets in the result).
func ThreeSum(nums []int) [][]int { //Brute force
	var res [][]int

	//sort first to handle uniqueness
	slices.Sort(nums)

	resMap := map[[3]int]bool{}
	for i := range nums {
		for j := i + 1; j < len(nums)-1; j++ {
			for k := j + 1; k < len(nums); k++ {
				triplets := [3]int{nums[i], nums[j], nums[k]}
				if nums[i]+nums[j]+nums[k] == 0 && !resMap[triplets] {
					resMap[triplets] = true
					res = append(res, []int{nums[i], nums[j], nums[k]})
				}
			}
		}
	}

	return res
}

// Given an array of integers, find all unique triplets [a, b, c] such that a + b + c == 0 (no duplicate triplets in the result).
func ThreeSumV2(nums []int) [][]int { //Optimized & simpler, but space O(n)
	var res [][]int

	//sort first to handle uniqueness
	slices.Sort(nums)

	seen := map[[3]int]bool{}
	for i := 0; i < len(nums)-2; i++ {
		left, right := i+1, len(nums)-1
		for left < right {
			sum := nums[i] + nums[left] + nums[right]
			if sum == 0 {
				key := [3]int{nums[i], nums[left], nums[right]}
				//record match
				if !seen[key] {
					seen[key] = true
					res = append(res, []int{nums[i], nums[left], nums[right]})
				}

				//move inward
				left++
				right--
			} else if sum < 0 {
				//need bigger sum
				left++
			} else {
				//need smaller sum
				right--
			}
		}
	}

	return res
}

// Given an array of integers, find all unique triplets [a, b, c] such that a + b + c == 0 (no duplicate triplets in the result).
func ThreeSumV3(nums []int) [][]int { //Most optimized, space O(1), time O(n2)
	var res [][]int

	//sort first to handle uniqueness
	slices.Sort(nums)

	for i := 0; i < len(nums)-2; i++ {
		if i > 0 && nums[i] == nums[i-1] {
			continue
		}

		left, right := i+1, len(nums)-1
		for left < right {
			sum := nums[i] + nums[left] + nums[right]
			if sum == 0 {
				//record match
				res = append(res, []int{nums[i], nums[left], nums[right]})

				//move inward
				left++
				right--

				for left < right && nums[left] == nums[left-1] { //if the new left is same as prev left, skip it
					left++
				}

				for left < right && nums[right] == nums[right+1] { //if the new right is same as prev right, skip it
					right--
				}
			} else if sum < 0 {
				//need bigger sum
				left++
			} else {
				//need smaller sum
				right--
			}
		}
	}

	return res
}

type ListNode struct {
	Val  int
	Next *ListNode
}

// Given the head of a linked list, determine if it contains a cycle (some node's next pointer eventually loops back to a previous node instead of reaching nil).
func DetectCycleInLinkedList(listNode *ListNode) bool { //Brute force, track all visited, Space O(n)
	seen := map[*ListNode]bool{}
	currentNode := listNode
	for currentNode != nil {
		if seen[currentNode] {
			return true
		}

		seen[currentNode] = true
		currentNode = currentNode.Next
	}

	return false
}

// Given the head of a linked list, determine if it contains a cycle (some node's next pointer eventually loops back to a previous node instead of reaching nil).
func DetectCycleInLinkedListV2(listNode *ListNode) bool { //Optimized, slowNode moves one step every 2 iterations, Space O(1)
	slowNode, fastNode := listNode, listNode
	slowNodePace := 1
	for fastNode != nil {
		// fmt.Printf("slowNodePace:%d, slowNode:%+v, fastNode:%+v", slowNodePace, slowNode, fastNode)

		if slowNodePace%2 == 0 {
			slowNode = slowNode.Next
		}
		fastNode = fastNode.Next

		slowNodePace++

		if slowNode == fastNode {
			return true
		}
	}

	return false
}

// Given the head of a linked list, determine if it contains a cycle (some node's next pointer eventually loops back to a previous node instead of reaching nil).
func DetectCycleInLinkedListV3(listNode *ListNode) bool { //Optimized, fastNode moves twice as fast, Space O(1)
	slowNode, fastNode := listNode, listNode
	for fastNode != nil && fastNode.Next != nil {
		slowNode = slowNode.Next
		fastNode = fastNode.Next.Next

		if slowNode == fastNode {
			return true
		}
	}

	return false
}

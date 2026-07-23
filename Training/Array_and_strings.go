package training

import "math"

func ReverseString(s string) string { //less efficient, more memory allocation
	var res string

	for i := len(s) - 1; i >= 0; i-- {
		res += string(s[i])
	}

	return res
}

func ReverseStringV2(s string) string { //more efficient, less memory allocation
	res := make([]rune, 0, len(s))

	for i := len(s) - 1; i >= 0; i-- {
		res = append(res, rune(s[i]))
	}

	return string(res)
}

func ReverseStringV3(s string) string { //most efficient, least memory allocation
	res := []rune(s)

	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		res[i], res[j] = res[j], res[i]
	}

	return string(res)
}

func IsAnagram(s1 string, s2 string) bool { //2 map, 3 loops
	lenS1 := len(s1)
	lenS2 := len(s2)

	if lenS1 != lenS2 {
		return false
	}

	mapS1 := make(map[rune]int, lenS1)
	mapS2 := make(map[rune]int, lenS2)
	for i, _ := range s1 {
		mapS1[rune(s1[i])]++
		mapS2[rune(s2[i])]++
	}

	for kS1, vS1 := range mapS1 {
		vS2, ok := mapS2[kS1]
		if !ok {
			return false
		}

		if vS1 != vS2 {
			return false
		}
	}

	return true
}

func IsAnagramV2(s1 string, s2 string) bool { //technically still O(1) space, 3 loops
	if len(s1) != len(s2) {
		return false
	}

	mapString := make(map[rune]int, len(s1))
	for _, v := range s1 {
		mapString[rune(v)]++
	}

	for _, v := range s2 {
		mapString[rune(v)]--
	}

	for _, v := range mapString {
		if v != 0 {
			return false
		}
	}

	return true
}

func IsAnagramV3(s1 string, s2 string) bool { //true O(1) space, 3 loops
	if len(s1) != len(s2) {
		return false
	}

	var arrayString [26]int
	for _, v := range s1 {
		arrayString[v-'a']++
	}

	for _, v := range s2 {
		arrayString[v-'a']--
	}

	for _, v := range arrayString {
		if v != 0 {
			return false
		}
	}

	return true
}

func FirstNonRepeatingCharacter(s string) rune {
	var res rune

	charCount := map[rune]int{}
	for _, v := range s {
		charCount[v]++
	}

	for _, v := range s {
		if charCount[v] == 1 {
			return v
		}
	}

	return res
}

func RemoveDuplicatesFromSortedArrayInPlace(input []int) int {
	if len(input) == 0 {
		return 0
	}

	writeIndex := 0
	for readIndex := 1; readIndex < len(input); readIndex++ { // 1,1,2,2,3,3
		//writeIndex = 0, readIndex = 1, skip, readIndex++
		//writeIndex = 0, readIndex = 2, pass,  writeIndex++, input[0] = 2, readIndex++
		//writeIndex = 1, readIndex = 3, skip
		//writeIndex = 1, readIndex = 4, pass, writeIndex++, input[1] = 3,  readIndex++
		//writeIndex = 2, readIndex = 5, skip

		if input[writeIndex] != input[readIndex] {
			writeIndex++
			input[writeIndex] = input[readIndex]
		}
	}

	return writeIndex + 1
}

func MergeTwoSortedArraysInPlace(nums1 []int, m int, nums2 []int, n int) { //check value
	for i := len(nums1) - 1; i >= 0; i-- {
		lastNum1 := math.MinInt
		if m > 0 {
			lastNum1 = nums1[m-1]
		}

		lastNum2 := math.MinInt
		if n > 0 {
			lastNum2 = nums2[n-1]
		}

		if lastNum1 > lastNum2 {
			nums1[i] = lastNum1
			m--
		} else {
			nums1[i] = lastNum2
			n--
		}
	}
}

func MergeTwoSortedArraysInPlaceV2(nums1 []int, m int, nums2 []int, n int) { //check index
	i1 := m - 1
	i2 := n - 1
	for i := len(nums1) - 1; i >= 0; i-- {
		if i2 < 0 || (i1 >= 0 && nums1[i1] > nums2[i2]) { //nums2 will always be smaller length
			nums1[i] = nums1[i1]
			i1--
		} else {
			nums1[i] = nums2[i2]
			i2--
		}
	}
}

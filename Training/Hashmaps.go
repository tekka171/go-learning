package training

import (
	"slices"
	"strings"
)

func TwoSum(nums []int, target int) []int { //brute-force
	for i := range nums {
		for j := i + 1; j < len(nums); j++ {
			if nums[i]+nums[j] == target {
				return []int{i, j}
			}
		}
	}

	return []int{}
}

func TwoSumV2(nums []int, target int) []int { //hashmap
	res := map[int]int{}
	for i := range nums {
		diff := target - nums[i]
		if v, ok := res[diff]; ok {
			return []int{v, i}
		}
		res[nums[i]] = i
	}

	return []int{}
}

func FindDuplicates(nums []int) []int {
	duplicatesMap := map[int]int{}
	for i := range nums {
		duplicatesMap[nums[i]]++
	}

	res := []int{}
	for k, v := range duplicatesMap {
		if v > 1 {
			res = append(res, k)
		}
	}

	return res
}

func GroupAnagrams(input []string) [][]string { //Sorted string as key
	res := [][]string{}

	groupedString := map[string][]string{}
	for i := range input {
		key := []rune(input[i])
		slices.Sort(key)

		groupedString[string(key)] = append(groupedString[string(key)], input[i])
	}

	for _, v := range groupedString {
		res = append(res, v)
	}

	return res
}

func GroupAnagramsV2(input []string) [][]string { //Frequency count as key
	res := [][]string{}

	groupedString := map[[26]int][]string{}
	for i := range input {
		key := [26]int{}
		inputStr := strings.ToLower(input[i]) //optional: to handle uppercase
		for j := range inputStr {
			key[inputStr[j]-'a']++
		}

		groupedString[key] = append(groupedString[key], input[i])
	}

	for _, v := range groupedString {
		res = append(res, v)
	}

	return res
}

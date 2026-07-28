package training

// Problem: Compute n! (factorial) — the product of all positive integers up to n.
// Example:
// 5! = 5 × 4 × 3 × 2 × 1 = 120
// 0! = 1 (by definition)
func Factorial(n int) int { //Recursive, but Space O(n), since every recursive is a new stack
	if n == 0 {
		return 1
	}

	return Factorial(n-1) * n
}

func FactorialIterative(n int) int { //Iterative, but Space O(1)
	res := 1

	for i := n; i > 0; i-- {
		res *= i
	}

	return res
}

// Problem: Compute the nth Fibonacci number, where:
//
// F(0) = 0
// F(1) = 1
// F(n) = F(n-1) + F(n-2) for n >= 2
//
// Example: 0, 1, 1, 2, 3, 5, 8, 13, 21, ...
// F(6) = 8
func Fibonacci(n int) int { //Time O(2^n), Space O(n), since every recursive is a new stack
	if n == 0 {
		return 0
	}

	if n == 1 {
		return 1
	}

	return Fibonacci(n-1) + Fibonacci(n-2)
}

func FibonacciMemo(n int, memo map[int]int) int { //Time O(n), Space O(n), recursive + caching
	if val, ok := memo[n]; ok {
		return val
	}

	if n == 0 {
		return 0
	}

	if n == 1 {
		return 1
	}

	res := FibonacciMemo(n-1, memo) + FibonacciMemo(n-2, memo)
	memo[n] = res

	return res
}

func FibonacciIterative(n int) int { //Time O(n), Space O(1), no recursion, no stack growth
	if n == 0 {
		return 0
	}

	if n == 1 {
		return 1
	}

	prev, curr := 0, 1
	for i := 2; i <= n; i++ {
		prev, curr = curr, prev+curr
	}

	return curr
}

// Problem: Given a set of distinct integers, return all possible subsets (the power set) — including the empty set and the set itself.
//
// Example:
//
// Input: [1, 2, 3]
// Output: [[], [1], [2], [1,2], [3], [1,3], [2,3], [1,2,3]] (order doesn't matter)
func GenerateAllSubsets(nums []int) [][]int { //Backtrack pattern
	var res [][]int
	var current []int

	var backtrack func(index int)
	backtrack = func(index int) {
		// base case here
		if index == len(nums) {
			subset := make([]int, len(current))
			copy(subset, current)
			res = append(res, subset)
			return
		}

		// decision 1: exclude
		backtrack(index + 1)

		// decision 2: include, recurse, then undo
		current = append(current, nums[index])
		backtrack(index + 1)
		current = current[:len(current)-1]
	}

	backtrack(0)

	return res
}

// Problem: Given the head of a singly linked list, reverse it — so 1 → 2 → 3 → nil becomes 3 → 2 → 1 → nil. Return the new head.
func ReverseLinkedList(head *ListNode) *ListNode {
	if head == nil {
		return nil
	}

	if head.Next == nil {
		return head
	}

	newHead := ReverseLinkedList(head.Next)
	head.Next.Next = head
	head.Next = nil
	return newHead

	//dive
	//ReverseLinkedList(1) -> ReverseLinkedList(2) - > ReverseLinkedList(3) - > ReverseLinkedList(4) -> ReverseLinkedList(nil)
	//go up
	//ReverseLinkedList(4) -> base case -> return 4 ---> {4-nil}
	//ReverseLinkedList(3) -> 4.next = 3; 3.next = nil -> return 4 ---> {4-3-nil}
	//ReverseLinkedList(2) -> 3.next = 2; 2.next = nil -> return 4 ---> {4-3-2-nil}
	//ReverseLinkedList(1) -> 2.next = 1; 1.next = nil -> return 4 ---> {4-3-2-1-nil}
}

// Problem: Given the head of a singly linked list, reverse it — so 1 → 2 → 3 → nil becomes 3 → 2 → 1 → nil. Return the new head.
func ReverseLinkedListIterative(head *ListNode) *ListNode {
	var prevNode *ListNode
	currNode := head
	for currNode != nil {
		nextNode := currNode.Next
		currNode.Next = prevNode
		prevNode = currNode
		currNode = nextNode
	}

	return prevNode

	// 1 2 3 4 nil, prevNode = nil, currNode = 1
	// loop:
	// nextNode = 2, 1.Next = nil, prevNode = 1, currNode = 2 ---> {1->nil}
	// nextNode = 3, 2.Next = 1, prevNode = 2, currNode = 3 ---> {2->1->nil}
	// nextNode = 4, 3.Next = 2, prevNode = 3, currNode = 4 ---> {3->2->1->nil}
	// nextNode = nil, 4.Next = 3, prevNode = 4, currNode = nil ---> {4->3->2->1->nil}
}

// sum of digits of a number (e.g., 1234 → 1+2+3+4=10)
func SumDigits(n int) int { //the last digit of 'n' + the sum of all digits in whats left after removing the last digit
	if n == 0 {
		return 0
	}
	return SumDigits(n/10) + n%10
}

func SumDigitsIterative(n int) int {
	sum := 0
	for n > 0 {
		sum += n % 10
		n /= 10
	}

	return sum
}

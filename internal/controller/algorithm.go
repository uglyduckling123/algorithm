package controller

import "C"
import (
	"fmt"
	"mango/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AlgorithmHandler struct {
	UserService service.UserService
}

func NewAlgorithmHandler(userService service.UserService) *AlgorithmHandler {
	return &AlgorithmHandler{
		UserService: userService,
	}
}

// Register 注册路由
func (v *AlgorithmHandler) Register(router *gin.RouterGroup) {
	userRouter := router.Group("/algorithm")
	{
		userRouter.POST("/sumUpToTarget", v.sumUpToTarget)
		userRouter.POST("/sumUpToTargetHashMap", v.sumUpToTargetHashMap)
		userRouter.POST("/buySold", v.buySold)
		userRouter.POST("/maxSubarray", v.maxSubarray)
		userRouter.POST("/findMinimumInRotatedArray", v.findMinimumInRotatedArray)
		userRouter.POST("/containerWithMostWater", v.containerWithMostWater)
		userRouter.POST("/numberOfOneBits", v.numberOfOneBits)
		userRouter.POST("/climbStairs", v.climbStairs)
		userRouter.POST("/coinCharge", v.coinCharge)
		userRouter.POST("/longestIncreasingSubsequence", v.longestIncreasingSubsequence)
		userRouter.POST("/twoStringLongestCommonSubsequence", v.twoStringLongestCommonSubsequence)
		userRouter.POST("/backTrack", v.backTrack)
		userRouter.POST("/rubHouse", v.rubHouse)
		userRouter.POST("/decodeLetter", v.decodeLetter)
		userRouter.POST("/cloneGraph", v.cloneGraph)
		userRouter.POST("/courseTopology", v.courseTopology)

	}
}

type SumUpToTargetRequest struct {
	Target int `json:"target" form:"target" binding:"required"`
}

/*
given an array of integers, return indexes of two numbers that sum up to the request target
array :=[]int{1,3,4,5}  target := 6  return 0,3
*/
func (v *AlgorithmHandler) sumUpToTarget(c *gin.Context) {
	var request SumUpToTargetRequest
	err := c.ShouldBind(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, nil)
		return
	}
	givenArray := []int{1, 3, 4, 5, 8}
	for i := 0; i < len(givenArray); i++ {
		for j := i + 1; j < len(givenArray); j++ {
			if givenArray[i]+givenArray[j] == request.Target {
				fmt.Printf("find index i:%d and j:%d\n", i, j)
				c.JSON(http.StatusOK, gin.H{
					"found":   true,
					"indices": []int{i, j},
					"values":  []int{givenArray[i], givenArray[j]},
				})
				return
			}
		}
	}
	c.JSON(http.StatusCreated, nil)
}

/*
given an array of integers, return indexes of two numbers that sum up to the request target
array :=[]int{1,3,4,5}  target := 6  return 0,3
*/
func (v *AlgorithmHandler) sumUpToTargetHashMap(c *gin.Context) {
	var request SumUpToTargetRequest
	err := c.ShouldBind(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, nil)
		return
	}
	givenArray := []int{1, 3, 4, 5, 8}
	hashMap := make(map[int]int, len(givenArray))
	for key, value := range givenArray {
		hashMap[value] = key
	}
	for i := 0; i < len(givenArray); i++ {
		findTarget := request.Target - givenArray[i]
		if value, exist := hashMap[findTarget]; exist {
			fmt.Printf("find index i:%d and j:%d\n", i, value)
			c.JSON(http.StatusOK, gin.H{
				"found":   true,
				"indices": []int{i, value},
				"values":  []int{givenArray[i], givenArray[value]},
			})
			return
		}
	}
	c.JSON(http.StatusCreated, nil)
}

/*
slide window
given an array  the elements are the price of stock one day, you can only buy and sell once to find
maximum the profit

	array :=[]int{7,1,5,3,6,4}   the profit is 5(1,6)
*/
func (v *AlgorithmHandler) buySold(c *gin.Context) {
	givenArray := []int{7, 1, 5, 3, 6, 4}
	var leftPoint, rightPoint int
	var maxProfit int
	for key, _ := range givenArray {
		rightPoint = key
		if givenArray[leftPoint] < givenArray[rightPoint] {
			maxProfit = max(maxProfit, givenArray[rightPoint]-givenArray[leftPoint])
		} else {
			leftPoint = rightPoint
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"maxProfit": maxProfit,
		"leftPoint": leftPoint,
	})
}

/*
leetcode 53
slide window
given an integer array  ,find the contiguous subarray that has the largest sum and return the sum

	array :=[]int{1, -4, 2, 3, 6, 4}   the profit is 2, 3, 6, 4 = 13
*/
func (v *AlgorithmHandler) maxSubarray(c *gin.Context) {
	givenArray := []int{1, -4, 2, 3, 6, 4}
	var sum int

	for _, value := range givenArray {
		sum += value
		if sum <= 0 {
			sum = 0
		}

	}
	c.JSON(http.StatusOK, gin.H{
		"maxProfit": sum,
	})
}

/*
leetcode 153 二分查找
假设一个按照升序排列的数组在某个点上进行了旋转（例如，数组 [0,1,2,4,5,6,7] 可能变成 [4,5,6,7,0,1,2]）。请找出其中最小的元素。
*/
func (v *AlgorithmHandler) findMinimumInRotatedArray(c *gin.Context) {
	givenArray := []int{4, 5, 6, 7, 0, 1, 2}
	var left int
	var right = len(givenArray) - 1

	for left < right {
		mid := left + (right-left)/2
		// 如果中间元素大于最右元素，说明最小值在右半部分
		if givenArray[mid] > givenArray[right] {
			left = mid + 1
		} else {
			// 否则最小值在左半部分（包括中间元素）
			right = mid
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"maxProfit": givenArray[left],
	})
}

/*
leetcode 11 双指针
container with the most water
双指针 左指针在最左，右指针在最右，比较两边指针值的高度，如果左边高度比右边高则移动右指针，相反则移动左指针。移动指针宽度肯定变小，
移动高度低的 有可能面积会变大，所以要移动高度低的指针
*/
func (v *AlgorithmHandler) containerWithMostWater(c *gin.Context) {
	givenArray := []int{1, 5, 6, 7, 4, 3, 4}
	var left int
	var right = len(givenArray) - 1
	var maxArea int
	for left < right {
		if givenArray[left] < givenArray[right] {
			maxAreaNew := (right - left) * givenArray[left]
			maxArea = max(maxArea, maxAreaNew)
			left++
		} else {
			maxAreaNew := (right - left) * givenArray[right]
			maxArea = max(maxArea, maxAreaNew)
			right--
		}

	}

	//brute force
	var maxAreaBF int
	for i := 0; i < len(givenArray); i++ {
		for j := 1; j < len(givenArray); j++ {
			if givenArray[i] < givenArray[j] {
				maxAreaNew := (j - i) * givenArray[i]
				maxAreaBF = max(maxAreaBF, maxAreaNew)
			} else {
				maxAreaNew := (j - i) * givenArray[j]
				maxAreaBF = max(maxAreaBF, maxAreaNew)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"maxArea":   maxArea,
		"maxAreaBF": maxAreaBF,
	})
}

/*
leetcode 191
given an integer and  return the the number of 1 bits
1:dynamic programming
2:逐位检查
*/
func (v *AlgorithmHandler) numberOfOneBits(c *gin.Context) {
	//1：dynamic programming n&(n-1) 将数字最右边的1变成0
	dynamic := 23
	var number int
	for dynamic != 0 {
		dynamic = dynamic & (dynamic - 1)
		number++
	}

	//2：逐位检查
	request := 23
	var count int
	for request > 0 {
		if request&1 != 0 {
			count++
		}
		//向又移动一位
		request >>= 1
	}
	c.JSON(http.StatusOK, gin.H{
		"count":  count,
		"number": number,
	})
}

/*
leetcode 70 climb stairs

	it takes n steps to reach the top  ,every time  you can climb either one or two steps  ,how many

ways can  you climb to  the top?
*/
func (v *AlgorithmHandler) climbStairs(c *gin.Context) {
	/*1：fn(n) 是n个台阶 有多少种方法的函数
		fn(n) = fn(n-1)+fn(n-2)
		fn(n-1) = fn(n-2)-fn(n-3)
		fn(n-2) = fn(n-3)-fn(n-4)
	    如果用递归计算 会存在很多重复的计算 时间复杂度大概为 2^n*/
	result := fibonacci(5)

	//2:可以用dynamic programming 处理  extra space O(n)  runtime O(N)
	//fn(1)=1 fn(2)=2 fn(3)=f(1)+fn(3) fn(n)=fn(n-1)+fn(n-2)
	n := 5
	dp := make([]int, n+1)
	dp[1] = 1
	dp[2] = 2
	for i := 3; i <= n; i++ {
		dp[i] = dp[i-1] + dp[i-2]
	}
	dynamic := dp[n]

	//3:还有一种空间复杂度为 O(1)
	A := 1
	B := 2
	for i := 3; i <= n; i++ {
		A, B = B, (A + B)
	}
	c.JSON(http.StatusOK, gin.H{
		"fibonacci": result,
		"dynamic":   dynamic,
		"B":         B,
	})
}
func fibonacci(n int) int {
	if n <= 1 {
		return n
	}
	if n <= 2 {
		return n
	}
	return fibonacci(n-1) + fibonacci(n-2)
}

/*
leetcode 322 背包问题  动态规划
动态规划：考虑所有可能性，可以回头看，类似走迷宫，一般会记录中间状态，一点是最优解）
贪心算法：一条路走到黑，不一定是最优解，只考虑当前最优解，一般不需要记录中间状态。可以从后往前推，也可以从前往后推，有终止值
一个切片 里边包含 各种整数的零钱，给一个值amount，返回相加等于m的最少元素个数
想要找到fn(m)的最小值，最后一步一定要取一个零钱 这个零钱可能是任何一个
fn(m) = min(fn(m),fn(m-1)+1,fn(m-2)+1,....) 知道对应的最小值就可以得出结果
*/
func (v *AlgorithmHandler) coinCharge(c *gin.Context) {
	amount := 10
	dp := make([]int, amount+1)
	for i := range dp {
		dp[i] = amount + 1
	}
	dp[0] = 0
	var charge = []int{1, 2, 5, 2, 5, 10}
	for i := 1; i < len(dp); i++ {
		for _, value := range charge {
			if i >= value {
				dp[i] = min(dp[i], dp[amount-value]+1)
			}
		}
	}
	// 检查是否能凑齐目标金额
	if dp[amount] > amount {
		dp[amount] = -1
	}

	c.JSON(http.StatusOK, gin.H{
		"coinCharge": dp[amount],
	})
}

/*
leetcode 300   动态规划
given an integer array nums, find the longest increasing subsequence
Input: nums = [10,9,2,5,3,7,101,18]
Output: 4   [2,3,7,101]
*/
func (v *AlgorithmHandler) longestIncreasingSubsequence(c *gin.Context) {
	array := []int{10, 9, 2, 5, 3, 7, 101, 18}
	var fn = make(map[int]int, len(array))
	for k, _ := range array {
		fn[k] = 1
	}
	var maxLength int
	for i := 0; i < len(array); i++ {
		for j := 0; j <= i; j++ {
			if array[i] > array[j] {
				fn[i] = max(fn[i], fn[j]+1)
			}
		}
		maxLength = max(maxLength, fn[i])
	}
	c.JSON(http.StatusOK, gin.H{
		"longestIncreasingSubsequence": maxLength,
	})
}

/*
leetcode 1143   动态规划
given two string text1 text2, return the common longest  subsequence length
Input: text1 = Input: text1 = "abcde", text2 = "ace"
Output: 3
假如 text1 的长度是 i  text2的长度是 j
dp[i][j] 代表 text1第i个字符 text2第j个字符 最长相同的长度
如果 最后一位相同 那么 dp[i][j]=dp[i-1][j-1]+1
如果最后一位 不相同 dp[i][j] = max(dp[i][j-1],dp[i-1][j])
*/
func (v *AlgorithmHandler) twoStringLongestCommonSubsequence(c *gin.Context) {
	text1 := "abcde"
	text2 := "ace"
	m, n := len(text1), len(text2)

	//初始化值
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if text1[i-1] == text2[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else {
				dp[i][j] = max(dp[i][j-1], dp[i-1][j])
			}
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"twoStringLongestCommonSubsequence": dp[m][n],
	})
}

/*
leetcode 39   backTrack
three key points
one : 路径path:记录作出选择的路径
two: 选择要素：当时可以做选择的要素 candidates[start:]
tree:结束条件：找到结果或者超过结果

			func backtrack(path,candidate){
			    if 满足条件{
			       result.add(路径)
		          return
			    }
			    for 选择列表 {
		            选择要素
		            backTrack(path,candidate)
	                撤销选择要素
		        }
			}
*/
func (v *AlgorithmHandler) backTrack(c *gin.Context) {
	candidates1 := []int{2, 3, 6, 7}
	target1 := 7
	result := [][]int{}
	path := []int{}
	//必包函数 必须先声明 再赋值不然 函数内不能调用
	var backTrackDetail func(start, currentSum int)
	backTrackDetail = func(start int, currentSum int) {
		// 终止条件
		if currentSum == target1 {
			// 需要深拷贝，否则后续修改会影响结果
			temp := make([]int, len(path))
			copy(temp, path)
			result = append(result, temp)
			return
		}
		if currentSum > target1 {
			return // 剪枝
		}
		for i := start; i < len(candidates1); i++ {
			//选择元素
			path = append(path, candidates1[i])
			//关键：传入 i 而不是 i+1，因为可以重复选择
			backTrackDetail(i, candidates1[i]+currentSum)
			path = path[:len(path)-1] //回溯
		}
	}
	backTrackDetail(0, 0)

	c.JSON(http.StatusOK, gin.H{
		"backTrack": result,
	})
}

/*
leetcode 198  dynamic programming
given an integer array ,each term represent the money that the house has , if two near  house were  broken
the alert system will trigger , return the most money the thief can rub
*/
func (v *AlgorithmHandler) rubHouse(c *gin.Context) {
	house := []int{2, 7, 9, 3, 1}
	var sum int
	dp := make([]int, len(house))
	dp[0] = 2
	dp[1] = 7
	for i := 2; i < len(house); i++ {
		dp[i] = max(dp[i-1], dp[i-2]+house[i])
		sum = max(sum, dp[i])
	}
	c.JSON(http.StatusOK, gin.H{
		"rubHouse": sum,
	})
}

/*
leetcode 91  dynamic programming
given a string which is made of (0-9) number , decode the numbers to letters  such as '1'->A ,'2'->'B' '26'->Z ，
how many ways it can decode into
*/
func (v *AlgorithmHandler) decodeLetter(c *gin.Context) {
	var text = "226"
	dp := make([]int, len(text)+1)
	dp[0] = 1
	dp[1] = 1
	for i := 2; i <= len(text); i++ {
		//singe letter
		if text[i-1] != '0' {
			dp[i] += dp[i-1]
		}
		// 双数字解码 - 修正字符转换
		twoDigit := int(text[i-2]-'0')*10 + int(text[i-1]-'0')
		if twoDigit >= 10 && twoDigit <= 26 {
			dp[i] += dp[i-2]
		}

	}
	c.JSON(http.StatusOK, gin.H{
		"decodeLetter": dp[len(text)],
	})
}

/*
leetcode 133 cloneGraph
dfs : 深度遍历 主要用递归
bfs   遍历相邻节点  队列 入队 出队
*/
func (v *AlgorithmHandler) cloneGraph(c *gin.Context) {
	node := createTestGraph()
	visited := make(map[*Node]*Node, 0)
	cloneNode := cloneGraphDFS(node, visited)
	bfs := cloneGraphBFS(node)
	c.JSON(http.StatusOK, gin.H{
		"node": len(node.Neighbors),
		"dfs":  len(cloneNode.Neighbors),
		"bfs":  len(bfs.Neighbors),
	})
}

type Node struct {
	Val       int
	Neighbors []*Node
}

func createTestGraph() *Node {
	node1 := &Node{Val: 1}
	node2 := &Node{Val: 2}
	node3 := &Node{Val: 3}
	node4 := &Node{Val: 4}

	node1.Neighbors = []*Node{node2, node4}
	node2.Neighbors = []*Node{node1, node3}
	node3.Neighbors = []*Node{node2, node4}
	node4.Neighbors = []*Node{node1, node3}

	return node1
}

func cloneGraphDFS(node *Node, visited map[*Node]*Node) *Node {
	//查看是否访问过
	if value, exist := visited[node]; exist {
		//返回新的值clone
		return value
	}
	clone := &Node{
		Val:       node.Val,
		Neighbors: make([]*Node, 0),
	}
	// key 为原始值 node ，value 为新clone
	visited[node] = clone
	for _, value := range node.Neighbors {
		neighbors := cloneGraphDFS(value, visited)
		clone.Neighbors = append(clone.Neighbors, neighbors)
	}
	return clone
}

func cloneGraphBFS(node *Node) *Node {
	visited := make(map[*Node]*Node)
	//1首先初始化一个值
	visited[node] = &Node{
		Val:       node.Val,
		Neighbors: []*Node{},
	}
	//2然后压入队列
	queue := []*Node{node}
	for len(queue) > 0 {
		//3：出队列
		currentQueue := queue[0]
		queue = queue[1:]
		//4：遍历出队列元素 子节点
		for _, value := range currentQueue.Neighbors {
			//5：子节点不存在 依次压入队列
			if _, exist := visited[value]; !exist {
				queue = append(queue, value)
				//5：添加访问过的子节点
				visited[value] = &Node{Val: value.Val, Neighbors: []*Node{}}
			}
			//赋值当前节点的 临节点
			visited[currentQueue].Neighbors = append(visited[currentQueue].Neighbors, &Node{Val: value.Val, Neighbors: []*Node{}})
		}

	}
	return visited[node]
}

/*
leetcode 207 courseTopology  拓扑排序
本学期有 m 门课程要修，但是课程之前 有限制
先修课程按数组 prerequisites 给出，
其中 prerequisites[i] = [ai, bi] ，表示如果要学习课程 ai 则必须先学习课程 bi。b->a
问 能否学习完 m 门课程
解题思路：
1：构建图：构建prerequisites 课程间的关联关系
2：计算入度：计算每门课程的前置数量
3：bfs 遍历 ：从入度为0 的课程开始
4：判断是否有环，无环  能全部学完
*/
func (v *AlgorithmHandler) courseTopology(c *gin.Context) {
	var coursesNum = 4
	prerequisites := [][]int{{1, 0}, {2, 0}, {3, 1}, {3, 2}}
	//1构建图
	graphy := make(map[int][]int, 0)
	//2:构建入度
	inDegree := make(map[int]int, 0)
	for _, value := range prerequisites {
		// value[1] → value[0]：必须先修value[1]才能修value[0]
		graphy[value[1]] = append(graphy[value[1]], value[0])
		inDegree[value[0]]++
	}
	//3: bfs 遍历 ：3从入度为0 的课程开始
	//初始化队列  把入度为0 的压入队列
	queue := make([]int, 0)
	for i := 0; i < coursesNum; i++ {
		if inDegree[i] == 0 {
			queue = append(queue, i)
		}

	}
	var visited = make([]int, 0)
	//3: bfs 遍历 ：从入度为0 的课程开始
	for len(queue) > 0 {
		//出队列
		currentQueue := queue[0]
		visited = append(visited, currentQueue)

		queue = queue[1:]
		//如果有相邻节点 相邻节点 入度-1后如果 为0 压入
		if _, exist := graphy[currentQueue]; exist {
			for _, value := range graphy[currentQueue] {
				inDegree[value] = inDegree[value] - 1
				if inDegree[value] == 0 {
					queue = append(queue, value)
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"coursesNum": coursesNum,
		"visited":    len(visited),
	})
}

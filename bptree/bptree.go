package bptree

import (
	"container/list"
	"errors"
	"fmt"
	"reflect"
)

/*
* B+ Tree说明:
* 在B+树中的节点通常被表示为一组有序的元素和子指针。
* 如果此B+树的阶数是m，则除了根之外的每个节点都包含最少 m/2 个元素最多 m-1 个元素，对于任意的结点有最多 m 个子指针。
* 对于所有内部节点，子指针的数目总是比元素的数目多一个。所有叶子都在相同的高度上，叶结点本身按关键字大小从小到大链接。
 */

const (
	ORDER     = 5         // 树的阶数:每个节点最大数据量=阶数-1
	MAX_LIMIT = ORDER - 1 // 节点的key数量上限
	MIN_LIMIT = ORDER / 2 // 节点的key数量下限
)

var (
	err        error
	splitIndex = 0        // 节点分裂的索引值(由order计算而得)
	queue      *list.List // 遍历Tree队列(链表) 只用于print调试
)

// Tree :B+树
type Tree struct {
	Root *Node
}

// Node :树节点
/*
*非叶子节点Pointers指向Node: len(Key) == len(Pointers)-1
*叶子节点Pointers指向数据Record: len(Key) == len(Pointers)
 */
type Node struct {
	Keys     []int         // 索引切片
	Pointers []interface{} // 指针切片 非叶子节点指向下一个节点 叶子节点指向数据Record
	Parent   *Node         // 父节点
	IsLeaf   bool          // 是否叶子节点
	Count    int           // 当前节点的key数量
	Next     *Node         // 叶子节点双向链表 下一节点指针
	Prev     *Node         // 叶子节点双向链表 上一节点指针
}

// Record :数据记录
type Record struct {
	Value []byte
}

// NewTree :构造函数 同时初始化分裂节点的index
func NewTree() *Tree {
	// 设置节点分裂的index(只与order相关)
	length := ORDER - 1
	if length%2 == 0 {
		splitIndex = length / 2
	} else {
		splitIndex = length/2 + 1
	}

	return &Tree{}
}

// *********************** Insert部分 ***********************
/*
1. 首先，查找要插入其中的节点的位置。接着把值插入这个节点中。
2. 如果没有节点处于违规状态则处理结束。
3. 如果某个节点有过多元素，则把它分裂为两个节点，每个都有最小数目的元素。
在树上递归向上继续这个处理直到到达根节点，如果根节点被分裂，则创建一个新根节点。
为了使它工作，元素的最小和最大数目典型的必须选择为使最小数不小于最大数的一半。
*/

// Insert :插入key
func (t *Tree) Insert(key int, value []byte) error {
	// Find没有err 说明key查找有返回 重复Insert
	if _, err := t.Find(key); err == nil {
		return errors.New("key already exists")
	}

	pointer, err := newRecord(value)
	if err != nil {
		return err
	}

	if t.Root == nil {
		return t.initRoot(key, pointer)
	}

	leaf := t.findLeaf(key)
	if leaf.Count < MAX_LIMIT {
		// 当前节点未排满 直接insert
		return t.insertIntoNode(leaf, key, pointer)
	}
	// 当前排满了 分裂然后insert
	return t.splitAndInsertIntoLeaf(leaf, key, pointer)
}

// 正常insert 直接插入一个数据至节点中
func (t *Tree) insertIntoNode(n *Node, key int, pointer interface{}) error {
	// 1.插入新记录
	newKeys, newPointers := setIntoNode(*n, key, pointer)
	n.Keys = newKeys
	n.Pointers = newPointers
	// 2.节点长度+1
	n.Count++
	return nil
}

// 分裂 然后insert至叶子节点
func (t *Tree) splitAndInsertIntoLeaf(leaf *Node, key int, pointer *Record) error {
	// 创建新Leaf
	newLeaf, _ := newLeaf()
	// 1.临时插入后的keys和pointers
	tempKeys, tempPointers := setIntoNode(*leaf, key, pointer)

	// 插入后的左节点(原节点)
	leaf.Keys = tempKeys[0:splitIndex]
	leaf.Pointers = tempPointers[0:splitIndex]
	leaf.Count = len(leaf.Keys)
	// 插入后的右节点
	newLeaf.Keys = tempKeys[splitIndex:]
	newLeaf.Pointers = tempPointers[splitIndex:]
	newLeaf.Count = len(newLeaf.Keys)
	newLeaf.Parent = leaf.Parent
	// 双向链表处理（链表指针都是分裂或者合并而来）
	leaf.Next = newLeaf
	newLeaf.Prev = leaf

	return t.insertIntoParent(leaf, tempKeys[splitIndex], newLeaf)
}

// 分裂后 插入父节点
func (t *Tree) insertIntoParent(left *Node, key int, right *Node) error {
	parent := left.Parent

	// 父节点为空 说明是根节点分裂 则拉高为新的根节点 结束
	if parent == nil {
		return t.insertIntoNewRoot(left, key, right)
	}

	// 父节点未填满 则直接塞入
	if parent.Count < MAX_LIMIT {
		t.insertIntoNode(parent, key, right)
		return nil
	}

	// 父节点塞满了 需要分裂然后插入
	return t.splitAndInsertIntoNode(parent, key, right)
}

// 分裂 然后insert至非叶子节点
func (t *Tree) splitAndInsertIntoNode(n *Node, key int, pointer *Node) error {
	// 创建新Node
	newNode, _ := newNode()
	// 1.临时插入后的keys和pointers
	tempKeys, tempPointers := setIntoNode(*n, key, pointer)

	// 插入后的左节点(原节点)
	n.Keys = tempKeys[0:splitIndex]
	n.Pointers = tempPointers[0 : splitIndex+1] //! 非叶子节点points数量比keys多1
	n.Count = len(n.Keys)
	// 插入后的右节点
	newNode.Keys = tempKeys[splitIndex+1:] // 这里分裂非叶节点 往parent插入的key 子节点不保留
	newNode.Pointers = tempPointers[splitIndex+1:]
	newNode.Count = len(newNode.Keys)
	newNode.Parent = n.Parent

	return t.insertIntoParent(n, tempKeys[splitIndex], newNode)
}

// 根节点分裂：需要创建一个新root
func (t *Tree) insertIntoNewRoot(left *Node, key int, right *Node) error {
	t.Root, err = newNode()
	if err != nil {
		return err
	}

	t.Root.Keys = append(t.Root.Keys, key)
	t.Root.Pointers = append(t.Root.Pointers, left)
	t.Root.Pointers = append(t.Root.Pointers, right)
	t.Root.Count++
	t.Root.Parent = nil
	left.Parent = t.Root
	right.Parent = t.Root
	return nil
}

// 初始化根节点
func (t *Tree) initRoot(key int, pointer *Record) error {
	t.Root, err = newLeaf()
	if err != nil {
		return err
	}
	t.Root.Keys = append(t.Root.Keys, key)
	t.Root.Pointers = append(t.Root.Pointers, pointer)
	// t.Root.Pointers[MAX_LIMIT] = nil
	t.Root.Parent = nil
	t.Root.Count++
	return nil
}

// 获取某个key在node中应该插入的位置
func (n *Node) getInsertIndex(key int) (i int) {
	for i < n.Count && n.Keys[i] < key {
		i++
	}
	return
}

// 获取叶子节点某key的Record
func (n *Node) getRecord(key int) *Record {
	if !n.IsLeaf {
		return nil
	}

	i := n.getKeyIndex(key)
	if i > -1 {
		return n.Pointers[i].(*Record)
	}
	return nil
}

// 获取节点某key的index
func (n *Node) getKeyIndex(key int) int {
	for i, k := range n.Keys {
		if k == key {
			return i
		}
	}
	return -1
}

// 获取节点某pointer的index
func (n *Node) getPointerIndex(pointer interface{}) int {
	for i, p := range n.Pointers {
		if p == pointer {
			return i
		}
	}
	return -1
}

// 塞入节点 返回新的keys和pointers
func setIntoNode(n Node, key int, pointer interface{}) ([]int, []interface{}) {
	// 1.找到插入点
	i := n.getInsertIndex(key)
	// 2.插入新记录 返回新keys和pointers(不改变原node数据,所以node不是指针)
	keys := append([]int{}, n.Keys[0:i]...)
	keys = append(keys, key)
	keys = append(keys, n.Keys[i:]...)
	// 非叶子节点 pointers的索引比keys多1
	if !n.IsLeaf {
		i++
	}
	pointers := append([]interface{}{}, n.Pointers[0:i]...)
	pointers = append(pointers, pointer)
	pointers = append(pointers, n.Pointers[i:]...)
	return keys, pointers
}

// *********************** Delete部分 ***********************
/*
1. 首先，查找要删除的值。接着从包含它的节点中删除这个值。
2. 如果没有节点处于违规状态则处理结束。
3. 如果节点处于违规状态则有两种可能情况：
	[1] 它的兄弟节点，就是同一个父节点的子节点，可以把一个或多个它的子节点转移到当前节点，而把它返回为合法状态。
	如果是这样，在更改父节点和两个兄弟节点的分离值之后处理结束。
	[2] 它的兄弟节点由于处在低边界上而没有额外的子节点。
	在这种情况下把两个兄弟节点合并到一个单一的节点中，而且我们递归到父节点上，因为它被删除了一个子节点。
	持续这个处理直到当前节点是合法状态或者到达根节点，在其上根节点的子节点被合并而且合并后的节点成为新的根节点。
*/

// Delete :删除key
func (t *Tree) Delete(key int) error {
	// keyRecord, err := t.Find(key)
	// if err != nil {
	// 	return err
	// }
	keyLeaf := t.findLeaf(key)
	if keyLeaf != nil {
		keyRecord := keyLeaf.getRecord(key)
		if keyRecord != nil {
			t.deleteKey(keyLeaf, key, keyRecord, -1)
			return nil
		}
	}

	return errors.New("the delete key not found")
}

// deleteKey :删除key 需要多次调用 所以封装单独func
func (t *Tree) deleteKey(n *Node, key int, pointer interface{}, keyIndex int) {
	n = t.removeKeyFromNode(n, key, pointer, keyIndex)

	// 删除的是root节点的key
	if n == t.Root {
		t.adjustRoot()
		return
	}

	// 删除之后node中的key数量合理 处理结束
	if n.Count >= MIN_LIMIT {
		return
	}

	// 其他情况
	neighbour, neighbourIndex := n.getNeighbour()
	// // neighbourIndex==-1时 neighbourKeyIndex=0（第一个合并如第二个）
	// neighbourKeyIndex := neighbourIndex
	// if neighbourIndex == -1 {
	// 	// 第二节点keyindex==0
	// 	neighbourKeyIndex = 0
	// }
	// neighbourKey := n.Parent.Keys[neighbourKeyIndex]

	// 如果加和数量合理就合并 不合理就重新分配（相当于合并再分裂）
	if neighbour.Count+n.Count < MAX_LIMIT {
		// 合并
		n.mergeToNode(neighbour, neighbourIndex, key, t)
	} else {
		// 借调key 保持平衡
		n.borrowFromNode(neighbour, neighbourIndex)
	}

	return
}

// borrowFromNode ：删除key之后 从邻节点借key
/*
neighbour 邻节点（被借调节点）
neighbourKeyIndex 邻节点在父节点的keyindex
neighbourKey 邻节点在父节点的key
*/
func (n *Node) borrowFromNode(neighbour *Node, neighbourIndex int) {
	if !n.IsLeaf {
		fmt.Println("redistributeNodes的不是leaf节点！！！！")
	}

	// neighbourIndex==-1时 neighbourKeyIndex=0（第一个从第二个借）
	neighbourKeyIndex := neighbourIndex - 1
	if neighbourIndex == -1 {
		// 第二节点keyindex==0
		neighbourKeyIndex = 0
	}

	if neighbourIndex == -1 {
		// neighbourIndex==-1 左从右借（第一个从第二个借） 右边第一个append给左边最后一个
		n.Keys = append(n.Keys, neighbour.Keys[0])
		n.Pointers = append(n.Pointers, neighbour.Pointers[0])
		// 删除邻节点被借调的key和pointer
		neighbour.Keys = neighbour.Keys[1:]
		neighbour.Pointers = neighbour.Pointers[1:]
		//父节点更新
		n.Parent.Keys[neighbourKeyIndex] = neighbour.Keys[0]
	} else {
		// neighbourIndex != -1 右从左借 左边最后一个 push给右边（append参数倒置）
		n.Keys = append([]int{neighbour.Keys[len(neighbour.Keys)-1]}, n.Keys...)
		n.Pointers = append([]interface{}{neighbour.Pointers[len(neighbour.Pointers)-1]}, n.Pointers...)
		// 删除邻节点被借调的key和pointer
		neighbour.Keys = neighbour.Keys[:len(neighbour.Keys)-1]
		neighbour.Pointers = neighbour.Pointers[:len(neighbour.Pointers)-1]
		//父节点更新
		nKeyIndex := neighbourKeyIndex + 1 // 右从左借 更新右在parent的key 所以+1
		n.Parent.Keys[nKeyIndex] = n.Keys[0]
	}

	n.Count++
	neighbour.Count--
	return
}

// mergeToNode ：删除key之后 合并入邻节点
/*
neighbour 邻节点（被借调节点）
neighbourKeyIndex 邻节点在父节点的keyindex
neighbourKey 邻节点在父节点的key
*/
func (n *Node) mergeToNode(neighbour *Node, neighbourIndex int, deletekey int, t *Tree) {
	// neighbourIndex==-1 左往右合并 反之右往左合并
	if neighbourIndex == -1 {
		// 左往右合并
		neighbour.Keys = append(append([]int{}, n.Keys...), neighbour.Keys...)
		neighbour.Pointers = append(append([]interface{}{}, n.Pointers...), neighbour.Pointers...)
		// 双向链表维护
		if n.IsLeaf && neighbour.IsLeaf {
			neighbour.Prev = n.Prev
		}

		// 左往右合并 neighbourKey只能是parent的第一个key
		neighbourKey := n.Parent.Keys[0]

		// 递归删除父节点的key和pointer（key是neighbour的，pointer是n的）
		t.deleteKey(n.Parent, neighbourKey, n, -1)
	} else {
		// 右往左合并
		neighbour.Keys = append(neighbour.Keys, n.Keys...)
		neighbour.Pointers = append(neighbour.Pointers, n.Pointers...)
		// 双向链表维护
		if n.IsLeaf && neighbour.IsLeaf {
			neighbour.Next = n.Next
		}

		// 递归删除父节点的key和pointer（key是n的，pointer是n的）
		t.deleteKey(n.Parent, deletekey, n, neighbourIndex)
	}

	neighbour.Count += n.Count
}

// removeKeyFromNode :执行remove key 返回被删除的node
func (t *Tree) removeKeyFromNode(n *Node, key int, pointer interface{}, keyIndex int) *Node {
	i := -1
	if keyIndex > -1 {
		i = keyIndex
	} else {
		i = n.getKeyIndex(key)
	}

	j := n.getPointerIndex(pointer)

	if i > -1 && j > -1 {
		// 删除key和pointer
		n.Keys = append(n.Keys[:i], n.Keys[i+1:]...)
		n.Pointers = append(n.Pointers[:j], n.Pointers[j+1:]...)
		// 节点计数-1
		n.Count--
	}

	return n
}

// root中的key被删之后,重新调整root
func (t *Tree) adjustRoot() {
	var newRoot *Node

	// root还剩key 不做任何处理
	if t.Root.Count > 0 {
		return
	}

	if t.Root.IsLeaf {
		// 被删空的root还是leaf 说明树空了 赋值nil即可
		newRoot = nil
	} else {
		// root被删空 说明之前只有一个key 子节点设置成新root即可
		newRoot, _ = t.Root.Pointers[0].(*Node)
		newRoot.Parent = nil
	}
	t.Root = newRoot

	return
}

// 根据节点获取相邻节点 (首节点的相邻是右节点，此时index返回-1,其他的相邻是左节点，返回对应index)
func (n *Node) getNeighbour() (neigh *Node, i int) {
	for i = 0; i <= n.Parent.Count; i++ {
		if reflect.DeepEqual(n.Parent.Pointers[i], n) {
			if i == 0 {
				i = -1
				neigh, _ = n.Parent.Pointers[1].(*Node)
			} else {
				i = i - 1
				neigh, _ = n.Parent.Pointers[i].(*Node)
			}
			break
		}
	}
	return neigh, i
}

// *********************** Find和Print等 ***********************

// Find :查找key
func (t *Tree) Find(key int) (*Record, error) {
	leaf := t.findLeaf(key)
	if leaf == nil {
		return nil, errors.New("key not found")
	}

	i := leaf.getKeyIndex(key)
	if i == -1 {
		return nil, errors.New("key not found")
	}

	r, _ := leaf.Pointers[i].(*Record)

	return r, nil
}

// 查找key对应的leaf,isParentContain：父节点是否包含此key
func (t *Tree) findLeaf(key int) *Node {
	n := t.Root
	// 空树
	if n == nil {
		// fmt.Println("Empty Tree")
		return nil
	}

	for i := 0; !n.IsLeaf; i = 0 {
		for i < n.Count {
			if n.Keys[i] <= key {
				i++
			} else {
				break
			}
		}

		// 切换至下层节点
		n, _ = n.Pointers[i].(*Node)
	}

	return n
}

// PrintTree :打印输出Tree(层次遍历 借助队列实现)
func (t *Tree) PrintTree() {
	var n *Node
	count := 0

	if t.Root == nil {
		fmt.Printf("Empty tree.\n")
		return
	}

	queue = list.New()
	queue.PushBack(t.Root)
	for queue.Len() > 0 {
		// 队列中取值
		n = queue.Front().Value.(*Node)
		queue.Remove(queue.Front()) // remove front element
		if n == nil {
			continue
		}

		// 遇到每层级第一个node 先换行
		if n.Parent != nil && n == n.Parent.Pointers[0] {
			level := t.getStepsToRoot(n)
			if level > 0 {
				fmt.Printf("\n")
			}
		}

		// 打印node的keys
		for i := 0; i < len(n.Keys); i++ {
			if i == 0 {
				fmt.Printf("[")
			}

			fmt.Printf("%d", n.Keys[i])

			if i < len(n.Keys)-1 {
				fmt.Printf(", ")
			}

			if i == len(n.Keys)-1 {
				fmt.Printf("] ")
				if n.IsLeaf && queue.Len() > 0 {
					fmt.Printf("-> ")
				}
			}

			if n.IsLeaf {
				count++
			}
		}

		// 非叶节点的子节点压入队列
		if !n.IsLeaf {
			for i := 0; i <= n.Count; i++ {
				if i > len(n.Pointers)-1 {
					fmt.Println(n)
				}
				c, _ := n.Pointers[i].(*Node)
				queue.PushBack(c)
			}
		}
	}

	height := t.height()
	fmt.Printf("\n阶数%d - 高度%d - key数%d\n", ORDER, height, count)
}

// PrintLeaves :打印输出所有叶子节点
func (t *Tree) PrintLeaves() {
	if t.Root == nil {
		fmt.Printf("Empty tree.\n")
		return
	}

	n := t.Root
	// 找到最小leaf
	for !n.IsLeaf {
		n, _ = n.Pointers[0].(*Node)
	}

	for n != nil {
		fmt.Printf("[")
		for i, key := range n.Keys {
			fmt.Printf("%d", key)
			if i < len(n.Keys)-1 {
				fmt.Printf(", ")
			}
		}
		fmt.Printf("]")
		if n.Next != nil {
			fmt.Printf(" -> ")
		}
		n = n.Next
	}
	fmt.Printf("\n")
}

// FindAndPrint :查找并打印
func (t *Tree) FindAndPrint(key int) {
	r, err := t.Find(key)

	if err != nil || r == nil {
		fmt.Printf("Record not found under key %d.\n", key)
	} else {
		fmt.Printf("Record at %d -- key %d, value %s.\n", r, key, r.Value)
	}
}

// FindAndPrintRange :查找并打印范围数据
func (t *Tree) FindAndPrintRange(keyMin, keyMax int) (count int) {
	keys, records := t.FindRange(keyMin, keyMax)
	if len(keys) == 0 {
		fmt.Println("None found")
	} else {
		fmt.Printf("Find %d keys\n", len(keys))
		for i, key := range keys {
			fmt.Printf("Key: %d  Location: %d  Value: %s\n",
				key,
				records[i],
				records[i].Value,
			)
		}
	}

	return len(keys)
}

// FindRange :范围查找
func (t *Tree) FindRange(keyMin int, keyMax int) ([]int, []*Record) {
	n := t.findLeaf(keyMin)
	if n == nil {
		return nil, nil
	}

	keys := make([]int, 0, 0)
	records := make([]*Record, 0, 0)

	for n != nil && n.Keys[len(n.Keys)-1] < keyMax {
		for i, key := range n.Keys {
			if key > keyMax || key < keyMin {
				continue
			}

			keys = append(keys, key)
			records = append(records, n.Pointers[i].(*Record))
		}

		n = n.Next
	}

	return keys, records
}

// 获取tree高度
func (t *Tree) height() int {
	h := 1
	n := t.Root
	if n == nil {
		return 0
	}

	for ; !n.IsLeaf; h++ {
		n, _ = n.Pointers[0].(*Node)
	}
	return h
}

// 回到根节点的步数
func (t *Tree) getStepsToRoot(child *Node) int {
	length := 0
	c := child
	for c != t.Root {
		c = c.Parent
		length++
	}
	return length
}

// ************************* constructor *************************

func newRecord(value []byte) (*Record, error) {
	r := &Record{}
	r.Value = value
	return r, nil
}

func newNode() (*Node, error) {
	n := &Node{}
	n.Keys = make([]int, 0)             //! 每个节点的最大数据量比阶数少1(MAX_LIMIT) 这里cap是考虑临时插入的情况+1
	n.Pointers = make([]interface{}, 0) //! 每个节点的最大指针数量==阶数 这里cap是考虑临时插入的情况+1
	n.IsLeaf = false
	n.Count = 0
	n.Parent = nil
	n.Next = nil
	return n, nil
}

func newLeaf() (*Node, error) {
	leaf, err := newNode()
	if err != nil {
		return nil, err
	}
	leaf.IsLeaf = true
	return leaf, nil
}

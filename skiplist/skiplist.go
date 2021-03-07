package skiplist

import (
	"bytes"
	"fmt"
	"math/rand"
)

// 定义了最大的层级限制
const MaxLevel = 16

// 定义生成层级的因子
const LevelFactor = 0.5

type SkipList struct {
	level  int   // 当前最高层级
	length int   // key数量
	size   int   // Skiplist字节大小
	head   *Node // 头节点是伪节点 不计数 不参与计算
}

type Node struct {
	key     []byte
	value   []byte
	forward []*Node
}

func NewSkipList() *SkipList {
	sk := &SkipList{
		head: &Node{
			forward: make([]*Node, MaxLevel),
		},
	}
	for i := 0; i < MaxLevel; i++ {
		sk.head.forward[i] = nil
	}
	//rand.Seed(time.Now().UnixNano())
	rand.Seed(123)
	return sk
}

func (sl *SkipList) Put(key, value []byte) {
	route := make([]*Node, MaxLevel)
	p := sl.head
	for i := sl.level - 1; i >= 0; i-- {
		for p.forward[i] != nil {
			r := bytes.Compare(key, p.forward[i].key)
			if r == -1 {
				break
			} else if r == 0 {
				sl.size += len(value) - len(p.forward[i].value)
				p.forward[i].value = value
				return
			} else if r == 1 {
				p = p.forward[i]
			}
		}

		route[i] = p
	}

	level := sl.randomLevel()
	node := &Node{
		key:     key,
		value:   value,
		forward: make([]*Node, level),
	}
	if level > sl.level {
		for i := sl.level; i < level; i++ {
			route[i] = sl.head
		}
		sl.level = level
	}
	for i := 0; i < level; i++ {
		node.forward[i] = route[i].forward[i]
		route[i].forward[i] = node
	}
	sl.size += len(key) + len(value) + 8
	sl.length++
}

func (sl *SkipList) Get(key []byte) []byte {
	p := sl.head
	for i := sl.level - 1; i >= 0; i-- {
		for p.forward[i] != nil {
			r := bytes.Compare(key, p.forward[i].key)
			if r == -1 {
				break
			} else if r == 0 {
				return p.forward[i].value
			} else if r == 1 {
				p = p.forward[i]
			}
		}
	}
	return nil
}

func (sl *SkipList) Delete(key []byte) {
	update := make([]*Node, MaxLevel)
	var q *Node
	p := sl.head
	for i := sl.level - 1; i >= 0; i-- {
		for p.forward[i] != nil {
			r := bytes.Compare(key, p.forward[i].key)
			if r == -1 {
				break
			} else if r == 0 {
				q = p.forward[i]
				break
			} else if r == 1 {
				p = p.forward[i]
			}
		}
		update[i] = p
	}
	if q != nil {
		sl.size += len(key) - len(p.value) - 8
		sl.length--
		for i := 0; i < len(q.forward); i++ {
			update[i].forward[i] = q.forward[i]
		}
	}
}

func (sl *SkipList) Iterator() *SKIterator {
	ski := &SKIterator{
		sk:   sl,
		node: &Node{},
	}
	return ski
}

func (sl *SkipList) Print() {
	for i := sl.level - 1; i >= 0; i-- {
		p := sl.head.forward[i]
		fmt.Printf("[level-%d] ", i+1)
		for p != nil {
			fmt.Printf("%v ", p.key)
			p = p.forward[i]
		}

		fmt.Println()
	}
}

func (sl *SkipList) Size() int {
	return sl.size
}

func (sl *SkipList) Length() int {
	return sl.length
}

/** 这里随机产生一个 层级
在 LevelFactor 是 0.5 的情况下
1 级的概率是 50%
2 级的概率是 25%
3 级的概率是 12.5%, 以此类推
*/
func (sl SkipList) randomLevel() int {
	l := 1
	// 使用随机数来决定层级
	for rand.Float64() < LevelFactor && l+1 < MaxLevel {
		l++
	}

	// 如果层级比当前层级高2级或以上，按照高一级处理，避免浪费
	if l > sl.level+1 {
		l = sl.level + 1
	}
	return l
}

// 迭代器
type SKIterator struct {
	sk   *SkipList
	node *Node
}

func (ski *SKIterator) First() {
	if ski.sk != nil {
		ski.node = ski.sk.head.forward[0]
	}
}

func (ski *SKIterator) Next() {
	if ski.node.forward != nil {
		ski.node = ski.node.forward[0]
	}
}

func (ski *SKIterator) End() bool {
	if ski.node == nil {
		return true
	}
	return false
}

func (ski *SKIterator) Key() []byte {
	if ski.node != nil {
		return ski.node.key
	}
	return nil
}

func (ski *SKIterator) Value() []byte {
	if ski.node != nil {
		return ski.node.value
	}
	return nil
}

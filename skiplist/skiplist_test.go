package skiplist

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"
)

var sl *SkipList

func TestMain(m *testing.M) {
	//fmt.Println("write setup code here...") // 测试之前的做一些设置
	setUpSkipList()

	// 如果 TestMain 使用了 flags，这里应该加上flag.Parse()
	retCode := m.Run() // 执行测试

	//fmt.Println("write teardown code here...") // 测试之后做一些拆卸工作
	os.Exit(retCode) // 退出测试
}

func TestSkipList_Add(t *testing.T) {
	tests := []struct {
		name string
		key  []byte
		want int
	}{
		// TODO: Add test cases.
		{name: "156", key: []byte{156}},
		{name: "10", key: []byte{10}},
		{name: "89", key: []byte{89}},
		{name: "5", key: []byte{5}},
		{name: "6", key: []byte{6}},
		{name: "55", key: []byte{55}},
		{name: "34", key: []byte{34}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sl.Put(tt.key, []byte{})
			sl.Print()
		})
	}
}

func TestSkipList_Delete(t *testing.T) {
	tests := []struct {
		name string
		key  []byte
		want int
	}{
		// TODO: Add test cases.
		{name: "10", key: []byte{10}},
		{name: "11", key: []byte{11}},
		{name: "677", key: []byte{77}},
		{name: "34", key: []byte{34}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sl.Delete(tt.key)
			sl.Print()
		})
	}
}

func TestSkipList_Find(t *testing.T) {
	_sl := NewSkipList()
	for i := 0; i < 1000000; i++ {
		key := make([]byte, 4)
		binary.BigEndian.PutUint32(key, rand.Uint32())
		_sl.Put(key, []byte{})
	}

	finder := make([]byte, 4)
	binary.BigEndian.PutUint32(finder, 1e8)
	tests := []struct {
		name string
		key  []byte
		want []byte
	}{
		// TODO: Add test cases.
		{name: "1e8", key: finder, want: []byte("aris")},
	}

	for _, tt := range tests {
		_sl.Put(tt.key, tt.want)
	}
	//_sl.Print()

	start := time.Now()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := _sl.Get(tt.key); bytes.Compare(got, tt.want) != 0 {
				t.Errorf("Find() = %v, want %v", got, tt.want)
			}
		})
	}
	tc := time.Since(start)
	fmt.Printf("Find in %d length skiplist cost [%dms]\n", _sl.length, tc.Nanoseconds()/1e6)

}

// 前置方法 组装一个SkipList
func setUpSkipList() {
	var data = []Node{
		{key: []byte{1}, value: []byte("beijing")},
		{key: []byte{3}, value: []byte("shanghai")},
		{key: []byte{9}, value: []byte("tianjin")},
		{key: []byte{10}, value: []byte("hefei")},
		{key: []byte{11}, value: []byte("guangzhou")},
		{key: []byte{12}, value: []byte("xian")},
		{key: []byte{34}, value: []byte("changchun")},
		{key: []byte{45}, value: []byte("shenyang")},
	}

	sl = NewSkipList()
	for _, node := range data {
		sl.Put(node.key, node.value)
	}
	sl.Print()
}

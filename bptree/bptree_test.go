package bptree

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"time"
)

// ************** insert test **************

func TestInsertNilRoot(t *testing.T) {
	tree := NewTree()

	key := 1
	value := []byte("test")

	err := tree.Insert(key, value)

	if err != nil {
		t.Errorf("%s", err)
	}

	r, err := tree.Find(key)
	if err != nil {
		t.Errorf("%s\n", err)
	}

	if r == nil {
		t.Errorf("returned nil \n")
	}

	if !reflect.DeepEqual(r.Value, value) {
		t.Errorf("expected %v and got %v \n", value, r.Value)
	}
}

func TestInsert(t *testing.T) {
	tree := NewTree()

	key := 1
	value := []byte("test")

	err := tree.Insert(key, value)
	if err != nil {
		t.Errorf("%s", err)
	}

	r, err := tree.Find(key)
	if err != nil {
		t.Errorf("%s\n", err)
	}

	if r == nil {
		t.Errorf("returned nil \n")
	}

	if !reflect.DeepEqual(r.Value, value) {
		t.Errorf("expected %v and got %v \n", value, r.Value)
	}
}

func TestInsertSameKeyTwice(t *testing.T) {
	tree := NewTree()

	key := 1
	value := []byte("test")

	err := tree.Insert(key, value)
	if err != nil {
		t.Errorf("%s", err)
	}

	err = tree.Insert(key, append(value, []byte("world1")...))
	if err == nil {
		t.Errorf("expected error but got nil")
	}

	r, err := tree.Find(key)
	if err != nil {
		t.Errorf("%s\n", err)
	}

	if r == nil {
		t.Errorf("returned nil \n")
	}

	if !reflect.DeepEqual(r.Value, value) {
		t.Errorf("expected %v and got %v \n", value, r.Value)
	}

	if tree.Root.Count > 1 {
		t.Errorf("expected 1 key and got %d", tree.Root.Count)
	}
}

func TestInsertSameValueTwice(t *testing.T) {
	tree := NewTree()

	key := 1
	value := []byte("test")

	err := tree.Insert(key, value)
	if err != nil {
		t.Errorf("%s", err)
	}
	err = tree.Insert(key+1, value)
	if err != nil {
		t.Errorf("%s", err)
	}

	r, err := tree.Find(key)
	if err != nil {
		t.Errorf("%s\n", err)
	}

	if r == nil {
		t.Errorf("returned nil \n")
	}

	if !reflect.DeepEqual(r.Value, value) {
		t.Errorf("expected %v and got %v \n", value, r.Value)
	}

	if tree.Root.Count <= 1 {
		t.Errorf("expected more than 1 key and got %d", tree.Root.Count)
	}
}

func TestFindNilRoot(t *testing.T) {
	tree := NewTree()

	r, err := tree.Find(1)
	if err == nil {
		t.Errorf("expected error and got nil")
	}

	if r != nil {
		t.Errorf("expected nil got %s \n", r)
	}
}

func TestFind(t *testing.T) {
	tree := NewTree()

	key := 1
	value := []byte("test")

	err := tree.Insert(key, value)
	if err != nil {
		t.Errorf("%s", err)
	}

	r, err := tree.Find(key)
	if err != nil {
		t.Errorf("%s\n", err)
	}

	if r == nil {
		t.Errorf("returned nil \n")
	}

	if !reflect.DeepEqual(r.Value, value) {
		t.Errorf("expected %v and got %v \n", value, r.Value)
	}
}

func TestFindRange(t *testing.T) {
	tree := NewTree()
	r := initRand()
	keys := []int{3, 13, 23, 33, 43, 53, 63, 73, 83, 93, 103, 113, 123}

	for _, key := range keys {
		value := getRandString(r, 4)
		_ = tree.Insert(key, []byte(value))
	}

	count := tree.FindAndPrintRange(50, 100)
	if count != 5 {
		t.Errorf("find range failed,expect find 5 keys,but get %d", count)
	}
}

// ????????????key ?????????tree
func TestBatchInsertAndPrintTree(t *testing.T) {
	tree := NewTree()
	_range := 1000
	num := 10
	r := initRand()

	for i := 0; i < num; i++ {
		key := r.Intn(_range)
		value := getRandString(r, 4)
		_ = tree.Insert(key, []byte(value))
	}

	tree.PrintTree()
}

// ????????????key ????????????????????????
func TestBatchInsertAndPrintLeaves(t *testing.T) {
	tree := NewTree()
	_range := 1000
	num := 10
	r := initRand()

	for i := 0; i < num; i++ {
		key := r.Intn(_range)
		value := getRandString(r, 4)
		_ = tree.Insert(key, []byte(value))
	}

	tree.PrintLeaves()
	// tree.PrintTree()
}

// ************** delete test **************

// ????????????
func TestSingleDelete(t *testing.T) {
	tree := NewTree()
	_range := 1000
	num := 10
	r := initRand()
	var curKey int

	//  1.insert 10????????????
	for i := 0; i < num; i++ {
		key := r.Intn(_range)
		value := getRandString(r, 4)
		_ = tree.Insert(key, []byte(value))
	}
	// 2.insert ???key
	curKey = r.Intn(_range)
	_ = tree.Insert(curKey, []byte(getRandString(r, 4)))

	tree.PrintTree()

	// 3. delete ???key
	fmt.Printf("Now delete key: %d\n", curKey)
	err = tree.Delete(curKey)
	if err != nil {
		t.Errorf("TestSingleDelete failed")
	}

	tree.PrintTree()
}

// ???????????????
func TestDeleteRoot(t *testing.T) {
	tree := NewTree()
	_range := 1000
	num := 10
	r := initRand()

	for i := 0; i < num; i++ {
		key := r.Intn(_range)
		value := getRandString(r, 4)
		_ = tree.Insert(key, []byte(value))
	}

	tree.PrintTree()

	for _, key := range tree.Root.Keys {
		fmt.Printf("-------------------------\nNow delete root key [%d]\n", key)
		err = tree.Delete(key)
		tree.PrintTree()
		if err != nil {
			t.Errorf("TestDeleteRoot failed:%s", err)
		}
		break
	}
}

// ??????????????????????????????
func TestMultiDelete(t *testing.T) {
	tree := NewTree()
	_range := 1000
	num := 10
	r := initRand()
	keys := make([]int, 0)

	//  1.insert ???????????????
	for i := 0; i < num; i++ {
		key := r.Intn(_range)
		keys = append(keys, key)
		value := getRandString(r, 4)
		_ = tree.Insert(key, []byte(value))
	}

	tree.PrintTree()

	// 3. ??????delete
	for _, key := range keys {
		fmt.Printf("-------------------------\nNow delete key [%d]\n", key)
		err = tree.Delete(key)
		if err != nil {
			t.Errorf("TestMultiDelete failed:%s", err)
		}

		tree.PrintTree()
	}

}

func TestDeleteNilTree(t *testing.T) {
	tree := NewTree()

	key := 1

	err := tree.Delete(key)
	if err == nil {
		t.Errorf("expected error and got nil")
	}

	r, err := tree.Find(key)
	if err == nil {
		t.Errorf("expected error and got nil")
	}

	if r != nil {
		t.Errorf("returned struct after delete \n")
	}
}

func TestDeleteNotFound(t *testing.T) {
	tree := NewTree()

	key := 1
	value := []byte("test")

	err := tree.Insert(key, value)
	if err != nil {
		t.Errorf("%s", err)
	}

	r, err := tree.Find(key)
	if err != nil {
		t.Errorf("%s\n", err)
	}

	if r == nil {
		t.Errorf("returned nil \n")
	}

	if !reflect.DeepEqual(r.Value, value) {
		t.Errorf("expected %v and got %v \n", value, r.Value)
	}

	err = tree.Delete(key + 1)
	if err == nil {
		t.Errorf("expected error and got nil")
	}

	r, err = tree.Find(key + 1)
	if err == nil {
		t.Errorf("expected error and got nil")
	}
}

// ************** util func **************
// ???????????????
func getRandomInt(max int) int {
	//??????????????????????????????
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max)
}

func initRand() *rand.Rand {
	return rand.New(rand.NewSource(time.Now().Unix()))
}

// RandString ?????????????????????
func getRandString(r *rand.Rand, len int) string {
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		b := r.Intn(26) + 65
		bytes[i] = byte(b)
	}
	return string(bytes)
}

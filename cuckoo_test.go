package cuckoo

import (
	"fmt"
	"strconv"
	"testing"
	// Nothing yet
	)

func TestNewTable(t *testing.T) {
	ct := NewTable()
	if ct == nil {
		t.Fatal("Could not allocate table")
	}
}

func TestBasic(t *testing.T) {
	ct := NewTable()
	for i := 0; i < 1000; i++ {
		istr := strconv.Itoa(i)
		ct.Put(keytype(istr), valuetype(fmt.Sprintf("testvalue-%d", i)))
	}
	for i := 0; i < 1000; i++ {
		istr := strconv.Itoa(i)
		v, found := ct.Get(keytype(istr))
		if (!found) {
			t.Fatal("Could not find key", istr)
		}
		vs := string(v)
		if vs != fmt.Sprintf("testvalue-%d", i) {
			t.Fatalf("Wrong value, expected testvalue-%d got %s", i, vs)
		}
	}

}
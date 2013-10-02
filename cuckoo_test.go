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

func TestFill(t *testing.T) {
	ct := NewTablePowerOfTwo(10)
	// Should be able to hold at least 960 elements, but will have to
	// cuckoo a lot to fill those last bits.  Stress test the cuckooing.
	limit := 960
	for i := 0; i < limit; i++ {
		istr := strconv.Itoa(i)
		ct.Put(keytype(istr), valuetype(fmt.Sprintf("testvalue-%d", i)))
	}
	for i := 0; i < limit; i++ {
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
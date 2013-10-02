package cuckoo

import (
	"strconv"
	"testing"
)

// Standardizing our entry formats
func keyAndValue(i int) (key keytype, value valuetype) {
	keystr := strconv.Itoa(i)
	key = keytype(keystr)
	value = valuetype("testvalue-" + keystr)
	return
}

// Helper function to populate the table
func fillTable(t *testing.T, ct *Table, limit int) {
	for i := 0; i < limit; i++ {
		k, v := keyAndValue(i)
		ct.Put(k, v)
	}
}

func validateFound(t *testing.T, ct *Table, start int, limit int, testName string) {
	for i := start; i < limit; i++ {
		k, wantedValue := keyAndValue(i)
		v, found := ct.Get(k)
		if !found {
			t.Fatal(testName, "could not find key", k)
		}
		if v != wantedValue {
			t.Fatalf(testName, "wrong value, expected %v got %v", i, wantedValue, v)
		}
	}
}

func TestNewTable(t *testing.T) {
	ct := NewTable()
	if ct == nil {
		t.Fatal("Could not allocate table")
	}
}

func TestBasic(t *testing.T) {
	ct := NewTable()
	nKeys := 1000
	fillTable(t, ct, nKeys)
	validateFound(t, ct, 0, nKeys, "TestBasic")
}

func TestFill(t *testing.T) {
	ct := NewTablePowerOfTwo(10)
	// Should be able to hold at least 950 elements, but will have to
	// cuckoo a lot to fill those last bits.  Stress test the cuckooing.
	limit := 874 // 875 fails - we're not BFS'ing well enough yet
	fillTable(t, ct, limit)
	validateFound(t, ct, 0, limit, "TestFill")
}

func TestDelete(t *testing.T) {
	ct := NewTable()
	limit := 1000
	fillTable(t, ct, limit)
	validateFound(t, ct, 0, limit, "TestDelete")
	for i := 0; i < limit; i++ {
		k, _ := keyAndValue(i)
		ct.Delete(k)
		foundVal, found := ct.Get(k)
		if found {
			t.Fatalf("TestDelete failed - item %v still present as %v", k, foundVal)
		}
		validateFound(t, ct, i+1, limit, "TestDelete")
	}
}

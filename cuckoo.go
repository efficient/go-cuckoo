// Copyright (c) 2013 David G. Andersen.  All rights reserved.
// Use of this source is goverened by the Apache open source license,
// a copy of which can be found in the LICENSE file.  Please contact
// the author if you would like a copy under another license.  We're
// happy to help out.

// https://github.com/efficient/go-cuckoo

// Package cuckoo provides an implementation of a high-performance,
// memory efficient hash table that supports fast and safe concurrent
// access by multiple threads.
// The default version of the hash table uses string keys and
// interface{} values.  For faster performance and fewer annoying
// typecasting issues, copy this code and change the valuetype
// appropriately.
package cuckoo

import (
	//"math/rand"
	//"time"
	//"sync/atomic"
	"fmt"
	"hash/crc64" // xxx - use city eventually.
)

const (
	SLOTS_PER_BUCKET = 4 // This is kinda hardcoded all over the place
	DEFAULT_START_POWER = 16 // 2^16 keys to start with.
	N_LOCKS = 2048
	MAX_REACH = 500 // number of buckets to examine before full
)

type keytype string
type valuetype string

type kvtype struct {
	keyhash uint64
	key *keytype
	value *valuetype
}

type Table struct {
	storage []kvtype
	locks [N_LOCKS]int32
	hashpower uint
	cheatmap map[keytype]valuetype // hee.  Getting the tests working for now.
	bucketMask uint64
	// For hashing
	crcTab *crc64.Table
}

func NewTable() *Table {
	return NewTablePowerOfTwo(DEFAULT_START_POWER)
}

func (t *Table) sizeTable(twopower uint) {
	t.hashpower = twopower-2
	t.bucketMask = (1<<t.hashpower)-1
}

func NewTablePowerOfTwo(twopower uint) *Table {
	t := &Table{}
	t.sizeTable(twopower)
	// storage holds items, but is organized into N/4 fully
	// associative buckets conceptually, so the hashpower differs
	// from the storage size.
	t.storage = make([]kvtype, 1<<twopower)
	t.cheatmap = make(map[keytype]valuetype)
	t.crcTab = crc64.MakeTable(crc64.ECMA)
	return t
}

func (t *Table) getKeyhash(k keytype) uint64 {
	return ((1<<63) | crc64.Checksum([]byte(k), t.crcTab))
}

var _ = fmt.Println

func (t *Table) indexes(keyhash uint64) (i1, i2 uint64){
	tag := (keyhash & 0xff)+1;
	i1 = (keyhash >> 8)  & t.bucketMask
	i2 = (i1 ^ (tag*0x5bd1e995)) & t.bucketMask
	return
}

func (t *Table) tryBucketRead(k keytype, keyhash uint64, bucket uint64) (valuetype, bool) {
	storageOffset := bucket*4
	buckets := t.storage[storageOffset:storageOffset+4]
	for _, b := range buckets {
		if b.keyhash == keyhash {
			if *b.key == k {
				return *b.value, true
			}
		}
	}
	return valuetype(0), false
}

func (t Table)hasSpace(bucket uint64) (bool, int) {
	storageOffset := bucket*4
	buckets := t.storage[storageOffset:storageOffset+4]
	for i, b := range buckets {
		if b.keyhash == 0 {
			return true, i
		}
	}
	return false, 0
}

func (t Table) insert(k keytype, v valuetype, keyhash uint64, bucket uint64, slot int) {
	b := &(t.storage[bucket*4+uint64(slot)])
	b.keyhash = keyhash
	b.key = &k
	b.value = &v
}

func (t *Table) Get(k keytype) (v valuetype, found bool) {
	keyhash := t.getKeyhash(k)
	i1, i2 := t.indexes(keyhash)
	v, found = t.tryBucketRead(k, keyhash, i1)
	if (!found) {
		v, found = t.tryBucketRead(k, keyhash, i2)
	}
	return
}

// func (t *Table) slotSearchBFS(i1, i2 uint64) {
// 	reach := 0
// 	var path [4]uint64 // List of actual storage offsets
// 	for depth := 0; depth < 4; depth++ {
// 		if hasit, where := t.hasSpace(i1); hasit {
			
// 	}
// }

func (t *Table) Put(k keytype, v valuetype) error {
	keyhash := t.getKeyhash(k)
	i1, i2 := t.indexes(keyhash)
	if hasSpace, where := t.hasSpace(i1); hasSpace {
		t.insert(k, v, keyhash, i1, where)
	} else if hasSpace, where := t.hasSpace(i2); hasSpace {
		t.insert(k, v, keyhash, i2, where)
	} else {
			//path := t.slotSearchBFS(i1, i2)
		panic("Have to cuckoo, but this isn't implemented yet!")
	}
	return nil
}
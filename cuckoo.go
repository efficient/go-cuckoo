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
	"log"
)

const (
	SLOTS_PER_BUCKET    = 4  // This is kinda hardcoded all over the place
	DEFAULT_START_POWER = 16 // 2^16 keys to start with.
	N_LOCKS             = 2048
	MAX_REACH           = 500 // number of buckets to examine before full
	MAX_PATH_DEPTH      = 5   // must be ceil(log4(MAX_REACH))
)

type keytype string
type valuetype string

type kvtype struct {
	keyhash uint64
	key     *keytype
	value   *valuetype
}

type Table struct {
	storage    []kvtype
	locks      [N_LOCKS]int32
	hashpower  uint
	cheatmap   map[keytype]valuetype // hee.  Getting the tests working for now.
	bucketMask uint64
	// For hashing
	crcTab *crc64.Table
}

func NewTable() *Table {
	return NewTablePowerOfTwo(DEFAULT_START_POWER)
}

func (t *Table) sizeTable(twopower uint) {
	t.hashpower = twopower - 2
	t.bucketMask = (1 << t.hashpower) - 1
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
	return ((1 << 63) | crc64.Checksum([]byte(k), t.crcTab))
}

var _ = fmt.Println
var _ = log.Fatal

func (t *Table) altIndex(bucket, keyhash uint64) uint64 {
	tag := (keyhash & 0xff) + 1
	return (bucket ^ (tag * 0x5bd1e995)) & t.bucketMask
}

func (t *Table) indexes(keyhash uint64) (i1, i2 uint64) {
	tag := (keyhash & 0xff) + 1
	i1 = (keyhash >> 8) & t.bucketMask
	i2 = (i1 ^ (tag * 0x5bd1e995)) & t.bucketMask
	return
}

func (t *Table) tryBucketRead(k keytype, keyhash uint64, bucket uint64) (valuetype, bool) {
	storageOffset := bucket * 4
	buckets := t.storage[storageOffset : storageOffset+4]
	for _, b := range buckets {
		if b.keyhash == keyhash {
			if *b.key == k {
				return *b.value, true
			}
		}
	}
	return valuetype(0), false
}

func (t Table) hasSpace(bucket uint64) (bool, int) {
	storageOffset := bucket * 4
	buckets := t.storage[storageOffset : storageOffset+4]
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
	if !found {
		v, found = t.tryBucketRead(k, keyhash, i2)
	}
	if !found {
		fmt.Printf("key %s not found.  Indexes %d and %d\n", string(k), i1, i2)
	}

	return
}

type pathEnt struct {
	bucket     uint64
	depth      int
	parent     int
	parentslot int
}

func (t *Table) slotSearchBFS(i1, i2 uint64) (success bool, path [MAX_PATH_DEPTH]uint64, depth int) {
	var queue [500]pathEnt
	queue_head := 0
	queue_tail := 0

	queue[queue_tail].bucket = i1
	queue_tail++

	queue[queue_tail].bucket = i2
	queue_tail++

	for dfspos := 0; dfspos < MAX_REACH; dfspos++ {
		candidate := queue[queue_head]
		candidate_pos := queue_head
		queue_head++
		//log.Printf("BFS examining %v ", candidate)
		if hasit, where := t.hasSpace(candidate.bucket); hasit {
			// log.Printf("BFS found space at bucket %d slot %d (parent %d slot %d)   candidate: %v",
			// 	candidate.bucket, where, candidate.parent, candidate.parentslot, candidate)
			cd := candidate.depth
			path[candidate.depth] = candidate.bucket*SLOTS_PER_BUCKET + uint64(where)
			//log.Printf("path %d = %v", candidate.depth, path[candidate.depth])
			parentslot := candidate.parentslot
			for i := 0; i < cd; i++ {
				candidate = queue[candidate.parent]
				path[candidate.depth] = candidate.bucket*SLOTS_PER_BUCKET + uint64(parentslot)
				parentslot = candidate.parentslot
				//log.Printf("path %d = %v  (%v)", candidate.depth, path[candidate.depth], candidate)
			}
			return true, path, cd
		} else {
			bStart := candidate.bucket * SLOTS_PER_BUCKET
			for i := 0; queue_tail < MAX_REACH && i < SLOTS_PER_BUCKET; i++ {
				buck := bStart + uint64(i)
				kh := t.storage[buck].keyhash
				ai := t.altIndex(candidate.bucket, kh)
				//log.Printf("  enqueue %d (%d) ((%v)) - %d", candidate.bucket, buck, *t.storage[buck].key, ai)
				queue[queue_tail].bucket = ai
				queue[queue_tail].depth = candidate.depth + 1
				queue[queue_tail].parent = candidate_pos
				queue[queue_tail].parentslot = i
				queue_tail++
			}
		}
	}
	return false, path, 0

}

func (t *Table) swap(x, y uint64) {
	// Needs to be made conditional on matching the path for the
	// concurrent version...
	// xkey, ykey := "none", "none"
	// if t.storage[x].key != nil { xkey = string(*t.storage[x].key) }
	// if t.storage[y].key != nil { ykey = string(*t.storage[y].key) }
	//log.Printf("swap %d to %d  (keys %v and %v)  (now: %v and %v)\n", x, y, xkey, ykey, t.storage[x], t.storage[y])
	t.storage[x], t.storage[y] = t.storage[y], t.storage[x]
	//log.Printf("  after %v and %v", t.storage[x], t.storage[y])
}

func (t *Table) Put(k keytype, v valuetype) error {
	keyhash := t.getKeyhash(k)
	i1, i2 := t.indexes(keyhash)
	if hasSpace, where := t.hasSpace(i1); hasSpace {
		t.insert(k, v, keyhash, i1, where)
	} else if hasSpace, where := t.hasSpace(i2); hasSpace {
		t.insert(k, v, keyhash, i2, where)
	} else {
		//log.Printf("BFSing for %v indexes %d %d\n", k, i1, i2)
		found, path, depth := t.slotSearchBFS(i1, i2)
		if !found {
			panic("Crap, table really full, search failed")
		}
		for i := depth; i > 0; i-- {
			t.swap(path[i], path[i-1])
		}
		t.insert(k, v, keyhash, path[0]/4, int(path[0]%4))
		//log.Printf("Insert at %d (now %v)", path[0], t.storage[path[0]])
	}
	return nil
}

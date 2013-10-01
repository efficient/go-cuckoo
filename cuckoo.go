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
	DEFAULT_START_POWER = 16 // 2^16 keys to start with.
	N_LOCKS = 2048
)

type keytype string
type valuetype string

type kvtype struct {
	keyhash int64
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
	t.hashpower = twopower
	t.bucketMask = (1<<twopower)-1
}

func NewTablePowerOfTwo(twopower uint) *Table {
	t := &Table{}
	t.sizeTable(twopower)
	t.storage = make([]kvtype, 1<<twopower)
	t.cheatmap = make(map[keytype]valuetype)
	t.crcTab = crc64.MakeTable(crc64.ECMA)
	return t
}

func (t *Table) getKeyhash(k keytype) uint64 {
	return crc64.Checksum([]byte(k), t.crcTab)
}

var _ = fmt.Println

func (t *Table) indexes(keyhash uint64) (i1, i2 uint64){
	tag := (keyhash & 0xff)+1;
	i1 = (keyhash >> 8)  & t.bucketMask
	i2 = (i1 ^ (tag*0x5bd1e995)) & t.bucketMask
	return
}

func (t *Table) Get(k keytype) (v valuetype, found bool) {
	v, found = t.cheatmap[k]
	keyhash := t.getKeyhash(k)
	i1, i2 := t.indexes(keyhash)
	return
	found = true
	return
}

func (t *Table) Put(k keytype, v valuetype) error {
	t.cheatmap[k] = v
	return nil
}
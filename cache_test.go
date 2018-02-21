/*
ltcache.go is released under the MIT License <http://www.opensource.org/licenses/mit-license.php
Copyright (C) ITsysCOM GmbH. All Rights Reserved.

A LRU cache with TTL capabilities.

*/

package ltcache

import (
	"math/rand"
	"testing"
	"time"
)

var testCIs = []*cachedItem{
	&cachedItem{itemID: "1", value: "one"},
	&cachedItem{itemID: "2", value: "two"},
	&cachedItem{itemID: "3", value: "three"},
	&cachedItem{itemID: "4", value: "four"},
	&cachedItem{itemID: "5", value: "five"},
}
var lastEvicted string

func TestSetGetRemNoIndexes(t *testing.T) {
	cache := New(UnlimitedCaching, 0, false,
		func(itmID string, v interface{}) { lastEvicted = itmID })
	for _, ci := range testCIs {
		cache.Set(ci.itemID, ci.value, nil)
	}
	if len(cache.cache) != 5 {
		t.Errorf("Wrong intems in cache: %+v", cache.cache)
	}
	if cache.lruIdx.Len() != 0 {
		t.Errorf("Wrong items in lru index: %+v", cache.lruIdx)
	}
	if len(cache.lruRefs) != 0 {
		t.Errorf("Wrong items in lru references: %+v", cache.lruRefs)
	}
	if cache.ttlIdx.Len() != 0 {
		t.Errorf("Wrong items in ttl index: %+v", cache.ttlIdx)
	}
	if len(cache.ttlRefs) != 0 {
		t.Errorf("Wrong items in ttl index: %+v", cache.ttlRefs)
	}
	if val, has := cache.Get("2"); !has {
		t.Error("item not in cache")
	} else if val.(string) != "two" {
		t.Errorf("wrong item value: %v", val)
	}
	if ks := cache.Items(); len(ks) != 5 {
		t.Errorf("wrong keys: %+v", ks)
	}
	cache.Set("2", "twice", nil)
	if val, has := cache.Get("2"); !has {
		t.Error("item not in cache")
	} else if val.(string) != "twice" {
		t.Errorf("wrong item value: %v", val)
	}
	if lastEvicted != "" {
		t.Error("lastEvicted var should be empty")
	}
	cache.Remove("2")
	if lastEvicted != "2" { // onEvicted should populate this var
		t.Error("lastEvicted var should be 2")
	}
	if _, has := cache.Get("2"); has {
		t.Error("item still in cache")
	}
	if len(cache.cache) != 4 {
		t.Errorf("Wrong intems in cache: %+v", cache.cache)
	}
	if cache.Len() != 4 {
		t.Errorf("Wrong intems in cache: %+v", cache.cache)
	}
	cache.Clear()
	if cache.Len() != 0 {
		t.Errorf("Wrong intems in cache: %+v", cache.cache)
	}
}

func TestSetGetRemLRU(t *testing.T) {
	cache := New(3, 0, false, nil)
	for _, ci := range testCIs {
		cache.Set(ci.itemID, ci.value, nil)
	}
	if len(cache.cache) != 3 {
		t.Errorf("Wrong intems in cache: %+v", cache.cache)
	}
	if cache.lruIdx.Len() != 3 {
		t.Errorf("Wrong items in lru index: %+v", cache.lruIdx)
	}
	if cache.lruIdx.Front().Value.(*cachedItem).itemID != "5" {
		t.Errorf("Wrong order of items in the lru index: %+v", cache.lruIdx)
	} else if cache.lruIdx.Back().Value.(*cachedItem).itemID != "3" {
		t.Errorf("Wrong order of items in the lru index: %+v", cache.lruIdx)
	}
	if len(cache.lruRefs) != 3 {
		t.Errorf("Wrong items in lru references: %+v", cache.lruRefs)
	}
	if cache.ttlIdx.Len() != 0 {
		t.Errorf("Wrong items in ttl index: %+v", cache.ttlIdx)
	}
	if len(cache.ttlRefs) != 0 {
		t.Errorf("Wrong items in ttl index: %+v", cache.ttlRefs)
	}
	if _, has := cache.Get("2"); has {
		t.Error("item still in cache")
	}
	// rewrite and reposition 3
	cache.Set("3", "third", nil)
	if val, has := cache.Get("3"); !has {
		t.Error("item not in cache")
	} else if val.(string) != "third" {
		t.Errorf("wrong item value: %v", val)
	}
	if cache.lruIdx.Len() != 3 {
		t.Errorf("Wrong items in lru index: %+v", cache.lruIdx)
	}
	if cache.lruIdx.Front().Value.(*cachedItem).itemID != "3" {
		t.Errorf("Wrong order of items in the lru index: %+v", cache.lruIdx)
	} else if cache.lruIdx.Back().Value.(*cachedItem).itemID != "4" {
		t.Errorf("Wrong order of items in the lru index: %+v", cache.lruIdx)
	}
	cache.Set("2", "second", nil)
	if val, has := cache.Get("2"); !has {
		t.Error("item not in cache")
	} else if val.(string) != "second" {
		t.Errorf("wrong item value: %v", val)
	}
	if cache.lruIdx.Len() != 3 {
		t.Errorf("Wrong items in lru index: %+v", cache.lruIdx)
	}
	if cache.lruIdx.Front().Value.(*cachedItem).itemID != "2" {
		t.Errorf("Wrong order of items in the lru index: %+v", cache.lruIdx)
	} else if cache.lruIdx.Back().Value.(*cachedItem).itemID != "5" {
		t.Errorf("Wrong order of items in the lru index: %+v", cache.lruIdx)
	}
	// 4 should have been removed
	if _, has := cache.Get("4"); has {
		t.Error("item still in cache")
	}
	cache.Remove("2")
	if _, has := cache.Get("2"); has {
		t.Error("item still in cache")
	}
	if len(cache.cache) != 2 {
		t.Errorf("Wrong intems in cache: %+v", cache.cache)
	}
	if cache.lruIdx.Len() != 2 {
		t.Errorf("Wrong items in lru index: %+v", cache.lruIdx)
	}
	if len(cache.lruRefs) != 2 {
		t.Errorf("Wrong items in lru references: %+v", cache.lruRefs)
	}
	if cache.lruIdx.Front().Value.(*cachedItem).itemID != "3" {
		t.Errorf("Wrong order of items in the lru index: %+v", cache.lruIdx)
	} else if cache.lruIdx.Back().Value.(*cachedItem).itemID != "5" {
		t.Errorf("Wrong order of items in the lru index: %+v", cache.lruIdx)
	}
	cache.Clear()
	if cache.Len() != 0 {
		t.Errorf("Wrong intems in cache: %+v", cache.cache)
	}
}

func TestSetGetRemTTLDynamic(t *testing.T) {
	cache := New(UnlimitedCaching, time.Duration(10*time.Millisecond), false, nil)
	for _, ci := range testCIs {
		cache.Set(ci.itemID, ci.value, nil)
	}
	if len(cache.cache) != 5 {
		t.Errorf("Wrong intems in cache: %+v", cache.cache)
	}
	if cache.lruIdx.Len() != 0 {
		t.Errorf("Wrong items in lru index: %+v", cache.lruIdx)
	}
	if len(cache.lruRefs) != 0 {
		t.Errorf("Wrong items in lru references: %+v", cache.lruRefs)
	}
	if cache.ttlIdx.Len() != 5 {
		t.Errorf("Wrong items in ttl index: %+v", cache.ttlIdx)
	}
	if len(cache.ttlRefs) != 5 {
		t.Errorf("Wrong items in ttl index: %+v", cache.ttlRefs)
	}
	time.Sleep(time.Duration(6 * time.Millisecond))
	if _, has := cache.Get("2"); !has {
		t.Error("item not in cache")
	}
	time.Sleep(time.Duration(6 * time.Millisecond))
	if cache.Len() != 1 {
		t.Errorf("Wrong items in cache: %+v", cache.cache)
	}
	if cache.ttlIdx.Len() != 1 {
		t.Errorf("Wrong items in ttl index: %+v", cache.ttlIdx)
	}
	if len(cache.ttlRefs) != 1 {
		t.Errorf("Wrong items in ttl index: %+v", cache.ttlRefs)
	}
}

func TestSetGetRemTTLStatic(t *testing.T) {
	cache := New(UnlimitedCaching, time.Duration(10*time.Millisecond), true, nil)
	for _, ci := range testCIs {
		cache.Set(ci.itemID, ci.value, nil)
	}
	if cache.Len() != 5 {
		t.Errorf("Wrong intems in cache: %+v", cache.cache)
	}
	time.Sleep(time.Duration(6 * time.Millisecond))
	if _, has := cache.Get("2"); !has {
		t.Error("item not in cache")
	}
	time.Sleep(time.Duration(6 * time.Millisecond))
	if cache.Len() != 0 {
		t.Errorf("Wrong items in cache: %+v", cache.cache)
	}
}

func TestSetGetRemLRUttl(t *testing.T) {
	nrItems := 3
	cache := New(nrItems, time.Duration(10*time.Millisecond), false, nil)
	for _, ci := range testCIs {
		cache.Set(ci.itemID, ci.value, nil)
	}
	if cache.Len() != nrItems {
		t.Errorf("Wrong intems in cache: %+v", cache.cache)
	}
	if cache.lruIdx.Len() != nrItems {
		t.Errorf("Wrong items in lru index: %+v", cache.lruIdx)
	}
	if len(cache.lruRefs) != nrItems {
		t.Errorf("Wrong items in lru references: %+v", cache.lruRefs)
	}
	if cache.ttlIdx.Len() != nrItems {
		t.Errorf("Wrong items in ttl index: %+v", cache.ttlIdx)
	}
	if len(cache.ttlRefs) != nrItems {
		t.Errorf("Wrong items in ttl index: %+v", cache.ttlRefs)
	}
	time.Sleep(time.Duration(6 * time.Millisecond))
	cache.Remove("4")
	cache.Set("3", "third", nil)
	nrItems = 2
	if cache.Len() != nrItems {
		t.Errorf("Wrong intems in cache: %+v", cache.cache)
	}
	if cache.lruIdx.Len() != nrItems {
		t.Errorf("Wrong items in lru index: %+v", cache.lruIdx)
	}
	if len(cache.lruRefs) != nrItems {
		t.Errorf("Wrong items in lru references: %+v", cache.lruRefs)
	}
	if cache.ttlIdx.Len() != nrItems {
		t.Errorf("Wrong items in ttl index: %+v", cache.ttlIdx)
	}
	if len(cache.ttlRefs) != nrItems {
		t.Errorf("Wrong items in ttl index: %+v", cache.ttlRefs)
	}
	time.Sleep(time.Duration(6 * time.Millisecond)) // timeout items which were not modified
	nrItems = 1
	if cache.Len() != nrItems {
		t.Errorf("Wrong intems in cache: %+v", cache.cache)
	}
	if cache.lruIdx.Len() != nrItems {
		t.Errorf("Wrong items in lru index: %+v", cache.lruIdx)
	}
	if len(cache.lruRefs) != nrItems {
		t.Errorf("Wrong items in lru references: %+v", cache.lruRefs)
	}
	if cache.ttlIdx.Len() != nrItems {
		t.Errorf("Wrong items in ttl index: %+v", cache.ttlIdx)
	}
	if len(cache.ttlRefs) != nrItems {
		t.Errorf("Wrong items in ttl index: %+v", cache.ttlRefs)
	}
}

func TestCacheDisabled(t *testing.T) {
	cache := New(DisabledCaching, time.Duration(10*time.Millisecond), false, nil)
	for _, ci := range testCIs {
		cache.Set(ci.itemID, ci.value, nil)
		if _, has := cache.Get(ci.itemID); has {
			t.Errorf("Wrong intems in cache: %+v", cache.cache)
		}
	}
	if cache.Len() != 0 {
		t.Errorf("Wrong intems in cache: %+v", cache.cache)
	}
	if cache.lruIdx.Len() != 0 {
		t.Errorf("Wrong items in lru index: %+v", cache.lruIdx)
	}
	if len(cache.lruRefs) != 0 {
		t.Errorf("Wrong items in lru references: %+v", cache.lruRefs)
	}
	if cache.ttlIdx.Len() != 0 {
		t.Errorf("Wrong items in ttl index: %+v", cache.ttlIdx)
	}
	if len(cache.ttlRefs) != 0 {
		t.Errorf("Wrong items in ttl index: %+v", cache.ttlRefs)
	}
	cache.Remove("4")
}

// BenchmarkSetSimpleCache 	10000000	       228 ns/op
func BenchmarkSetSimpleCache(b *testing.B) {
	cache := New(UnlimitedCaching, 0, false, nil)
	rand.Seed(time.Now().UTC().UnixNano())
	min, max := 0, len(testCIs)-1 // so we can have random index
	for n := 0; n < b.N; n++ {
		ci := testCIs[rand.Intn(max-min)+min]
		cache.Set(ci.itemID, ci.value, nil)
	}
}

// BenchmarkGetSimpleCache 	20000000	        99.7 ns/op
func BenchmarkGetSimpleCache(b *testing.B) {
	cache := New(UnlimitedCaching, 0, false, nil)
	for _, ci := range testCIs {
		cache.Set(ci.itemID, ci.value, nil)
	}
	rand.Seed(time.Now().UTC().UnixNano())
	min, max := 0, len(testCIs)-1 // so we can have random index
	for n := 0; n < b.N; n++ {
		ci := testCIs[rand.Intn(max-min)+min]
		cache.Get(ci.itemID)
	}
}

// BenchmarkSetLRU         	 5000000	       316 ns/op
func BenchmarkSetLRU(b *testing.B) {
	cache := New(3, 0, false, nil)
	rand.Seed(time.Now().UTC().UnixNano())
	min, max := 0, len(testCIs)-1 // so we can have random index
	for n := 0; n < b.N; n++ {
		ci := testCIs[rand.Intn(max-min)+min]
		cache.Set(ci.itemID, ci.value, nil)
	}
}

// BenchmarkGetLRU         	20000000	       114 ns/op
func BenchmarkGetLRU(b *testing.B) {
	cache := New(3, 0, false, nil)
	for _, ci := range testCIs {
		cache.Set(ci.itemID, ci.value, nil)
	}
	rand.Seed(time.Now().UTC().UnixNano())
	min, max := 0, len(testCIs)-1 // so we can have random index
	for n := 0; n < b.N; n++ {
		ci := testCIs[rand.Intn(max-min)+min]
		cache.Get(ci.itemID)
	}
}

// BenchmarkSetTTL         	50000000	        30.4 ns/op
func BenchmarkSetTTL(b *testing.B) {
	cache := New(0, time.Duration(time.Millisecond), false, nil)
	rand.Seed(time.Now().UTC().UnixNano())
	min, max := 0, len(testCIs)-1 // so we can have random index
	for n := 0; n < b.N; n++ {
		ci := testCIs[rand.Intn(max-min)+min]
		cache.Set(ci.itemID, ci.value, nil)
	}
}

// BenchmarkGetTTL         	20000000	        88.4 ns/op
func BenchmarkGetTTL(b *testing.B) {
	cache := New(0, time.Duration(5*time.Millisecond), false, nil)
	for _, ci := range testCIs {
		cache.Set(ci.itemID, ci.value, nil)
	}
	rand.Seed(time.Now().UTC().UnixNano())
	min, max := 0, len(testCIs)-1 // so we can have random index
	for n := 0; n < b.N; n++ {
		ci := testCIs[rand.Intn(max-min)+min]
		cache.Get(ci.itemID)
	}
}

// BenchmarkSetLRUttl      	 5000000	       373 ns/op
func BenchmarkSetLRUttl(b *testing.B) {
	cache := New(3, time.Duration(time.Millisecond), false, nil)
	rand.Seed(time.Now().UTC().UnixNano())
	min, max := 0, len(testCIs)-1 // so we can have random index
	for n := 0; n < b.N; n++ {
		ci := testCIs[rand.Intn(max-min)+min]
		cache.Set(ci.itemID, ci.value, nil)
	}
}

// BenchmarkGetLRUttl      	10000000	       187 ns/op
func BenchmarkGetLRUttl(b *testing.B) {
	cache := New(3, time.Duration(5*time.Millisecond), false, nil)
	for _, ci := range testCIs {
		cache.Set(ci.itemID, ci.value, nil)
	}
	rand.Seed(time.Now().UTC().UnixNano())
	min, max := 0, len(testCIs)-1 // so we can have random index
	for n := 0; n < b.N; n++ {
		ci := testCIs[rand.Intn(max-min)+min]
		cache.Get(ci.itemID)
	}
}

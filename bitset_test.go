// Copyright 2011 Will Fitzgerald. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bitset

import (
	"math"
	"math/rand"
	"testing"
)

func TestEmptyBitset(t *testing.T) {
	b := New(0)
	if l := b.Len(); l != 0 {
		t.Errorf("Empty set should be of length 0, not %d", l)
	}
}

func TestBitsetNew(t *testing.T) {
	v := New(16)
	if v.Test(0) != false {
		t.Errorf("Unable to make a bit set and read its 0th value.")
	}
}

func TestBitsetHuge(t *testing.T) {
	v := New(math.MaxUint32)
	if v.Test(0) != false {
		t.Errorf("Unable to make a huge bit set and read its 0th value.")
	}
}

func TestLen(t *testing.T) {
	v := New(1000)
	if l := v.Len(); l != 1000 {
		t.Errorf("Len should be 1000, but is %d.", l)
	}
}

func TestBitsetIsClear(t *testing.T) {
	v := New(1000)
	for i := uint32(0); i < 1000; i++ {
		if v.Test(i) != false {
			t.Errorf("Bit %d is set, and it shouldn't be.", i)
		}
	}
}

func TestExtendOnBoundary(t *testing.T) {
	v := New(32)
	v.Set(32)
	if found := v.Test(31); found {
		t.Error("31 shouldn't have been found")
	}
	if found := v.Test(32); !found {
		t.Error("32 set and then not found")
	}
	if found := v.Test(33); found {
		t.Error("33 shouldn't have been found")
	}
}

func TestExpand(t *testing.T) {
	v := New(0)
	for i := uint32(1000); i > 0; i-- {
		v.Set(i)
		if found := v.Test(i); !found {
			t.Errorf("%d set and then not found", i)
		}
	}
}

func TestBitsetAndGet(t *testing.T) {
	v := New(1000)
	v.Set(100)
	if v.Test(100) != true {
		t.Errorf("Bit %d is clear, and it shouldn't be.", 100)
	}
}

func TestChain(t *testing.T) {
	b := New(1000)
	b.Set(100)
	b.Set(99)
	b.Clear(99)
	if b.Test(100) != true {
		t.Errorf("Bit %d is clear, and it shouldn't be.", 100)
	}
}

func TestOutOfBoundsLong(t *testing.T) {
	v := New(64)
	v.Set(1000)
}

func TestOutOfBoundsClose(t *testing.T) {
	v := New(65)
	v.Set(66)
}

func TestCount(t *testing.T) {
	tot := uint32(64*4 + 11) // just some multi unit64 number
	v := New(tot)
	checkLast := true
	for i := uint32(0); i < tot; i++ {
		sz := v.Count()
		if sz != i {
			t.Errorf("Count reported as %d, but it should be %d", sz, i)
			checkLast = false
			break
		}
		v.Set(i)
	}
	if checkLast {
		sz := v.Count()
		if sz != tot {
			t.Errorf("After all bits set, size reported as %d, but it should be %d", sz, tot)
		}
	}
}

// test setting every 3rd bit, just in case something odd is happening
func TestCount2(t *testing.T) {
	tot := uint32(64*4 + 11) // just some multi unit64 number
	v := New(tot)
	for i := uint32(0); i < tot; i += 3 {
		sz := v.Count()
		if sz != i/3 {
			t.Errorf("Count reported as %d, but it should be %d", sz, i)
			break
		}
		v.Set(i)
	}
}

func TestEqual(t *testing.T) {
	a := New(100)
	b := New(99)
	c := New(100)
	if a.Equal(b) {
		t.Error("Sets of different sizes should be not be equal")
	}
	if !a.Equal(c) {
		t.Error("Two empty sets of the same size should be equal")
	}
	a.Set(99)
	c.Set(0)
	if a.Equal(c) {
		t.Error("Two sets with differences should not be equal")
	}
	c.Set(99)
	a.Set(0)
	if !a.Equal(c) {
		t.Error("Two sets with the same bits set should be equal")
	}
}

func TestUnion(t *testing.T) {
	a := New(100)
	b := New(200)
	for i := uint32(1); i < 100; i += 2 {
		a.Set(i)
		b.Set(i - 1)
	}
	for i := uint32(100); i < 200; i++ {
		b.Set(i)
	}
	c := a.Union(b)
	d := b.Union(a)
	if c.Count() != 200 {
		t.Errorf("Union should have 200 bits set, but had %d", c.Count())
	}
	if !c.Equal(d) {
		t.Errorf("Union should be symmetric")
	}
}

func TestIntersection(t *testing.T) {
	a := New(100)
	b := New(200)
	for i := uint32(1); i < 100; i += 2 {
		a.Set(i)
		b.Set(i - 1)
		b.Set(i)
	}
	for i := uint32(100); i < 200; i++ {
		b.Set(i)
	}
	c := a.Intersection(b)
	d := b.Intersection(a)
	if c.Count() != 50 {
		t.Errorf("Intersection should have 50 bits set, but had %d", c.Count())
	}
	if !c.Equal(d) {
		t.Errorf("Intersection should be symmetric")
	}
}

func TestDifference(t *testing.T) {
	a := New(100)
	b := New(200)
	for i := uint32(1); i < 100; i += 2 {
		a.Set(i)
		b.Set(i - 1)
	}
	for i := uint32(100); i < 200; i++ {
		b.Set(i)
	}
	c := a.Difference(b)
	d := b.Difference(a)
	if c.Count() != 50 {
		t.Errorf("a-b Difference should have 50 bits set, but had %d", c.Count())
	}
	if d.Count() != 150 {
		t.Errorf("b-a Difference should have 150 bits set, but had %d", c.Count())
	}
	if c.Equal(d) {
		t.Errorf("Difference, here, should not be symmetric")
	}
}

func TestSymmetricDifference(t *testing.T) {
	a := New(100)
	b := New(200)
	for i := uint32(1); i < 100; i += 2 {
		a.Set(i)     // 01010101010 ... 0000000
		b.Set(i - 1) // 11111111111111111000000
		b.Set(i)
	}
	for i := uint32(100); i < 200; i++ {
		b.Set(i)
	}
	c := a.SymmetricDifference(b)
	d := b.SymmetricDifference(a)
	if c.Count() != 150 {
		t.Errorf("a^b Difference should have 150 bits set, but had %d", c.Count())
	}
	if d.Count() != 150 {
		t.Errorf("b^a Difference should have 150 bits set, but had %d", c.Count())
	}
	if !c.Equal(d) {
		t.Errorf("SymmetricDifference should be symmetric")
	}
}

func TestComplement(t *testing.T) {
	a := New(50)
	b := a.Complement()
	if b.Count() != 50 {
		t.Errorf("Complement failed, size should be 50, but was %d", b.Count())
	}
	a = New(50)
	a.Set(10)
	a.Set(20)
	a.Set(42)
	b = a.Complement()
	if b.Count() != 47 {
		t.Errorf("Complement failed, size should be 47, but was %d", b.Count())
	}
}

func BenchmarkSet(b *testing.B) {
	b.StopTimer()
	r := rand.New(rand.NewSource(0))
	sz := int32(100000)
	s := New(uint32(sz))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		s.Set(uint32(r.Int31n(sz)))
	}
}

func BenchmarkGetTest(b *testing.B) {
	b.StopTimer()
	r := rand.New(rand.NewSource(0))
	sz := int32(100000)
	s := New(uint32(sz))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		s.Test(uint32(r.Int31n(sz)))
	}
}

func BenchmarkSetExpand(b *testing.B) {
	b.StopTimer()
	sz := uint32(100000)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		s := New(0)
		s.Set(sz)
	}
}

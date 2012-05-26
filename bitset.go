// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bitset

import (
	"bytes"
	"fmt"
	"math"
)

const (
	lWord     uint32 = 32
	lLog2Word uint32 = 5
	allBits   uint32 = 0xffffffff
)

func wordsNeeded(n uint32) uint32 {
	if n == 0 {
		return 1
	} else if n == math.MaxUint32 {
		return math.MaxUint32 >> lLog2Word
	}
	return (n + (lWord - 1)) >> lLog2Word
}

type Bitset struct {
	n uint32
	b []uint32
}

// Returns the current size of the bitset.
func (b *Bitset) Len() uint32 {
	return b.n
}

// Test whether bit i is set.
func (b *Bitset) Test(i uint32) bool {
	if i >= b.n {
		return false
	}
	return ((b.b[i>>lLog2Word] & (1 << (i & (lWord - 1)))) != 0)
}

// Set bit i to 1.
func (b *Bitset) Set(i uint32) {
	if i >= b.n {
		nsize := wordsNeeded(i + 1)
		l := uint32(len(b.b))
		if nsize > l {
			nb := make([]uint32, nsize-l)
			b.b = append(b.b, nb...)
		}
		b.n = i + 1
	}
	b.b[i>>lLog2Word] |= (1 << (i & (lWord - 1)))
}

// Set bit i to 0.
func (b *Bitset) Clear(i uint32) {
	if i >= b.n {
		return
	}
	b.b[i>>lLog2Word] &^= 1 << (i & (lWord - 1))
}

// Flip bit i.
func (b *Bitset) Flip(i uint32) {
	if i >= b.n {
		b.Set(i)
	}
	b.b[i>>lLog2Word] ^= 1 << (i & (lWord - 1))
}

// Clear all bits in the bitset.
func (b *Bitset) ClearAll() {
	for i := range b.b {
		b.b[i] = 0
	}
}

// Get the number of words used in the bitset.
func (b *Bitset) wordCount() uint32 {
	return wordsNeeded(b.n)
}

// Clone the bitset.
func (b *Bitset) Clone() *Bitset {
	c := New(b.n)
	copy(c.b, b.b)
	return c
}

// Copy the bitset into another bitset, returning the size of the destination
// bitset.
func (b *Bitset) Copy(c *Bitset) (n uint32) {
	copy(c.b, b.b)
	n = c.n
	if b.n < c.n {
		n = b.n
	}
	return
}

// http://en.wikipedia.org/wiki/Hamming_weight                                     
const (
	m1 uint32 = 0x55555555 // 0101...
	m2 uint32 = 0x33333333 // 00110011...
	m4 uint32 = 0x0f0f0f0f // 00001111...
)

func popCountUint32(x uint32) uint32 {
	x -= (x >> 1) & m1             // put count of each 2 bits into those 2 bits
	x = (x & m2) + ((x >> 2) & m2) // put count of each 4 bits into those 4 bits 
	x = (x + (x >> 4)) & m4        // put count of each 8 bits into those 8 bits 
	x += x >> 8                    // put count of each 16 bits into their lowest 8 bits
	x += x >> 16                   // put count of each 32 bits into their lowest 8 bits
	return x & 0x7f
}

// Get the number of set bits in the bitset.
func (b *Bitset) Count() uint32 {
	sum := uint32(0)
	for _, w := range b.b {
		sum += popCountUint32(w)
	}
	return sum
}

// Test if two bitsets are equal. Returns true if both bitsets are the same
// size and all the same bits are set in both bitsets.
func (b *Bitset) Equal(c *Bitset) bool {
	if b.n != c.n {
		return false
	}
	for p, v := range b.b {
		if c.b[p] != v {
			return false
		}
	}
	return true
}

// Bitset &^ (and or); difference between receiver and another set.
func (b *Bitset) Difference(ob *Bitset) (result *Bitset) {
	result = b.Clone() // clone b (in case b is bigger than ob)
	szl := ob.wordCount()
	l := uint32(len(b.b))
	for i := uint32(0); i < l; i++ {
		if i >= szl {
			break
		}
		result.b[i] = b.b[i] &^ ob.b[i]
	}
	return
}

func sortByLength(a *Bitset, b *Bitset) (ap *Bitset, bp *Bitset) {
	if a.n <= b.n {
		ap, bp = a, b
	} else {
		ap, bp = b, a
	}
	return
}

// Bitset & (and); intersection of receiver and another set.
func (b *Bitset) Intersection(ob *Bitset) (result *Bitset) {
	b, ob = sortByLength(b, ob)
	result = New(b.n)
	for i, w := range b.b {
		result.b[i] = w & ob.b[i]
	}
	return
}

// Bitset | (or); union of receiver and another set.
func (b *Bitset) Union(ob *Bitset) (result *Bitset) {
	b, ob = sortByLength(b, ob)
	result = ob.Clone()
	szl := ob.wordCount()
	l := uint32(len(b.b))
	for i := uint32(0); i < l; i++ {
		if i >= szl {
			break
		}
		result.b[i] = b.b[i] | ob.b[i]
	}
	return
}

// Bitset ^ (xor); symmetric difference of receiver and another set.
func (b *Bitset) SymmetricDifference(ob *Bitset) (result *Bitset) {
	b, ob = sortByLength(b, ob)
	// ob is bigger, so clone it
	result = ob.Clone()
	szl := b.wordCount()
	l := uint32(len(b.b))
	for i := uint32(0); i < l; i++ {
		if i >= szl {
			break
		}
		result.b[i] = b.b[i] ^ ob.b[i]
	}
	return
}

// Return true if the bitset's length is a multiple of the word size.
func (b *Bitset) isEven() bool {
	return (b.n % lWord) == 0
}

// Clean last word by setting unused bits to 0.
func (b *Bitset) cleanLastWord() {
	if !b.isEven() {
		b.b[wordsNeeded(b.n)-1] &= (allBits >> (lWord - (b.n % lWord)))
	}
}

// Return the (local) complement of a bitset (up to n bits).
func (b *Bitset) Complement() (result *Bitset) {
	b.String()
	result = New(b.n)
	for i, w := range b.b {
		result.b[i] = ^(w)
	}
	result.cleanLastWord()
	return
}

// Returns true if all bits in the bitset are set.
func (b *Bitset) All() bool {
	return b.Count() == b.n
}

// Returns true if no bit in the bitset is set.
func (b *Bitset) None() bool {
	for _, w := range b.b {
		if w > 0 {
			return false
		}
	}
	return true
}

// Return true if any bit in the bitset is set.
func (b *Bitset) Any() bool {
	return !b.None()
}

// Get a string representation of the words in the bitset.
func (b *Bitset) String() string {
	buffer := bytes.NewBufferString("")
	for i := int(wordsNeeded(b.n) - 1); i >= 0; i-- {
		fmt.Fprintf(buffer, "%032b.", b.b[i])
	}
	return string(buffer.Bytes())
}

// Make a new bitset with a starting capacity of n bits. The bitset expands
// automatically.
func New(n uint32) *Bitset {
	return &Bitset{n, make([]uint32, wordsNeeded(n))}
}

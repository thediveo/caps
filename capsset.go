// Copyright 2023 Harald Albrecht.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build linux

package caps

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/exp/slices"
)

// CapabilitiesSet is a set of capabilities.
//
// CapabilitiesSet is independent of any kernel version and its particular set
// width. Instead, it manages capabilities in a dynamically (re)sizing set
// (actually a slice).
type CapabilitiesSet []uint32

// NewCapabilitiesSet returns a new capabilities set. This is more of a
// convenience for those who prefer the "New..." pattern.
func NewCapabilitiesSet() CapabilitiesSet {
	return CapabilitiesSet{}
}

// AllCapabilities returns a new set with all capabilities that the kernel
// supports we're currently running on.
func AllCapabilities() CapabilitiesSet {
	maxindex, maxbitno := wordBitIndices(lastCapability)
	c := make(CapabilitiesSet, maxindex+1)
	for idx := 0; idx < maxindex; idx++ {
		c[idx] = ^uint32(0)
	}
	c[maxindex] = ^uint32(0) >> (31 - maxbitno)
	return c
}

// Clone a set of capabilities into a new and independent set.
func (c CapabilitiesSet) Clone() CapabilitiesSet {
	cl := make(CapabilitiesSet, len(c))
	copy(cl, c)
	return cl
}

// Clear clears all capabilities from this set.
func (c *CapabilitiesSet) Clear() {
	*c = CapabilitiesSet{}
}

// Add (set) one or more effective capabilities identified by their numbers to a set.
func (c *CapabilitiesSet) Add(capno int, morecapnos ...int) {
	capnos := append([]int{capno}, morecapnos...)
	for _, capno := range capnos {
		wordindex, bitno := wordBitIndices(capno)
		c.ensure(wordindex)
		(*c)[wordindex] |= uint32(1) << bitno
	}
}

// Drop (remove) one or more capabilities identified by their numbers to a set.
func (c *CapabilitiesSet) Drop(capno int, morecapnos ...int) {
	capnos := append([]int{capno}, morecapnos...)
	for _, capno := range capnos {
		wordindex, bitno := wordBitIndices(capno)
		if wordindex >= len(*c) {
			continue // no need to expand if the cap isn't in the set anyway.
		}
		(*c)[wordindex] &= ^(uint32(1) << bitno)
	}
}

// Has returns true if the set contains the specified capability (as identified
// by its number).
func (c CapabilitiesSet) Has(capno int) bool {
	wordindex, bitno := wordBitIndices(capno)
	if wordindex >= len(c) {
		return false
	}
	return c[wordindex]&(uint32(1)<<bitno) != 0
}

// Names returns the names of the capabilities in this set, sorted by increasing
// bit number.
func (c CapabilitiesSet) Names() []string {
	names := []string{}
	for idx, w := range c {
		for bit := 0; bit <= 31; bit++ {
			if w&(uint32(1)<<bit) != 0 {
				capno := idx*32 + bit
				name := CapabilityNameByNumber[capno]
				if name == "" {
					name = "CAP_" + strconv.Itoa(capno)
				}
				names = append(names, name)
			}
		}
	}
	return names
}

// SortedNames returns the names of the capabilities in this set in
// lexicographic order, but with "anonymous" capabilities (CAP_ddd) always
// sorted last.
func (c CapabilitiesSet) SortedNames() []string {
	names := c.Names()
	slices.SortFunc(names, cmpCapName)
	return names
}

// cmpCapName orders capability names lexicographically, but with "anonymous"
// capability names coming only after all known capability names.
func cmpCapName(a, b string) int {
	unknownA := isAnonymousCapability(a)
	unknownB := isAnonymousCapability(b)
	if unknownA != unknownB {
		// if one xor the other is an anonymous ("number") capability then we
		// want to go the properly named cap before the number cap.
		switch unknownA {
		case true:
			return 1
		case false:
			return -1
		}
	}
	return strings.Compare(a, b)
}

// isAnonymousCapability returns true if the specified (uppercase) capability
// name is an unknown capability in the form of "CAP_" followed only by digits.
func isAnonymousCapability(name string) bool {
	if !strings.HasPrefix(name, "CAP_") {
		return false
	}
	for idx := len("CAP_"); idx < len(name); idx++ {
		if !unicode.IsDigit(rune(name[idx])) {
			return false
		}
	}
	return true
}

// String returns a textual representation of the capabilities in this set,
// alphabetically sorted by capability (symbol) names.
func (c CapabilitiesSet) String() string {
	names := c.Names()
	slices.Sort(names)
	return strings.Join(names, ", ")
}

// Hex returns the hexadecimal representation of this capabilities set.
func (c CapabilitiesSet) Hex() string {
	h := ""
	size := capDataElements
	if l := len(c); l > size {
		size = l
	}
	for idx := size - 1; idx >= 0; idx-- {
		v := uint32(0)
		if idx < len(c) {
			v = c[idx]
		}
		h = h + fmt.Sprintf("%08x", v)
	}
	return h
}

// CapabilitiesFromHex parses the given hexadecimal string into a capabilities
// set. If the string representation is invalid then an error is returned
// instead, together with a zero capabilities set.
func CapabilitiesFromHex(h string) (CapabilitiesSet, error) {
	b, err := hex.DecodeString(h)
	if err != nil {
		return nil, err
	}
	b = append([]byte{0x00, 0x00, 0x00}[:(4-len(b)&3)&3], b...)
	c := CapabilitiesSet(make([]uint32, 0, len(b)>>2))
	for idx := len(b) - 4; idx >= 0; idx -= 4 {
		c = append(c,
			(uint32(b[idx])<<24)+
				(uint32(b[idx+1])<<16)+
				(uint32(b[idx+2])<<8)+
				uint32(b[idx+3]))
	}
	return c, nil
}

// returns the word element index as well as the bit number corresponding with
// the specified capability (bit) number.
func wordBitIndices(capno int) (wordindex, bitno int) {
	if capno < 0 {
		panic(fmt.Sprintf("invalid negative capability bit number %d", capno))
	}
	return capno >> 5, capno & 31
}

// ensures that are enough elements up to and including the element at
// wordoffset.
func (c *CapabilitiesSet) ensure(wordindex int) {
	if wordindex >= len(*c) {
		*c = append(*c, make([]uint32, wordindex-len(*c)+1)...)
	}
}

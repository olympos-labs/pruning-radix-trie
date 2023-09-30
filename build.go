// Copyright 2023 Jean Niklas L'orange.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ptrie

import (
	"sort"
	"unicode/utf8"
)

// We can build the trie incrementally, but this is simpler. I'll keep it like
// this unless there are some good reasons for not having it like this.

// FromItems returns an immutable PTrie from the given items. The item terms
// must be distinct.
func FromItems[T any](items []Item[T]) *PTrie[T] {
	// sort by largest first. That way we should have good cache locality (...
	// statistically anyway) _and_ it simplifies the building algorithm by a lot.
	sort.Slice(items, func(i, j int) bool {
		return items[i].Rank > items[j].Rank
	})

	var root PTrie[T]
	for _, item := range items {
		root.insertItem(item)
	}

	return &root
}

// insertItem inserts item into pt. The item's rank must be equal or lower
// than the max rank of pt.
func (pt *PTrie[T]) insertItem(item Item[T]) {
	pt.maxRank = item.Rank
	cur := pt
	var scratch [utf8.UTFMax]byte
	for _, r := range item.Term {
		bs := utf8.AppendRune(scratch[:0], r)
		for i, b := range bs {
			isLastByte := i+1 == len(bs)
			cur = cur.focusByteOrNew(b, isLastByte, item.Rank)
		}
	}

	cur.item = item
}

func (pt *PTrie[T]) focusByteOrNew(c byte, isLastByte bool, curRank uint) *PTrie[T] {
	cur := pt.FocusByte(c)
	if cur != nil {
		return cur
	}
	cur = &PTrie[T]{
		char:       c,
		parent:     pt,
		depth:      pt.depth + 1,
		insideRune: !isLastByte,
		maxRank:    curRank,
	}
	pt.children = append(pt.children, cur)
	return cur
}

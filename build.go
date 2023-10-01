// Copyright 2023 Jean Niklas L'orange.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ptrie

import (
	"sort"
)

// We can build the trie incrementally, but this is simpler. I'll keep it like
// this for now.

// FromItems returns an immutable PTrie from the given items. The item terms
// must be distinct.
func FromItems[T any](items []Item[T]) *PTrie[T] {
	// sort by highest first. That way we should have good cache locality (...
	// statistically anyway) _and_ it simplifies the building algorithm by a lot.
	sort.Slice(items, func(i, j int) bool {
		return items[i].Rank > items[j].Rank
	})

	var root PTrie[T]
	for _, item := range items {
		root.insertItem(item.Term, item)
	}

	return &root
}

// insertItem inserts item into pt. The item's rank must be equal or lower
// than the max rank of pt.
func (pt *PTrie[T]) insertItem(term string, item Item[T]) {
	pt.maxRank = max(pt.maxRank, item.Rank)
	for i, child := range pt.children {
		c, numCommon := child.compare(term)
		switch c {
		case cmpEqual:
			// can happen if longer words make a node. Put the item in, should be
			// sufficient.
			child.item = item
			return
		case cmpNoMatch:
			// move on to next child
		case cmpSubkey:
			// item term is smaller, replace child with new node
			child.suffixLen -= numCommon
			pt.children[i] = &PTrie[T]{
				children:  []*PTrie[T]{child},
				maxRank:   child.maxRank,
				item:      item,
				suffixLen: len(term) - numCommon,
			}
			return
		case cmpSuperkey:
			// item term is larger, recurse
			child.insertItem(term[numCommon:], item)
			return
		case cmpSharedPrefix:
			// shared prefix, but differing suffixes. Replace child with a parent node
			// containing them both.
			child.suffixLen -= numCommon
			newChild := &PTrie[T]{
				maxRank:   item.Rank,
				item:      item,
				suffixLen: len(term) - numCommon,
			}
			pt.children[i] = &PTrie[T]{
				children: []*PTrie[T]{child, newChild},
				maxRank:  child.maxRank,
				item: Item[T]{
					Term: term[:numCommon],
				},
				suffixLen: numCommon,
			}
			return
		}
	}

	// not part of any existing node, so append new node instead:
	pt.children = append(pt.children,
		&PTrie[T]{
			maxRank:   item.Rank,
			item:      item,
			suffixLen: len(term),
		},
	)
}

func (pt *PTrie[T]) compare(term string) (cmp, int) {
	m := min(len(term), pt.suffixLen)
	ptTerm := pt.term()
	numCommon := 0
	for i := 0; i < m; i++ {
		if term[i] != ptTerm[i] {
			break
		}
		numCommon++
	}
	switch {
	case numCommon == 0:
		return cmpNoMatch, numCommon
	case len(term) == pt.suffixLen && numCommon == m:
		return cmpEqual, numCommon
	case len(term) < pt.suffixLen && numCommon == m:
		return cmpSubkey, numCommon
	case pt.suffixLen < len(term) && numCommon == m:
		return cmpSuperkey, numCommon
	case numCommon < m:
		return cmpSharedPrefix, numCommon
	default:
		panic("unhandled case, logic error")
	}
}

func (pt *PTrie[T]) term() string {
	return pt.item.Term[len(pt.item.Term)-pt.suffixLen:]
}

type cmp int

const (
	cmpEqual cmp = iota + 1
	cmpNoMatch
	cmpSubkey
	cmpSuperkey
	cmpSharedPrefix
)

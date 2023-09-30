// Copyright 2023 Jean Niklas L'orange.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package ptrie implements a Pruning Radix Trie.
package ptrie

import "unicode/utf8"

// Item is an element store in a PTrie. If T is immutable, so is the item.
type Item[T any] struct {
	Value T
	Term  string
	Rank  uint
}

// PTrie is a Pruning Radix Trie.
type PTrie[T any] struct {
	parent   *PTrie[T]
	children []*PTrie[T]
	depth    int
	maxRank  uint // highest rank in this subtrie
	item     Item[T]

	char       byte
	insideRune bool
}

// FocusString returns the subtrie of pt with the given prefix.
func (pt *PTrie[T]) FocusString(prefix string) *PTrie[T] {
	cur := pt
	for _, b := range []byte(prefix) {
		cur = cur.FocusByte(b)
		if cur == nil {
			return nil
		}
	}
	return cur
}

// FocusRune returns the subtrie of pt with c as its prefix.
func (pt *PTrie[T]) FocusRune(c rune) *PTrie[T] {
	bs := utf8.AppendRune(nil, c)
	cur := pt
	for _, b := range bs {
		cur = cur.FocusByte(b)
		if cur == nil {
			return nil
		}
	}
	return cur
}

// FocusByte returns the subtrie of pt with c as its prefix.
func (pt *PTrie[T]) FocusByte(c byte) *PTrie[T] {
	for _, child := range pt.children {
		if child.char == c {
			return child
		}
	}
	return nil
}

// UnfocusByte returns the parent of pt, the equivalent of removing the last
// byte in the prefix.
func (pt *PTrie[T]) UnfocusByte() *PTrie[T] {
	return pt.parent
}

// UnfocusRune returns the first ancestor of pt that is a sequence of valid
// utf-8 characters, the equivalent of removing the last rune in the prefix.
func (pt *PTrie[T]) UnfocusRune() *PTrie[T] {
	cur := pt.parent
	for cur.insideRune {
		cur = pt.parent
	}
	return cur
}

// FindTopK returns the k items with the highest rank in pt with the prefix
// `term`, where the highest ranked item is the first element etc. There may be
// less than k elements in the result.
func (pt *PTrie[T]) FindTopK(prefix string, k int) []Item[T] {
	return pt.FindTopKFast(prefix, make([]Item[T], 0, k))
}

func (pt *PTrie[T]) FindTopKFast(prefix string, result []Item[T]) []Item[T] {
	best := result[:0]

	// first, walk pt until we have term sliced off:
	cur := pt.FocusString(prefix)
	if cur == nil {
		return nil
	}

	// then, recursively walk the current subtrie
	best = cur.walk(best)

	return best
}

func (pt *PTrie[T]) walk(result Items[T]) []Item[T] {
	if pt.containsItem() {
		result = result.insert(pt)
	}
	for _, child := range pt.children {
		if !result.mustWalk(child) {
			// since we sort children by max rank, any child after us will also
			// return false for mustWalk, so terminate early.
			break
		}
		result = child.walk(result)
	}
	return result
}

func (pt *PTrie[T]) containsItem() bool {
	return pt.item.Rank != 0
}

type Items[T any] []Item[T]

func (result Items[T]) worst() Item[T] {
	return result[len(result)-1]
}

func (result Items[T]) insert(subTrie *PTrie[T]) Items[T] {
	if !result.shouldInsert(subTrie) {
		return result
	}

	item := subTrie.item

	// perform an insertion sort. Fast enough for small k
	for i := range result {
		if result[i].Rank < item.Rank {
			result[i], item = item, result[i]
		}
	}
	if len(result) < cap(result) {
		result = append(result, item)
	}
	return result
}

func (result Items[T]) shouldInsert(subTrie *PTrie[T]) bool {
	// insert if we don't have k results yet
	if len(result) < cap(result) {
		return true
	}
	if result.worst().Rank < subTrie.item.Rank {
		return true
	}
	return false
}

func (result Items[T]) mustWalk(subTrie *PTrie[T]) bool {
	// must walk if we don't have k results yet
	if len(result) < cap(result) {
		return true
	}
	// must walk if there's at least one element with better rank in the subtree
	// than the worst element in the current resultset
	if result.worst().Rank < subTrie.maxRank {
		return true
	}

	return false
}

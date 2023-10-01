// Copyright 2023 Jean Niklas L'orange.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package ptrie implements a Pruning Radix Trie.
package ptrie

// Item is an element store in a PTrie. If T is immutable, so is the item.
type Item[T any] struct {
	Value T
	Term  string
	Rank  uint
}

// PTrie is a Pruning Radix Trie.
type PTrie[T any] struct {
	item      Item[T]
	children  []*PTrie[T]
	maxRank   uint // highest rank in this subtrie
	suffixLen int
}

// FindTopK returns the k items with the highest rank in pt with the prefix
// `term`, where the highest ranked item is the first element etc. There may be
// less than k elements in the result.
func (pt *PTrie[T]) FindTopK(prefix string, k int) []Item[T] {
	return pt.FindTopKFast(prefix, make([]Item[T], 0, k))
}

// FindTopKFast works like [PTrie.FindTopK], but is able to reuse an already
// allocated result slice for the items. The slice will be filled up to its
// capacity if there are enough elements that match the prefix.
func (pt *PTrie[T]) FindTopKFast(prefix string, result []Item[T]) []Item[T] {
	best := result[:0]

	// first, walk pt until we have term sliced off:
	lca := pt.lcaScan(prefix)
	if lca == nil {
		return nil
	}
	best = lca.walk(best)

	return best
}

// lcaScan returns the lowest common ancestor subtrie containing all items with
// the given prefix.
func (pt *PTrie[T]) lcaScan(prefix string) *PTrie[T] {
	for _, child := range pt.children {
		c, numCommon := child.compare(prefix)
		switch c {
		case cmpNoMatch:
			// check other children
		case cmpEqual:
			return child
		case cmpSubkey:
			return child
		case cmpSuperkey:
			return child.lcaScan(prefix[numCommon:])
		case cmpSharedPrefix:
			return nil
		}
	}
	return pt
}

func (pt *PTrie[T]) hasPrefix(prefix string) bool {
	if pt.suffixLen < len(prefix) {
		return false
	}
	ptTerm := pt.term()
	for i := 0; i < len(prefix); i++ {
		if prefix[i] != ptTerm[i] {
			return false
		}
	}
	return true
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

	// append the item, or set to last (least) element
	if len(result) < cap(result) {
		result = append(result, item)
	} else {
		result[len(result)-1] = item
	}

	// guard for indexing in loop, can skip if only 1 result so far
	if len(result) > 1 {
		// reverse bubble sort with early stop
		for i := len(result)-2; i>=0; i-- {
			if result[i].Rank < result[i+1].Rank {
				result[i], result[i+1] = result[i+1], result[i]
			} else {
				break
			}
		}
	}

	return result
}

func (result Items[T]) shouldInsert(subTrie *PTrie[T]) bool {
	// insert if we don't have k results yet
	if len(result) < cap(result) {
		return true
	}
	// or if we're better than the worst result
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

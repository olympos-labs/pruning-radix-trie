// Copyright 2023 Jean Niklas L'orange.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ptrie

import (
	"bufio"
	"compress/gzip"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPTrie(t *testing.T) {
	items := readItems(t, "min-terms.txt.gz")

	trie := FromItems(items)

	for m := 1; m <= len("microsoft"); m++ {
		target := "microsoft"[:m]
		printFound(t, trie, target)
	}
}

func printFound(t *testing.T, pt *PTrie[empty], prefix string) {
	t.Log("matches for", prefix)
	found := pt.FindTopK(prefix, 10)
	for _, elem := range found {
		t.Logf("%s: %d", elem.Term, elem.Rank)
	}
	t.Log()
}

// go test -bench=. -benchmem
//
// goos: linux
// goarch: amd64
// pkg: olympos.io/container/pruning-radix-trie
// cpu: AMD Ryzen 7 5800X 8-Core Processor
// BenchmarkMicrosoft/m-16  	 2117851	       567.7 ns/op	       0 B/op	       0 allocs/op
// BenchmarkMicrosoft/mi-16 	 2167874	       560.4 ns/op	       0 B/op	       0 allocs/op
// BenchmarkMicrosoft/mic-16         	 2885617	       415.1 ns/op	       0 B/op	       0 allocs/op
// BenchmarkMicrosoft/micr-16        	 2208138	       538.9 ns/op	       0 B/op	       0 allocs/op
// BenchmarkMicrosoft/micro-16       	 2225856	       539.2 ns/op	       0 B/op	       0 allocs/op
// BenchmarkMicrosoft/micros-16      	 3643194	       334.2 ns/op	       0 B/op	       0 allocs/op
// BenchmarkMicrosoft/microso-16     	 8094313	       147.8 ns/op	       0 B/op	       0 allocs/op
// BenchmarkMicrosoft/microsof-16    	14714454	        82.43 ns/op	       0 B/op	       0 allocs/op
// BenchmarkMicrosoft/microsoft-16   	14215828	        84.88 ns/op	       0 B/op	       0 allocs/op

func BenchmarkMicrosoft(b *testing.B) {
	items := readItems(b, "min-terms.txt.gz")

	trie := FromItems(items)
	for m := 1; m <= len("microsoft"); m++ {
		target := "microsoft"[:m]
		b.Run(target, func(b *testing.B) {
			resultSlice := make([]Item[empty], 0, 10)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				resultSlice = trie.FindTopKFast(target, resultSlice)
			}
		})
	}
}

type empty struct{}

func readItems(t require.TestingT, fname string) []Item[empty] {
	var items []Item[empty]

	f, err := os.Open(fname)
	require.NoError(t, err)
	defer func() { require.NoError(t, f.Close()) }()
	gf, err := gzip.NewReader(f)
	require.NoError(t, err)
	defer func() { require.NoError(t, gf.Close()) }()
	scanner := bufio.NewScanner(gf)
	for scanner.Scan() {
		term, after, found := strings.Cut(strings.TrimSpace(scanner.Text()), "\t")
		require.True(t, found)
		count, err := strconv.Atoi(after)
		require.NoError(t, err)
		items = append(items, Item[empty]{
			Term: term,
			Rank: uint(count),
		})
	}

	return items
}

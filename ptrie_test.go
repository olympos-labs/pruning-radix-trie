// Copyright 2023 Jean Niklas L'orange.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ptrie

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPTrie(t *testing.T) {
	items := readItems(t, "min-terms.txt")

	trie := FromItems(items)

	for m := 1; m <= len("microsoft"); m++ {
		target := "microsoft"[:m]
		printFound(trie, target)
	}
}

// go test -bench=. -benchmem
//
// goos: linux
// goarch: amd64
// pkg: olympos.io/container/ptrie
// cpu: AMD Ryzen 7 5800X 8-Core Processor
// BenchmarkMicrosoft/m-16  	 1249870	       963.8 ns/op	       0 B/op	       0 allocs/op
// BenchmarkMicrosoft/mi-16 	 2029866	       590.1 ns/op	       0 B/op	       0 allocs/op
// BenchmarkMicrosoft/mic-16         	 3008647	       399.3 ns/op	       0 B/op	       0 allocs/op
// BenchmarkMicrosoft/micr-16        	 1780680	       673.8 ns/op	       0 B/op	       0 allocs/op
// BenchmarkMicrosoft/micro-16       	 1775973	       673.7 ns/op	       0 B/op	       0 allocs/op
// BenchmarkMicrosoft/micros-16      	 3389410	       351.1 ns/op	       0 B/op	       0 allocs/op
// BenchmarkMicrosoft/microso-16     	 9875887	       121.9 ns/op	       0 B/op	       0 allocs/op
// BenchmarkMicrosoft/microsof-16    	63202345	        18.74 ns/op	       0 B/op	       0 allocs/op
// BenchmarkMicrosoft/microsoft-16   	67311615	        17.46 ns/op	       0 B/op	       0 allocs/op

func BenchmarkMicrosoft(b *testing.B) {
	items := readItems(b, "min-terms.txt")

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

func printFound(pt *PTrie[empty], prefix string) {
	fmt.Println("matches for", prefix)
	found := pt.FindTopK(prefix, 10)
	for _, elem := range found {
		fmt.Printf("%s: %d\n", elem.Term, elem.Rank)
	}
	fmt.Println()
}

type empty struct{}

func readItems(t require.TestingT, fname string) []Item[empty] {
	var items []Item[empty]

	f, err := os.Open(fname)
	require.NoError(t, err, "if the file's not found, gunzip the file")
	defer f.Close()
	scanner := bufio.NewScanner(f)
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

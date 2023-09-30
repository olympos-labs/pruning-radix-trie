# Pruning Radix Trie

A Go port of the [Pruning Radix
Trie](https://github.com/wolfgarbe/PruningRadixTrie), an augmented patricia trie
that orders nodes based upon the maximum _rank_ of the items in their subtree.
This allows it to quickly search and find _k_ elements with the highest rank.
(You decide what the rank is for each item.)

The most obvious use case would be autocomplete: Finding the top 10 used words
with a given prefix is cheap with this data structure.

## Implementation Notes

This implementation is an uncompressed radix trie, not yet a Patricia trie. I'm
working on that part.

## Thoughts

Both this and the original implementation is better suited for bigger machines
(e.g. servers), as they trade no memory allocations for higher memory usage. It
may be more appropriate to trade more allocations for lower memory usage and
ship it to the client instead: API calls are much more heavy than memory
lookups, after all.

## License

Copyright © 2023 Jean Niklas L'orange

Distributed under the BSD 3-clause license, which is available in the file
LICENSE.

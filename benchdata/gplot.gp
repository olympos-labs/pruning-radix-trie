set terminal png size 800,600 font 'TeX Gyre Pagella'

set output "performance.png"

#set key inside left top box linetype -1 linewidth 1.000
set key inside right top box linetype -1 linewidth 1.000 width 6
set bars 2.0

set boxwidth 0.15

set yrange [0:]

set title 'Lookup time for top 10 ranked results from a \~6 million term dictionary'
set ylabel 'ns/op'
set xlabel 'prefix length on the word "microsoft"'
set format y '%.0f'

plot "uncompressed.dat" using 1:2 title 'Uncompressed trie' with linespoints, \
     "compressed-v1.dat" using 1:2 title 'Compressed trie (v1)' with linespoints, \
     "compressed.dat" using 1:2 title 'Compressed trie (current)' with linespoints

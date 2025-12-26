package main

import (
	"fmt"
	"hash/fnv"
	"math"
)

type Sketch struct {
	width  int
	depth  int
	table  [][]int
	hashes []uint64
}

func CountMinSketch(width, depth int) *Sketch {
	return &Sketch{
		width:  width,
		depth:  depth,
		table:  make([][]int, depth),
		hashes: make([]uint64, depth),
	}
}

func (cms *Sketch) Init() {
	for i := range cms.table {
		cms.table[i] = make([]int, cms.width)
		cms.hashes[i] = uint64(i + 1)
	}
}

func (cms *Sketch) HashIndex(s string, seed uint64) int {
	h := fnv.New64a()
	h.Write([]byte(s))
	v := h.Sum64()
	mixed := v + seed*0x5bd1e995 // небольшая простая "соль"
	return int(mixed % uint64(cms.width))
}

func (cms *Sketch) Add(s string, count int) {
	fmt.Printf("Добавляем элемент '%s'", s)
	for i := 0; i < cms.depth; i++ {
		idx := cms.HashIndex(s, cms.hashes[i])
		cms.table[i][idx] += count
		fmt.Printf(" \n Хэш %d -> индекс %d, новое значение %d\n", i+1, idx, cms.table[i][idx])
	}
}

func (cms *Sketch) Count(s string) int {
	min := math.MaxInt32
	for i := 0; i < cms.depth; i++ {
		idx := cms.HashIndex(s, cms.hashes[i])
		if cms.table[i][idx] < min {
			min = cms.table[i][idx]
		}
	}
	return min
}

func main() {

	width := 10
	depth := 3
	cms := CountMinSketch(width, depth)
	cms.Init()

	keys := []string{"apple", "banana", "apple", "orange", "banana", "apple"}

	// Наивный счётчик
	seen := make(map[string]int)

	for _, key := range keys {

		seen[key]++
		cms.Add(key, 1)

	}

	fmt.Println("\nПроверка частот ")
	for _, key := range []string{"apple", "banana", "orange", "grape"} {
		fmt.Printf("Элемент '%s': Наивный счётчик=%d, CMS=%d\n", key, seen[key], cms.Count(key))
	}
}

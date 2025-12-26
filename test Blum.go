package main

import (
	"fmt"
	"hash/fnv"
	"runtime"
)

type BloomFilter struct {
	bitSet    []uint64
	size      int
	hashCount int
}

func NewBloomFilter(size, hashCount int) *BloomFilter {
	bitSetSize := size / 64
	if size%64 != 0 {
		bitSetSize++
	}
	return &BloomFilter{
		bitSet:    make([]uint64, bitSetSize),
		size:      size,
		hashCount: hashCount,
	}
}

func (bf *BloomFilter) Add(item []byte, verbose bool) {
	h1 := fnv.New64a()
	h2 := fnv.New64a()
	h1.Write(item)
	h2.Write(item)
	v1 := h1.Sum64()
	v2 := h2.Sum64()

	for i := 0; i < bf.hashCount; i++ {
		mixed := v1 + uint64(i)*v2
		idx := int(mixed % uint64(bf.size))
		cell := idx / 64   // номер uint64 в массиве
		bitPos := idx % 64 // позиция бита внутри ячейки
		bf.bitSet[cell] |= 1 << bitPos

		if verbose {
			fmt.Printf("Хэш %d для '%s': mixed=%d, общий индекс=%d, ячейка=%d, бит=%d, устанавливаем 1\n",
				i, string(item), mixed, idx, cell, bitPos)
		}
	}
}

func (bf *BloomFilter) Naiv(item []byte, verbose bool) bool {
	h1 := fnv.New64a()
	h2 := fnv.New64a()
	h1.Write(item)
	h2.Write(item)
	v1 := h1.Sum64()
	v2 := h2.Sum64()

	for i := 0; i < bf.hashCount; i++ {
		mixed := v1 + uint64(i)*v2
		idx := int(mixed % uint64(bf.size))
		bitSet := bf.bitSet[idx/64]&(1<<(idx%64)) != 0

		if verbose {
			fmt.Printf("Проверка хэш %d для '%s': индекс=%d, бит=%v\n", i, string(item), idx, bitSet)
		}

		if !bitSet {
			return false
		}
	}
	return true
}

func main() {
	fmt.Println("Фильтр Блума")

	var n int
	n = 5
	hashCount := 3
	keys := make([]string, n)
	for i := 0; i < n; i++ {
		keys[i] = fmt.Sprintf("element-%d", i)

	}

	// Создаем фильтр Блума
	filter := NewBloomFilter(n*10, hashCount)

	verbose := true
	fmt.Println("\nДобавляем элементы в фильтр:")
	for i, key := range keys {
		filter.Add([]byte(key), verbose && i < 10) // подробный вывод только для первых 10 элементов
	}

	fmt.Println("\nПроверяем элементы в фильтре:")
	realCount := 0
	foundCount := 0
	for i, key := range keys {
		inBloom := filter.Naiv([]byte(key), verbose && i < 10)
		realCount++
		if inBloom {
			foundCount++
		}
	}

	fmt.Printf("\nРеальное количество элементов: %d\n", realCount)
	fmt.Printf("Алгоритм нашел в фильтре: %d\n", foundCount)

	// Память фильтра
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("Память фильтра: %d байт\n", len(filter.bitSet)*8)
}

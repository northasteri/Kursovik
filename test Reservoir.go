package main

import (
	"fmt"
	"math/rand"
	"time"
)

type Reservoir struct {
	k      int
	count  int64
	sample []int
}

// Создание нового резервуара
func NReservoir(k int) *Reservoir {
	return &Reservoir{
		k:      k,
		count:  0,
		sample: make([]int, 0, k),
	}
}

// Добавление элемента в резервуар
func (r *Reservoir) Add(x int, verbose bool) {
	r.count++
	if len(r.sample) < r.k {
		r.sample = append(r.sample, x)
		if verbose {
			fmt.Printf("Добавляем %d: Резервуар = %v\n", x, r.sample)
		}
		return
	}
	j := rand.Int63n(r.count)
	if j < int64(r.k) {
		if verbose {
			fmt.Printf("Заменяем элемент %d на %d (j=%d)\n", r.sample[j], x, j)
		}
		r.sample[j] = x
		if verbose {
			fmt.Printf("Резервуар = %v\n", r.sample)
		}
	} else if verbose {
		fmt.Printf("Пропускаем %d (j=%d)\n", x, j)
	}
}

func (r *Reservoir) Sample() []int {
	return r.sample
}

func main() {
	rand.Seed(time.Now().UnixNano())

	var n int

	n = 10

	k := 5 // маленький резервуар для наглядности
	res := NReservoir(k)

	verbose := n <= 20 // показываем пошагово только для малых потоков
	fmt.Printf("Резервуарная выборка (k=%d) для потока из %d элементов\n\n", k, n)

	// Поток данных: элементы от 1 до n
	for i := 1; i <= n; i++ {
		res.Add(i, verbose)
	}

	sample := res.Sample()
	fmt.Printf("\nИтоговая выборка (%d элементов): %v\n", len(sample), sample)

}

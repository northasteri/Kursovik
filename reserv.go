package main

import (
	"fmt"
	"math"
	"math/rand"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type Reservoir struct {
	k      int
	count  int64
	sample []int
}

func NReservoir(k int) *Reservoir {
	return &Reservoir{
		k:      k,
		count:  0,
		sample: make([]int, 0, k),
	}
}

func (r *Reservoir) Add(x int) {
	r.count++
	if len(r.sample) < r.k {
		r.sample = append(r.sample, x)
		return
	}
	j := rand.Int63n(r.count)
	if j < int64(r.k) {
		r.sample[j] = x
	}
}

func (r *Reservoir) Sample() []int {
	return r.sample
}

// Наивный алгоритм
func NaiveSample(data []int, k int) []int {
	copyData := make([]int, len(data))
	copy(copyData, data) // создаём отдельную память
	rand.Shuffle(len(copyData), func(i, j int) {
		copyData[i], copyData[j] = copyData[j], copyData[i]
	})
	return copyData[:k]
}

func main() {
	fmt.Println("Reservoir sampling")

	var n1 int
	for {
		fmt.Print("\nВведите количество элементов (до 15 млн): ")
		var input string
		if _, err := fmt.Scanln(&input); err != nil {
			fmt.Println("Некорректный ввод")
			continue
		}
		if _, err := strconv.Atoi(input); err != nil || strings.TrimSpace(input) == "" {
			fmt.Println("Некорректный ввод")
			continue
		}
		n1, _ = strconv.Atoi(input)
		if n1 <= 0 || n1 > 15_000_000 {
			fmt.Println("Некорректный ввод")
			continue
		}
		break
	}

	k := 100_000
	n := int64(n1)

	// создаем массив данных
	data := make([]int, n)
	for i := int64(0); i < n; i++ {
		data[i] = int(i)
	}

	// Создаем объект резервуара
	res := NReservoir(k)

	// Reservoir sampling
	startR := time.Now()
	for _, value := range data {
		res.Add(value)
	}
	resSample := res.Sample()
	timeR := time.Since(startR)

	// Наивный алгоритм
	startNaive := time.Now()
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)

	naiveSample := NaiveSample(data, k)

	timeNaive := time.Since(startNaive)
	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)
	naiveMemory := m2.Alloc - m1.Alloc

	p := float64(k) / float64(n)
	AbsError := math.Sqrt(float64(k) * p * (1 - p))
	relError := math.Sqrt((1 - p) / (float64(k) * p))

	fmt.Printf("Средняя абсолютная ошибка Reservoir: %.2f\n", AbsError)
	fmt.Printf("Средняя относительная ошибка Reservoir: %.2f%%\n", relError*100)

	fmt.Printf("Память Reservoir sampling: %d байт\n", k*8)
	fmt.Printf("Память Naive: %d байт\n", naiveMemory)

	fmt.Printf("Наивный алгоритм: %v\n", timeNaive)
	fmt.Printf("Reservoir sampling: %v\n", timeR)

	fmt.Printf("Reservoir sampling first 10: %v\n", resSample[:10])
	fmt.Printf("Naive first 10: %v\n", naiveSample[:10])
}

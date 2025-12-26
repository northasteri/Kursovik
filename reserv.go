// Объявление пакета main - точка входа в программу на Go
package main

// Импорт необходимых пакетов:
import (
	"fmt" // для форматированного ввода-вывода
	"math"
	"math/rand" // для генерации случайных чисел
	"runtime"
	"time"
)

type Reservoir struct {
	k      int   // сколько элементов нужно выбрать
	count  int64 //  счетчик обработанных элементов
	sample []int // массив выбранных элементов
}

func NReservoir(k int) *Reservoir {
	return &Reservoir{
		k:      k,
		count:  0,
		sample: make([]int, 0, k), // с емкостью k, длиной 0
	}
}

// Add добавляет новый элемент в резервуарную выборку

func (r *Reservoir) Add(x int) {
	// Увеличиваем счетчик обработанных элементов
	r.count++

	// Если резервуар еще не заполнен полностью (меньше k элементов)
	if len(r.sample) < r.k {
		// Добавляем элемент в конец
		r.sample = append(r.sample, x)
		return
	}

	// случайное число от 0 до count-1
	j := rand.Int63n(r.count)

	// Если случайное число меньше размера выборки k
	if j < int64(r.k) {
		// Заменяем элемент в позиции j
		r.sample[j] = x
	}
}

// текущая выборка из резервуара
func (r *Reservoir) Sample() []int {
	return r.sample
}

// Наивный алгоритм: собираем все элементы, затем выбираем k случайных
// data - исходный массив данных
// k - количество элементов для выборки
func NaiveSample(data []int, k int) []int {

	// Создаем копию исходных данных, чтобы не изменять оригинальный массив
	Data := make([]int, len(data))
	copy(Data, data) // копируем элементы из data в copyData

	// rand.Shuffle случайным образом перемешивает последовательность
	//генерирует пары случайных индексов (i, j) функция определяет, как поменять
	//местами элементы с этими индексами
	rand.Shuffle(len(Data), func(i, j int) {
		// Меняем местами элементы с индексами i и j
		Data[i], Data[j] = Data[j], Data[i]
	})

	// Возвращаем первые k элементов перемешанного массива
	return Data[:k]
}

func calc(counts []int, expected float64) (absMean, relMean float64) {
	n := float64(len(counts))
	var sumAbs, sumRel float64
	for _, c := range counts {
		abs := math.Abs(float64(c) - expected)
		rel := abs / expected
		sumAbs += abs
		sumRel += rel
	}
	absMean = sumAbs / n
	relMean = sumRel / n
	return
}

func main() {

	fmt.Printf("Reservoir sampling\n")
	var n1 int // n1 - общее количество элементов в потоке
	for {
		fmt.Print("Введите количество входящих элементов(выбираем 100тыс): ")
		if _, err := fmt.Scan(&n1); err == nil && n1 >= 100_000 {
			break
		}
		fmt.Println("Некорректный ввод")
	}

	k := 100_000 // k сколько элементов выбрать
	n := int64(n1)
	// Симулируем поток данных: создаем массив из n элементов
	data := make([]int, n) // создаем слайс целых чисел длиной n

	// Заполняем массив значениями от 0 до n-1
	for i := int64(0); i < n; i++ {
		data[i] = int(i) // преобразуем int64 в int
	}

	// резерв

	startR := time.Now()
	// Создаем новый экземпляр резервуарной выборки
	res := NReservoir(k)

	// Обрабатываем все элементы потока через резервуарную выборку
	for _, value := range data {
		res.Add(value) // добавляем каждый элемент в резервуар
	}

	// Получаем итоговую выборку из резервуара
	resSample := res.Sample()
	timeR := time.Since(startR)

	// наивн

	startNaive := time.Now()

	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)

	naiveSample := NaiveSample(data, k)

	timeNaive := time.Since(startNaive)

	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)
	naiveMemory := m2.Alloc - m1.Alloc

	// 1. Проверка вероятности попадания каждого элемента
	runs := 1000 // Количество запусков для статистики
	selectedCount := make([]int, n)

	startFreq := time.Now()
	for run := 0; run < runs; run++ {
		// Создаем новый reservoir для каждого запуска
		res := NReservoir(k)

		// Обрабатываем поток (важно: тот же самый поток!)
		for _, value := range data {
			res.Add(value)
		}

		// Отмечаем выбранные элементы
		for _, val := range res.Sample() {
			selectedCount[val]++
		}
	}
	timeFreq := time.Since(startFreq)

	// 2. Рассчитываем отклонение от ожидаемой вероятности
	expectedProb := float64(k) / float64(n)
	absErrorSum := 0.0
	relErrorSum := 0.0
	validElements := 0

	for i := 0; i < n; i++ {
		if i < n { // Все элементы потока
			actualProb := float64(selectedCount[i]) / float64(runs)
			expected := expectedProb

			absError := math.Abs(actualProb - expected)
			relError := absError / expected

			absErrorSum += absError
			relErrorSum += relError
			validElements++
		}
	}

	avgAbsErrorProb := absErrorSum / float64(validElements)
	avgRelErrorProb := (relErrorSum / float64(validElements)) * 100

	fmt.Printf("\n=== Проверка вероятности попадания ===\n")
	fmt.Printf("Ожидаемая вероятность для каждого элемента: %.6f\n", expectedProb)
	fmt.Printf("Средняя абсолютная ошибка вероятности: %.6f\n", avgAbsErrorProb)
	fmt.Printf("Средняя относительная ошибка вероятности: %.2f%%\n", avgRelErrorProb)
	fmt.Printf("Время на %d запусков для статистики: %v\n", runs, timeFreq)

	// 3. Проверка распределения выбранных элементов (хи-квадрат)
	chi2 := 0.0
	totalSelected := 0
	for _, count := range selectedCount {
		totalSelected += count
	}
	expectedCount := float64(totalSelected) / float64(n)

	for i := 0; i < n; i++ {
		observed := float64(selectedCount[i])
		chi2 += math.Pow(observed-expectedCount, 2) / expectedCount
	}

	fmt.Printf("\n=== Хи-квадрат тест на равномерность ===\n")
	fmt.Printf("Хи-квадрат: %.2f\n", chi2)
	fmt.Printf("Степени свободы: %d\n", n-1)
	// Для n > 1000, хи-квадрат ≈ n ± √(2n)
	expectedChi2 := float64(n)
	stdChi2 := math.Sqrt(2 * float64(n))
	fmt.Printf("Ожидаемый хи-квадрат: %.0f ± %.0f\n", expectedChi2, stdChi2)

	// 4. Быстрая проверка первых 100 элементов
	fmt.Printf("\n=== Проверка первых 100 элементов ===\n")
	for i := 0; i < 100 && i < n; i++ {
		actualProb := float64(selectedCount[i]) / float64(runs)
		fmt.Printf("Элемент %d: выбрано %d раз (вероятность %.4f, ожидается %.4f)\n",
			i, selectedCount[i], actualProb, expectedProb)
	}

	fmt.Printf("Память Reservoir sampling : %d байт\n", k*8)
	fmt.Printf("Память Naive: %d байт\n", naiveMemory)

	fmt.Printf("Наивный алгоритм:  %v\n", timeNaive)
	fmt.Printf("Reservoir sampling:  %v\n", timeR)

	// Выводим первые 10 элементов из резервуарной выборки
	fmt.Printf("Reservoir sampling first 10: %v\n", resSample[:10])

	// Выводим первые 10 элементов из наивной выборки
	fmt.Printf("Naive first 10: %v\n", naiveSample[:10])
}

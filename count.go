package main

// fmt — для форматированного ввода-вывода
// hash/fnv — для хэш-функции FNV-1a
// math — для математических констант и функций
import (
	"fmt"
	"hash/fnv"
	"math"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type Sketch struct {
	width  int      // ширина таблицы (количество столбцов)
	depth  int      // глубина таблицы (количество строк, хэш-функций)
	table  [][]int  // двумерный массив счётчиков [depth][width]
	hashes []uint64 // массив для каждой хэш-функции
}

func CountMinSketch(width, depth int) *Sketch {
	return &Sketch{
		width:  width,                 // инициализация ширины
		depth:  depth,                 // инициализация глубины
		table:  make([][]int, depth),  // выделение памяти для строк таблицы
		hashes: make([]uint64, depth), // выделение памяти для массива соль
	}
}

// Инициализация таблицы и хэшей после создания структуры
func (cms *Sketch) Init() {
	//по строкам
	for i := range cms.table {
		//Все элементы инициализируются нулями
		cms.table[i] = make([]int, cms.width) // создание строки таблицы
		//вычисляем значение соли (seed)
		//преобразуем в  64-битное целое число
		cms.hashes[i] = uint64(i + 1)
	}
}

// генерирует хэш от строки и соли, возвращает индекс в таблице
func (cms *Sketch) HashIndex(s string, seed uint64) int {
	h1 := fnv.New64a()
	h2 := fnv.New64a()
	h1.Write([]byte(s))
	h2.Write([]byte(s))
	v1 := h1.Sum64()
	v2 := h2.Sum64()

	// Генерируем разные индексы для каждого хеша, используя seed
	mixed := v1 + seed*v2
	return int(mixed % uint64(cms.width))
}

// Add увеличивает счётчик для элемента на заданное количество
func (cms *Sketch) Add(s string, count int) {
	// cms.depth - количество хеш-функций
	for i := 0; i < cms.depth; i++ {
		idx := cms.HashIndex(s, cms.hashes[i]) // вычисляем индекс
		cms.table[i][idx] += count             // увеличиваем счётчик в ячейке
	}
}

// Count возвращает оценочное минимальное количество вхождений элемента
func (cms *Sketch) Count(s string) int {
	min := math.MaxInt32 // 2 миллиарда
	for i := 0; i < cms.depth; i++ {
		idx := cms.HashIndex(s, cms.hashes[i]) // вычисляем индекс
		if cms.table[i][idx] < min {           // если значение в ячейке меньше текущего минимума
			min = cms.table[i][idx] // обновляем минимум
		}
	}
	return min // возвращаем минимальное значение (оценка частоты)
}

func main() {

	fmt.Printf("Count-min-sketch\n")

	var n int
	for {
		fmt.Print("\nВведите количество элементов(до 10 млн): ")
		var input string
		if _, err := fmt.Scanln(&input); err != nil {
			fmt.Println("Некорректный ввод")
			continue
		}
		// Проверяем, что строка состоит только из цифр
		if _, err := strconv.Atoi(input); err != nil || strings.TrimSpace(input) == "" {
			fmt.Println("Некорректный ввод")
			continue
		}
		n, _ = strconv.Atoi(input)
		if n <= 0 || n > 10_000_000 {
			fmt.Println("Некорректный ввод")
			continue
		}
		break
	}
	width := n  // ширина таблицы (столбцы)
	depth := 15 // глубина таблицы (хэш-функции)

	// Создание и инициализация Count-Min Sketch
	cms := CountMinSketch(width, depth)
	cms.Init()

	seen := make(map[string]int) // наивный

	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)

	startNaive := time.Now()

	// Создаём массив всех ключей
	keys := make([]string, n)
	for i := 0; i < n; i++ {
		keys[i] = fmt.Sprintf("element-%d", i%5_000_000)
	}

	copyKeys := make([]string, len(keys))
	copy(copyKeys, keys)

	seen = make(map[string]int, 5_000_000)
	for _, key := range copyKeys {
		seen[key]++
	}

	timeNaive := time.Since(startNaive)
	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)
	naiveMemory := m2.Alloc - m1.Alloc

	cmsMemory := width * depth * 8

	startCMS := time.Now()
	for i := 0; i < n; i++ {
		key := fmt.Sprintf("element-%d", i%2_000_000)
		cms.Add(key, 1) // добавляем элемент в CMS, на 1 увеличиваем счет
	}
	timeCMS := time.Since(startCMS)

	var countMismatch int // Счётчик количества элементов, где оценка CMS не совпала с точным значением
	var totalCount int    // Общее количество уникальных элементов

	// Перебираем все уникальные ключи и их точные значения из наивного счётчика
	for key, exactCount := range seen {
		cmsCount := cms.Count(key)  // Получаем оценку частоты элемента из Count-Min Sketch
		if cmsCount != exactCount { // Если оценка не совпадает с точным значением
			countMismatch++ // Увеличиваем счётчик ошибок
		}
		totalCount++ // Увеличиваем общий счётчик элементов
	}

	absError := float64(countMismatch)                             // Абсолютная ошибка — количество элементов с ошибкой
	relError := float64(countMismatch) / float64(totalCount) * 100 // Относительная ошибка — процент элементов с ошибкой

	fmt.Printf("\nСредняя абсолютная ошибка: %.f\n", absError)
	fmt.Printf("Средняя относительная ошибка: %.2f%%\n", relError)

	fmt.Printf("Наивный алгоритм:  %v\n", timeNaive)
	fmt.Printf("Count-Min Sketch:  %v\n", timeCMS)
	fmt.Printf("CMS память: %d байт\n", cmsMemory)
	fmt.Printf("Память наивного алгоритма: %d байт\n", naiveMemory)

}

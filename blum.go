package main

import (
	"fmt"
	"hash/fnv"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type BloomFilter struct {
	bitSet []uint64 // массив 64-битных целых чисел для хранения битов

	size      int // Общий размер битового массива в битах (количество доступных битов)
	hashCount int // Количество хэш-функций, используемых для каждого элемента (параметр k)
}

func NewBloomFilter(size, hashCount int) *BloomFilter {

	bitSetSize := size / 64
	if size%64 != 0 {
		bitSetSize++
	}

	return &BloomFilter{
		bitSet:    make([]uint64, bitSetSize), //  массив нужного размера
		size:      size,                       //  общий размер в битах
		hashCount: hashCount,
	}
}

func (bf *BloomFilter) Add(item []byte) {

	h1 := fnv.New64a()
	h2 := fnv.New64a()
	// Передаём данные элемента в хэш-функцию
	h1.Write(item)
	h2.Write(item)
	// возвращает готовое число
	v1 := h1.Sum64()
	v2 := h2.Sum64()

	// hashCount различных индексов
	for i := 0; i < bf.hashCount; i++ {

		// Комбинируем базовый хэш с номером итерации для получения различных значений
		// mixed = v1 + i*v1 = v1*(i+1) - создаёт линейную последовательность
		mixed := v1 + uint64(i)*v2

		// Вычисляем индекс в диапазоне [0, bf.size-1] с помощью операции остатка от деления
		// % bf.size гарантирует, что индекс не выйдет за пределы битового массива
		idx := int(mixed % uint64(bf.size))

		// Устанавливаем соответствующий бит в массиве:
		// 1. idx/64 - определяем, в каком элементе массива uint64 находится нужный бит
		// 2. idx%64 - определяем позицию бита
		// 3. 1 << (idx % 64) - создаём битовую маску с единицей в нужной позиции
		// 4. |= (побитовое ИЛИ с присваиванием) - устанавливает бит в 1
		bf.bitSet[idx/64] |= 1 << (idx % 64)
	}

}

func (bf *BloomFilter) Nalich(item []byte) bool {

	h1 := fnv.New64a()
	h2 := fnv.New64a()
	// Передаём данные элемента в хэш-функцию
	h1.Write(item)
	h2.Write(item)
	// возвращает готовое число
	v1 := h1.Sum64()
	v2 := h2.Sum64()

	// Проверяем все hashCount битов, которые должны быть установлены для этого элемента
	for i := 0; i < bf.hashCount; i++ {
		// Вычисляем индекс по точно такой же формуле, как в методе Add
		mixed := v1 + uint64(i)*v2
		idx := int(mixed % uint64(bf.size))

		// Проверяем, установлен ли бит по вычисленному индексу:
		// 1. bf.bitSet[idx/64] - получаем нужный элемент массива uint64
		// 2. &(1 << (idx % 64)) - применяем битовую маску для проверки конкретного бита
		// 3. == 0 - если результат равен 0, значит бит НЕ установлен
		if bf.bitSet[idx/64]&(1<<(idx%64)) == 0 {
			// Нашли хотя бы один не установленный бит
			return false
		}
	}
	return true
}

func main() {

	fmt.Printf("Фильтр Блума\n")
	var n int
	for {
		fmt.Print("\nВведите количество элементов(до 20 млн): ")
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
		if n <= 0 || n > 20_000_000 {
			fmt.Println("Некорректный ввод")
			continue
		}
		break
	}
	// битов, которые должны быть установлены для этого элемента
	hashCount := 5

	keys := make([]string, n)

	if n < 5_000_000 {
		for i := 0; i < n; i++ {
			keys[i] = fmt.Sprintf("element-%d", i%300_000)
		}
	} else {
		for i := 0; i < n; i++ {
			keys[i] = fmt.Sprintf("element-%d", i%5_000_000)
		}
	}

	// фильтр Блума

	// размер битового массива будет 10n бит, то есть в 10 раз больше, чем количество элементов
	filter := NewBloomFilter(n*10, hashCount)

	startB := time.Now()

	// Добавляем все элементы в фильтр Блума
	for _, key := range keys {
		// Преобразуем строку в []byte и добавляем в фильтр
		filter.Add([]byte(key))
	}

	timeB := time.Since(startB)

	// наивный
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)

	startN := time.Now()
	naive := make(map[string]bool)

	for _, key := range keys {

		naive[key] = true
	}

	timeN := time.Since(startN)

	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)

	naiveMemory := m2.Alloc - m1.Alloc

	// Тест: считаем ложноположительные для ключей, которых точно нет
	falsePositives := 0
	trueNegatives := 0
	tests := 0

	for i := 0; i < n; i++ {
		key := fmt.Sprintf("missing-%d", i)
		inBloom := filter.Nalich([]byte(key))
		inNaive := naive[key]

		if inBloom && !inNaive {
			falsePositives++
		} else if !inBloom && !inNaive {
			trueNegatives++
		}
		tests++
	}
	absError := float64(falsePositives)
	relError := float64(falsePositives) / float64(tests) * 100

	fmt.Printf("\nСредняя абсолютная ошибка: %.f\n", absError)
	fmt.Printf("Средняя относительная ошибка: %.2f%%\n", relError)

	fmt.Printf("Время фильтра Блума: %v\n", timeB)
	fmt.Printf("Время naive: %v\n", timeN)

	// Выводим использование памяти
	fmt.Printf("Память naive: %d байт\n", naiveMemory)

	fmt.Printf("Память фильтра Блума: %d байт\n", len(filter.bitSet)*8)

	fmt.Println("element-1:", filter.Nalich([]byte("element-1")))

	fmt.Println("element-455000:", filter.Nalich([]byte("element-455000")))
}

package main

import (
	"fmt"
	"hash/fnv"
	"math"
	"math/rand"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	m = 4096 //  количество ячеек памяти, определяет точность алгоритма
)

type HyperLogLog struct {
	registers [m]byte
}

func NewHyperLogLog() *HyperLogLog {
	return &HyperLogLog{}
}

// hash возвращает два 32-битных числа из строки
func hash(s string) (uint32, uint32) {
	h := fnv.New64a()  // создаёт 64-битный хэш FNV-1a
	h.Write([]byte(s)) // преобразует строку в байты
	v := h.Sum64()     // получает 64-битный хэш-код

	// Разделяем 64-битный хэш на два 32-битных числа
	w := uint32(v >> 32)
	z := uint32(v)

	// Перемешивание
	w ^= z<<13 | z>>(32-13)
	z ^= w<<7 | w>>(32-7)

	return w, z
}

// Считает количество нулей слева в двоичной записи числа

func countLeadingZeros(x uint32) byte {
	for i := 0; i < 32; i++ {
		//x >> (31-i) - сдвиг вправо
		if (x>>(31-i))&1 == 1 {
			return byte(i)
		}
	}
	return 32
}

// Хэш

func (hll *HyperLogLog) Add(s string) {
	h1, h2 := hash(s)

	// определяем индекс регистра
	idx := h1 % m

	// вычисляем значение для обновленного регистра
	// h2 считаем, сколько ведущих нулей, прибавляем 1
	rho := countLeadingZeros(h2) + 1

	//Если новое значение больше, чем то, что уже хранится в регистре
	if rho > hll.registers[idx] {
		hll.registers[idx] = rho
	}
}

// делает оценку количества уникальных элементов
func (hll *HyperLogLog) Estimate() float64 {
	//среднее значений регистров
	sum := 0.0
	// Проходим по всем регистрам
	for _, val := range hll.registers {

		//1 / 2^val вероятность увидеть данный хэш
		//Каждый бит может быть с вероятностью 1/2
		//сумма вероятностей по всем регистрам

		sum += 1 / math.Pow(2, float64(val))
	}
	//Используем гармоническое среднее
	//если среднее ариф то одна большая оценка испортит всё
	estimate := alpha(m) * m * m / sum

	// коррекция для малых значений, чем больше нулевых регистров — тем
	// меньше реальность
	if estimate <= 5*m/2 {
		zeros := 0
		//игнорируем индекс регистра
		for _, val := range hll.registers {
			if val == 0 { // не видел ни одного элемента
				zeros++
			}
		}
		if zeros != 0 {
			estimate = float64(m) * math.Log(float64(m)/float64(zeros))
		}
	}

	return estimate
}

// alpha — поправочный коэффициент, компенсирующий систематическую ошибку
func alpha(m int) float64 {
	switch m {
	case 16:
		return 0.673
	case 32:
		return 0.697
	case 64:
		return 0.709
	default:
		return 0.7213 / (1 + 1.079/float64(m))
	}
}

func main() {

	hll := NewHyperLogLog()

	fmt.Printf("HyperLogLog\n")
	var n int

	for {
		fmt.Print("\nВведите количество элементов: ")
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
		if n <= 0 {
			fmt.Println("Некорректный ввод")
			continue
		}
		break
	}

	startHP := time.Now()
	for i := 0; i < n; i++ {
		key := strconv.Itoa(rand.Intn(n))
		//struct{} занимает 0 байтов в памяти, bool занимает 1 байт
		//strconv.Itoa(...) — преобразует число в строку
		hll.Add(key)
	}
	timeHP := time.Since(startHP)

	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)

	startNaive := time.Now()
	seen := make(map[string]struct{})
	for i := 0; i < n; i++ {
		key := strconv.Itoa(rand.Intn(n))
		//struct{} занимает 0 байтов в памяти, bool занимает 1 байт
		//strconv.Itoa(...) — преобразует число в строку
		seen[key] = struct{}{}
	}
	timeNaive := time.Since(startNaive)

	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)
	naiveMemory := m2.Alloc - m1.Alloc

	// Реальное количество уникальных элементов
	exact := float64(len(seen))
	// Оценка HyperLogLog
	estimated := hll.Estimate()

	// Абсолютная ошибка
	absError := math.Abs(estimated - exact)
	// Относительная ошибка
	relError := absError / exact * 100

	fmt.Printf("\nСредняя абсолютная ошибка: %.2f\n", absError)
	fmt.Printf("Средняя относительная ошибка: %.2f%%\n", relError)

	fmt.Printf("Память HyperLogLog: %d байт\n", m)
	fmt.Printf("Память Naive: %d байт\n", naiveMemory)

	fmt.Printf("Наивный алгоритм:  %v\n", timeNaive)
	fmt.Printf("Hyper:  %v\n", timeHP)

	fmt.Printf("Реальное количество уникальных элементов: %d\n", len(seen))
	fmt.Printf("Оценка уникальных элементов HyperLogLog: %.2f\n", hll.Estimate())
}

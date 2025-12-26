package main

import (
	"fmt"
	"hash/fnv"
	"math"
)

const (
	m = 8 // Количество регистров (маленькое для наглядности)
)

// HyperLogLog структура для вероятностного подсчёта уникальных элементов
type HyperLogLog struct {
	registers [m]byte // Массив регистров, каждый хранит максимальное значение ρ
}

// NewHyperLogLog создаёт новый экземпляр HyperLogLog
func NewHyperLogLog() *HyperLogLog {
	return &HyperLogLog{}
}

// hash генерирует два 32-битных хеша из строки
func hash(s string) (uint32, uint32) {
	h := fnv.New64a()  // Создаём 64-битный хеш FNV-1a
	h.Write([]byte(s)) // Преобразуем строку в байты и передаём в хеш
	v := h.Sum64()

	w := uint32(v >> 32)
	z := uint32(v)

	// Перемешивание для увеличения случайности
	w ^= z<<13 | z>>(32-13)
	z ^= w<<7 | w>>(32-7)

	return w, z
}

// countLeadingZeros считает количество ведущих нулей в 32-битном числе
func countLeadingZeros(x uint32) byte {
	for i := 0; i < 32; i++ { // Проходим по всем 32 битам
		if (x>>(31-i))&1 == 1 { // Проверяем бит с позиции 31-i
			return byte(i) // Возвращаем количество нулей до первой 1
		}
	}
	return 32 // Если все биты нулевые, возвращаем 32
}

// добавляет элемент
func (hll *HyperLogLog) Add(s string) {
	h1, h2 := hash(s)
	idx := h1 % m
	rho := countLeadingZeros(h2) + 1

	fmt.Printf("Добавляем '%s':\n", s)
	fmt.Printf("  hash64 = 0x%016x\n", h1<<32|h2)                 // Полный 64-битный хеш
	fmt.Printf("  h1 = 0x%08x (%d) %% %d = %d\n", h1, h1, m, idx) // Первый хеш и индекс
	fmt.Printf("  h2 = 0x%08x (двоичный: %032b)\n", h2, h2)       // Второй хеш в двоичном виде
	fmt.Printf("  ведущих нулей = %d -> rho = %d\n", countLeadingZeros(h2), rho)

	old := hll.registers[idx] // Текущее значение регистра
	if rho > old {            // Если новое значение больше старого
		fmt.Printf("  Регистр[%d]: %d -> %d\n", idx, old, rho)
		hll.registers[idx] = rho // Обновляем регистр
	} else { // Если значение не больше
		fmt.Printf("  Регистр[%d]: %d >= %d (не обновляем)\n", idx, old, rho)
	}
}

// printRegisters выводит текущее состояние всех регистров
func (hll *HyperLogLog) printRegisters() {
	fmt.Printf("  Регистры: [")
	for i, v := range hll.registers { // Проходим по всем регистрам
		if i > 0 { // Если не первый элемент
			fmt.Print(", ") // Добавляем разделитель
		}
		fmt.Printf("%d", v) // Выводим значение регистра
	}
	fmt.Printf("]\n")
}

// Estimate оценивает количество уникальных элементов
func (hll *HyperLogLog) Estimate() float64 {
	fmt.Println("Оценка:")

	sum := 0.0
	for _, val := range hll.registers { // Для каждого регистра
		prob := 1 / math.Pow(2, float64(val)) // Вычисляем 1/2^значение
		fmt.Printf("  1/2^%d = %.3f\n", val, prob)
		sum += prob // Суммируем все вероятности
	}

	alpha := alpha(m) // Получаем поправочный коэффициент
	fmt.Printf("  alpha(%d) = %.4f\n", m, alpha)
	fmt.Printf("  сумма = %.3f\n", sum)

	// Основная формула оценки
	estimate := alpha * float64(m) * float64(m) / sum
	fmt.Printf("  %.4f * %d * %d / %.3f = %.2f\n", alpha, m, m, sum, estimate)

	// Считаем количество нулевых регистров
	zeros := 0
	for _, val := range hll.registers {
		if val == 0 {
			zeros++
		}
	}
	fmt.Printf("  Нулевых регистров = %d\n", zeros)

	// Применяем коррекцию для малых значений (Linear Counting)
	if estimate <= 5*float64(m)/2 && zeros != 0 {
		estimate = float64(m) * math.Log(float64(m)/float64(zeros))
		fmt.Printf("  Linear Counting: %d * ln(%d/%d) = %.2f\n", m, m, zeros, estimate)
	}

	return estimate
}

// alpha возвращает поправочный коэффициент для заданного m
func alpha(m int) float64 {
	switch m { // Оптимальные значения для стандартных m
	case 16:
		return 0.673
	case 32:
		return 0.697
	case 64:
		return 0.709
	default: // Общая формула для других значений
		return 0.7213 / (1 + 1.079/float64(m))
	}
}

func main() {
	fmt.Println("Иллюстрация работы HyperLogLog (m =", m, ")")
	fmt.Println()

	hll := NewHyperLogLog()
	fmt.Println("Начальное состояние:")
	hll.printRegisters() // Выводим начальные регистры

	elements := []string{"apple", "banana", "da", "date", "dat",
		"t", "cat", "dog"}

	// Добавляем все элементы по очереди
	for _, elem := range elements {
		hll.Add(elem)        // Добавляем элемент
		hll.printRegisters() // Показываем состояние после каждого добавления
	}

	exact := float64(len(elements)) // Точное количество элементов
	estimated := hll.Estimate()     // Оценка HyperLogLog

	fmt.Printf("\nРезультаты:\n")
	fmt.Printf("Всего уникальных элементов: %.0f\n", exact)
	fmt.Printf("Оценка HyperLogLog: %.2f\n", estimated)

}

package cmd

import (
	"flag"
	"fmt"
	"log"
)

// parse запускает процесс парсинга в зависимости от параметров.
func parse(parseType, maxRecipes, concurrency int) {
	switch parseType {
	case 1:
		fmt.Println("Выполняется полный парсинг...")
		// Здесь вызывается функция для парсинга категорий и рецептов
	case 2:
		fmt.Println("Выполняется парсинг только категорий...")
		// Здесь вызывается функция для парсинга только категорий
	default:
		fmt.Println("Неверный тип парсинга. Используйте 1 для полного парсинга или 2 для парсинга только категорий.")
	}
}

// RunParser запускает парсер с переданными аргументами.
func RunParser(args []string) error {
	var (
		parseType   int // Тип парсинга (1 - полный, 2 - только категории)
		maxRecipes  int // Количество рецептов для парсинга в каждой категории
		concurrency int // Количество одновременных потоков (горутин)
	)

	fs := flag.NewFlagSet("run", flag.ExitOnError)
	fs.IntVar(&parseType, "type", 1, "Тип парсинга: 1 - полный, 2 - только категории")
	fs.IntVar(&maxRecipes, "recipes", 10, "Количество рецептов для каждой категории")
	fs.IntVar(&concurrency, "concurrency", 5, "Количество одновременных потоков")

	if err := fs.Parse(args); err != nil {
		return err
	}

	log.Println("Запуск парсера с параметрами:")
	log.Printf("Тип парсинга: %d\n", parseType)
	log.Printf("Количество рецептов: %d\n", maxRecipes)
	log.Printf("Количество одновременных потоков: %d\n", concurrency)

	parse(parseType, maxRecipes, concurrency)

	return nil
}

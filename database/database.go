package database

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/seniorcat/scraper/entity"
)

// DBServiceInterface определяет методы для работы с базой данных
type DBServiceInterface interface {
	SaveCategories(ctx context.Context, categories []entity.Category) error
	SaveRecipes(ctx context.Context, recipes []entity.Recipe) error
}

// DBService предоставляет доступ к методам работы с базой данных
type DBService struct {
	Pool             *pgxpool.Pool
	CategorySaveChan chan []entity.Category // Канал для сохранения категорий
	RecipeSaveChan   chan []entity.Recipe   // Канал для сохранения рецептов
}

// NewDBService инициализирует соединение с базой данных PostgreSQL и запускает воркеры
func NewDBService(connString string) (*DBService, error) {
	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		return nil, err
	}

	dbService := &DBService{
		Pool:             pool,
		CategorySaveChan: make(chan []entity.Category, 10), // Инициализация каналов
		RecipeSaveChan:   make(chan []entity.Recipe, 10),
	}

	// Запуск горутин для асинхронного сохранения данных
	go dbService.saveCategoriesWorker()
	go dbService.saveRecipesWorker()

	return dbService, nil
}

// saveCategoriesWorker - воркер для асинхронного сохранения категорий
func (db *DBService) saveCategoriesWorker() {
	for categories := range db.CategorySaveChan {
		ctx := context.Background()
		if err := db.SaveCategories(ctx, categories); err != nil {
			log.Printf("Failed to save categories: %v", err)
		}
	}
}

// saveRecipesWorker - воркер для асинхронного сохранения рецептов
func (db *DBService) saveRecipesWorker() {
	for recipes := range db.RecipeSaveChan {
		ctx := context.Background()
		if err := db.SaveRecipes(ctx, recipes); err != nil {
			log.Printf("Failed to save recipes: %v", err)
		}
	}
}

// SaveCategories сохраняет список категорий в базу данных
func (db *DBService) SaveCategories(ctx context.Context, categories []entity.Category) error {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, category := range categories {
		_, err = tx.Exec(ctx, "INSERT INTO categories (name, href) VALUES ($1, $2)", category.Name, category.Href)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// SaveRecipes сохраняет список рецептов в базу данных
func (db *DBService) SaveRecipes(ctx context.Context, recipes []entity.Recipe) error {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, recipe := range recipes {
		_, err = tx.Exec(ctx, "INSERT INTO recipes (name, href) VALUES ($1, $2)", recipe.Name, recipe.Href)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// CreateTables создаёт таблицы в базе данных
func (db *DBService) CreateTables(ctx context.Context) error {
	_, err := db.Pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS categories (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			href TEXT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS recipes (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			href TEXT NOT NULL
		);
	`)
	return err
}

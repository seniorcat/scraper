package entity

import (
	"fmt"
	"strings"
)

// Recipe хранит информацию о рецепте
type Recipe struct {
	Name string
	Href string
}

// Validate проверяет данные рецепта на корректность
func (r *Recipe) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("recipe name is empty")
	}
	if r.Href == "" {
		return fmt.Errorf("recipe href is empty")
	}
	return nil
}

// Normalize нормализует данные рецепта (например, удаляет лишние пробелы и приводит к нижнему регистру)
func (r *Recipe) Normalize() {
	r.Name = normalizeRecipeName(r.Name)
	r.Href = normalizeHref(r.Href)
}

// normalizeRecipeName нормализует название рецепта
func normalizeRecipeName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

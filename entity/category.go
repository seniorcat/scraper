package entity

import (
	"fmt"
	"strings"
)

// Category хранит информацию о категории
type Category struct {
	Name string
	Href string
}

// Validate проверяет данные категории на корректность
func (c *Category) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("category name is empty")
	}
	if c.Href == "" {
		return fmt.Errorf("category href is empty")
	}
	return nil
}

// Normalize нормализует данные категории (например, удаляет лишние пробелы и приводит к нижнему регистру)
func (c *Category) Normalize() {
	c.Name = normalizeCategoryName(c.Name)
	c.Href = normalizeHref(c.Href)
}

// normalizeCategoryName нормализует название категории
func normalizeCategoryName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

// normalizeHref нормализует ссылку
func normalizeHref(href string) string {
	return strings.TrimSpace(href)
}

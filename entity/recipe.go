package entity

import "fmt"

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

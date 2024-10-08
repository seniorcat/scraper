package entity

import "fmt"

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

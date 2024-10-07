package cmd

import (
	"fmt"
	"os"
)

type CLI struct {
	Commands map[string]Command
}

// Command - структура, описывающая команду CLI
type Command struct {
	Name        string
	Description string
	Action      func(args []string)
}

// NewCLI создает новый экземпляр CLI
func NewCLI() *CLI {
	return &CLI{
		Commands: make(map[string]Command),
	}
}

// RegisterCommand регистрирует новую команду
func (cli *CLI) RegisterCommand(name string, description string, action func(args []string)) {
	cli.Commands[name] = Command{
		Name:        name,
		Description: description,
		Action:      action,
	}
}

// Run запускает CLI и вызывает нужную команду на основе аргументов
func (cli *CLI) Run() {
	if len(os.Args) < 2 {
		cli.PrintHelp()
		return
	}

	commandName := os.Args[1]
	command, exists := cli.Commands[commandName]
	if !exists {
		fmt.Printf("Неизвестная команда: %s\n", commandName)
		cli.PrintHelp()
		return
	}

	command.Action(os.Args[2:])
}

// printHelp выводит справку по всем доступным командам
func (cli *CLI) PrintHelp() {
	fmt.Println("Доступные команды:")
	for _, cmd := range cli.Commands {
		fmt.Printf(" %s: %s\n", cmd.Name, cmd.Description)
	}
}

package main

import (
    "bufio"
    "fmt"
    "os"
    "strings"
)

func cleanInput(text string) []string {
    text = strings.ToLower(text)
    text = strings.TrimSpace(text)
    return strings.Fields(text)
}
func startRepl(cfg *config) {
    scanner := bufio.NewScanner(os.Stdin)
    for {
		fmt.Print("Pokedex > ")

		if !scanner.Scan() {
			return
		}

		input := scanner.Text()
		words := cleanInput(input)

		if len(words) == 0 {
			continue
		}

		commandName := words[0]
		cmd, ok := cfg.commands[commandName]

		if !ok {
			fmt.Println("Unknown command")
			continue
		}

		err := cmd.callback(cfg, words[1:]...)
		if err != nil {
			fmt.Println("Error:", err)
		}
	}
}
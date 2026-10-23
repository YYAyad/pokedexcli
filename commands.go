package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"

	"example.com/pokedexcli/internal/pokecache"
)

type config struct {
	commands            map[string]cliCommand
	nextLocationURL     *string
	previousLocationURL *string
	cache               *pokecache.Cache
	pokedex             map[string]Pokemon
}

type cliCommand struct {
	name        string
	description string
	callback    func(*config, ...string) error
}

type locationAreaResponse struct {
	Count    int     `json:"count"`
	Next     *string `json:"next"`
	Previous *string `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
	} `json:"results"`
}
type Pokemon struct {
	Name           string `json:"name"`
	BaseExperience int    `json:"base_experience"`
	Height         int    `json:"height"`
	Weight         int    `json:"weight"`
	Stats          []struct {
		BaseStat int `json:"base_stat"`
		Stat     struct {
			Name string `json:"name"`
		} `json:"stat"`
	} `json:"stats"`
	Types []struct {
		Type struct {
			Name string `json:"name"`
		} `json:"type"`
	} `json:"types"`
}
type pokemonResponse struct {
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
		} `json:"pokemon"`
	} `json:"pokemon_encounters"`
}

func commandExit(cfg *config, args ...string) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(cfg *config, args ...string) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	fmt.Println()

	for _, cmd := range cfg.commands {
		fmt.Printf("%s: %s\n", cmd.name, cmd.description)
	}

	fmt.Println()
	return nil
}

func getCommands() map[string]cliCommand {
	return map[string]cliCommand{
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback:    commandHelp,
		},
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
		"map": {
			name:        "map",
			description: "Display location areas",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "Display previous location areas",
			callback:    commandMapBack,
		},
		"explore": {
			name:        "explore",
			description: "Explore a location area",
			callback:    commandExplore,
		},
		"catch": {
			name:        "catch",
			description: "Catch a Pokemon",
			callback:    commandCatch,
		},
		"inspect": {
			name:        "inspect",
			description: "Inspect a caught Pokemon",
			callback:    commandInspect,
		},
		"pokedex": {
			name:        "pokedex",
			description: "List caught Pokemon",
			callback:    commandPokedex,
		},
	}
}

func commandMap(cfg *config, args ...string) error {
	url := "https://pokeapi.co/api/v2/location-area"

	if cfg.nextLocationURL != nil {
		url = *cfg.nextLocationURL
	}

	body, err := getLocationAreaResponse(cfg, url)
	if err != nil {
		return err
	}

	var data locationAreaResponse
	err = json.Unmarshal(body, &data)
	if err != nil {
		return err
	}

	cfg.nextLocationURL = data.Next
	cfg.previousLocationURL = data.Previous

	for _, area := range data.Results {
		fmt.Println(area.Name)
	}

	return nil
}

func commandMapBack(cfg *config, args ...string) error {
	if cfg.previousLocationURL == nil {
		fmt.Println("you're on the first page")
		return nil
	}

	url := *cfg.previousLocationURL

	body, err := getLocationAreaResponse(cfg, url)
	if err != nil {
		return err
	}

	var data locationAreaResponse
	err = json.Unmarshal(body, &data)
	if err != nil {
		return err
	}

	cfg.nextLocationURL = data.Next
	cfg.previousLocationURL = data.Previous

	for _, area := range data.Results {
		fmt.Println(area.Name)
	}

	return nil
}

func getLocationAreaResponse(cfg *config, url string) ([]byte, error) {
	if body, ok := cfg.cache.Get(url); ok {
		return body, nil
	}

	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	cfg.cache.Add(url, body)
	return body, nil
}

func commandExplore(cfg *config, args ...string) error {
	if len(args) == 0 {
		return fmt.Errorf("please provide a location area name")
	}

	areaName := args[0]
	fmt.Printf("Exploring %s...\n", areaName)

	url := "https://pokeapi.co/api/v2/location-area/" + areaName + "/"

	body, err := getLocationAreaResponse(cfg, url)
	if err != nil {
		return err
	}

	var data pokemonResponse
	err = json.Unmarshal(body, &data)
	if err != nil {
		return err
	}

	fmt.Println("Found Pokemon:")
	for _, encounter := range data.PokemonEncounters {
		fmt.Printf(" - %s\n", encounter.Pokemon.Name)
	}

	return nil
}
func commandCatch(cfg *config, args ...string) error {
	if len(args) == 0 {
		return fmt.Errorf("please provide a Pokemon name")
	}

	pokemonName := args[0]
	fmt.Printf("Throwing a Pokeball at %s...\n", pokemonName)
	url := "https://pokeapi.co/api/v2/pokemon/" + pokemonName + "/"

	body, err := getLocationAreaResponse(cfg, url)
	if err != nil {
		return err
	}

	var pokemon Pokemon
	err = json.Unmarshal(body, &pokemon)
	if err != nil {
		return err
	}
	catchChance := 100 - pokemon.BaseExperience
	if catchChance < 10 {
		catchChance = 10
	}

	if rand.Intn(100) < catchChance {
		fmt.Printf("%s was caught!\n", pokemon.Name)
		cfg.pokedex[pokemon.Name] = pokemon
	} else {
		fmt.Printf("%s escaped!\n", pokemon.Name)
	}

	return nil
}

func commandInspect(cfg *config, args ...string) error {
	if len(args) == 0 {
		return fmt.Errorf("please provide a Pokemon name")
	}

	pokemonName := args[0]
	pokemon, ok := cfg.pokedex[pokemonName]

	if !ok {
		fmt.Println("you have not caught that pokemon")
		return nil
	}
	fmt.Printf("Name: %s\n", pokemon.Name)
	fmt.Printf("Height: %d\n", pokemon.Height)
	fmt.Printf("Weight: %d\n", pokemon.Weight)
	fmt.Println("Stats:")
	for _, stat := range pokemon.Stats {
		fmt.Printf("  -%s: %d\n", stat.Stat.Name, stat.BaseStat)
	}
	fmt.Println("Types:")
	for _, pokemonType := range pokemon.Types {
		fmt.Printf("  - %s\n", pokemonType.Type.Name)
	}
	return nil
}

func commandPokedex(cfg *config, args ...string) error {
	fmt.Println("Your Pokedex:")

	for _, pokemon := range cfg.pokedex {
		fmt.Printf(" - %s\n", pokemon.Name)
	}

	return nil
}

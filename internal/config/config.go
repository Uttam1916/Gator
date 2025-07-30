package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// the json filename
const gatorconfigjson string = ".gatorconfig.json"

type Config struct {
	Db_url           string `json:"db_url"`
	Currret_username string `json:"current_username"`
}

func Read() Config {
	path, err := getFilePath(gatorconfigjson)
	if err != nil {
		fmt.Errorf("Error getting path")
	}
	// open the file into a stream for the decoder to use
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Error opening file")
	}
	// decode and return the file
	var c Config
	err = json.NewDecoder(file).Decode(&c)
	if err != nil {
		fmt.Println("Error error decoding file")
	}

	return c
}

func (c Config) SetUser(current_username string) error {
	// get path
	path, err := getFilePath(gatorconfigjson)
	if err != nil {
		fmt.Errorf("Error getting path")
	}
	c.Currret_username = current_username
	// create prettified marshaled data
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		fmt.Errorf("Error marshaling config data")
	}
	//write to file
	err = os.WriteFile(path, data, 0644)
	if err != nil {
		fmt.Errorf("Error writing file")
	}
	return nil
}

func getFilePath(filename string) (string, error) {
	// get the home directory
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("Error finding home directory")
	}
	path := filepath.Join(home, filename)
	return path, nil
}

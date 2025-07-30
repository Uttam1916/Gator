package config

import (
	"fmt"
	"os"
)

type Config struct {
	Db_url string `json:"db_url"`
}

func Read() Config {
	fmt.Print(os.UserHomeDir())
	return Config{}
}

Read()
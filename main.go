package main

import (
	"fmt"

	"github.com/Uttam1916/Gator/internal/config"
)

var cfg config.Config

func main() {
	cfg = config.Read()
	cfg.SetUser("bobom")
	fmt.Println(config.Read())

}

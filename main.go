package main

import (
	"fmt"
	"os"

	"github.com/Uttam1916/Gator/internal/config"
)

type state struct {
	configpointer *config.Config
}

type command struct {
	name      string
	arguments []string
}

type commands struct {
	command_map map[string]func(*state, command) error
}

// declare config variables
var ste state
var cfg config.Config
var comms commands

func main() {
	if len(os.Args) < 2 {
		fmt.Println("not enough arguements")
		os.Exit(1)
	}
	cmd := command{
		name:      os.Args[1],
		arguments: os.Args[2:],
	}
	comms = commands{
		command_map: make(map[string]func(*state, command) error),
	}
	cfg = config.Read()
	ste.configpointer = &cfg

	//initialize command handlers
	comms.register("login", handlerLogin)

	err := comms.run(&ste, cmd)
	if err != nil {
		fmt.Printf("error:%v \n", err)
	}

}

func (comm commands) run(s *state, cmd command) error {
	handler, ok := comm.command_map[cmd.name]
	if !ok {
		return fmt.Errorf("unknown command: %s", cmd.name)
	}
	return handler(s, cmd)
}

func (comm *commands) register(name string, f func(*state, command) error) {
	comm.command_map[name] = f
}

func handlerLogin(s *state, c command) error {
	if len(c.arguments) < 1 {
		return fmt.Errorf("login requires a username")
	}
	username := c.arguments[0]
	err := s.configpointer.SetUser(username)
	if err != nil {
		return err
	}
	fmt.Printf("User set to %s\n", username)
	return nil
}

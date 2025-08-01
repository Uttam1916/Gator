package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Uttam1916/Gator/internal/config"
	"github.com/Uttam1916/Gator/internal/database"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type state struct {
	configpointer *config.Config
	db            *database.Queries
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
	//initialize state
	dbURL := "postgres://postgres:postgres@localhost:5432/gator?sslmode=disable"
	conn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("could not connect to DB:", err)
	}
	dbQueries := database.New(conn)

	cfg = config.Read()

	ste := state{
		db:            dbQueries,
		configpointer: &cfg,
	}

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
	//initialize command handlers
	comms.register("login", handlerLogin)
	comms.register("register", handlerRegister)
	comms.register("users", handlerUsers)
	err = comms.run(&ste, cmd)
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
	_, err := s.db.GetUserByName(context.Background(), username)
	if err != nil {
		fmt.Println("cant log into non existent user")
		os.Exit(1)
	}

	err = s.configpointer.SetUser(username)
	if err != nil {
		return err
	}
	fmt.Printf("User set to %s\n", username)
	return nil
}

func handlerRegister(s *state, c command) error {
	if len(c.arguments) < 1 {
		fmt.Println("register requires a username")
		os.Exit(1)
	}
	//check if user already exists
	_, err := s.db.GetUserByName(context.Background(), c.arguments[0])
	if err == nil {
		fmt.Println("user already exists")
		os.Exit(1)
	}
	usertobecreated := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      c.arguments[0],
	}
	_, err = s.db.CreateUser(context.Background(), usertobecreated)
	if err != nil {
		return fmt.Errorf("couldnt create new user: ", err)
	}
	err = s.configpointer.SetUser(c.arguments[0])
	if err != nil {
		return fmt.Errorf("couldnt set new user: ", err)
	}

	fmt.Printf("User '%s' registered successfully\n", c.arguments[0])
	return nil
}

func handlerUsers(s *state, c command) error {
	if len(c.arguments) > 1 {
		return fmt.Errorf("this function does not require arguements")
	}
	users, err := s.db.GetAllUsersName(context.Background())
	if err != nil {
		return fmt.Errorf("couldnt retrive usernames ")
	}
	for _, user := range users {
		if user == s.configpointer.Currret_username {
			fmt.Printf("* %s (current user)\n", user)
		} else {
			fmt.Printf("* %s\n", user)
		}
	}
	return nil
}

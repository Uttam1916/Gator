package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
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

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Items       []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
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

	ste = state{
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
	comms.register("agg", handlerAgg)
	comms.register("addfeed", handlerAddFeed)
	comms.register("feeds", handlerFeeds)

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
		return fmt.Errorf("couldnt retrive usernames \n")
	}
	for _, user := range users {
		if user == s.configpointer.Current_username {
			fmt.Printf("* %s (current user)\n", user)
		} else {
			fmt.Printf("* %s\n", user)
		}
	}
	return nil
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	// create the request
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("couldnt form request\n")
	}
	req.Header.Set("User-Agent", "gator")
	// use client to make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error recieving response\n")
	}
	defer resp.Body.Close()
	// obtain and convert xml into a struct
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading body\n")
	}
	var RSSresp RSSFeed

	err = xml.Unmarshal(body, &RSSresp)
	if err != nil {
		return nil, fmt.Errorf("couldnt convert xml into go struct\n")
	}
	//clean up the struct feilds
	RSSresp.Channel.Title = html.UnescapeString(RSSresp.Channel.Title)
	RSSresp.Channel.Description = html.UnescapeString(RSSresp.Channel.Description)
	for i := range RSSresp.Channel.Items {
		RSSresp.Channel.Items[i].Title = html.UnescapeString(RSSresp.Channel.Items[i].Title)
		RSSresp.Channel.Items[i].Description = html.UnescapeString(RSSresp.Channel.Items[i].Description)
	}

	return &RSSresp, nil
}

func handlerAgg(s *state, c command) error {
	rss, err := fetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
	if err != nil {
		return fmt.Errorf("error reading go struct\n")
	}
	fmt.Println(rss)
	return nil
}
func handlerAddFeed(s *state, c command) error {
	if len(c.arguments) < 2 {
		return fmt.Errorf("this function requires url and name")
	}
	// create feed struct and get userID being tied to feed
	userid, err := s.db.GetUserIdByName(context.Background(), s.configpointer.Current_username)
	if err != nil {
		return fmt.Errorf("error obtaining user id\n")
	}
	feedinfo := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      c.arguments[0],
		Url:       c.arguments[1],
		UserID:    userid,
	}

	feed, err := s.db.CreateFeed(context.Background(), feedinfo)
	if err != nil {
		return fmt.Errorf("error adding feed\n")
	}
	fmt.Println("New feed added:")
	fmt.Printf("  ID        : %s\n", feed.ID)
	fmt.Printf("  Name      : %s\n", feed.Name)
	fmt.Printf("  URL       : %s\n", feed.Url)
	fmt.Printf("  User ID   : %s\n", feed.UserID)
	fmt.Printf("  Created At: %s\n", feed.CreatedAt.Format(time.RFC3339))
	fmt.Printf("  Updated At: %s\n", feed.UpdatedAt.Format(time.RFC3339))
	return nil
}

func handlerFeeds(s *state, c command) error {
	feeds, err := s.db.ReturnAllFeedsWithUsers(context.Background())
	if err != nil {
		return fmt.Errorf("couldnt retrieve feed data from database\n")
	}
	for _, feed := range feeds {
		fmt.Println("----------")
		fmt.Printf("Feed Name : %s\n", feed.Name)
		fmt.Printf("Feed URL  : %s\n", feed.Url)
		fmt.Printf("Created By: %s\n", feed.Username)
	}
	return nil
}

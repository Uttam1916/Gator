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
	comms.register("addfeed", middlewareLogin(handlerAddFeed))
	comms.register("feeds", handlerFeeds)
	comms.register("follow", middlewareLogin(handlerFollow))
	comms.register("following", middlewareLogin(handlerFollowing))
	comms.register("unfollow", middlewareLogin(handlerUnfollow))

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
		return fmt.Errorf("couldnt create new user\n")
	}
	err = s.configpointer.SetUser(c.arguments[0])
	if err != nil {
		return fmt.Errorf("couldnt set new user\n")
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

// middleware higher order function to check login
func middlewareLogin(handler func(s *state, c command, user database.User) error) func(*state, command) error {
	return func(s *state, c command) error {
		if s.configpointer.Current_username == "" {
			return fmt.Errorf("No user logged in\n")
		}
		user, err := s.db.GetUserByName(context.Background(), s.configpointer.Current_username)
		if err != nil {
			return fmt.Errorf("error obtaining user info\n")
		}
		return handler(s, c, user)
	}
}

func handlerAddFeed(s *state, c command, user database.User) error {
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
	// ---- Automatically follow the feed ----
	feedfollowparams := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    userid,
		FeedID:    feed.ID,
	}

	feedfollow, err := s.db.CreateFeedFollow(context.Background(), feedfollowparams)
	if err != nil {
		return fmt.Errorf("feed created but follow failed: %v", err)
	}

	fmt.Printf("User '%s' is now following feed '%s'\n", feedfollow.UserName, feedfollow.FeedName)

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

func handlerFollow(s *state, c command, user database.User) error {
	userid, err := s.db.GetUserIdByName(context.Background(), s.configpointer.Current_username)
	if err != nil {
		return fmt.Errorf("error obtaining user id\n")
	}
	feedid, err := s.db.GetFeedIdFromUrl(context.Background(), c.arguments[0])
	if err != nil {
		return fmt.Errorf("error obtaining feed id\n")
	}
	feedfollowparams := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    userid,
		FeedID:    feedid,
	}
	feedfollow, err := s.db.CreateFeedFollow(context.Background(), feedfollowparams)
	if err != nil {
		return fmt.Errorf("error obtaining feed follow details\n")
	}
	fmt.Printf("Feed Name : %v\n", feedfollow.FeedName)
	fmt.Printf("User Name : %v\n", feedfollow.UserName)
	return nil
}

func handlerFollowing(s *state, c command, user database.User) error {
	userid, err := s.db.GetUserIdByName(context.Background(), s.configpointer.Current_username)
	if err != nil {
		return fmt.Errorf("error obtaining user id\n")
	}
	feedfollows, err := s.db.GetFeedFollowsForUser(context.Background(), userid)
	fmt.Println("Feed follows for current user:")
	for i, feedfollow := range feedfollows {
		fmt.Printf(" %v.\n", 1+i)
		fmt.Printf("Feed Name : %v\n", feedfollow.FeedName)
		fmt.Printf("User Name : %v\n", feedfollow.UserName)
	}
	return nil
}

func handlerUnfollow(s *state, c command, user database.User) error {
	if len(c.arguments) < 1 {
		return fmt.Errorf("this function requires url\n")
	}
	params := database.DeleteFeedFollowByUserAndURLParams{
		Name: s.configpointer.Current_username,
		Url:  c.arguments[0],
	}
	err := s.db.DeleteFeedFollowByUserAndURL(context.Background(), params)
	if err != nil {
		return fmt.Errorf("error unfollowing feed\n")
	}
	return nil
}

func scrapeFeeds(s *state) error {
	nextfeed, err := s.db.GetNextFeed(context.Background())
	if err != nil {
		return fmt.Errorf("error getting next feed\n")
	}
	err = s.db.MarkFetchedFeed(context.Background(), nextfeed.ID)
	if err != nil {
		return fmt.Errorf("error marking fetched feed\n")
	}
	feed, err := fetchFeed(context.Background(), nextfeed.Url)
	if err != nil {
		return fmt.Errorf("error getting feed from database\n")
	}
	printRSSFeed(feed)
	return nil
}

func printRSSFeed(feed *RSSFeed) {
	fmt.Println("📡 Feed Info")
	fmt.Println("┌──────────────────────────────────────────────")
	fmt.Printf("│ Title       : %s\n", feed.Channel.Title)
	fmt.Printf("│ Link        : %s\n", feed.Channel.Link)
	fmt.Printf("│ Description : %s\n", feed.Channel.Description)
	fmt.Println("└──────────────────────────────────────────────")

	fmt.Printf("\n📰 %d Posts:\n", len(feed.Channel.Items))
	for i, item := range feed.Channel.Items {
		fmt.Printf("  ── [%d] ────────────────────────────────────────\n", i+1)
		fmt.Printf("  📌 Title      : %s\n", item.Title)
		fmt.Printf("  🔗 Link       : %s\n", item.Link)
		fmt.Printf("  📝 Summary    : %s\n", item.Description)
		fmt.Printf("  📅 Published  : %s\n", item.PubDate)
	}
}

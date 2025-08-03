# Gator 

Gator is a command-line application written in Go for managing RSS feeds, following users, aggregating feed content, and browsing posts. It supports user registration, login, feed management, and scraping of RSS content.

## Prerequisites

To run Gator locally, you must have:

- [Go](https://go.dev/dl/) (version 1.20+ recommended)
- [PostgreSQL](https://www.postgresql.org/) running on your machine

Make sure PostgreSQL is running with the following connection (or modify it in `main.go`):

## Installation

To install the Gator CLI:

```bash
go install github.com/Uttam1916/Gator@latest
```
## Configuration

Create a config file (.gator_config.json) in your home directory. This file is used to store the currently logged-in username. It will be automatically updated during login and registration.

Example file content:
```bash
 {
   "current_username": "your-username"
 }
```
## Running Gator

Gator is used via commands. Each command may require arguments. You can run the binary as 
```bash
gator <command> [arguments...]
```

### Available Commands

register <username>: Register a new user.

login <username>: Log in as an existing user.

users: List all registered users.

addfeed <name> <url>: Add a new RSS feed (and follow it).

feeds: List all available feeds.

follow <url>: Follow an existing feed by its URL.

following: List feeds you're following.

unfollow <url>: Unfollow a feed by its URL.

agg <duration>: Continuously scrape feeds every <duration>, e.g., 30s, 2m.

browse [limit]: View your latest followed posts (default limit is 2).

help: Show documentation for all available commands.
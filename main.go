package main

import (
	"fmt"
	"os"

	"github.com/yevhenshymotiuk/stickers-storage-bot/bot"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
	}
}

func run() error {
	bot, err := bot.NewBot()
	if err != nil {
		return err
	}

	bot.CheckForUpdates()

	return nil
}

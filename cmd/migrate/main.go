package main

import (
	"context"
	"fmt"
	"os"

	"github.com/11DingKing/autumn-grain-resilience/internal/config"
	"github.com/11DingKing/autumn-grain-resilience/internal/storage"
)

func main() {
	db, err := storage.Open(config.Load().DatabaseURL)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer db.Close()
	if err := storage.Migrate(context.Background(), db); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

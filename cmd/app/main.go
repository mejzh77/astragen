package main

import (
	"context"
	"log"
	"os"

	"github.com/mejzh77/astragen/internal/gsheets"
)

func main() {
	ctx := context.Background()
	creds, err := os.ReadFile("credentials.json")
	check(err)
	sheetService, err := gsheets.NewService(ctx, creds)
	check(err)
	err = gsheets.RunSync(ctx)
}

func check(err error) {
	if err != nil {
		log.Fatalf("%v", err)
	}
}

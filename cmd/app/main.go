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
	s, err := gsheets.NewService(ctx, creds)
	check(err)
	err = s.RunSync(ctx)
	check(err)
}

func check(err error) {
	if err != nil {
		log.Fatalf("%v", err)
	}
}

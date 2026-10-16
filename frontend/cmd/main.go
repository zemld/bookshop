package main

import (
	"os"

	"bookshop/frontend/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		os.Exit(1)
	}
}

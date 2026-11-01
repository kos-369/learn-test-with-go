package main

import (
	"log"
	"os"

	poker "github.com/kos-369/learn-test-with-go/http_server"
)

const dbFileName = "game.db.json"

func main() {
	store, close, err := poker.FileSystemPlayerStoreFromFile(dbFileName)
	if err != nil {
		log.Fatal(err)
	}
	defer close()

	game := poker.NewCLI(store, os.Stdin)
	game.PlayPoker()
}

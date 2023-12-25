package main

import (
	"encoding/json"
	"hockey-teams-parser/internal/parser"
	"log"
	"os"
	"time"
)

func main() {
	startTime := time.Now()
	teams := parser.CollectTeams()
	log.Printf("elapsed time %.2fs", time.Since(startTime).Seconds())
	log.Printf("parsed %d teams", len(teams))
	data, err := json.Marshal(teams)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	log.Println("dump teams to file!")
	err = os.WriteFile("teams.json", data, 0644)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	log.Println("all operations complete!")
}

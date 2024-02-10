package main

import (
	"github.com/Hydoc/guess-dev/backend/internal"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
)

const (
	defaultGuesses            = "1,2,3,4,5"
	defaultGuessesDescription = "Bis zu 4 Std.,Bis zu 8 Std.,Bis zu 3 Tagen,Bis zu 5 Tagen,Mehr als 5 Tage"
)

func main() {
	var possibleGuesses, possibleGuessesDescription string
	possibleGuesses, ok := os.LookupEnv("POSSIBLE_GUESSES")
	possibleGuessesDescription, okDesc := os.LookupEnv("POSSIBLE_GUESSES_DESC")
	if !ok {
		log.Println("can not find env POSSIBLE_GUESSES")
		possibleGuesses = defaultGuesses
	}
	if !okDesc {
		log.Println("can not find env POSSIBLE_GUESSES_DESC")
		possibleGuessesDescription = defaultGuessesDescription
	}
	log.Println("using possible guesses", possibleGuesses)
	log.Println("using possible guesses description", possibleGuessesDescription)

	upgrader := &websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	config, err := internal.NewGuessConfig(possibleGuesses, possibleGuessesDescription)
	if err != nil {
		log.Fatal(err)
		return
	}
	app := internal.NewApplication(http.NewServeMux(), upgrader, config)
	go app.ListenForRoomDestroy()
	router := app.ConfigureRouting()
	log.Fatal(http.ListenAndServe(":8080", router))
}

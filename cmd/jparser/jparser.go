package main

import (
	"log"

	"github.com/ilyakaznacheev/jparser"
)

func main() {
	// create new server instance
	s := jparser.NewServer(1)
	// run server (it handles keyboard interrupt by default)
	log.Fatal(s.Start())
}

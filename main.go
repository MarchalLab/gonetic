package main

import (
	"fmt"

	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/MarchalLab/gonetic/cmd"
)

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	fmt.Println("Welcome to GoNetic")
	cmd.Execute()
}

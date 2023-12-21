package main

import (
	"fmt"
	"os"
	"proj1/feed"
	"proj1/server"
	"strconv"
)

func main() {
	var Mode string
	var ConsumersCount int

	if len(os.Args) < 2 {
		Mode = "s"
	} else if len(os.Args) == 2 {
		Mode = "p"
		ConsumersCount, _ = strconv.Atoi(os.Args[1])
	} else {
		fmt.Println("Usage: twitter <number of consumers>\n<number of consumers> = the number of goroutines (i.e., consumers) to be part of the parallel version.")
	}

	config := server.NewConfig(Mode, ConsumersCount)

	userFeed := feed.NewFeed()
	server.Run(*config, userFeed)
}

package main

import (
	"fmt"
	"log"

	socketIO "../../goSocketIOClient"
)

func main() {
	conn, err := socketIO.Dial("https://ukhas.net/logtail")
	if err != nil {
		log.Fatal(err)
	}

	for {
		msg := <-conn.Output
		fmt.Println(msg.Event)
	}

}

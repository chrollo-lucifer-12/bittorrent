package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/codecrafters-io/bittorrent-starter-go/app/bencode"
)

var _ = json.Marshal

func main() {
	decoder := bencode.NewDecoder()
	command := os.Args[1]
	if command == "decode" {
		bencodedValue := os.Args[2]
		decodedValue, err := decoder.Decode(bencodedValue)
		if err != nil {
			fmt.Println(err)
			return
		}
		jsonOutput, _ := json.Marshal(decodedValue)
		fmt.Println(string(jsonOutput))
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}

package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/codecrafters-io/bittorrent-starter-go/app/bencode"
	"github.com/codecrafters-io/bittorrent-starter-go/app/metainfo"
)

var _ = json.Marshal

func main() {
	decoder := bencode.NewDecoder()
	encoder := bencode.NewEncoder()
	command := os.Args[1]
	if command == "decode" {
		bencodedValue := os.Args[2]
		bencodedValueBytes := []byte(bencodedValue)
		decodedValue, err := decoder.Decode(bencodedValueBytes)
		if err != nil {
			fmt.Println(err)
			return
		}
		jsonOutput, _ := json.Marshal(decodedValue)
		fmt.Println(string(jsonOutput))
	} else if command == "info" {
		filename := os.Args[2]
		opts := metainfo.MetaInfoOpts{
			Filename: filename,
			Decoder:  decoder,
			Encoder:  encoder,
		}
		metaInfo := metainfo.NewMetaInfo(opts)
		metaInfo.Parse()
		metaInfo.Hash()
		metaInfo.CalculatePieceHashes()
		metaInfo.DiscoverPeers()
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}

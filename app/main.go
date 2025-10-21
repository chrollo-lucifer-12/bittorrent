package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/codecrafters-io/bittorrent-starter-go/app/bencode"
	"github.com/codecrafters-io/bittorrent-starter-go/app/magnetlink"
	"github.com/codecrafters-io/bittorrent-starter-go/app/metainfo"
	"github.com/codecrafters-io/bittorrent-starter-go/app/p2p"
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

		tcpOpts := p2p.TCPConnectionOpts{
			MetaInfo: metaInfo,
			MetaInfoDownloadFunc: func(t *p2p.TCPConnection) error {
				var err error
				peers, err := t.MetaInfo.DiscoverPeers()
				if err != nil || len(peers) == 0 {
					return err
				}
				go t.DialPeer(peers[0])
				return nil
			},
		}
		t := p2p.NewTCPConnection(tcpOpts)
		err := t.DownloadFile()
		if err != nil {
			fmt.Println(err)
		}
	} else if command == "magnet_parse" {
		magnetLink := os.Args[2]
		opts := magnetlink.MagnetLinkOpts{
			MagnetLink: magnetLink,
		}
		magnetLinkHandler := magnetlink.NewMagentLink(opts)
		info, err := magnetLinkHandler.Parse()
		if err != nil {
			fmt.Printf("error in parsing : %s\n", err)
		}
		fmt.Println("parsed info : ", info)
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}

	select {}
}

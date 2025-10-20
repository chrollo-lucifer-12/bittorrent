package p2p

import (
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/codecrafters-io/bittorrent-starter-go/app/metainfo"
)

type TCPConnectionOpts struct {
	MetaInfo *metainfo.MetaInfo
}

type TCPConnection struct {
	TCPConnectionOpts
	listener net.Listener
	peers    []string
}

func NewTCPConnection(opts TCPConnectionOpts) *TCPConnection {
	return &TCPConnection{
		TCPConnectionOpts: opts,
	}
}

func (t *TCPConnection) ListenAndAccept() error {
	var err error
	t.peers, err = t.MetaInfo.DiscoverPeers()
	if err != nil || len(t.peers) == 0 {
		return err
	}
	//	fmt.Printf("Found peers: %s\n", t.peers)
	conn, err := net.DialTimeout("tcp", t.peers[0], 3*time.Second)
	if err != nil {
		return err
	}
	conn.SetDeadline(time.Now().Add(3 * time.Second))
	defer conn.SetDeadline(time.Time{})
	request := make([]byte, 0, 68)
	request = append(request, 19)
	request = append(request, []byte("BitTorrent protocol")...)
	request = append(request, 0, 0, 0, 0, 0, 0, 0, 0)
	request = append(request, t.MetaInfo.InfoHash[:]...)
	request = append(request, []byte(t.MetaInfo.GetPeerId())...)
	_, err = conn.Write(request)
	if err != nil {
		return err
	}
	response := make([]byte, 68)
	_, err = io.ReadAtLeast(conn, response, 68)
	if err != nil {
		return err
	}
	fmt.Println(hex.EncodeToString(response[48:68]))
	return nil
}

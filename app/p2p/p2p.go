package p2p

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"
	"strings"

	"github.com/codecrafters-io/bittorrent-starter-go/app/filestore"
	"github.com/codecrafters-io/bittorrent-starter-go/app/metainfo"
)

const BlockSize = 4 * 1024

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

func (t *TCPConnection) DownloadFile() error {
	var err error
	t.peers, err = t.MetaInfo.DiscoverPeers()
	if err != nil || len(t.peers) == 0 {
		return err
	}
	//	fmt.Printf("Found peers: %s\n", t.peers)
	go t.DialPeer(t.peers[0])
	return nil
}

func (t *TCPConnection) handshake(conn net.Conn) {
	request := make([]byte, 0, 68)
	request = append(request, 19)
	request = append(request, []byte("BitTorrent protocol")...)
	request = append(request, 0, 0, 0, 0, 0, 0, 0, 0)
	request = append(request, t.MetaInfo.InfoHash[:]...)
	request = append(request, []byte(t.MetaInfo.GetPeerId())...)
	_, _ = conn.Write(request)
	handshake := make([]byte, 68)
	_, err := io.ReadFull(conn, handshake)
	if err != nil {
		log.Println("Failed to read handshake:", err)
		return
	}
	fmt.Println("Handshake received:", hex.EncodeToString(handshake))
}

func readMessage(conn net.Conn) (id byte, payload []byte, err error) {
	lengthBuf := make([]byte, 4)
	if _, err = io.ReadFull(conn, lengthBuf); err != nil {
		return
	}

	length := binary.BigEndian.Uint32(lengthBuf)
	if length == 0 {
		return 0, nil, nil
	}

	msgBuf := make([]byte, length)
	if _, err = io.ReadFull(conn, msgBuf); err != nil {
		return
	}

	id = msgBuf[0]
	payload = msgBuf[1:]
	return
}

func sendInterested(conn net.Conn) error {
	msg := []byte{0, 0, 0, 1, 2}
	_, err := conn.Write(msg)
	return err
}

func sendRequest(conn net.Conn, pieceIndex, begin, length int) error {
	buf := make([]byte, 17)
	binary.BigEndian.PutUint32(buf[0:4], 13)
	buf[4] = 6
	binary.BigEndian.PutUint32(buf[5:9], uint32(pieceIndex))
	binary.BigEndian.PutUint32(buf[9:13], uint32(begin))
	binary.BigEndian.PutUint32(buf[13:17], uint32(length))

	_, err := conn.Write(buf)
	return err
}

func receivePiece(conn net.Conn) (pieceIndex int, begin int, block []byte, err error) {
	id, payload, err := readMessage(conn)
	if err != nil {
		return
	}
	if id != 7 {
		return 0, 0, nil, fmt.Errorf("expected piece message, got ID %d", id)
	}

	pieceIndex = int(binary.BigEndian.Uint32(payload[0:4]))
	begin = int(binary.BigEndian.Uint32(payload[4:8]))
	block = payload[8:]
	return pieceIndex, begin, block, nil
}

func (t *TCPConnection) DialPeer(peer string) {
	conn, _ := net.Dial("tcp", t.peers[0])
	defer conn.Close()

	t.handshake(conn)

	for {
		id, _, err := readMessage(conn)
		if err != nil {
			return
		}
		if id == 5 {
			break
		}
	}

	sendInterested(conn)

	for {
		id, _, err := readMessage(conn)
		if err != nil {
			return
		}
		if id == 1 {
			break
		}
	}

	fileName := t.MetaInfo.GetName()
	dirName := strings.Split(fileName, ".")[0]

	fileStoreOpts := filestore.FileStoreOpts{
		Dirname: dirName,
	}

	fileStore := filestore.NewFileStore(fileStoreOpts)

	pieceLength := t.MetaInfo.GetPieceLength()
	pieceIndex := 0

	for begin := 0; begin < int(pieceLength); begin += BlockSize {
		blockLen := BlockSize
		if begin+BlockSize > int(pieceLength) {
			blockLen = int(pieceLength) - begin
		}

		if err := sendRequest(conn, pieceIndex, begin, blockLen); err != nil {
			log.Println("Failed to send request:", err)
			return
		}

		idx, b, block, err := receivePiece(conn)
		if err != nil {
			log.Println("Failed to receive piece:", err)
			return
		}

		if idx != pieceIndex || b != begin {
			log.Println("Received wrong piece or offset")
			return
		}

		fileName := fmt.Sprintf("piece_%d_block_%d", idx, b)
		fileStore.AddFile(fileName, block)

		fmt.Printf("Received block at offset %d (length %d)\n", b, len(block))
	}
}

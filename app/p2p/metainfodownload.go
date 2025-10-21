package p2p

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"
	"strings"

	"github.com/codecrafters-io/bittorrent-starter-go/app/filestore"
)

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

func (t *TCPConnection) validatePiece(pieceIndex int, data []byte) bool {
	dataHash := sha1.Sum(data)
	start := pieceIndex * 20
	end := start + 20
	expectedHash := t.MetaInfo.PieceHashes[start:end]

	match := bytes.Equal(dataHash[:], expectedHash)
	fmt.Printf("piece %d: expected %x, received %x, match: %v\n", pieceIndex, expectedHash, dataHash, match)
	return match
}

func (t *TCPConnection) DialPeer(peer string) {
	conn, _ := net.Dial("tcp", peer)
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

	totalPieces := int(t.MetaInfo.GetNumPieces())
	pieceLength := int(t.MetaInfo.GetPieceLength())

	for pieceIndex := 0; pieceIndex < totalPieces; pieceIndex++ {
		currentPieceLength := pieceLength
		if pieceIndex == totalPieces-1 {
			lastPieceSize := int(t.MetaInfo.GetLength()) - pieceLength*(totalPieces-1)
			currentPieceLength = lastPieceSize
		}

		pieceBuffer := make([]byte, currentPieceLength)
		received := 0

		for begin := 0; begin < currentPieceLength; begin += BlockSize {
			blockLen := BlockSize
			if begin+BlockSize > currentPieceLength {
				blockLen = currentPieceLength - begin
			}
			if err := sendRequest(conn, pieceIndex, begin, blockLen); err != nil {
				return
			}
			idx, b, block, err := receivePiece(conn)
			if err != nil {
				return
			}
			if idx != pieceIndex || b != begin {
				return
			}
			copy(pieceBuffer[b:b+len(block)], block)
			received += len(block)
		}

		if t.validatePiece(pieceIndex, pieceBuffer) {
			fmt.Printf("Piece %d validated successfully!\n", pieceIndex)
			fileStore.AddFile(fmt.Sprintf("piece_%d", pieceIndex), pieceBuffer)
		} else {
			fmt.Printf("Piece %d failed validation\n", pieceIndex)
		}
	}

	fileStore.CombineFiles()
}

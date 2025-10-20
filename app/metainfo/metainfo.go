package metainfo

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/codecrafters-io/bittorrent-starter-go/app/bencode"
)

type TorrentInfo struct {
	length      int
	name        string
	pieceLength int
	pieces      any
}

type TorrentFile struct {
	announce string
	info     TorrentInfo
}

type MetaInfoOpts struct {
	Filename string
	Decoder  *bencode.Decoder
	Encoder  *bencode.Encoder
}

type MetaInfo struct {
	filename    string
	decoder     *bencode.Decoder
	encoder     *bencode.Encoder
	peerId      string
	TorrentFile *TorrentFile
	InfoHash    [20]byte
	PieceHashes []byte
}

func NewMetaInfo(opts MetaInfoOpts) *MetaInfo {
	peerID := make([]byte, 20)
	rand.Read(peerID)

	return &MetaInfo{
		filename: opts.Filename,
		decoder:  opts.Decoder,
		encoder:  opts.Encoder,
		peerId:   string(peerID),
	}
}

func (m *MetaInfo) GetNumPieces() int {
	fileLength := m.GetLength()
	pieceLength := m.GetPieceLength()
	totalPieces := (fileLength + pieceLength - 1) / pieceLength
	return totalPieces
}

func (m *MetaInfo) GetLength() int {
	return m.TorrentFile.info.length
}

func (m *MetaInfo) GetPieceLength() int {
	return m.TorrentFile.info.pieceLength
}

func (m *MetaInfo) GetName() string {
	return m.TorrentFile.info.name
}

func (m *MetaInfo) GetPeerId() string {
	return m.peerId
}

func (m *MetaInfo) calculatePieceHashes() {
	pieces := m.TorrentFile.info.pieces.([]byte)
	m.PieceHashes = pieces
}

func (m *MetaInfo) readFile() string {
	data, _ := os.ReadFile(m.filename)
	return string(data)
}

func (m *MetaInfo) DiscoverPeers() ([]string, error) {
	var res []string
	m.parse()
	m.hash()
	m.calculatePieceHashes()
	trackerURL := m.TorrentFile.announce
	info_hash := m.InfoHash
	peer_id := string(m.peerId)
	port := "6881"
	uploaded := "0"
	downloaded := "0"
	left := m.TorrentFile.info.length
	compact := "1"
	u, _ := url.Parse(trackerURL)
	query := u.Query()
	query.Set("info_hash", string(info_hash[:]))
	query.Set("peer_id", string(peer_id))
	query.Set("port", port)
	query.Set("uploaded", uploaded)
	query.Set("downloaded", downloaded)
	query.Set("left", strconv.Itoa(left))
	query.Set("compact", compact)

	u.RawQuery = query.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	decoded, err := m.decoder.Decode(body)
	if err != nil {
		return res, err
	}

	dict := decoded.(map[string]any)
	peersRaw, ok := dict["peers"].([]byte)
	if !ok {
		fmt.Println("No peers returned by tracker")
		return res, nil
	}

	for i := 0; i < len(peersRaw); i += 6 {
		ip := net.IP(peersRaw[i : i+4])
		port := binary.BigEndian.Uint16(peersRaw[i+4 : i+6])
		res = append(res, fmt.Sprintf("%s:%d", ip, port))
	}

	return res, nil
}

func (m *MetaInfo) hash() error {
	if m.TorrentFile == nil {
		err := m.parse()
		return err
	}
	info := make(map[string]any)
	info["pieces"] = m.TorrentFile.info.pieces
	info["length"] = m.TorrentFile.info.length
	info["name"] = m.TorrentFile.info.name
	info["piece length"] = m.TorrentFile.info.pieceLength
	bencodedInfo, err := m.encoder.Encode(info)
	if err != nil {
		return err
	}
	hash := sha1.Sum(bencodedInfo)
	m.InfoHash = hash
	return nil
}

func (m *MetaInfo) parse() error {
	data := m.readFile()

	decodedDataAny, err := m.decoder.Decode([]byte(data))
	if err != nil {
		return err
	}

	decodedData, ok := decodedDataAny.(map[string]any)
	if !ok {
		return fmt.Errorf("invalid torrent file format")
	}

	announceRaw, ok := decodedData["announce"]
	if !ok {
		return fmt.Errorf("announce field missing")
	}

	announce, ok := announceRaw.([]byte)
	if !ok {
		return fmt.Errorf("announce field not a []byte")
	}

	infoRaw, ok := decodedData["info"]
	if !ok {
		return fmt.Errorf("info field missing")
	}

	infoMap, ok := infoRaw.(map[string]any)
	if !ok {
		return fmt.Errorf("info field not a dictionary")
	}
	fmt.Println(infoMap)
	lengthVal, ok := infoMap["length"]
	if !ok {
		return fmt.Errorf("length field missing")
	}
	length, ok := lengthVal.(int)
	if !ok {
		return fmt.Errorf("length field not an int")
	}

	nameVal, ok := infoMap["name"]
	if !ok {
		return fmt.Errorf("name field missing")
	}
	name, ok := nameVal.([]byte)
	if !ok {
		return fmt.Errorf("name field not a []byte")
	}

	pieceLengthVal, ok := infoMap["piece length"]
	if !ok {
		return fmt.Errorf("piece length field missing")
	}
	pieceLength, ok := pieceLengthVal.(int)
	if !ok {
		return fmt.Errorf("piece length not an int")
	}

	torrentInfo := TorrentInfo{
		length:      length,
		name:        string(name),
		pieceLength: pieceLength,
		pieces:      infoMap["pieces"],
	}

	torrentFile := TorrentFile{
		announce: string(announce),
		info:     torrentInfo,
	}
	m.TorrentFile = &torrentFile
	return nil
}

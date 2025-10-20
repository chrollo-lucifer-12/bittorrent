package metainfo

import (
	"crypto/sha1"
	"fmt"
	"os"

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
	TorrentFile *TorrentFile
	InfoHash    string
	PieceHashes string
}

func NewMetaInfo(opts MetaInfoOpts) *MetaInfo {
	return &MetaInfo{
		filename: opts.Filename,
		decoder:  opts.Decoder,
		encoder:  opts.Encoder,
	}
}

func (m *MetaInfo) CalculatePieceHashes() {
	pieces := m.TorrentFile.info.pieces.([]byte)
	hashHex := fmt.Sprintf("%x", pieces)
	m.PieceHashes = hashHex
}

func (m *MetaInfo) readFile() string {
	data, _ := os.ReadFile(m.filename)
	return string(data)
}

func (m *MetaInfo) Hash() error {
	if m.TorrentFile == nil {
		err := m.Parse()
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
	m.InfoHash = fmt.Sprintf("%x", hash)
	return nil
}

func (m *MetaInfo) Parse() error {
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

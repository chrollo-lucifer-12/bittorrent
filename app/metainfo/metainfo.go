package metainfo

import (
	"os"

	"github.com/codecrafters-io/bittorrent-starter-go/app/bencode"
)

type TorrentInfo struct {
	length      int
	name        string
	pieceLength int
	pieces      string
}

type TorrentFile struct {
	announce string
	info     TorrentInfo
}

type MetaInfoOpts struct {
	Filename string
	Decoder  *bencode.Decoder
}

type MetaInfo struct {
	filename string
	decoder  *bencode.Decoder
}

func NewMetaInfo(opts MetaInfoOpts) *MetaInfo {
	return &MetaInfo{
		filename: opts.Filename,
		decoder:  opts.Decoder,
	}
}

func (m *MetaInfo) readFile() string {
	data, _ := os.ReadFile(m.filename)
	return string(data)
}

func (m *MetaInfo) Info() TorrentFile {
	data := m.readFile()
	decodedDataAny, _ := m.decoder.Decode([]byte(data))
	decodedData, _ := decodedDataAny.(map[string]any)
	decodedDataInfo := decodedData["info"].(map[string]any)
	torrentInfo := TorrentInfo{
		length: decodedDataInfo["length"].(int),
		name:   decodedDataInfo["name"].(string),
	}
	torrentFile := TorrentFile{
		info:     torrentInfo,
		announce: decodedData["announce"].(string),
	}
	return torrentFile
}

package p2p

import (
	"github.com/codecrafters-io/bittorrent-starter-go/app/magnetlink"
	"github.com/codecrafters-io/bittorrent-starter-go/app/metainfo"
)

const BlockSize = 16 * 1024

type TCPConnectionOpts struct {
	MetaInfo             *metainfo.MetaInfo
	MagnetLink           *magnetlink.MagnetLink
	MetaInfoDownloadFunc func(*TCPConnection) error
}

type TCPConnection struct {
	TCPConnectionOpts
}

func NewTCPConnection(opts TCPConnectionOpts) *TCPConnection {
	return &TCPConnection{
		TCPConnectionOpts: opts,
	}
}

func (t *TCPConnection) DownloadFile() error {
	if t.MetaInfoDownloadFunc != nil {
		return t.MetaInfoDownloadFunc(t)
	}
	return nil
}

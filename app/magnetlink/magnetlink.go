package magnetlink

import (
	"net/url"
	"strings"
)

type MagnetLinkInfo struct {
	Infohash   string
	Filename   string
	TrackerUrl string
}

type MagnetLinkOpts struct {
	MagnetLink string
}

type MagnetLink struct {
	raw        string
	infohash   string
	filename   string
	trackerUrl string
}

func NewMagentLink(opts MagnetLinkOpts) *MagnetLink {
	return &MagnetLink{
		raw: opts.MagnetLink,
	}
}

func (m *MagnetLink) Parse() (MagnetLinkInfo, error) {
	u, err := url.Parse(m.raw)
	info := MagnetLinkInfo{}
	if err != nil {
		return info, err
	}
	q := u.Query()
	m.infohash = strings.TrimPrefix(q.Get("xt"), "urn:btih:")
	m.filename = q.Get("dn")
	m.trackerUrl = q.Get("tr")
	info.Filename = m.filename
	info.Infohash = m.infohash
	info.TrackerUrl = m.trackerUrl
	return info, nil
}

func (m *MagnetLink) GetInfo() MagnetLinkInfo {
	return MagnetLinkInfo{
		Infohash:   m.infohash,
		Filename:   m.filename,
		TrackerUrl: m.trackerUrl,
	}
}

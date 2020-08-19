package model

import (
	"encoding/json"
)

type Sync struct {
	Categories Categories
	FullUpdate bool
	Rid        uint32
	Torrents   ArrTorrent
}

type Categories struct{}

type ArrTorrent []Torrent

type Torrent struct {
	Name      string         `json:"name"`
	Hash      string         `json:"hash"`
	Size      uint64         `json:"size"`
	Status    string         `json:"state"`
	Progress  float64        `json:"progress"`
	DownSpeed uint           `json:"dlspeed"`
	UpSpeed   uint           `json:"upspeed"`
	Estimate  uint64         `json:"eta"`
	Encode    EncodeProgress `json:"encode"`
}

type EncodeProgress struct {
	Name     string  `json:"name"`
	Progress float32 `json:"progress"`
	Status   string  `json:"status"`
}

func (t *ArrTorrent) UnmarshalJSON(b []byte) error {
	var result map[string]Torrent
	if err := json.Unmarshal(b, &result); err != nil {
		return err
	}

	for k, v := range result {
		v.Hash = k
		*t = append(*t, v)
	}
	return nil
}

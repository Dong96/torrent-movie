package api

import (
	"backend/model"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/websocket"
)

//Describe action request to qBittorrent client
type qbtAction int

const (
	Login qbtAction = iota
	AddNewTorrent
	Delete
	GetTorrents
	PauseTorrent
	ResumeTorrent
	SyncData

	Encode
	EncodeProgress
	StopEncode
)

/*
	Path: /magnet
	Method: POST
*/
func magnetHandle(w http.ResponseWriter, r *http.Request) error {
	r.ParseMultipartForm(0)
	magnet := r.FormValue("urls")
	log.Println("Maget: ", magnet)
	err := addNewMagnetRequest(magnet)
	if err != nil {
		return err
	}
	return nil
}

/*
	Path: torrents/delete?hash=&delData=
	Method: DELETE
*/
func deleteHandle(w http.ResponseWriter, r *http.Request) error {
	param := r.URL.Query()["hashes"]
	hashes := strings.Join(param, "|")
	delData := r.URL.Query().Get("delData")

	if delData == "" {
		delData = "true"
	}

	api := formingAPI(Delete)
	u, err := url.Parse(api)
	if err != nil {
		return err
	}
	q := u.Query()
	q.Set("hashes", hashes)
	q.Set("deleteFiles", delData)
	u.RawQuery = q.Encode()

	_, err = client.Get(u.String())
	if err != nil {
		return err
	}
	return nil
}

/*
	Path: /torrents
	Method: GET
*/
func torrentsHandle(w http.ResponseWriter, r *http.Request) error {
	api := formingAPI(GetTorrents)

	res, err := client.Get(api)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("Bad request to qbserver: %v", res.StatusCode)
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	var data []model.Torrent
	if err := json.Unmarshal(body, &data); err != nil {
		return err
	}

	err = json.NewEncoder(w).Encode(data)
	return err
}

/*
	Path: /torrents/pause?hashes=
	Method: GET
*/
func pauseTorrentHandle(w http.ResponseWriter, r *http.Request) error {
	param := r.URL.Query()["hashes"]
	hashes := strings.Join(param, "|")
	api := formingAPI(PauseTorrent)
	u, err := url.Parse(api)
	if err != nil {
		return err
	}
	q := u.Query()
	q.Set("hashes", hashes)
	u.RawQuery = q.Encode()
	_, err = client.Get(u.String())
	return err
}

/*
	Path: /torrents/resume?hashes=
	Method: GET
*/
func resumeTorrentsHandle(w http.ResponseWriter, r *http.Request) error {
	param := r.URL.Query()["hashes"]
	hashes := strings.Join(param, "|")
	api := formingAPI(ResumeTorrent)
	u, err := url.Parse(api)
	if err != nil {
		return err
	}
	q := u.Query()
	q.Set("hashes", hashes)
	u.RawQuery = q.Encode()
	_, err = client.Get(u.String())
	return err
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func wsHandle(w http.ResponseWriter, r *http.Request) error {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	go func(conn *websocket.Conn) {
		defer func() {
			conn.Close()
		}()

		for data := range updateChan {
			if err := conn.WriteJSON(data); err != nil {
				log.Println("Websocket Error: ", err)
				break
			}
		}
	}(conn)
	return nil
}

func addNewMagnetRequest(magnet string) error {
	api := formingAPI(AddNewTorrent)

	data := map[string]string{
		"urls": magnet,
	}
	req, err := newPostRequest(api, data)
	if err != nil {
		return err
	}

	res, err := client.Do(req)
	if err != nil {
		ms := fmt.Sprintf("Could not send request: %v", err)
		log.Println(ms)
		return fmt.Errorf(ms)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		ms := fmt.Sprintf("Failed to connect to qBittorrent, Status code: %v", res.StatusCode)
		log.Printf(ms)
		return fmt.Errorf(ms)
	}
	log.Printf("Request %v success.", api)
	return nil
}

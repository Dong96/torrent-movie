package api

import (
	"backend/azureci"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/mux"
)

func startEncodeServer() error {
	return azureci.StartEncodeService()
}

func encodeHandle(w http.ResponseWriter, r *http.Request) (err error) {
	if err := azureci.StartEncodeService(); err != nil {
		return err
	}
	w.WriteHeader(http.StatusCreated)

	go func() {
		for {
			err, state := azureci.GetStateOfService()
			if err != nil {
				log.Println(err)
			}
			if state == "Succeeded" {
				break
			}
			time.Sleep(time.Second)
		}

		vars := mux.Vars(r)
		movie := vars["movie"]

		api := formingEncodeAPI(Encode)
		u, err := url.Parse(api)
		if err != nil {
			return
		}
		u, err = u.Parse(movie)
		if err != nil {
			return
		}

		log.Println("Send request to encode service")
		resp, err := client.Get(u.String())
		if err != nil {
			log.Println(err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			bytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Println(err)
				return
			}
			log.Printf("Bad response: %v - %v", resp.Status, string(bytes))
		}

	}()
	return
}

func stopEncodeHandle(w http.ResponseWriter, r *http.Request) (err error) {
	vars := mux.Vars(r)
	movie := vars["movie"]

	api := formingAPI(StopEncode)
	u, err := url.Parse(api)
	if err != nil {
		return
	}
	u.Parse(movie)
	resp, err := client.Get(u.String())
	if err != nil {
		return
	}
	defer resp.Body.Close()

	return
}

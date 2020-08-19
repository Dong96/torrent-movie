package api

import (
	"backend/logger"
	"backend/model"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
)

var client = &http.Client{
	Timeout:   15 * time.Second,
	Transport: &transport{underlyingTransport: http.DefaultTransport},
}

type transport struct {
	underlyingTransport http.RoundTripper
}

type ClientError struct {
	Root     error  `json:"-"`
	Response string `json:"error"`
	Status   int    `json:"-"`
}

func (e *ClientError) Error() string {
	return e.Root.Error()
}

func (e *ClientError) ResponseBody() ([]byte, error) {
	body, err := json.Marshal(e)
	if err != nil {
		return nil, fmt.Errorf("Error while parsing response body: %v", err)
	}
	return body, nil
}

type wrapHandle func(w http.ResponseWriter, r *http.Request) error

func (h wrapHandle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h(w, r)
	if err == nil {
		return
	}

	clientError, ok := err.(*ClientError)
	if !ok {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	body, err := clientError.ResponseBody()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(clientError.Status)
	w.Write(body)
}

func StartServer() {
	handler := createHandler()

	server := &http.Server{
		Addr:         viper.GetString("port"),
		WriteTimeout: time.Second * 10,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Second * 60,
		Handler:      handler,
	}
	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Could not start server: %v", err)
		}
	}()

	go checkServerRunning()

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGTERM)
	<-c

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Println(err)
		os.Exit(1)
	}

	log.Println("Shutdown successfull.")
	os.Exit(0)
}

func createHandler() http.Handler {
	var h http.Handler = matchingRoutes()
	h = loggingHandler(h)
	h = corsSetting(h)
	return h
}

func matchingRoutes() *mux.Router {
	r := mux.NewRouter().StrictSlash(true)
	// r.Use(loggingMiddleware)
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hi!"))
	})
	r.Handle("/magnet", wrapHandle(magnetHandle)).Methods(http.MethodPost)
	r.PathPrefix("/movie").Handler(http.StripPrefix("/movie", handlerFile()))

	s := r.PathPrefix("/torrents").Subrouter()
	s.Handle("/", wrapHandle(torrentsHandle)).Methods(http.MethodGet)
	s.Handle("/pause", wrapHandle(pauseTorrentHandle)).Methods(http.MethodGet)
	s.Handle("/resume", wrapHandle(resumeTorrentsHandle)).Methods(http.MethodGet)
	s.Handle("/delete", wrapHandle(deleteHandle)).Methods(http.MethodDelete)
	s.Handle("/ws", wrapHandle(wsHandle))

	encodeRouter := r.PathPrefix("/encode").Subrouter()
	encodeRouter.Handle("/{movie}", wrapHandle(encodeHandle)).Methods(http.MethodPost)
	encodeRouter.Handle("/stop/{movie}", wrapHandle(stopEncodeHandle)).Methods(http.MethodGet)

	return r
}

func handlerFile() http.Handler {
	return http.FileServer(http.Dir("/downloads"))
}

func loggingHandler(h http.Handler) http.Handler {
	return handlers.CombinedLoggingHandler(logger.MultiLogWriter(), h)
}

func corsSetting(h http.Handler) http.Handler {
	headersOK := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Range"})
	originsOK := handlers.AllowedOrigins([]string{"*"})
	methodsOK := handlers.AllowedMethods([]string{"GET", "POST", "OPTIONS", "DELETE", "PUT"})
	return handlers.CORS(originsOK, headersOK, methodsOK)(h)
}

func checkServerRunning() {
	addr := "http://localhost" + viper.GetString("port")
	for {
		time.Sleep(time.Second)

		log.Println("Checking server started...")
		res, err := http.Get(addr)
		if err != nil {
			log.Println("Failed:", err)
			continue
		}
		res.Body.Close()
		if res.StatusCode != http.StatusOK {
			log.Println("Not OK:", res.StatusCode)
			continue
		}

		break
	}
	log.Println("SERVER UP AND RUNNING AT: ", addr)
	loginRequest()
	pollingUpdate()
}

func (t *transport) RoundTrip(r *http.Request) (*http.Response, error) {
	if strings.Contains(r.URL.String(), viper.GetString("qbittorrent.base")) {
		r.Header.Set("Cookie", "SID="+viper.GetString("cookieTorrent"))
	}
	return t.underlyingTransport.RoundTrip(r)
}

// func loggingMiddleware(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		log.Println(r.RequestURI)
// 		next.ServeHTTP(w, r)
// 	})
// }

func newPostRequest(url string, data map[string]string) (*http.Request, error) {
	writer, body := createMultipartFormData(data)
	req, err := http.NewRequest(http.MethodPost, url, &body)
	if err != nil {
		log.Println("Could not create request: ", err)
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, nil
}

func createMultipartFormData(data map[string]string) (*multipart.Writer, bytes.Buffer) {
	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)

	for k, v := range data {
		writer.WriteField(k, v)
	}
	return writer, buffer
}

func formingAPI(action qbtAction) string {
	api := viper.GetString("qbittorrent.base")
	switch action {
	case Login:
		api += viper.GetString("qbittorrent.login")
	case AddNewTorrent:
		api += viper.GetString("qbittorrent.torrents.add")
	case Delete:
		api += viper.GetString("qbittorrent.torrents.delete")
	case GetTorrents:
		api += viper.GetString("qbittorrent.torrents.info")
	case PauseTorrent:
		api += viper.GetString("qbittorrent.torrents.pause")
	case ResumeTorrent:
		api += viper.GetString("qbittorrent.torrents.resume")
	case SyncData:
		api += viper.GetString("qbittorrent.sync")
	}
	return api
}

func formingEncodeAPI(action qbtAction) string {
	api := viper.GetString("encode.base")
	switch action {
	case Encode:
		api += viper.GetString("encode.encode")
	case EncodeProgress:
		api += viper.GetString("encode.progress")
	case StopEncode:
		api += viper.GetString("encode.stop")
	}
	return api
}

var updateChan chan model.Sync

func pollingUpdate() {
	c := time.Tick(1500 * time.Millisecond)
	api := formingAPI(SyncData)
	// eAPI := formingEncodeAPI(EncodeProgress)
	threshold := 5
	count := 0
	updateChan = make(chan model.Sync)
	for range c {
		resp, err := client.Get(api)
		if err != nil {
			log.Println(err)
			count++
			if count == threshold {
				log.Fatal("Exceed the request update's error threshold!")
			}
			continue
		}
		count = 0

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)
			return
		}
		var data model.Sync
		if err = json.Unmarshal(body, &data); err != nil {
			log.Println(err)
			continue
		}

		ep := getEncodeProgress(data.Torrents)
		for i, t := range data.Torrents {
			for _, e := range ep {
				if t.Name == e.Name {
					data.Torrents[i].Encode = e
				}
			}
		}

		updateChan <- data
		resp.Body.Close()
	}
}

func getEncodeProgress(torrent []model.Torrent) []model.EncodeProgress {
	var data []model.EncodeProgress
	url, err := url.Parse(formingEncodeAPI(EncodeProgress))

	q := url.Query()
	for _, t := range torrent {
		q.Add("movies", t.Name)
	}
	url.RawQuery = q.Encode()

	res, err := client.Get(url.String())
	if err != nil {
		log.Println(err)
		return data
	}
	if res.StatusCode != http.StatusOK {
		log.Println(res.StatusCode)
		return data
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
		return data
	}
	if err := json.Unmarshal(body, &data); err != nil {
		log.Println(err)
		return data
	}
	return data
}

func loginRequest() {
	api := formingAPI(Login)

	data := map[string]string{
		"username": viper.GetString("username"),
		"password": viper.GetString("password"),
	}
	req, err := newPostRequest(api, data)
	if err != nil {
		log.Println(err)
		return
	}

	attempt := 5
	var res *http.Response
	for i := 0; i < attempt; i++ {
		r, err := client.Do(req)
		if err != nil {
			log.Println(err)
			if i == attempt-1 {
				log.Fatal("Cannot connect to qbittorrent server!")
			}
			time.Sleep(time.Second)
			continue
		}
		res = r
		break
	}

	if res.StatusCode != http.StatusOK {
		log.Fatal("Cannot login to qBittorrent server, response: ", res.StatusCode)
		return
	}
	defer res.Body.Close()

	for _, cookie := range res.Cookies() {
		if cookie.Name == "SID" {
			viper.Set("cookieTorrent", cookie.Value)
		}
	}
	log.Println("Login request to torrent server successfull!")
}

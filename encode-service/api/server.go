package api

import (
	"context"
	"encode-service/encode"
	"encode-service/logger"
	"encode-service/model"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func StartServer() {
	handler := createHandler()

	port := ":80"

	server := &http.Server{
		Addr:         port,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      handler,
	}
	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal("Could not start server: ", err)
		}
	}()
	log.Println("Server start at :", port)

	autoShutdownService()

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGTERM)
	<-c

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
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

func handlerFile() http.Handler {
	return http.FileServer(http.Dir("/downloads"))
}

var router *mux.Router

func matchingRoutes() *mux.Router {
	r := mux.NewRouter().StrictSlash(true)
	r.PathPrefix("/movie").Handler(http.StripPrefix("/movie", handlerFile()))
	r.Handle("/encode/{movie}", wrapHandle(encodeHandle)).Methods(http.MethodGet)
	r.Handle("/progress", wrapHandle(progressHandle))
	r.Handle("/encode/stop/{movie}", wrapHandle(stopEncodeHandle)).Methods(http.MethodGet)
	return r
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

type wrapHandle func(w http.ResponseWriter, r *http.Request) error

func (h wrapHandle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h(w, r)
	if err == nil {
		return
	}

	clientError, ok := err.(model.ClientError)
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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(clientError.Status)
	w.Write(body)
}

func autoShutdownService() {
	c := make(chan bool)
	go func() {
		interval := 1 * time.Second
		for range time.Tick(interval) {
			if encode.TaskMap().Len() > 0 {
				c <- true
			}
		}
	}()
	go func() {
		duration := 15 * time.Second
		timer := time.NewTimer(duration)
		for {
			select {
			case <-timer.C:
				log.Println("Timeout, send shutdown signal!")
				syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
			case <-c:
				if !timer.Stop() {
					<-timer.C
				}
				timer.Reset(duration)
			}
		}
	}()
}

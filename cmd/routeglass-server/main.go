package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/herzadigital/routeglass/internal/config"
	"github.com/herzadigital/routeglass/internal/httpapi"
)

func main() {
	cfg, err := config.Load("/etc/routeglass/config.json")
	if err != nil {
		log.Fatal(err)
	}

	addr := "127.0.0.1:8080"
	srv := &http.Server{Addr: addr, Handler: httpapi.New(cfg)}

	go func() {
		log.Printf("RouteGlass listening on http://%s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c
}

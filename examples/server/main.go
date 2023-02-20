package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/hexastack-dev/devkit-go/server"
)

func handleHello(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello"))
}

func handleCreatePanic(w http.ResponseWriter, r *http.Request) {
	panic(fmt.Errorf("server will recover from panic and log the error"))
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleHello)
	mux.HandleFunc("/panic", handleCreatePanic)
	srv := server.New(mux, nil)

	log.Println("Server started at port 8080")
	srv.ListenAndServe(":8080")
}

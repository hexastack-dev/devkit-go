package main

import (
	"context"
	"fmt"
	stdlog "log"
	"net/http"
	"os"
	"time"

	"github.com/hexastack-dev/devkit-go/log"
	"github.com/hexastack-dev/devkit-go/server"
	"github.com/hexastack-dev/devkit-go/shutdown"
)

func handleHello(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello"))
}

func main() {
	logger := log.NewSimpleLogger(stdlog.Writer(), log.InfoLogLevel)
	log.SetLogger(logger)

	h := http.HandlerFunc(handleHello)

	srv := server.New(h, nil)

	lsn := map[string]shutdown.Listener{
		"server": shutdown.ListenerFunc(shutdown.ListenerFunc(func(ctx context.Context) error {
			// return fmt.Errorf("oopsie")
			return srv.Shutdown(ctx)
		})),
	}
	sh := shutdown.New(10*time.Second, lsn)

	go srv.ListenAndServe(":8080")
	log.Info("Server started at port 8080")
	_, err := sh.Wait()

	if err != nil {
		log.Info(fmt.Sprintf("Shut down didn't exit cleanly: %v", err))
		os.Exit(1)
	}
	log.Info("Shuted down successfuly")
}

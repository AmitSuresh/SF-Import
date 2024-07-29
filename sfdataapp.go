package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

const (
	VERSION        = "0.0.1"
	shutdownTime   = 6 * time.Second
	httpServerAddr = "0.0.0.0:9090"
)

var (
	clientKey     string
	clientSecret  string
	username      string
	password      string
	securityToken string
	instanceURL   string
)

func main() {
	l, _ := zap.NewProduction()
	defer l.Sync()

	sm := mux.NewRouter()

	pR := sm.PathPrefix("/api").Subrouter()
	getR := pR.Methods(http.MethodGet).Subrouter()
	getR.HandleFunc("/getpick", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("test")) })

	httpServer := &http.Server{
		Addr:         httpServerAddr,
		Handler:      sm,
		IdleTimeout:  120 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 4 * time.Second,
		ErrorLog:     zap.NewStdLog(l),
	}

	go func() {
		l.Info("[INFO]", zap.Any("starting http server on: ", httpServerAddr))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			l.Fatal("[ERROR] error starting server", zap.Error(err))
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	l.Info("Shutdown signal received, shutting down servers...")

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTime)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		l.Error("error during http server shutdown", zap.Error(err))
	}

	l.Info("server shutdown complete.")

}

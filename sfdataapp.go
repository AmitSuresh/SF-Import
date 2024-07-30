package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AmitSuresh/sfdataapp/handlers"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

const (
	shutdownTime   = 15 * time.Second
	httpServerAddr = "localhost:9090"
)

var (
	clientID     string
	clientSecret string
	username     string
	instanceURL  string
	sfEnv        string
	keyPath      string
	version      string
)

func main() {
	l, _ := zap.NewProduction()
	defer l.Sync()

	err := godotenv.Load()
	if err != nil {
		l.Error("error loading .env file")
	}

	clientID = os.Getenv("clientID")
	clientSecret = os.Getenv("clientSecret")
	username = os.Getenv("username")
	instanceURL = os.Getenv("instanceURL")
	sfEnv = os.Getenv("sfEnv")
	keyPath = os.Getenv("keyPath")
	version = os.Getenv("version")

	h, err := handlers.GetHandler(clientID, clientSecret, username, instanceURL, version, keyPath, sfEnv, l)
	if err != nil {
		l.Fatal("error creating a new handler", zap.Error(err))
	}

	sm := mux.NewRouter()

	pR := sm.PathPrefix("/api").Subrouter()
	getR := pR.Methods(http.MethodGet).Subrouter()
	getR.HandleFunc("/queryrecords", h.QueryRecords)
	getR.HandleFunc("/querypicklist", h.GetPickBasedMappingRec)

	postR := pR.Methods(http.MethodPost).Subrouter()
	postR.HandleFunc("/insertmappedrecords", h.CreateMappedRecords)
	postR.HandleFunc("/insertbulkmappedrecords", h.CreateBulkMappedRecords)

	httpServer := &http.Server{
		Addr:         httpServerAddr,
		Handler:      sm,
		IdleTimeout:  120 * time.Second,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
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

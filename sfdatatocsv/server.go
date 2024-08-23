package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/AmitSuresh/sfdataapp/sfdatatocsv/handlers"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

var (
	httpServerAddr string
	l              *zap.Logger
	httpServer     *http.Server
	shutdownTime   time.Duration
	cfg            *handlers.Config
)

func init() {

	l, _ = zap.NewProduction()

	err := godotenv.Load()
	if err != nil {
		l.Fatal("error loading .env file")
	}

	httpServerAddr = fmt.Sprintf("%s:%s", os.Getenv("ipAddr"), os.Getenv("port"))
	t, err := strconv.Atoi(os.Getenv("shutdownTime"))
	if err != nil {
		l.Fatal("error converting time")
	}
	shutdownTime = time.Duration(t) * time.Second
	cfg = &handlers.Config{
		JsonDirPath: os.Getenv("csvDirPath"),
	}
	h, err := handlers.GetHandler(l, cfg)
	if err != nil {
		l.Error("error initializing handler", zap.Error(err))
	}

	sm := mux.NewRouter()

	getR := sm.Methods(http.MethodGet).Subrouter()
	getR.HandleFunc("/getmeasurecalcsmap", h.GetMeasureCalcs)
	getR.HandleFunc("/getmclitosearch", h.GetMCLIToSearch)
	getR.HandleFunc("/getmcliquery", h.Getmcliquery)
	getR.HandleFunc("/getallmcli", h.GetMCLI)

	httpServer = &http.Server{
		Addr:         httpServerAddr,
		Handler:      sm,
		IdleTimeout:  120 * time.Second,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		ErrorLog:     zap.NewStdLog(l),
	}

}

func main() {
	defer l.Sync()
	go func() {
		err := httpServer.ListenAndServe()
		if err != nil {
			l.Fatal("error starting server", zap.Error(err))
		}
	}()

	l.Info("server is running at", zap.Any("address:", httpServer.Addr))

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTime)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		l.Error("error during http server shutdown", zap.Error(err))
	}

	l.Info("shutting down server!")
}

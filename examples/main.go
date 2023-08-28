package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	sfimport "github.com/AmitSuresh/SF-Import"
)

var (
	clientKey     string
	clientSecret  string
	username      string
	password      string
	securityToken string
	instanceURL   string
)

func init() {
	flag.StringVar(&clientKey, "k", "", "clientKey")
	flag.StringVar(&clientSecret, "s", "", "clientSecret")
	flag.StringVar(&username, "u", "", "username")
	flag.StringVar(&password, "p", "", "password")
	flag.StringVar(&securityToken, "t", "", "security token")
	flag.StringVar(&instanceURL, "i", "", "instance URL")
	flag.Parse()
}

//	go run main.go -k "3MVG9fe4g9fhX0E55XnC2xUev_f7hmPow7ARpGooVrrPSFl4HwxbGi9ttS0EhFBMcFXeWFlQDbmAhqOPSoJJu" -s "C2B9136CD8964542F64EADC35E6DF833B37195DBE310D7F2635A1C16E7953661" -u "amit_suresh@batch88.com" -p "IAm100%MoreAnnoyed" -t "wDOLX6KJnWUJlVpOlIviZcnvo" -i "https://login.salesforce.com"

func main() {

	sesh, err := sfimport.CreateSession(clientKey, clientSecret, username, password, securityToken, instanceURL)
	if err != nil {
		log.Println("Error initializing connection.")
		return
	}
	err = sesh.InitiateConnection()
	if err != nil {
		fmt.Println("Error opening session: ", err)
	}

	fmt.Println("App is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}

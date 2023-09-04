package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	sfimport "github.com/AmitSuresh/sfdataapp"
)

var (
	clientKey     string
	clientSecret  string
	username      string
	password      string
	securityToken string
	instanceURL   string
	objectAPI     string
)

func init() {
	flag.StringVar(&clientKey, "k", "", "clientKey")
	flag.StringVar(&clientSecret, "s", "", "clientSecret")
	flag.StringVar(&username, "u", "", "username")
	flag.StringVar(&password, "p", "", "password")
	flag.StringVar(&securityToken, "t", "", "security token")
	flag.StringVar(&instanceURL, "i", "", "instance URL")
	flag.StringVar(&objectAPI, "o", "", "object API")
	flag.Parse()
}

func main() {

	sesh, err := sfimport.CreateSession(clientKey, clientSecret, username, password, securityToken, instanceURL)
	if err != nil {
		log.Println("Error creating session.")
		return
	}
	err = sesh.InitiateConnection()
	if err != nil {
		fmt.Println("Error initiating connection: ", err)
	}
	//var fieldMapping sfimport.FieldAPILabelMapping
	fieldMapping, err := sesh.BuildDynamicMapping(objectAPI)
	if err != nil {
		fmt.Println("error retrieving object data and mapping: ", err)
	}
	fmt.Println(fieldMapping)

	fmt.Println("App is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}

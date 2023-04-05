package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/craigcondit/webdav-server/pkg/webdav"
	"gopkg.in/yaml.v3"
)

func main() {
	contentRoot, ok := os.LookupEnv("CONTENT_ROOT")
	if !ok {
		contentRoot = "./sandbox"
	}

	listenAddr, ok := os.LookupEnv("LISTEN_ADDR")
	if !ok {
		listenAddr = ":8080"
	}

	usersFile, ok := os.LookupEnv("USERS_FILE")
	if !ok {
		usersFile = "./conf/users.yaml"
	}
	users := make(map[string]string)
	data, err := os.ReadFile(usersFile)
	if err != nil {
		log.Fatal(err)
	}
	if err = yaml.Unmarshal(data, &users); err != nil {
		log.Fatal(err)
	}

	server := webdav.NewWebDavServer(contentRoot, listenAddr, users)
	server.Start()

	done := make(chan struct{})
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		close(done)
	}()
	<-done
	server.Stop()
}

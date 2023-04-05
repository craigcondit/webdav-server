package webdav

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/tg123/go-htpasswd"
	"golang.org/x/net/webdav"
)

type WebDavServer struct {
	server      *http.Server
	contentRoot string
}

func NewWebDavServer(contentRoot string, listenAddr string, users map[string]string) *WebDavServer {
	handler := &webdav.Handler{
		FileSystem: webdav.Dir(contentRoot),
		LockSystem: webdav.NewMemLS(),
		Logger: func(r *http.Request, err error) {
			remoteAddr := r.RemoteAddr
			if remoteAddr == "" {
				remoteAddr = "-"
			}
			remoteUser := r.Header.Get("X-Remote-User")
			if remoteUser == "" {
				remoteUser = "-"
			}
			log.Default().Printf("%s %s %s %v\n", remoteAddr, remoteUser, r.Method, r.URL)
		},
	}
	htPasswdFile := ""
	for user, pw := range users {
		htPasswdFile = htPasswdFile + fmt.Sprintf("%s:%s\n", user, pw)
	}
	auth, err := htpasswd.NewFromReader(strings.NewReader(htPasswdFile), htpasswd.DefaultSystems, nil)
	if err != nil {
		log.Fatal(err)
	}
	basicAuth := NewBasicAuthenticator(auth, handler)
	mux := http.NewServeMux()
	mux.Handle("/", basicAuth)
	server := &http.Server{
		Addr:              listenAddr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}
	return &WebDavServer{
		server:      server,
		contentRoot: contentRoot,
	}
}

func (s *WebDavServer) Start() {
	log.Default().Printf("Starting WebDAV server on %s using content root '%s'.\n", s.server.Addr, s.contentRoot)
	go func() {
		if err := s.server.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				log.Fatal(err)
			}
		}
	}()
}

func (s *WebDavServer) Stop() {
	s.server.Close()
	if err := s.server.Shutdown(context.Background()); err != nil {
		log.Fatal(err)
	}
}

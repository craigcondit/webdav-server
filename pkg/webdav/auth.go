package webdav

import (
	"encoding/base64"
	"log"
	"net/http"
	"strings"

	"github.com/tg123/go-htpasswd"
)

type BasicAuthenticator struct {
	auth        *htpasswd.File
	nextHandler http.Handler
}

var _ http.Handler = &BasicAuthenticator{}

func NewBasicAuthenticator(auth *htpasswd.File, nextHandler http.Handler) *BasicAuthenticator {
	return &BasicAuthenticator{
		auth:        auth,
		nextHandler: nextHandler,
	}
}

func (b BasicAuthenticator) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	user, ok := b.authenticate(writer, request)
	if !ok {
		return
	}
	request.Header.Add("X-Remote-User", user)
	b.nextHandler.ServeHTTP(writer, request)
}

func (b BasicAuthenticator) authenticate(writer http.ResponseWriter, request *http.Request) (string, bool) {
	// don't authenticate OPTIONS calls
	if strings.ToLower(request.Method) == "options" {
		return "", true
	}

	s := strings.SplitN(request.Header.Get("Authorization"), " ", 2)
	if len(s) != 2 {
		// unauthenticated
		b.challenge(writer, request)
		return "", false
	}
	b64, err := base64.StdEncoding.DecodeString(s[1])
	if err != nil {
		b.challenge(writer, request)
		return "", false
	}
	pair := strings.SplitN(string(b64), ":", 2)
	if len(pair) != 2 {
		b.challenge(writer, request)
		return "", false
	}

	if !b.auth.Match(pair[0], pair[1]) {
		b.challenge(writer, request)
		return "", false
	}

	return pair[0], true
}

func (b BasicAuthenticator) challenge(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("WWW-Authenticate", `Basic realm="webdav"`)
	writer.WriteHeader(401)
	writer.Write([]byte("401 Unauthorized\n"))
	remoteAddr := request.RemoteAddr
	if remoteAddr == "" {
		remoteAddr = "-"
	}
	log.Default().Printf("%s - %s %v\n", remoteAddr, request.Method, request.URL)
}

package main

import (
	"io"
	// "io/ioutil"
	"encoding/hex"
	// "fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"flag"
	"strconv"
	"crypto/sha256"
)

// const PIN = 1234
const (
	defaultPort = 8088
	defaultPortUsage = "default server port, ':80', ':8088'..."
	defaultBackend = "http://127.0.0.1:8080"
	defaultBackendUsage = "default backend url, 'http://127.0.0.1:8080'"

	defaultPIN = "0000"
	defaultPINUsage = "default site PIN, '1234'"
)

func main() {
	port, backendURL, pin := parseArgs()
	log.Println("PIN is", pin)
	pinHash := sha256.Sum256([]byte(pin))

	// backendURL := "http://localdev:80"
	log.Println("Proxy URL:", backendURL)
	proxyURL, _ := url.Parse(backendURL)
	backendProxy := httputil.NewSingleHostReverseProxy(proxyURL)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// if this is a form submission, hash and save the supplied PIN
		if r.Method == http.MethodPost {
			r.ParseForm()
			if formPIN := r.PostFormValue("sitepin"); formPIN != "" {
				formPINhash := sha256.Sum256([]byte(formPIN))
				formPINhashStr := hex.EncodeToString(formPINhash[:])
				newCookie := http.Cookie{
					Name: "sitepin",
					Value: formPINhashStr,
				}
				http.SetCookie(w, &newCookie)
			}
		}

		// check the site PIN cookie if set
		if pinCookie, err := r.Cookie("sitepin"); err == nil {
			pinCookieValue, _ := hex.DecodeString(pinCookie.Value)
			var pinCookieValue32 [32]byte
			copy(pinCookieValue32[:], pinCookieValue)

			if pinCookieValue32 == pinHash {
				// It's a match! Proxy the request
				log.Println("Proxying request to", r.URL.Path)
				backendProxy.ServeHTTP(w, r)
			}
		}

		// serve the login form
		log.Println("Writing login form")
		io.WriteString(w, "foo!")
		return
	})
	// http.HandleFunc("/", backendProxy.ServeHTTP)
	portStr := ":"+strconv.Itoa(port)
	log.Println("Listening to port", portStr)
	http.ListenAndServe(portStr, nil)
	log.Printf("Quitting")
}

func parseArgs() (int, string, string) {
	// command line flags
	port := flag.Int("port", defaultPort, defaultPortUsage)
	url := flag.String("url", defaultBackend, defaultBackendUsage)
	pin := flag.String("pin", defaultPIN, defaultPINUsage)

	flag.Parse()

	return *port, *url, *pin
}

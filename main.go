package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/labstack/echo"
)

// var driver *agouti.WebDriver = nil
var configFile = flag.String("config", "config.toml", "TOML formatted config file path")
var server *echo.Echo = nil

func handleCtrlC(c chan os.Signal) {
	sig := <-c

	// handle ctrl+c event here
	fmt.Println("\nsignal: ", sig)
	if server != nil {
		server.Shutdown(time.Second * 1)
	}
	os.Exit(0)
}

type Config struct {
	ApiKeys []string
	Routes map[string]map[string]Route
}

type Route struct {
	URL string
	User string
	Pass string
}

func main() {
	flag.Parse()

	var config Config
	_, err := toml.DecodeFile(*configFile, &config)
	if err != nil {
		log.Fatalf("Failed to read config file %v.", *configFile)
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill)
	go handleCtrlC(c)

	server := echo.New()

	knownKeys := make(map[string]bool)
	for _, key := range config.ApiKeys {
		knownKeys[key] = true		// enable all keys initially
	}

	// client supplies the following URL-encoded query params:
	//   url (required) - the URL of the target resource
	//   user (optional) - basic auth username to authenticate against <url>
	//   pass (optional) - basic auth password to authenticate against <url>
	server.GET("/:key/get", func(c echo.Context) error {
		key := c.Param("key")
		url := c.QueryParam("url")
		username := c.QueryParam("user")
		password := c.QueryParam("pass")

		if keyEnabled, present := knownKeys[key]; !present || !keyEnabled {
			fmt.Printf("Access key not allowed: %v", key)
			return c.String(http.StatusUnauthorized, "Access key not allowed.")
		}

		// issue GET request
		client := http.Client{}
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Failed to create request.")
		}

		if username != "" && password != "" {
			req.SetBasicAuth(username, password)
		}

		res, err := client.Do(req)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Failed to submit request.")
		}

		// process response
		contentType := res.Header.Get("Content-Type")
		bodyByteSlice, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			return c.String(http.StatusInternalServerError, "Failed to read response from remote host.")
		}
		return c.Blob(res.StatusCode, contentType, bodyByteSlice)
	})

	server.GET("/:key/:route", func(c echo.Context) error {
		key := c.Param("key")
		routeName := c.Param("route")

		if keyEnabled, present := knownKeys[key]; !present || !keyEnabled {
			fmt.Printf("Access key not allowed: %v", key)
			return c.String(http.StatusUnauthorized, "Access key not allowed.")
		}

		routes, present := config.Routes[key]
		if !present {
			fmt.Println("Unknown route.")
			return c.String(http.StatusNotFound, "Unknown route.")
		}

		route, present := routes[routeName]
		if !present {
			fmt.Println("Unknown route.")
			return c.String(http.StatusNotFound, "Unknown route.")
		}

		url := route.URL
		username := route.User
		password := route.Pass

		// issue GET request
		client := http.Client{}
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Failed to create request.")
		}

		if username != "" && password != "" {
			req.SetBasicAuth(username, password)
		}

		res, err := client.Do(req)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Failed to submit request.")
		}

		// process response
		contentType := res.Header.Get("Content-Type")
		bodyByteSlice, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			return c.String(http.StatusInternalServerError, "Failed to read response from remote host.")
		}
		return c.Blob(res.StatusCode, contentType, bodyByteSlice)
	})

	server.Logger.Fatal(server.Start(":4444"))
}

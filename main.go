package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo"
)

// var driver *agouti.WebDriver = nil
var server *echo.Echo = nil

func handleCtrlC(c chan os.Signal) {
	sig := <-c

	// handle ctrl+c event here
	fmt.Println("\nsignal: ", sig)
	// if driver != nil {
	// 	driver.Stop()
	// }
	if server != nil {
		server.Shutdown(time.Second * 1)
	}
	os.Exit(0)
}

func main() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill)
	go handleCtrlC(c)

	//driver = agouti.Selenium()
	//capabilities := agouti.NewCapabilities().Browser("firefox")
	//page, err := driver.NewPage(agouti.Desired(capabilities))
	// driver = agouti.Selenium()
	// capabilities := agouti.NewCapabilities().Browser("firefox")
	// page, err := driver.NewPage(agouti.Desired(capabilities))
	// if err != nil {
	// 	fmt.Errorf("Unable to ")
	// }
	// page.Navigate("http://www.drudgereport.com")
	// page.Screenshot("~/Desktop/drudge.jpg")

	server := echo.New()

	knownKeys := make(map[string]bool)
	knownKeys["55555"] = true		// key 55555 is enabled

	// client supplies the following URL-encoded query params:
	//   url (required) - the URL of the target resource
	//   user (optional) - basic auth username to authenticate against <url>
	//   pass (optional) - basic auth password to authenticate against <url>
	server.GET("/raw/:key", func(c echo.Context) error {
		key := c.Param("key")
		url := c.QueryParam("url")
		username := c.QueryParam("user")
		password := c.QueryParam("pass")

		if keyEnabled, present := knownKeys[key]; !present || !keyEnabled {
			fmt.Println("Access key not allowed: %v", key)
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

	server.Logger.Fatal(server.Start(":4444"))
}

package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strconv"
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

	knownKeys := make(map[int]bool)
	knownKeys[55555] = true		// key 55555 is enabled

	// client supplies the following URL-encoded query params:
	//   url (required) - the URL of the target resource
	//   user (optional) - basic auth username to authenticate against <url>
	//   pass (optional) - basic auth password to authenticate against <url>
	server.GET("/raw/:key", func(c echo.Context) error {
		key, err := strconv.Atoi(c.Param("key"))
		if err != nil {
			fmt.Println("Access key could not be converted to int: %v", c.Param("key"))
			return c.String(http.StatusBadRequest, "Access key malformed.")
		}

		url := c.QueryParam("url")
		// username := c.QueryParam("user")
		// password := c.QueryParam("pass")

		if keyEnabled, present := knownKeys[key]; !present || !keyEnabled {
			fmt.Println("Access key not allowed: %v", key)
			return c.String(http.StatusUnauthorized, "Access key not allowed.")
		}

		res, err := http.Get(url)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Request failed.")
		}

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

// // taken from https://github.com/labstack/echo/blob/master/middleware/proxy.go
// func respondWithResponse(c echo.Context, in *http.Response) (err error) {
// 	// return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		out, _, err := c.Response().Hijack()
// 		if err != nil {
// 			c.Error(fmt.Errorf("proxy raw, hijack error=%v, url=%s", t.URL, err))
// 			return
// 		}
// 		defer out.Close()
//
// 		// Write header
// 		err = r.Write(out)
// 		if err != nil {
// 			he := echo.NewHTTPError(http.StatusBadGateway, fmt.Sprintf("proxy raw, request header copy error=%v, url=%s", t.URL, err))
// 			c.Error(he)
// 			return
// 		}
//
// 		errc := make(chan error, 2)
// 		cp := func(dst io.Writer, src io.Reader) {
// 			_, err := io.Copy(dst, src)
// 			errc <- err
// 		}
//
// 		go cp(out, in)
// 		go cp(in, out)
// 		err = <-errc
// 		if err != nil && err != io.EOF {
// 			c.Logger().Errorf("proxy raw, copy body error=%v, url=%s", t.URL, err)
// 		}
// 	// })
// 	return
// }

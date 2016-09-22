package main

import (
	"os"
	"net/http"
	"github.com/urfave/cli"
	"io/ioutil"
	"fmt"
	"time"
	"strings"
	"errors"
)


// Function will return true, if word exist in source code
func search_needed_words(response string, search string) bool {
	if strings.Contains(response, search) {
		return true
	} else {
		return false
	}
}


// Function will return true, if word does not exist in source code
func search_stop_words(response string, search string) bool {
	if strings.Contains(response, search) {
		return false
	} else {
		return true
	}
}

//Exception handler
func handle_exception(exception_str string, predict_str string) string {
	if strings.Contains(exception_str, predict_str) {
		return "handled"
	} else {
		return "unhandled"
	}
}


//Implemented full logic
func check_site(c *cli.Context) error {
	url := c.String("url")
	timeout := time.Duration(c.Int64("timeout")) * time.Second
	http_client := http.Client{
		Timeout: timeout,
	}
	start := time.Now()
	resp, err := http_client.Get(url)
	if err != nil {
		switch {
		case "handled" == handle_exception(err.Error(), "no such host"):
			return cli.NewExitError("No such host! Review your url.", 3)
		case "handled" == handle_exception(err.Error(), "connection refused"):
			return cli.NewExitError("Connection refused! Please try later.", 3)
		case "handled" == handle_exception(err.Error(), "Timeout exceeded"):
			return cli.NewExitError("Timeout exceeded! Connection aborted.", 3)
		default:
			return cli.NewExitError("Cannot create http request" + err.Error(), 3)
		}
	}

	if resp.StatusCode != 200 {
		fmt.Println("WARNING: Status code is", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	elapsed := time.Since(start)
	if err != nil {
		return errors.New("Cannot read body")
	}
	resp.Body.Close()

	if search_needed_words(string(body), c.String("search-word")) != true {
		error_string := "DISASTER: Site " + c.String("url") + " " + c.String("search-word") + " not in source!"
		return cli.NewExitError(error_string, 3)
	}


	if search_stop_words(string(body), c.String("stop-word")) != true {
		error_string := "DISASTER: Site " + c.String("url") + " " + c.String("stop-word") + " finding stop word!"
		return cli.NewExitError(error_string, 3)
	}

	fmt.Printf("OK: Site is %s reacheble for %s\n", c.String("url"), elapsed)
	return nil
}


// Implemented commandline application
// Usage: ./site_checker --url=http://unplag.com --timeout=1 --search-word=unplag --stop-word=exception
// Usage: ./site_checker --url=http://cf.ua --timeout=1 --search-word=cf --stop-word=exception
func main() {
	app := cli.NewApp()
	app.Name = "Mega Super Site Cheker"
	app.Version = "0.0.1"
	app.Usage = "check your site for availability!"

	app.Flags = []cli.Flag {
		cli.StringFlag{
			Name: "url, u",
			Value: "https://unplag.com",
			Usage: "url to check",
		},
		cli.StringFlag{
			Name: "timeout, t",
			Value: "5",
			Usage: "timeout for check",
		},
		cli.StringFlag{
			Name: "search-word, search",
			Value: "unplag",
			Usage: "word for searching in source",
		},
		cli.StringFlag{
			Name: "stop-word, stop",
			Value: "exception",
			Usage: "word in source that will not be present",
		},
	}

	app.Action = func(c *cli.Context) error {
		err := check_site(c)
		if err != nil {
			return cli.NewExitError(err.Error(), 3)
		}
		return nil
	}

	app.Run(os.Args)
}
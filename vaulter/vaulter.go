package main

import (
	"os"
	"fmt"
	"net/http"
	"errors"
	"strconv"
	"io/ioutil"
	"regexp"
	"github.com/urfave/cli"
	"github.com/fatih/color"
	"github.com/Jeffail/gabs"
)


var result_slice []string


func status_code_handler(status_code int) error {
	switch status_code  {
	case 200:
		return nil
	case 204:
		return nil
	case 400:
		return cli.NewExitError(color.RedString("Response: Invalid request"), 3)
	case 403:
		return cli.NewExitError(color.RedString("Response: Forbidden"), 3)
	case 404:
		return cli.NewExitError(color.RedString("Response: Invalid path"), 3)
	case 429:
		return cli.NewExitError(color.RedString("Response: Rate limit exceeded"), 3)
	case 500:
		return cli.NewExitError(color.RedString("Response: Internal server error"), 3)
	case 503:
		return cli.NewExitError(color.RedString("Response: Vault is down"), 3)
	}

	return errors.New("Response: Unknown status code " + strconv.Itoa(status_code))
}


func read_http_api(url, token string) (data []byte, err error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)

	//fmt.Println(color.YellowString(url))

	req.Header.Set("X-Vault-Token", token)
	res, req_err := client.Do(req)

	if req_err != nil {
		return []byte("0"), cli.NewExitError(color.RedString(req_err.Error()), 3)
	}

	resp_err := status_code_handler(res.StatusCode)

	if resp_err != nil {
		return []byte("0"), cli.NewExitError(resp_err.Error(), 3)
	}

	var body_data []byte


	body_data, _ = ioutil.ReadAll(res.Body)

	defer res.Body.Close()

	return body_data, nil
}


func verify_connection(c *cli.Context) error {
	vault_addr := c.String("vault_addr")
	vault_token := c.String("vault_token")

	body_data, err := read_http_api(vault_addr + "/v1/auth/token/lookup-self", vault_token)

	if err != nil {
		return cli.NewExitError(color.RedString(err.Error()), 3)
	}

	json_parsed, _ := gabs.ParseJSON(body_data)

	request_id, _ := json_parsed.Path("request_id").Data().(string)
	token, _ := json_parsed.Path("data.id").Data().(string)
	path, _ := json_parsed.Path("data.policies").Children()

	fmt.Println(color.GreenString("Lookup url: " + vault_addr + "/v1/auth/token/lookup-self"))

	fmt.Println(color.GreenString("Request id: " + request_id))
	fmt.Println(color.GreenString("Request token: " + token))

	for _, child := range path {
		fmt.Println(color.GreenString("Policies: " + child.Data().(string)))
	}

	return nil
}


func read_path(c *cli.Context, path string) error {
	vault_addr := c.String("vault_addr")
	vault_token := c.String("vault_token")

	body_data, err := read_http_api(vault_addr + "/v1/" + path, vault_token)

	if err != nil {
		return cli.NewExitError(color.RedString(err.Error()), 3)
	}

	json_parsed, _ := gabs.ParseJSON(body_data)

	resp, _ := json_parsed.Search("data").ChildrenMap()
	for key, child := range resp {
		fmt.Println(color.BlueString(key) + ":" ,child.Data().(string))
	}

	return nil
}


func make_tree(c *cli.Context, path string, is_silent bool)  (err error, return_path []string) {
	vault_addr := c.String("vault_addr")
	vault_token := c.String("vault_token")

	body_data, err := read_http_api(vault_addr + "/v1/" + path  + "?list=true", vault_token)

	if err != nil {
		return cli.NewExitError(color.RedString(err.Error()), 3), nil
	}

	json_parsed, _ := gabs.ParseJSON(body_data)

	path_obj, _ := json_parsed.Path("data.keys").Children()


	for _, child := range path_obj {
		//fmt.Println(color.GreenString("Keys: " + child.Data().(string)))
		new_path := path + child.Data().(string)
		if string(new_path[len(new_path)-1:]) != "/" {
			//make_tree(c, path + "/" + child.Data().(string))
			//fmt.Println(string(path[len(path)-1:]))
			result_slice = append(result_slice, new_path)
			if is_silent == false {
				fmt.Println(color.BlueString(new_path))
			}
		} else {
			//fmt.Println("Folder: " + new_path + " ---- " + string(path[len(path)-1:]))
			make_tree(c, path + child.Data().(string), is_silent)
			//fmt.Println(string(path[len(path)-1:]))
		}

	}

	return nil, result_slice
}


func make_search(c *cli.Context, backend string, regular_exp string)  (err error, return_path []string) {
	err, path_slice := make_tree(c, backend, true)

	for _, link := range path_slice {
		ok, err := regexp.MatchString(regular_exp, link)
		if ok {
			fmt.Println(color.YellowString("\n" + link))
			read_path(c, link)
		} else if err != nil {
			return cli.NewExitError(color.RedString(err.Error()), 3), nil
		}
	}

	return nil, nil
}


func main() {
	app := cli.NewApp()

	app.Name = "Vaulter"
	app.Version = "0.0.1"
	app.Usage = "friendliest console client for vault!"

	app.Commands = []cli.Command{
		{
			Name:    "verify",
			Aliases: []string{"v"},
			Usage:   "verify connection to vault",
			Action:  func(c *cli.Context) error {
				err := verify_connection(c)
				if err != nil {
					return cli.NewExitError(err.Error(), 3)
				}
				return nil
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name: "vault_addr, addr, a",
					Value: "http://localhost:8200",
					Usage: "connection for vault",
					EnvVar: "VAULT_ADDR",
				}, cli.StringFlag{
					Name: "vault_token, token, t",
					Usage: "vault token",
					EnvVar: "VAULT_TOKEN",
				},
			},
		},
		{
			Name:        "search",
			Aliases:     []string{"t"},
			Usage:       "search for regex pattern",
			Action:  func(c *cli.Context) error {
					backend := c.Args().Get(0)
					regular_exp := c.Args().Get(1)

					if backend == "" {
						return cli.NewExitError(color.RedString("Please, enter vault backend"), 3)
					}

					if regular_exp == "" {
						return cli.NewExitError(color.RedString("Please, enter regular expression"), 3)
					}

					if backend[len(backend)-1:] != "/" {
						backend += "/"
					}
					err, _ := make_search(c, backend, regular_exp)
					if err != nil {
						return cli.NewExitError(err.Error(), 3)
					}
					return nil
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name: "vault_addr, addr, a",
					Value: "http://localhost:8200",
					Usage: "connection for vault",
					EnvVar: "VAULT_ADDR",
				}, cli.StringFlag{
					Name: "vault_token, token, t",
					Usage: "vault token",
					EnvVar: "VAULT_TOKEN",
				},
			},
		}, {
			Name:        "read",
			Aliases:     []string{"t"},
			Usage:       "read from location",
			Action:  func(c *cli.Context) error {
					err := read_path(c, c.Args().Get(0))
					if err != nil {
						return cli.NewExitError(err.Error(), 3)
					}
					return nil
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name: "vault_addr, addr, a",
					Value: "http://localhost:8200",
					Usage: "connection for vault",
					EnvVar: "VAULT_ADDR",
				}, cli.StringFlag{
					Name: "vault_token, token, t",
					Usage: "vault token",
					EnvVar: "VAULT_TOKEN",
				},
			},
		}, {
			Name:        "tree",
			Aliases:     []string{"t"},
			Usage:       "show all vault secrets",
			Action:  func(c *cli.Context) error {
					path := c.Args().Get(0)

					if path == "" {
						return cli.NewExitError(color.RedString("Please, enter vault path"), 3)
					}

					if  path[len(path)-1:] != "/" {
						path += "/"
					}
					err, _ := make_tree(c, path, false)
					if err != nil {
						return cli.NewExitError(err.Error(), 3)
					}
					return nil
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name: "vault_addr, addr, a",
					Value: "http://localhost:8200",
					Usage: "connection for vault",
					EnvVar: "VAULT_ADDR",
				}, cli.StringFlag{
					Name: "vault_token, token, t",
					Usage: "vault token",
					EnvVar: "VAULT_TOKEN",
				},
			},
		},
	}

	app.Run(os.Args)
}

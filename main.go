package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/shunirr/extract-chrome-storage/core"

	"github.com/urfave/cli"
)

func getChromeDir() string {
	return os.Getenv("HOME") + "/Library/Application Support/Google/Chrome/Default"
}

func getCookiesPath(basePath string) string {
	return basePath + "/Cookies"
}

func getLocalStoragePath(basePath string) string {
	return basePath + "/Local Storage/leveldb"
}

type AppTypeEnum int

const (
	AppStore AppTypeEnum = iota
	NonAppStore
	Unknown
)

func getSlackDir(appType AppTypeEnum) string {
	slackDir := "/Library/Application Support/Slack/"
	if appType == AppStore {
		return os.Getenv("HOME") + "/Library/Containers/com.tinyspeck.slackmacgap/Data" + slackDir
	}
	return os.Getenv("HOME") + slackDir
}

func convertToJson(data map[string]string) (string, error) {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

func printChromeCookies(host string, cookiesPath *string) {
	if cookiesPath == nil {
		chromeDir := getChromeDir()
		path := getCookiesPath(chromeDir)
		cookiesPath = &path
	}
	cookies, err := core.GetChromeCookies(
		*cookiesPath,
		host,
		"Chrome",
	)
	if err != nil {
		panic(err)
	}

	json, err := convertToJson(cookies)
	if err != nil {
		panic(err)
	}

	fmt.Println(json)
}

func printSlackCookies(appType AppTypeEnum, cookiesPath *string) {
	var account string
	if appType == AppStore {
		account = "Slack App Store Key"
	} else {
		account = "Slack Key"
	}
	if cookiesPath == nil {
		chromeDir := getSlackDir(appType)
		path := getCookiesPath(chromeDir)
		cookiesPath = &path
	}
	cookies, err := core.GetChromeCookies(
		*cookiesPath,
		".slack.com",
		account,
	)
	if err != nil {
		panic(err)
	}

	json, err := convertToJson(cookies)
	if err != nil {
		panic(err)
	}

	fmt.Println(json)
}

func printChromeLocalStorage(host string, key string, levelDbPath *string) {
	if levelDbPath == nil {
		chromeDir := getChromeDir()
		path := getLocalStoragePath(chromeDir)
		levelDbPath = &path
	}

	storage, err := core.GetChromeLocalStorage(
		*levelDbPath,
		host,
		key,
	)
	if err != nil {
		panic(err)
	}
	fmt.Println(storage)
}

func printSlackLocalStorage(appType AppTypeEnum, host string, key string, levelDbPath *string) {
	if levelDbPath == nil {
		chromeDir := getSlackDir(appType)
		path := getLocalStoragePath(chromeDir)
		levelDbPath = &path
	}

	storage, err := core.GetChromeLocalStorage(
		*levelDbPath,
		host,
		key,
	)
	if err != nil {
		panic(err)
	}
	fmt.Println(storage)
}

func main() {
	app := &cli.App{
		Name:  "extract-chrome-storage",
		Usage: "This is a CLI tool to extract Chrome storage, such as cookies and local storage.",
		Commands: []cli.Command{
			{
				Name:  "cookie",
				Usage: "Extract cookies",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "browser",
						Usage: "'chrome' or 'slack'",
						Value: "chrome",
					},
					&cli.BoolFlag{
						Name:  "app-store",
						Usage: "If you installed browser app from App Store",
					},
					&cli.StringFlag{
						Name:     "domain",
						Usage:    "Domain of the cookie (e.g., '.example.com')",
						Required: true,
					},
					&cli.StringFlag{
						Name:  "cookies-path",
						Usage: "Path of the Cookies' SQLite (e.g., '/Users/USERNAME/Library/Application Support/Google/Chrome/Default/Cookies')",
					},
				},
				Action: func(c *cli.Context) error {
					browserType := strings.ToLower(c.String("browser"))
					domain := c.String("domain")
					isAppStore := c.Bool("app-store")
					cookiesPath := c.String("cookies-path")

					switch browserType {
					case "chrome":
						printChromeCookies(domain, &cookiesPath)
					case "slack":
						if cookiesPath != "" {
							printSlackCookies(Unknown, &cookiesPath)
							return nil
						}
						if isAppStore {
							printSlackCookies(AppStore, nil)
						} else {
							printSlackCookies(NonAppStore, nil)
						}
					}
					return nil
				},
			},
			{
				Name:  "local-storage",
				Usage: "Extract local storage",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "browser",
						Usage: "'chrome' or 'slack'",
						Value: "chrome",
					},
					&cli.BoolFlag{
						Name:  "app-store",
						Usage: "If you installed the browser app from the App Store",
					},
					&cli.BoolFlag{
						Name:  "http",
						Usage: "If you use an HTTP scheme for the origin",
					},
					&cli.StringFlag{
						Name:     "domain",
						Usage:    "Domain of the local storage (e.g., 'www.example.com')",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "key",
						Usage:    "Key name of the local storage (e.g., 'AppConfig')",
						Required: true,
					},
					&cli.StringFlag{
						Name:  "leveldb-path",
						Usage: "Path of the local storage's leveldb (e.g., '/Users/USERNAME/Library/Application Support/Google/Chrome/Default/Local Storage/leveldb')",
					},
				},
				Action: func(c *cli.Context) error {
					browserType := strings.ToLower(c.String("browser"))
					isAppStore := c.Bool("app-store")
					domain := c.String("domain")
					http := c.Bool("http")
					key := c.String("key")
					leveldbPath := c.String("leveldb-path")

					var hostWithScheme string
					if http {
						hostWithScheme = "http://" + domain
					} else {
						hostWithScheme = "https://" + domain
					}

					switch browserType {
					case "chrome":
						if leveldbPath != "" {
							printChromeLocalStorage(hostWithScheme, key, &leveldbPath)
						} else {
							printChromeLocalStorage(hostWithScheme, key, nil)
						}
					case "slack":
						if leveldbPath != "" {
							printSlackLocalStorage(AppStore, hostWithScheme, key, &leveldbPath)
							return nil
						}
						if isAppStore {
							printSlackLocalStorage(AppStore, hostWithScheme, key, nil)
						} else {
							printSlackLocalStorage(NonAppStore, hostWithScheme, key, nil)
						}
					}
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

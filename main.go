package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	gh "github.com/crosbymichael/octokat"
)

var (
	token  string
	logger = logrus.New()
)

// Load the users github token from
// $HOME/.github
func loadToken() (string, error) {
	path := filepath.Join(os.Getenv("HOME"), ".github")
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func newFile(r io.Reader) gh.File {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		logger.Fatal(err)
	}
	return gh.File{
		Content: string(data),
	}
}

func getFilesFromContext(context *cli.Context) map[string]gh.File {
	if len(context.Args()) == 0 {
		return map[string]gh.File{
			"stdin": newFile(os.Stdin),
		}
	}
	files := make(map[string]gh.File)
	for _, file := range context.Args() {
		f, err := os.Open(file)
		if err != nil {
			logger.Fatal(err)
		}
		name := filepath.Base(file)
		files[name] = newFile(f)
		f.Close()
	}
	return files
}

func main() {
	app := cli.NewApp()
	app.Name = "gistit"
	app.Usage = "post content to gist.github.com"
	app.Version = "2"
	app.Author = "@crosbymichael"
	app.Flags = []cli.Flag{
		cli.BoolFlag{Name: "secret,s", Usage: "set the gist to secret"},
		cli.StringFlag{Name: "description,d", Usage: "set the gist's description"},
		cli.StringFlag{Name: "token", Usage: "use the specified github access token"},
	}

	app.Before = func(context *cli.Context) (err error) {
		token = context.GlobalString("token")
		if token == "" {
			if token, err = loadToken(); err != nil {
				return err
			}
		}
		return nil
	}

	app.Action = func(context *cli.Context) {
		client := gh.NewClient()
		client.WithToken(token)

		files := getFilesFromContext(context)
		gist, err := client.CreateGist(
			context.GlobalString("description"),
			!context.GlobalBool("secret"), files)
		if err != nil {
			logger.Fatal(err)
		}
		fmt.Println(gist.HtmlUrl)
	}

	if err := app.Run(os.Args); err != nil {
		logger.Fatal(err)
	}
}

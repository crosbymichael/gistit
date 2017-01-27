package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/Sirupsen/logrus"
	gh "github.com/crosbymichael/octokat"
	"github.com/urfave/cli"
)

// Load the users github token from
// $HOME/.github
func loadToken() string {
	path := filepath.Join(os.Getenv("HOME"), ".github")
	b, err := ioutil.ReadFile(path)
	if err != nil {
		logrus.WithError(err).Fatal("load token")
	}
	return strings.Trim(string(b), "\n")
}

func newFile(r io.Reader) gh.File {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		logrus.Fatal(err)
	}
	return gh.File{
		Content: string(data),
	}
}

func getFilesFromContext(context *cli.Context) (map[string]gh.File, error) {
	if len(context.Args()) == 0 {
		return map[string]gh.File{
			"stdin": newFile(os.Stdin),
		}, nil
	}
	files := make(map[string]gh.File)
	for _, file := range context.Args() {
		f, err := os.Open(file)
		if err != nil {
			return nil, err
		}
		name := filepath.Base(file)
		files[name] = newFile(f)
		f.Close()
	}
	return files, nil
}

func main() {
	app := cli.NewApp()
	app.Name = "gistit"
	app.Usage = "post content to gist.github.com"
	app.Version = "3"
	app.Author = "@crosbymichael"
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "secret,s",
			Usage: "set the gist to secret",
		},
		cli.StringFlag{
			Name:  "description,d",
			Usage: "set the gist's description",
		},
		cli.StringFlag{
			Name:  "token",
			Usage: "use the specified github access token",
			Value: loadToken(),
		},
	}
	app.Action = func(context *cli.Context) error {
		client := gh.NewClient()
		client.WithToken(context.String("token"))

		files, err := getFilesFromContext(context)
		if err != nil {
			return err
		}
		gist, err := client.CreateGist(
			context.GlobalString("description"),
			!context.GlobalBool("secret"), files)
		if err != nil {
			return err
		}
		fmt.Println(gist.HtmlUrl)
		return err
	}
	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}

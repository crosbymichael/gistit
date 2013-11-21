package main

import (
	"flag"
	"fmt"
	gh "github.com/crosbymichael/octokat"
	"io/ioutil"
	"os"
	"path"
)

var (
	public      bool   // Is the gist public
	description string // What is the description for the gist
)

func init() {
	flag.BoolVar(&public, "p", true, "Create a public gist")
	flag.StringVar(&description, "m", "", "Description for the gist")
	flag.Parse()
}

// Load the users github token from
// $HOME/.github
func loadToken() (string, error) {
	p := path.Join(os.Getenv("HOME"), ".github")

	b, err := ioutil.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("Could not load your github token from %s", p)
		}
		return "", err
	}
	return string(b), nil
}

// Load each file passed and get the contents
func loadFiles() (map[string]gh.File, error) {
	var (
		name, p string
		m       = make(map[string]gh.File)
	)
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	for i := 0; i < flag.NArg(); i++ {
		name = flag.Arg(i)
		p = path.Join(cwd, name)

		contents, err := ioutil.ReadFile(p)
		if err != nil {
			return nil, err
		}
		m[name] = gh.File{Content: string(contents)}
	}
	return m, nil
}

func readFromStdin() (map[string]gh.File, error) {
	contents, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return nil, err
	}
	s := string(contents)

	if s == "" {
		return nil, fmt.Errorf("No content read from stdin")
	}

	return map[string]gh.File{
		"stdinfile": {Content: s},
	}, nil
}

func writeError(err error) {
	fmt.Fprintf(os.Stderr, "%s\n", err)
	os.Exit(1)
}

func main() {
	var (
		err   error
		files map[string]gh.File
		gist  gh.Gist
	)
	token, err := loadToken()
	if err != nil {
		writeError(err)
	}

	// Check if the user is passing one or more
	// files to the gist
	if flag.NArg() > 0 {
		files, err = loadFiles()
	} else { // Read from stdin to get content
		files, err = readFromStdin()
	}
	if err != nil {
		writeError(err)
	}

	// Create a new github client and load the token
	client := gh.NewClient()
	client.WithToken(token)

	if gist, err = client.CreateGist(description, public, files); err != nil {
		writeError(err)
	}
	// Print out the url to the gist
	fmt.Printf("%s\n", gist.HtmlUrl)
}

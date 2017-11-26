package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/BurntSushi/toml"
)

type project struct {
	Name     string
	Revision string
}

func (p project) url() (string, error) {
	projectURL, err := url.Parse(fmt.Sprintf("https://%s", p.Name))
	if err != nil {
		return "", err
	}
	if strings.ToLower(projectURL.Hostname()) == "github.com" {
		return projectURL.String() + ".git", nil
	}
	q := projectURL.Query()
	q.Set("go-get", "1")
	projectURL.RawQuery = q.Encode()
	req, err := http.NewRequest(http.MethodGet, projectURL.String(), nil)
	if err != nil {
		return "", err
	}
	client := new(http.Client)
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Unable to get VCS: %s", res.Status)
	}
	return parseHTML(res.Body)
}

type projects struct {
	Projects []project
}

func main() {
	var dependencies projects
	if _, err := toml.DecodeFile("Gopkg.lock", &dependencies); err != nil {
		log.Fatal(err)
	}
	for _, dep := range dependencies.Projects {
		projectURL, err := dep.url()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("  go_resource \"%s\" do\n", dep.Name)
		fmt.Printf("    url \"%s\",\n", projectURL)
		fmt.Printf("      :revision => \"%s\"\n", dep.Revision)
		fmt.Printf("  end\n")
		fmt.Println()
	}
}

func parseHTML(r io.Reader) (string, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return "", err
	}
	imp := findGoImport(doc)
	if imp == "" {
		return "", fmt.Errorf("Could not find go-import meta tag")
	}
	return imp + ".git", nil
}

func findGoImport(node *html.Node) string {
	if node.Type == html.ElementNode && node.DataAtom == atom.Meta {
		importEl := false
		for _, att := range node.Attr {
			if att.Key == "name" && att.Val == "go-import" {
				importEl = true
			}
		}
		if importEl {
			for _, att := range node.Attr {
				if att.Key == "content" {
					content := strings.Split(att.Val, " ")
					return content[len(content)-1]
				}
			}
		}
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		imp := findGoImport(c)
		if imp != "" {
			return imp
		}
	}
	return ""
}

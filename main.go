package main

import (
	"fmt"
	"log"

	"github.com/BurntSushi/toml"
)

type project struct {
	Name     string
	Revision string
}

func (p project) url() string {
	return fmt.Sprintf("https://%s.git", p.Name)
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
		fmt.Printf("  go_resource \"%s\" do\n", dep.Name)
		fmt.Printf("    url \"%s\",\n", dep.url())
		fmt.Printf("      :revision => \"%s\"\n", dep.Revision)
		fmt.Printf("  end\n")
		fmt.Println()
	}
}

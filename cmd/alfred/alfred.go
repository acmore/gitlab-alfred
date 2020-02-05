package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/acmore/gitlab-alfred/pkg/commands"
	"github.com/acmore/gitlab-alfred/pkg/provider"
	aw "github.com/deanishe/awgo"
	"github.com/joho/godotenv"
)

type Config struct {
	Token string `env:"GITLAB_TOKEN"`
	URL   string `env:"GITLAB_URL"`
}

var (
	wf     *aw.Workflow
	client provider.Provider
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Printf("failed to load .env, error: %v", err)
	}

	wf = aw.New()
}

func run() {
	wf.Args()
	flag.Parse()
	if flag.NArg() < 2 {
		panic(fmt.Errorf("invalid arguments %d", flag.NArg()))
	}
	args := flag.Args()

	cfg := &Config{}
	if err := wf.Config.To(cfg); err != nil {
		panic(err)
	}

	// Create a provider
	client = provider.NewGitlabProvider(cfg.URL, cfg.Token)

	cmd, subcmd := args[0], args[1]
	switch cmd {
	case "project":
		projCmd := commands.NewProjectCommand(wf, client)
		switch subcmd {
		case "list":
			projCmd.List(args)
		}
	}

	log.Printf("loaded %v", cfg)
}

func main() {
	wf.Run(run)
}

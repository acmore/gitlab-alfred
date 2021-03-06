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

func setKey(key, value string) {
	wf.Configure(aw.TextErrors(true))

	if err := wf.Config.Set(key, value, true).Do(); err != nil {
		wf.FatalError(err)
	}
}

func run() {
	wf.Args()
	flag.Parse()
	if flag.NArg() < 1 {
		panic(fmt.Errorf("invalid arguments %d", flag.NArg()))
	}
	args := flag.Args()

	cfg := &Config{}
	if err := wf.Config.To(cfg); err != nil {
		panic(err)
	}

	// Create a provider
	client = provider.NewGitlabProvider(cfg.URL, cfg.Token)

	cmd := args[0]
	args = args[1:]
	switch cmd {
	case "project":
		projCmd := commands.NewProjectCommand(wf, client)
		projCmd.Run(args)
	case "pipeline":
		pipelineCmd := commands.NewPipelineCommand(wf, client)
		pipelineCmd.Run(args)
	case "branch":
		branchCmd := commands.NewBranchCommand(wf, client)
		branchCmd.Run(args)
	case "issue":
		issueCmd := commands.NewIssueCommand(wf, client)
		issueCmd.Run(args)
	case "config":
		if len(args) < 2 {
			log.Println("missing key value")
			break
		}
		setKey(args[0], args[1])
	}
}

func main() {
	wf.Run(run)
}

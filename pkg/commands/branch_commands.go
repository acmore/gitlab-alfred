package commands

import (
	"fmt"
	"log"
	"time"

	"github.com/acmore/gitlab-alfred/pkg/provider"
	aw "github.com/deanishe/awgo"
	flag "github.com/spf13/pflag"
)

const (
	CacheKeyBranchFormat = "branches-%s"
)

type BranchCommand struct {
	wf     *aw.Workflow
	client provider.Provider
}

func NewBranchCommand(wf *aw.Workflow, client provider.Provider) *BranchCommand {
	return &BranchCommand{wf: wf, client: client}
}

func (c *BranchCommand) Run(args []string) {
	defer c.wf.SendFeedback()

	log.Printf("%v", args)

	var projectID, query string

	flagSet := flag.NewFlagSet("branch", flag.PanicOnError)
	flagSet.StringVar(&projectID, "project-id", "", "Project ID")
	flagSet.StringVar(&query, "query", "", "Query")

	flagSet.Parse(args)

	subcmd := "list"
	if flagSet.NArg() > 0 {
		subcmd = flagSet.Arg(0)
	}

	switch subcmd {
	case "list":
		c.List(projectID, query)
	}
}

func (c *BranchCommand) List(projectID string, query string) {
	log.Printf("List branch %s, %s", projectID, query)

	reload := func() (interface{}, error) {
		var branches []*provider.Branch

		pageSize := 100
		for page := 0; ; page++ {
			res, err := c.client.ListBranches(projectID, page, pageSize)
			if err != nil {
				return nil, err
			}
			branches = append(branches, res...)
			if len(res) < pageSize {
				break
			}
		}
		return branches, nil
	}

	var branches []*provider.Branch
	if err := c.wf.Cache.LoadOrStoreJSON(fmt.Sprintf(CacheKeyBranchFormat, projectID), 10*time.Second, reload, &branches); err != nil {
		panic(err)
	}

	for _, p := range branches {
		item := c.wf.Feedback.NewItem(p.Name).Var("branch_name", p.Name).Valid(true)
		if p.Merged {
			item.Subtitle(fmt.Sprintf("%s merged", p.Commit.Title))
		} else {
			item.Subtitle(p.Commit.Title)
		}
	}

	if len(query) > 0 {
		c.wf.Filter(query)
	}

	c.wf.WarnEmpty("Empty", "No Pipelines")
}

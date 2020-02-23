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
	CacheKeyIssuesFormat = "branches-%s"
)

type IssueCommand struct {
	wf     *aw.Workflow
	client provider.Provider
}

func NewIssueCommand(wf *aw.Workflow, client provider.Provider) *IssueCommand {
	return &IssueCommand{wf: wf, client: client}
}

func (c *IssueCommand) Run(args []string) {
	defer c.wf.SendFeedback()

	log.Printf("%v", args)

	var projectID, query string

	flagSet := flag.NewFlagSet("issue", flag.PanicOnError)
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

func (c *IssueCommand) List(projectID string, query string) {
	log.Printf("List issues %s, %s", projectID, query)

	c.wf.NewItem("Open").Subtitle("Opens project issue list").Valid(true).Arg("open_issues")
	c.wf.NewItem("New").Subtitle("Creates a new issue").Valid(true).Arg("new_issue")
	if len(query) > 0 {
		c.wf.Filter(query)
	}

	reload := func() (interface{}, error) {
		var issues []*provider.Issue

		pageSize := 100
		for page := 0; ; page++ {
			res, err := c.client.ListIssues(projectID, page, pageSize)
			if err != nil {
				return nil, err
			}
			issues = append(issues, res...)
			if len(res) < pageSize {
				break
			}
		}
		return issues, nil
	}

	var issues []*provider.Issue
	if err := c.wf.Cache.LoadOrStoreJSON(fmt.Sprintf(CacheKeyIssuesFormat, projectID), 30*time.Minute, reload, &issues); err != nil {
		panic(err)
	}

	for _, issue := range issues {
		c.wf.NewItem(issue.Title).Var("issue_id", issue.ID).Var("issue_url", issue.WebURL).Valid(true).
			Subtitle(fmt.Sprintf("%s %s", issue.State, issue.Assignee)).Autocomplete(issue.Title)
	}

	if len(query) > 0 {
		c.wf.Filter(query)
	}

	c.wf.WarnEmpty("Empty", "No Issues")
}

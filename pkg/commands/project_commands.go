package commands

import (
	"log"
	"time"

	"github.com/acmore/gitlab-alfred/pkg/provider"
	aw "github.com/deanishe/awgo"
)

const (
	TitleWarning    = "Failure"
	CacheKeyProject = "projects"
)

type ProjectCommand struct {
	client provider.Provider
	wf     *aw.Workflow
}

func NewProjectCommand(wf *aw.Workflow, client provider.Provider) *ProjectCommand {
	return &ProjectCommand{client: client, wf: wf}
}

func (c *ProjectCommand) Run(args []string) {
	defer c.wf.SendFeedback()

	if len(args) < 1 {
		c.wf.Warn("Warning", "Missing subcommand")
		return
	}
	var query string
	if len(args) > 1 {
		query = args[1]
	}

	subcmd := args[0]
	switch subcmd {
	case "list":
		c.List(query)
	}
}

func (c *ProjectCommand) List(query string) {
	reload := func() (interface{}, error) {
		var projects []*provider.Project
		page, pageSize := 0, 100
		for {
			items, err := c.client.ListProjects(page, pageSize)
			if err != nil {
				log.Printf("Failed to list projects, %v", err)
				return nil, err
			}
			projects = append(projects, items...)
			if len(items) < pageSize {
				break
			}
			page += 1
		}
		return projects, nil
	}

	var projects []*provider.Project
	err := c.wf.Cache.LoadOrStoreJSON(CacheKeyProject, 1*time.Hour, reload, &projects)
	if err != nil {
		c.wf.Warn(TitleWarning, err.Error())
		return
	}
	for _, p := range projects {
		c.wf.Feedback.NewItem(p.Name).
			Subtitle(p.WebURL).
			Var("project_id", p.ID).
			Var("project_name", p.Name).
			Var("project_url", p.WebURL).
			Valid(true).
			Autocomplete(p.Name)
	}

	// Filter result
	if len(query) > 0 {
		c.wf.Filter(query)
	}

	c.wf.WarnEmpty("Empty", "No projects")
}

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

func (c *ProjectCommand) List() {
	defer c.wf.SendFeedback()

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
		c.wf.Feedback.NewItem(p.Name).Subtitle(p.WebURL).Arg(p.WebURL)
	}
	c.wf.WarnEmpty("Empty", "No projects")
}

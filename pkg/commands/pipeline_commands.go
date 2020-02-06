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
	CacheKeyPipelineFormat = "pipelines-%s"
)

type PipelineCommand struct {
	wf     *aw.Workflow
	client provider.Provider
}

func NewPipelineCommand(wf *aw.Workflow, client provider.Provider) *PipelineCommand {
	return &PipelineCommand{wf: wf, client: client}
}

func (c *PipelineCommand) Run(args []string) {
	defer c.wf.SendFeedback()

	log.Printf("%v", args)

	var projectID, query, branch, app string

	flagSet := flag.NewFlagSet("pipeline", flag.ContinueOnError)
	flagSet.StringVar(&projectID, "project-id", "", "Project ID")
	flagSet.StringVar(&query, "query", "", "Query")
	flagSet.StringVar(&branch, "branch", "", "Branch to run pipeline")
	flagSet.StringVar(&app, "app", "", "App name")

	if err := flagSet.Parse(args); err != nil {
		log.Printf("Failed to parse command, err: %s", err.Error())
	}

	subcmd := "list"
	if flagSet.NArg() > 0 {
		subcmd = flagSet.Arg(0)
	}

	switch subcmd {
	case "list":
		c.List(projectID, query)
	case "run":
		c.Create(projectID, branch, app)
	}
}

func (c *PipelineCommand) List(projectID string, query string) {
	log.Printf("List pipeline %s, %s", projectID, query)

	reload := func() (interface{}, error) {
		var pipelines []*provider.Pipeline

		pageSize := 100
		for page := 1; ; page++ {
			res, err := c.client.ListPipelines(projectID, page, pageSize, "running")
			if err != nil {
				return nil, err
			}
			pipelines = append(pipelines, res...)
			if len(res) < pageSize {
				break
			}
		}

		for page := 1; ; page++ {
			res, err := c.client.ListPipelines(projectID, page, pageSize, "pending")
			if err != nil {
				return nil, err
			}
			pipelines = append(pipelines, res...)
			if len(res) < pageSize {
				break
			}
		}

		return pipelines, nil
	}

	if len(query) == 0 {
		// Show create option
		c.wf.Feedback.NewItem("Run").Subtitle("Runs a pipeline").Arg("run").Valid(true)
	}

	var pipelines []*provider.Pipeline
	if err := c.wf.Cache.LoadOrStoreJSON(fmt.Sprintf(CacheKeyPipelineFormat, projectID), 10*time.Second, reload, &pipelines); err != nil {
		panic(err)
	}

	for _, p := range pipelines {
		c.wf.Feedback.NewItem(p.Ref).
			Subtitle(p.Status).
			Var("pipeline_id", p.ID).
			Var("pipeline_ref", p.Ref).
			Var("pipeline_status", p.Status).
			Var("pipeline_url", p.WebURL).
			Valid(true)
	}

	if len(query) > 0 {
		c.wf.Filter(query)
	}

	c.wf.WarnEmpty("Empty", "No Pipelines")
}

func (c *PipelineCommand) Create(projectID, branch, app string) {
	log.Printf("Run a pipeline %s, %s, %s", projectID, branch, app)
	vars := make(map[string]string)
	if len(app) > 0 {
		vars["CI_BUILD_APP"] = app
	}
	p, err := c.client.CreatePipeline(projectID, branch, vars)
	if err != nil {
		c.wf.Warn("Failure", err.Error())
	}
	c.wf.NewItem(p.WebURL).Var("pipeline_id", p.ID).Var("pipeline_url", p.WebURL).Subtitle(p.Status).Valid(true)
}

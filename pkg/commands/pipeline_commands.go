package commands

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/acmore/gitlab-alfred/pkg/provider"
	aw "github.com/deanishe/awgo"
	flag "github.com/spf13/pflag"
)

const (
	CacheKeyPipelineFormat = "pipelines-%s"

	notifyTrigger = "tell application id \"com.runningwithcrayons.Alfred\" to run trigger \"notify\" in workflow \"com.acmore.gitlab\" with argument \"%s\""
)

func run(command string) (string, error) {
	cmd := exec.Command("osascript", "-e", command)
	output, err := cmd.CombinedOutput()
	prettyOutput := strings.Replace(string(output), "\n", "", -1)

	// Ignore errors from the user hitting the cancel button
	if err != nil && strings.Index(string(output), "User canceled.") < 0 {
		return "", errors.New(err.Error() + ": " + prettyOutput + " (" + command + ")")
	}

	return prettyOutput, nil
}

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

	var projectID, query, branch, app, pipelineID string

	flagSet := flag.NewFlagSet("pipeline", flag.ContinueOnError)
	flagSet.StringVar(&projectID, "project-id", "", "Project ID")
	flagSet.StringVar(&query, "query", "", "Query")
	flagSet.StringVar(&branch, "branch", "", "Branch to run pipeline")
	flagSet.StringVar(&app, "app", "", "App name")
	flagSet.StringVar(&pipelineID, "pipeline-id", "", "Pipeline ID")

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
	case "cancel":
		c.Cancel(projectID, pipelineID)
	case "watch":
		c.Watch(projectID, pipelineID)
	}
}

func (c *PipelineCommand) List(projectID string, query string) {
	log.Printf("List pipeline %s, %s", projectID, query)

	// Show create option
	c.wf.NewItem("Run").Subtitle("Runs a pipeline").Arg("run").Valid(true)
	c.wf.NewItem("Open").Subtitle("Open pipelines page").Arg("open").Valid(true)
	if len(query) > 0 {
		c.wf.Filter(query)
	}

	reload := func() (interface{}, error) {
		var pipelines []*provider.Pipeline

		pageSize := 100
		for page := 1; ; page++ {
			res, err := c.client.ListPipelines(projectID, page, pageSize, "")
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

	var pipelines []*provider.Pipeline
	if err := c.wf.Cache.LoadOrStoreJSON(fmt.Sprintf(CacheKeyPipelineFormat, projectID), 10*time.Second, reload, &pipelines); err != nil {
		c.wf.FatalError(err)
		return
	}

	for _, p := range pipelines {
		c.wf.NewItem(p.Ref).
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
		return
	}
	c.wf.NewItem(p.WebURL).Var("pipeline_id", p.ID).Var("pipeline_url", p.WebURL).Subtitle(p.Status).Valid(true)

	// Start a background process to watch
	cmd := exec.Command(os.Args[0], "pipeline", "watch", "--project-id", projectID, "--pipeline-id", p.ID)
	if err := c.wf.RunInBackground("watch", cmd); err != nil {
		c.wf.FatalError(err)
	}
}

func (c *PipelineCommand) Cancel(projectID, pipelineID string) {
	log.Printf("Cancel a pipeline %s, %s", projectID, pipelineID)

	c.wf.NewItem("Canceling pipeline").Subtitle(pipelineID).Valid(false)
	c.wf.SendFeedback()

	p, err := c.client.CancelPipeline(projectID, pipelineID)
	if err != nil {
		c.wf.Warn("Failure", err.Error())
		return
	}
	c.wf.NewItem("Pipeline is cancelled").Subtitle(p.WebURL).Var("pipeline_url", p.WebURL).Valid(true)
}

func (c *PipelineCommand) Watch(projectID, pipelineID string) {
	if len(projectID) == 0 || len(pipelineID) == 0 {
		return
	}

	for {
		p, err := c.client.GetPipeline(projectID, pipelineID)
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}
		if p.Status == provider.PipelineStatusPending || p.Status == provider.PipelineStatusRunning {
			time.Sleep(5 * time.Second)
		} else {
			arg := fmt.Sprintf("%s:%s:%s", projectID, pipelineID, p.Status)
			if s, err := run(fmt.Sprintf(notifyTrigger, arg)); err != nil {
				log.Printf("notify trigger failed: %v, output: %s", err, s)
			}
			break
		}
	}
}

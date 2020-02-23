package provider

import (
	"strconv"

	"github.com/xanzy/go-gitlab"
)

type GitlabProvider struct {
	baseURL string
	token   string
	client  *gitlab.Client
}

func NewGitlabProvider(baseURL, token string) Provider {
	client := gitlab.NewClient(nil, token)
	client.SetBaseURL(baseURL)
	return &GitlabProvider{baseURL: baseURL, token: token, client: client}
}

func (c *GitlabProvider) ListProjects(page, pageSize int) ([]*Project, error) {
	accessLevel := gitlab.GuestPermissions
	opt := &gitlab.ListProjectsOptions{MinAccessLevel: &accessLevel, ListOptions: gitlab.ListOptions{Page: page, PerPage: pageSize}}
	projects, _, err := c.client.Projects.ListProjects(opt)
	if err != nil {
		return nil, err
	}
	var result []*Project
	for _, project := range projects {
		result = append(result, &Project{
			ID:          strconv.Itoa(project.ID),
			Name:        project.Name,
			Namespace:   project.Namespace.Name,
			Path:        project.Path,
			WebURL:      project.WebURL,
			Description: project.Description,
		})
	}
	return result, nil
}

func (c *GitlabProvider) GetProject(projectId string) (*Project, error) {
	opt := &gitlab.GetProjectOptions{}
	project, _, err := c.client.Projects.GetProject(projectId, opt)
	if err != nil {
		return nil, err
	}
	return &Project{ID: projectId, Name: project.Name, Namespace: project.Namespace.Name, Path: project.Path, WebURL: project.WebURL, Description: project.Description}, nil
}

func (c *GitlabProvider) ListPipelines(projectID string, page, pageSize int, status string) ([]*Pipeline, error) {
	opt := &gitlab.ListProjectPipelinesOptions{ListOptions: gitlab.ListOptions{Page: page, PerPage: pageSize}}
	if len(status) > 0 {
		s := gitlab.BuildStateValue(status)
		opt.Status = &s
	}

	pipelines, _, err := c.client.Pipelines.ListProjectPipelines(projectID, opt)
	if err != nil {
		return nil, err
	}
	var result []*Pipeline
	for _, p := range pipelines {
		result = append(result, &Pipeline{ID: strconv.Itoa(p.ID), Status: p.Status, Hash: p.SHA, WebURL: p.WebURL, Ref: p.Ref, UpdatedAt: p.UpdatedAt, CreatedAt: p.CreatedAt})
	}
	return result, nil
}

func (c *GitlabProvider) ListBranches(projectID string, page, pageSize int) ([]*Branch, error) {
	opt := &gitlab.ListBranchesOptions{ListOptions: gitlab.ListOptions{Page: page, PerPage: pageSize}}
	branches, _, err := c.client.Branches.ListBranches(projectID, opt)
	if err != nil {
		return nil, err
	}
	var result []*Branch
	for _, p := range branches {
		pipeline := &Pipeline{}
		if p.Commit.LastPipeline != nil {
			pipeline.ID = strconv.Itoa(p.Commit.LastPipeline.ID)
			pipeline.Ref = p.Commit.LastPipeline.Ref
			pipeline.WebURL = p.Commit.LastPipeline.WebURL
		}
		commit := &Commit{
			ID:             p.Commit.ID,
			ShortID:        p.Commit.ShortID,
			Title:          p.Commit.Title,
			AuthorName:     p.Commit.AuthorName,
			AuthorEmail:    p.Commit.AuthorEmail,
			AuthoredDate:   p.Commit.AuthoredDate,
			CommitterName:  p.Commit.CommitterName,
			CommitterEmail: p.Commit.CommitterEmail,
			CommittedDate:  p.Commit.CommittedDate,
			CreatedAt:      p.Commit.CreatedAt,
			Message:        p.Commit.Message,
			ParentIDs:      p.Commit.ParentIDs,
			LastPipeline:   pipeline,
		}
		result = append(result, &Branch{
			Name:               p.Name,
			Protected:          p.Protected,
			Merged:             p.Merged,
			Default:            p.Default,
			DevelopersCanPush:  p.DevelopersCanPush,
			DevelopersCanMerge: p.DevelopersCanMerge,
			Commit:             commit,
		})
	}
	return result, nil
}

func (c *GitlabProvider) CreatePipeline(projectID, ref string, variables map[string]string) (*Pipeline, error) {
	opt := &gitlab.CreatePipelineOptions{Ref: &ref}
	for k, v := range variables {

		opt.Variables = append(opt.Variables, &gitlab.PipelineVariable{
			Key:          k,
			Value:        v,
			VariableType: string(gitlab.EnvVariableType),
		})
	}
	p, _, err := c.client.Pipelines.CreatePipeline(projectID, opt)
	if err != nil {
		return nil, err
	}
	pipeline := &Pipeline{
		ID:        strconv.Itoa(p.ID),
		Status:    p.Status,
		Ref:       p.Ref,
		WebURL:    p.WebURL,
		Hash:      p.SHA,
		UpdatedAt: p.UpdatedAt,
		CreatedAt: p.CreatedAt,
	}
	return pipeline, nil
}

func (c *GitlabProvider) CancelPipeline(projectID, pipelineID string) (*Pipeline, error) {
	id, _ := strconv.Atoi(pipelineID)
	p, _, err := c.client.Pipelines.CancelPipelineBuild(projectID, id)
	if err != nil {
		return nil, err
	}
	pipeline := &Pipeline{
		ID:        strconv.Itoa(p.ID),
		Status:    p.Status,
		Ref:       p.Ref,
		WebURL:    p.WebURL,
		Hash:      p.SHA,
		UpdatedAt: p.UpdatedAt,
		CreatedAt: p.CreatedAt,
	}
	return pipeline, nil
}

func (c *GitlabProvider) GetPipeline(projectID, pipelineID string) (*Pipeline, error) {
	id, _ := strconv.Atoi(pipelineID)
	p, _, err := c.client.Pipelines.GetPipeline(projectID, id)
	if err != nil {
		return nil, err
	}
	pipeline := &Pipeline{
		ID:        strconv.Itoa(p.ID),
		Status:    p.Status,
		Ref:       p.Ref,
		WebURL:    p.WebURL,
		Hash:      p.SHA,
		UpdatedAt: p.UpdatedAt,
		CreatedAt: p.CreatedAt,
	}
	return pipeline, nil
}

func (c *GitlabProvider) ListIssues(projectID string, page, pageSize int) ([]*Issue, error) {
	opt := &gitlab.ListProjectIssuesOptions{ListOptions: gitlab.ListOptions{Page: page, PerPage: pageSize}}
	issues, _, err := c.client.Issues.ListProjectIssues(projectID, opt)
	if err != nil {
		return nil, err
	}
	var result []*Issue
	for _, issue := range issues {
		item := &Issue{
			ID:          strconv.Itoa(issue.ID),
			ProjectID:   projectID,
			Description: issue.Description,
			State:       issue.State,
			Title:       issue.Title,
			UpdatedAt:   issue.UpdatedAt,
			CreatedAt:   issue.CreatedAt,
			ClosedAt:    issue.ClosedAt,
			WebURL:      issue.WebURL,
			Weight:      issue.Weight,
		}
		if issue.Assignee != nil {
			item.Assignee = issue.Assignee.Name
		}
		result = append(result, item)
	}
	return result, nil
}

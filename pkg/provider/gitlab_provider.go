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

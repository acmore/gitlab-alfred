package provider

type Provider interface {
	ListProjects(page, pageSize int) ([]*Project, error)
	GetProject(projectID string) (*Project, error)

	ListPipelines(projectID string, page, pageSize int, status string) ([]*Pipeline, error)
	ListBranches(projectID string, page, pageSize int) ([]*Branch, error)
	CreatePipeline(projectID, ref string, variables map[string]string) (*Pipeline, error)
}

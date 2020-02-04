package provider

type Provider interface {
	ListProjects(page, pageSize int) ([]*Project, error)
	GetProject(projectId string) (*Project, error)
}

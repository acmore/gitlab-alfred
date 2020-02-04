package provider

type Project struct {
	ID          string
	Name        string
	Namespace   string
	Path        string
	WebURL      string
	Description string
}

type Commit struct {
	Hash string
}

type Branch struct {
	Name string
}

type Pipeline struct {
	ID     string
	Status string
}

type Job struct {
	Status string
}

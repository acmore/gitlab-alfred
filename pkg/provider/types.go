package provider

import "time"

type Project struct {
	ID          string
	Name        string
	Namespace   string
	Path        string
	WebURL      string
	Description string
}

type Commit struct {
	ID             string
	ShortID        string
	Title          string
	AuthorName     string
	AuthorEmail    string
	AuthoredDate   *time.Time
	CommitterName  string
	CommitterEmail string
	CommittedDate  *time.Time
	CreatedAt      *time.Time
	Message        string
	ParentIDs      []string
	Status         string
	LastPipeline   *Pipeline
}

type Branch struct {
	Commit             *Commit
	Name               string
	Protected          bool
	Merged             bool
	Default            bool
	DevelopersCanPush  bool
	DevelopersCanMerge bool
}

type Pipeline struct {
	ID        string
	Status    string
	Ref       string
	Hash      string
	WebURL    string
	UpdatedAt *time.Time
	CreatedAt *time.Time
}

type Job struct {
	Status string
}

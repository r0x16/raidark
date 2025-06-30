package auth

// Permission represents a permission in the domain
type Permission struct {
	Owner        string
	Name         string
	CreatedTime  string
	DisplayName  string
	Description  string
	Users        []string
	Groups       []string
	Roles        []string
	Domains      []string
	Model        string
	Adapter      string
	ResourceType string
	Resources    []string
	Actions      []string
	Effect       string
	IsEnabled    bool
	Submitter    string
	Approver     string
	ApproveTime  string
	State        string
}

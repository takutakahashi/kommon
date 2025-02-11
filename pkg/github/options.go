package github

// Options contains GitHub specific options for comment retrieval
type Options struct {
	Owner    string
	Repo     string
	Number   int
	Token    string
	Type     string // "issue" or "pr"
}

// NewOptions creates a new Options instance with default values
func NewOptions(owner, repo string, number int, token string, commentType string) *Options {
	return &Options{
		Owner:  owner,
		Repo:   repo,
		Number: number,
		Token:  token,
		Type:   commentType,
	}
}
package github

// Options contains GitHub specific options for comment retrieval
type Options struct {
	Owner    string
	Repo     string
	PRNumber int
	// GitHub specific options can be added here
	// APIEndpoint string
	// IncludeReviewComments bool
	// etc...
}
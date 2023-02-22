package step

type GitHelper interface {
	GetRootPath() (string, error)
	GetRemoteUrl() (string, error)
	GetCurrentBranch() (string, error)
	GetBranchHash(string) (string, error)
	GetCurrentBranchHash() (string, error)
}

type GitHelperProvider interface {
	NewGitHelper(string) (GitHelper, error)
}

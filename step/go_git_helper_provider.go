package step

import (
	"fmt"
	"github.com/go-git/go-git/v5"
)

type GoGitHelperProvider struct{}

func NewGoGitHelperProvider() GoGitHelperProvider {
	return GoGitHelperProvider{}
}

func (g GoGitHelperProvider) NewGitHelper(projectPath string) (GitHelper, error) {
	repo, err := git.PlainOpen(projectPath)
	if err != nil {
		return GoGitHelper{}, fmt.Errorf("failed to open git repository error: %s", err)
	}

	return NewGoGitHelper(repo), nil
}

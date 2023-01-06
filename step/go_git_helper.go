package step

import (
	"github.com/go-git/go-git"
)

type GoGitHelper struct {
	repo *git.Repository
}

func NewGoGitHelper(repo *git.Repository) GoGitHelper {
	return GoGitHelper{repo: repo}
}

func (g GoGitHelper) GetRootPath() (string, error) {
	//TODO implement me
	panic("implement me")
}

func (g GoGitHelper) GetRemoteUrl() (string, error) {
	//TODO implement me
	panic("implement me")
}

func (g GoGitHelper) GetCurrentBranch() (string, error) {
	//TODO implement me
	panic("implement me")
}

func (g GoGitHelper) GetBranchHash(s string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (g GoGitHelper) GetCurrentBranchHash() (string, error) {
	//TODO implement me
	panic("implement me")
}

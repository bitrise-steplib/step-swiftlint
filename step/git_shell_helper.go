package step

import (
	"fmt"
	"github.com/bitrise-io/go-utils/v2/command"
	"strings"
)

type GitShellHelper struct {
	cmdFactory command.Factory
	opts       command.Opts
}

func NewGitShellHelper(cmdFactory command.Factory, projectPath string) GitShellHelper {
	opts := command.Opts{
		Dir: projectPath,
	}
	helper := GitShellHelper{
		cmdFactory: cmdFactory,
		opts:       opts,
	}

	return helper
}

func (g GitShellHelper) GetRootPath() (string, error) {
	//git rev-parse --show-toplevel
	cmd := g.cmdFactory.Create("git", []string{"rev-parse", "--show-toplevel"}, &g.opts)
	rootPath, err := cmd.RunAndReturnTrimmedOutput()
	if err != nil {
		return "", fmt.Errorf("command failed with error: %s", err)
	}
	return rootPath, nil
}

func (g GitShellHelper) GetRemoteUrl() (string, error) {
	//git config --get remote.origin.url
	cmd := g.cmdFactory.Create("git", []string{"config", "--get", "remote.origin.url"}, &g.opts)
	remoteURL, err := cmd.RunAndReturnTrimmedOutput()
	if err != nil {
		return "", fmt.Errorf("command failed with error: %s", err)
	}
	return strings.TrimSuffix(remoteURL, ".git"), nil
}

func (g GitShellHelper) GetCurrentBranch() (string, error) {
	// git rev-parse --abbrev-ref HEAD
	cmd := g.cmdFactory.Create("git", []string{"rev-parse", "--abbrev-ref", "HEAD"}, &g.opts)
	currentBranch, err := cmd.RunAndReturnTrimmedOutput()
	if err != nil {
		return "", fmt.Errorf("command failed with error: %s", err)
	}
	return currentBranch, nil
}

func (g GitShellHelper) GetBranchHash(branch string) (string, error) {
	// git rev-parse --abbrev-ref HEAD
	cmd := g.cmdFactory.Create("git", []string{"rev-parse", branch}, &g.opts)
	currentBranchHash, err := cmd.RunAndReturnTrimmedOutput()
	if err != nil {
		return "", fmt.Errorf("command failed with error: %s", err)
	}
	return currentBranchHash, nil
}

func (g GitShellHelper) GetCurrentBranchHash() (string, error) {
	currentBranch, err := g.GetCurrentBranch()
	if err != nil {
		return "", err
	}

	var hash string
	hash, err = g.GetBranchHash(currentBranch)
	if err != nil {
		return "", err
	}

	return hash, nil
}

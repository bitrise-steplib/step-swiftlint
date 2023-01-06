package step

import (
	"github.com/bitrise-io/go-utils/v2/command"
)

type GitShellHelperProvider struct {
	cmdFactory command.Factory
}

func NewGitShellHelperProvider(cmdFactory command.Factory) GitShellHelperProvider {
	return GitShellHelperProvider{
		cmdFactory: cmdFactory,
	}
}

func (g GitShellHelperProvider) NewGitHelper(projectPath string) (GitHelper, error) {
	return NewGitShellHelper(
		g.cmdFactory,
		projectPath,
	), nil
}

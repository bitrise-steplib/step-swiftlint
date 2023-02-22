package step

import (
	"fmt"
	"strings"
)

const (
	repositoryProviderGithub    = "github.com"
	repositoryProviderGitlab    = "gitlab.com"
	repositoryProviderBitbucket = "bitbucket.org"
)

type ParsedLineFormatter interface {
	format(line parsedLine) string
}

func ParsedLineFormatterFactory(config Config) ParsedLineFormatter {
	remoteURL := strings.ToLower(config.RepoState.RemoteURL)

	if strings.Contains(remoteURL, repositoryProviderGithub) {
		return GithubParsedLineFormatter{
			RemoteURL:         config.RepoState.RemoteURL,
			CurrentBranchHash: config.RepoState.CurrentBranchHash,
		}
	} else if strings.Contains(remoteURL, repositoryProviderGitlab) {
		return GitlabParsedLineFormatter{
			RemoteURL:         config.RepoState.RemoteURL,
			CurrentBranchHash: config.RepoState.CurrentBranchHash,
		}
	} else if strings.Contains(remoteURL, repositoryProviderBitbucket) {
		return BitbucketParsedLineFormatter{
			RemoteURL:         config.RepoState.RemoteURL,
			CurrentBranchHash: config.RepoState.CurrentBranchHash,
		}
	}

	return nil
}

type GithubParsedLineFormatter struct {
	RemoteURL         string
	CurrentBranchHash string
}

func (g GithubParsedLineFormatter) format(line parsedLine) string {
	return fmt.Sprintf("%s/blob/%s%s#L%d:%s", g.RemoteURL, g.CurrentBranchHash, line.relativeFilePath, line.lineNumber, line.message)
}

type GitlabParsedLineFormatter struct {
	RemoteURL         string
	CurrentBranchHash string
}

func (g GitlabParsedLineFormatter) format(line parsedLine) string {
	return fmt.Sprintf("%s/blob/%s%s#L%d:%s", g.RemoteURL, g.CurrentBranchHash, line.relativeFilePath, line.lineNumber, line.message)
}

type BitbucketParsedLineFormatter struct {
	RemoteURL         string
	CurrentBranchHash string
}

func (g BitbucketParsedLineFormatter) format(line parsedLine) string {
	return fmt.Sprintf("%s/blob/%s%s#L%d:%s", g.RemoteURL, g.CurrentBranchHash, line.relativeFilePath, line.lineNumber, line.message)
}

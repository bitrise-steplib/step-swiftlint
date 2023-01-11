package step

import (
	"fmt"
	"net/url"
	"strings"
)

const (
	repositoryProviderGithub    = "github"
	repositoryProviderGitlab    = "gitlab"
	repositoryProviderBitbucket = "bitbucket"
)

type ParsedLineFormatter interface {
	format(line parsedLine) string
}

func ParsedLineFormatterFactory(config Config) ParsedLineFormatter {
	remoteURL, err := url.ParseRequestURI(config.RepoState.RemoteURL)
	if err != nil {
		return nil
	}

	switch strings.ToLower(remoteURL.Host) {
	case repositoryProviderGithub:
		return GithubParsedLineFormatter{
			RemoteURL:         config.RepoState.RemoteURL,
			CurrentBranchHash: config.RepoState.CurrentBranchHash,
		}
	case repositoryProviderGitlab:
		return GitlabParsedLineFormatter{
			RemoteURL:         config.RepoState.RemoteURL,
			CurrentBranchHash: config.RepoState.CurrentBranchHash,
		}
	case repositoryProviderBitbucket:
		return BitbucketParsedLineFormatter{
			RemoteURL:         config.RepoState.RemoteURL,
			CurrentBranchHash: config.RepoState.CurrentBranchHash,
		}
	default:
		return nil
	}
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

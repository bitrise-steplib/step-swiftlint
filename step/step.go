package step

import (
	"fmt"
	"github.com/bitrise-io/go-steputils/v2/stepconf"
	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/log"
	"strings"
)

type Inputs struct {
	ProjectPath string `env:"project_path"`
	GenerateLog bool   `env:"generate_log,opt[true,false]"`
	DebugMode   bool   `env:"verbose_log,opt[true,false]"`
	StrictMode  bool   `env:"strict_mode,opt[true,false]"`
}

type Config struct {
	Inputs
}

type SwiftLinter struct {
	inputParser stepconf.InputParser
	logger      log.Logger
	cmdFactory  command.Factory
}

// NewSwiftLinter ...
func NewSwiftLinter(
	stepInputParser stepconf.InputParser,
	logger log.Logger,
	cmdFactory command.Factory,
) SwiftLinter {
	return SwiftLinter{
		inputParser: stepInputParser,
		logger:      logger,
		cmdFactory:  cmdFactory,
	}
}

func (s SwiftLinter) ProcessInputs() (Config, error) {
	var inputs Inputs
	if err := s.inputParser.Parse(&inputs); err != nil {
		return Config{}, fmt.Errorf("failed to parse inputs: %s", err)
	}

	stepconf.Print(inputs)

	config := Config{inputs}
	s.logger.EnableDebugLog(config.DebugMode)

	return config, nil
}

func (s SwiftLinter) EnsureDependencies() error {
	isInstalled, err := s.isSwiftLintInstalled()
	if err != nil {
		return err
	}

	if !isInstalled {
		err = s.installSwiftLint()
		if err != nil {
			return err
		}
	}

	return nil
}

func (s SwiftLinter) isSwiftLintInstalled() (bool, error) {
	s.logger.Println()
	s.logger.Infof("Checking if SwiftLint is installed")

	cmd := s.cmdFactory.Create("brew", []string{"list"}, nil)
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return false, fmt.Errorf("%s: error: %s", out, err)
	}

	return strings.Contains(out, "swiftlint"), nil
}

func (s SwiftLinter) installSwiftLint() error {
	s.logger.Println()
	s.logger.Infof("SwiftLint is not installed. Installing SwiftLint.")

	cmd := s.cmdFactory.Create("brew", []string{"install", "swiftlint"}, nil)
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: error: %s", out, err)
	}

	return nil
}

func (s SwiftLinter) Run(config Config) error {
	s.logger.Println()
	s.logger.Infof("Running SwiftLint")
	opts := command.Opts{
		Dir: config.Inputs.ProjectPath,
	}
	//git rev-parse --show-toplevel
	cmd := s.cmdFactory.Create("git", []string{"rev-parse", "--show-toplevel"}, &opts)
	rootPath, err := cmd.RunAndReturnTrimmedOutput()
	if err != nil {
		return fmt.Errorf("failed to get root git path error: %s", err)
	}

	//git config --get remote.origin.url
	cmd = s.cmdFactory.Create("git", []string{"config", "--get", "remote.origin.url"}, &opts)
	remoteURL, err := cmd.RunAndReturnTrimmedOutput()
	if err != nil {
		return fmt.Errorf("failed to get remote url error: %s", err)
	}
	remoteURL = strings.TrimSuffix(remoteURL, ".git")

	// git rev-parse --abbrev-ref HEAD
	cmd = s.cmdFactory.Create("git", []string{"rev-parse", "--abbrev-ref", "HEAD"}, &opts)
	currentBranch, err := cmd.RunAndReturnTrimmedOutput()
	if err != nil {
		return fmt.Errorf("failed to get current branch error: %s", err)
	}

	// git rev-parse --abbrev-ref HEAD
	cmd = s.cmdFactory.Create("git", []string{"rev-parse", currentBranch}, &opts)
	currentBranchHash, err := cmd.RunAndReturnTrimmedOutput()
	if err != nil {
		return fmt.Errorf("failed to get hash of current branch: %s", err)
	}

	parser := LinterParser{
		logger:           s.logger,
		cmdFactory:       s.cmdFactory,
		rootPath:         rootPath,
		repositoryURL:    remoteURL,
		currentBranchSHA: currentBranchHash,
	}

	opts = command.Opts{
		Stdout: parser,
		Stderr: parser,
		Dir:    config.Inputs.ProjectPath,
	}

	args := []string{}
	if config.StrictMode {
		args = append(args, "--strict")
	}

	cmd = s.cmdFactory.Create("swiftlint", args, &opts)
	return cmd.Run()
}

func (s SwiftLinter) ExportOutputs(config Config) error {

	return nil
}

package step

import (
	"bytes"
	"fmt"
	"github.com/bitrise-io/go-steputils/v2/stepconf"
	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/fileutil"
	"github.com/bitrise-io/go-utils/v2/log"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Input struct {
	ProjectPath string `env:"project_path"`
	GenerateLog bool   `env:"generate_log,opt[true,false]"`
	DebugMode   bool   `env:"verbose_log,opt[true,false]"`
	StrictMode  bool   `env:"strict_mode,opt[true,false]"`

	// Output export
	DeployDir string `env:"BITRISE_DEPLOY_DIR"`
}

type Config struct {
	Input
	RepoState GitRepositoryState
}

type GitRepositoryState struct {
	RootPath          string
	RemoteURL         string
	CurrentBranchHash string
}

type SwiftLinter struct {
	inputParser       stepconf.InputParser
	logger            log.Logger
	cmdFactory        command.Factory
	gitHelperProvider GitHelperProvider
}

// NewSwiftLinter ...
func NewSwiftLinter(
	stepInputParser stepconf.InputParser,
	logger log.Logger,
	cmdFactory command.Factory,
	gitHelperProvider GitHelperProvider,
) SwiftLinter {
	return SwiftLinter{
		inputParser:       stepInputParser,
		logger:            logger,
		cmdFactory:        cmdFactory,
		gitHelperProvider: gitHelperProvider,
	}
}

func (s SwiftLinter) ProcessInputs() (Config, error) {
	var input Input
	if err := s.inputParser.Parse(&input); err != nil {
		return Config{}, fmt.Errorf("failed to parse input: %s", err)
	}

	stepconf.Print(input)

	gitHelper, err := s.gitHelperProvider.NewGitHelper(input.ProjectPath)
	if err != nil {
		return Config{}, fmt.Errorf("failed to initialize git helper error: %s", err)
	}

	rootPath, err := gitHelper.GetRootPath()
	if err != nil {
		return Config{}, fmt.Errorf("failed to get root git path error: %s", err)
	}

	remoteURL, err := gitHelper.GetRemoteUrl()
	if err != nil {
		return Config{}, fmt.Errorf("failed to get remote url error: %s", err)
	}
	remoteURL = strings.TrimSuffix(remoteURL, ".git")

	currentBranchHash, err := gitHelper.GetCurrentBranchHash()
	if err != nil {
		return Config{}, fmt.Errorf("failed to get hash of current branch: %s", err)
	}

	config := Config{
		Input: input,
		RepoState: GitRepositoryState{
			RootPath:          rootPath,
			RemoteURL:         remoteURL,
			CurrentBranchHash: currentBranchHash,
		},
	}
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

// Result ...
type Result struct {
	SwiftLintLog []byte
}

func (s SwiftLinter) Run(config Config) (Result, error) {
	s.logger.Println()
	s.logger.Infof("Running SwiftLint")
	opts := command.Opts{
		Dir: config.Input.ProjectPath,
	}

	lineFormatter := ParsedLineFormatterFactory(config)
	var stdOut io.Writer = os.Stdout
	var stdErr io.Writer = os.Stderr

	if lineFormatter != nil {
		parser := LinterParser{
			Logger:        s.logger,
			CmdFactory:    s.cmdFactory,
			RepoState:     config.RepoState,
			LineFormatter: lineFormatter,
		}
		stdOut = parser
		stdErr = parser
	}

	var outBuffer bytes.Buffer
	multiOutWriter := io.MultiWriter(stdOut, &outBuffer)
	multiErrWriter := io.MultiWriter(stdErr, &outBuffer)

	opts = command.Opts{
		Stdout: multiOutWriter,
		Stderr: multiErrWriter,
		Dir:    config.Input.ProjectPath,
	}

	args := []string{}
	if config.StrictMode {
		args = append(args, "--strict")
	}

	cmd := s.cmdFactory.Create("swiftlint", args, &opts)
	err := cmd.Run()

	return Result{
		SwiftLintLog: outBuffer.Bytes(),
	}, err
}

func (s SwiftLinter) ExportOutputs(config Config, result Result) error {
	if !config.GenerateLog {
		return nil
	}
	logFileName := "raw-swiftlint-output.log"
	logPath := filepath.Join(config.DeployDir, logFileName)
	return fileutil.NewFileManager().WriteBytes(logPath, result.SwiftLintLog)
}

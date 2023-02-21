package step

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/go-steputils/v2/stepconf"
	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/fileutil"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-utils/v2/pathutil"
)

const (
	automaticBinarySearchSpecification = "auto"
	swiftlintBinaryName                = "swiftlint"
	cocoapodsSubdirectory              = "Pods/Swiftlint/"
)

type Input struct {
	ProjectPath    string `env:"project_path"`
	GenerateLog    bool   `env:"generate_log,opt[true,false]"`
	DebugMode      bool   `env:"verbose_log,opt[true,false]"`
	StrictMode     bool   `env:"strict_mode,opt[true,false]"`
	BinaryPath     string `env:"binary_path"`
	LintConfigPath string `env:"lint_config_path"`

	// Output export
	DeployDir string `env:"BITRISE_DEPLOY_DIR"`
}

type Config struct {
	Input
	RepoState          GitRepositoryState
	resolvedBinaryPath string
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
	pathModifier      pathutil.PathModifier
	pathChecker       pathutil.PathChecker
}

// NewSwiftLinter ...
func NewSwiftLinter(
	stepInputParser stepconf.InputParser,
	logger log.Logger,
	cmdFactory command.Factory,
	gitHelperProvider GitHelperProvider,
	pathModifier pathutil.PathModifier,
	pathChecker pathutil.PathChecker,
) SwiftLinter {
	return SwiftLinter{
		inputParser:       stepInputParser,
		logger:            logger,
		cmdFactory:        cmdFactory,
		gitHelperProvider: gitHelperProvider,
		pathModifier:      pathModifier,
		pathChecker:       pathChecker,
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
		resolvedBinaryPath: "",
	}
	s.logger.EnableDebugLog(config.DebugMode)

	return config, nil
}

func (s SwiftLinter) EnsureDependencies(config Config) (Config, error) {
	if config.BinaryPath == automaticBinarySearchSpecification {
		s.logger.Println()
		s.logger.Infof("Automatic binary search specified")
		pathToBinary := s.findSwiftLintBinary(config.ProjectPath)

		if len(pathToBinary) > 0 {
			s.logger.Infof("SwiftLint binary found: %s", pathToBinary)

			updatedConfig := Config{
				config.Input,
				config.RepoState,
				pathToBinary,
			}
			return updatedConfig, nil
		}

		s.logger.Warnf("Failed to locate SwiftLint binary in project directory (%s)", config.ProjectPath)
		isInstalled, pathToBinary, err := s.isSwiftLintInstalled()
		if err != nil {
			return Config{}, err
		}

		if !isInstalled {
			pathToBinary, err = s.installSwiftLint()
			if err != nil {
				return Config{}, err
			}
		}

		if err = s.printSwiftLintVersion(pathToBinary); err != nil {
			return Config{}, fmt.Errorf("Failed to get SwiftLint version: %w", err)
		}

		return Config{
			config.Input,
			config.RepoState,
			pathToBinary,
		}, nil
	}

	s.logger.Println()
	s.logger.Infof("Swiftlint binary path specified")
	s.logger.Infof("Expand path")
	pathToBinary, err := s.pathModifier.AbsPath(config.BinaryPath)
	if err != nil {
		return Config{}, fmt.Errorf("Failed to expand path (%s), error: %w", config.BinaryPath, err)
	}

	exists, err := s.pathChecker.IsPathExists(pathToBinary)
	if err != nil {
		return Config{}, fmt.Errorf("Failed to check if path exists (%s), error: %w", config.BinaryPath, err)
	}
	if !exists {
		return Config{}, fmt.Errorf("Binary at specified path (%s) does not exist", config.BinaryPath)
	}

	updatedConfig := Config{
		config.Input,
		config.RepoState,
		pathToBinary,
	}

	return updatedConfig, nil
}

func (s SwiftLinter) findSwiftLintBinary(root string) string {
	if pathToBinary := s.checkCommonSwiftLintLocations(root); len(pathToBinary) > 0 {
		return pathToBinary
	}

	var pathToBinary string
	swiftLintFound := errors.New("SwiftLint found")
	pathsToSkipSearching := s.getPathsToSkipSearching()
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if _, ok := pathsToSkipSearching[d.Name()]; ok {
				return filepath.SkipDir
			}
		}

		if !d.IsDir() && d.Name() == swiftlintBinaryName {
			pathToBinary = path
			return swiftLintFound
		}

		return nil
	})
	if err == swiftLintFound {
		return pathToBinary
	}

	return ""
}

func (s SwiftLinter) checkCommonSwiftLintLocations(root string) string {
	if path := s.checkCocoaPodsDirectory(root); len(path) > 0 {
		return path
	}
	return ""
}

func (s SwiftLinter) getPathsToSkipSearching() map[string]struct{} {
	return map[string]struct{}{
		".git": struct{}{},
	}
}

func (s SwiftLinter) checkCocoaPodsDirectory(root string) string {
	fullPath := filepath.Join(root, cocoapodsSubdirectory, swiftlintBinaryName)
	if s.checkFileExists(fullPath) {
		return fullPath
	}

	return ""
}

func (s SwiftLinter) checkFileExists(pathToFile string) bool {
	fullPath, err := s.pathModifier.AbsPath(pathToFile)
	if err != nil {
		return false
	}
	exists, err := s.pathChecker.IsPathExists(fullPath)
	if err != nil {
		return false
	}
	return exists
}

func (s SwiftLinter) isSwiftLintInstalled() (bool, string, error) {
	s.logger.Println()
	s.logger.Infof("Checking if SwiftLint is installed")

	pathToBinary, err := exec.LookPath(swiftlintBinaryName)
	if err != nil {
		return false, "", fmt.Errorf("error: %s", err)
	}

	return true, pathToBinary, nil
}

func (s SwiftLinter) printSwiftLintVersion(pathToBinary string) error {
	cmd := s.cmdFactory.Create(pathToBinary, []string{"--version"}, nil)

	versionOut, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return err
	}

	s.logger.Infof("SwiftLint version %s found at: %s", versionOut, pathToBinary)

	return nil
}

func (s SwiftLinter) installSwiftLint() (string, error) {
	s.logger.Println()
	s.logger.Infof("SwiftLint is not installed")
	s.logger.Infof("Installing SwiftLint")

	cmd := s.cmdFactory.Create("brew", []string{"install", "swiftlint"}, nil)
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s: error: %s", out, err)
	}

	pathToBinary, err := exec.LookPath(swiftlintBinaryName)
	if err != nil {
		return "", fmt.Errorf("error: %s", err)
	}

	return pathToBinary, nil
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

	if len(config.LintConfigPath) > 0 {
		args = append(args, "--config", config.LintConfigPath)
	}

	cmd := s.cmdFactory.Create(config.resolvedBinaryPath, args, &opts)
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

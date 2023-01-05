package step

import (
	"fmt"
	"github.com/bitrise-io/go-steputils/v2/stepconf"
	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/log"
	"os"
	"strings"
)

//const (
//	minSupportedXcodeMajorVersion = 9
//)

type Inputs struct {
	projectPath string
	generateLog bool
	debugMode   bool
}

type Config struct {
	Inputs
}

type SwiftLinter struct {
	inputParser stepconf.InputParser
	//pathProvider    pathutil.PathProvider
	//pathModifier    pathutil.PathModifier
	logger     log.Logger
	cmdFactory command.Factory
}

// NewSwiftLinter ...
func NewSwiftLinter(
	stepInputParser stepconf.InputParser,
	logger log.Logger,
	cmdFactory command.Factory,
	// pathModifier pathutil.PathModifier,
) SwiftLinter {
	return SwiftLinter{
		inputParser: stepInputParser,
		logger:      logger,
		cmdFactory:  cmdFactory,
		//pathModifier: pathModifier,
	}
}

func (s SwiftLinter) ProcessInputs() (Config, error) {
	var inputs Inputs
	if err := s.inputParser.Parse(&inputs); err != nil {
		return Config{}, fmt.Errorf("failed to parse inputs: %s", err)
	}

	stepconf.Print(inputs)

	config := Config{inputs}
	s.logger.EnableDebugLog(config.debugMode)

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
	s.logger.Infof("Checking if swiftlint is installed")

	cmd := s.cmdFactory.Create("brew", []string{"list"}, nil)
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return false, fmt.Errorf("%s: error: %s", out, err)
	}

	return strings.Contains(out, "swiftlint"), nil
}

func (s SwiftLinter) installSwiftLint() error {
	s.logger.Println()
	s.logger.Infof("Swiftlint is not installed. Installing Swiftlint.")

	cmd := s.cmdFactory.Create("brew", []string{"install", "swiftlint"}, nil)
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: error: %s", out, err)
	}

	return nil
}

func (s SwiftLinter) Run(config Config) error {
	opts := command.Opts{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Dir:    config.Inputs.projectPath,
	}
	cmd := s.cmdFactory.Create("swiftlint", []string{}, &opts)
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: error: %s", out, err)
	}

	return nil
}

func (s SwiftLinter) ExportOutputs() error {

	return nil
}

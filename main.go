package main

import (
	"fmt"
	"os"

	"github.com/bitrise-io/go-steputils/v2/stepconf"
	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/env"
	"github.com/bitrise-io/go-utils/v2/errorutil"
	. "github.com/bitrise-io/go-utils/v2/exitcode"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-utils/v2/pathutil"
	"github.com/bitrise-steplib/steps-swiftlint/step"
)

func main() {
	exitCode := int(run())
	os.Exit(exitCode)
}

func run() ExitCode {
	logger := log.NewLogger()
	buildStep := createStep(logger)

	//process config
	config, err := buildStep.ProcessInputs()
	if err != nil {
		logger.Errorf(errorutil.FormattedError(fmt.Errorf("Failed to process Step inputs: %w", err)))
		return Failure
	}

	//ensure deps
	config, err = buildStep.EnsureDependencies(config)
	if err != nil {
		logger.Errorf(errorutil.FormattedError(fmt.Errorf("Failed to install Step dependencies: %w", err)))
		return Failure
	}

	//run
	result, err := buildStep.Run(config)
	if err != nil {
		logger.Errorf(errorutil.FormattedError(fmt.Errorf("Failed to execute Step: %w", err)))
		return Failure
	}

	//export outputs
	err = buildStep.ExportOutputs(config, result)
	if err != nil {
		logger.Errorf(errorutil.FormattedError(fmt.Errorf("Failed to export Step outputs: %w", err)))
		return Failure
	}

	return Success
}

func createStep(logger log.Logger) step.SwiftLinter {
	envRepository := env.NewRepository()
	inputParser := stepconf.NewInputParser(envRepository)
	cmdFactory := command.NewFactory(envRepository)
	gitHelperProvider := step.NewGitShellHelperProvider(cmdFactory)
	pathModifier := pathutil.NewPathModifier()
	pathChecker := pathutil.NewPathChecker()
	buildStep := step.NewSwiftLinter(inputParser, logger, cmdFactory, gitHelperProvider, pathModifier, pathChecker)
	return buildStep
}

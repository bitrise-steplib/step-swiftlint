package step

import (
	"fmt"
	"github.com/bitrise-io/go-steputils/v2/stepconf"
	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-utils/v2/pathutil"
	"log"
)

type Inputs struct {
	projectPath string
	generateLog bool
	debugMode   bool
}

type Config struct {
	Inputs
}

type SwiftLinter struct {
	stepInputParser stepconf.InputParser
	pathProvider    pathutil.PathProvider
	pathModifier    pathutil.PathModifier
	logger          log.Logger
	cmdFactory      command.Factory
}

func (s SwiftLinter) ProcessInputs() (Config, error) {
	var inputs Inputs
	if err := s.stepInputParser.Parse(&inputs); err != nil {
		return Config{}, fmt.Errorf("failed to parse inputs: %s", err)
	}

	stepconf.Print(inputs)

	config := Config{inputs}
	s.logger.EnableDebugLog(config.debugMode)

	return config, nil
}

func (s SwiftLinter) EnsureDependencies() error {

}

func (s SwiftLinter) Run() error {

}

func (s SwiftLinter) ExportOutputs() error {

}

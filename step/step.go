package step

import (
	"fmt"
	"github.com/bitrise-io/go-steputils/v2/stepconf"
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
}

func (s SwiftLinter) ProcessInputs() (Config, error) {
	var inputs Inputs
	if err := s.stepInputParser.Parse(&inputs); err != nil {
		return Config{}, fmt.Errorf("failed to parse inputs: %s", err)
	}
}

func (s SwiftLinter) EnsureDependencies() error {

}

func (s SwiftLinter) Run() error {

}

func (s SwiftLinter) ExportOutputs() error {

}

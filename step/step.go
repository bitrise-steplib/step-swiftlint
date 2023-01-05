package step

type Config struct {
	projectPath string
	generateLog bool
	debugMode   bool
}

type SwiftLinter struct {
	config Config
}

func (s SwiftLinter) ProcessInputs() (Config, error) {
	
}

func (s SwiftLinter) EnsureDependencies() error {

}

func (s SwiftLinter) Run() error {

}

func (s SwiftLinter) ExportOutputs() error {

}


package step

import (
	"fmt"
	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/log"
	"strconv"
	"strings"
)

const (
	linterLoggingSeverityWarning = "warning"
	linterLoggingSeverityError   = "error"
)

type LinterParser struct {
	Logger        log.Logger
	CmdFactory    command.Factory
	RepoState     GitRepositoryState
	LineFormatter ParsedLineFormatter
}

type parsedLine struct {
	relativeFilePath string
	lineNumber       int
	columnNumber     int
	severity         string
	message          string
}

func (l LinterParser) Write(p []byte) (int, error) {
	err := l.parseAndLog(string(p))
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func (l LinterParser) parseAndLog(s string) error {
	splitStrings := strings.Split(s, "\n")
	parsedLines := []parsedLine{}
	for i := range splitStrings {
		line, err := l.parseLine(splitStrings[i])
		if err != nil { //we're just going to ignore any parsing errors, print the original line, and go next
			l.Logger.Printf(splitStrings[i])
			continue
		}
		var logger func(string, ...interface{})
		switch line.severity {
		case linterLoggingSeverityWarning:
			logger = l.Logger.Warnf
		case linterLoggingSeverityError:
			logger = l.Logger.Errorf
		default:
			logger = l.Logger.Printf
		}
		//github.com/<organization>/<repository>/blob/<branch_name>/README.md#L14: <linter message>
		logString := l.LineFormatter.format(line)
		logger(logString)
		parsedLines = append(parsedLines, line)
	}

	return nil
}

func (l LinterParser) parseLine(s string) (parsedLine, error) {
	// /Users/xxx/Documents/Sample Projects/Bitrise-iOS-Sample/BitriseTestUITests/BitriseTestUITests.swift:18:1: warning: Line Length Violation: Line should be 120 characters or less: currently 182 characters (line_length)
	split := strings.Split(s, ":")

	if len(split) < 5 {
		return parsedLine{}, fmt.Errorf("Unexpected format")
	}

	relPath := strings.TrimPrefix(split[0], l.RepoState.RootPath)

	line64, err := strconv.ParseInt(split[1], 10, 64)
	if err != nil {
		return parsedLine{}, fmt.Errorf("failed to parse line number: %s", err)
	}
	line := int(line64)

	col64, err := strconv.ParseInt(split[2], 10, 64)
	if err != nil {
		return parsedLine{}, fmt.Errorf("failed to parse column number: %s", err)
	}
	col := int(col64)

	sev := strings.TrimSpace(split[3])

	message := strings.Join(split[4:], ":")

	parsedLine := parsedLine{
		relativeFilePath: relPath,
		lineNumber:       line,
		columnNumber:     col,
		severity:         sev,
		message:          message,
	}

	return parsedLine, nil
}

// Package message provides a rich set of functions for displaying messages to the user.
package message

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/pterm/pterm"
	"github.com/sergi/go-diff/diffmatchpatch"
)

// LogLevel is the level of logging to display.
type LogLevel int

const (
	// WarnLevel level. Non-critical entries that deserve eyes.
	WarnLevel LogLevel = iota
	// InfoLevel level. General operational entries about what's going on inside the
	// application.
	InfoLevel
	// DebugLevel level. Usually only enabled when debugging. Very verbose logging.
	DebugLevel
	// TraceLevel level. Designates finer-grained informational events than the Debug.
	TraceLevel

	// TermWidth sets the width of full width elements like progressbars and headers
	TermWidth = 100
)

// NoProgress tracks whether spinner/progress bars show updates.
var NoProgress bool

// RuleLine creates a line of ━ as wide as the terminal
var RuleLine = strings.Repeat("━", TermWidth)

// LogWriter is the stream to write logs to.
var LogWriter io.Writer = os.Stderr

// logLevel holds the pterm compatible log level integer
var logLevel = InfoLevel

// logFile acts as a buffer for logFile generation
var logFile *os.File

// useLogFile controls whether to use the log file or not
var useLogFile bool

// failOutput is a prefix printer for fail messages
var failOutput pterm.PrefixPrinter

// DebugWriter represents a writer interface that writes to message.Debug
type DebugWriter struct{}

func (d *DebugWriter) Write(raw []byte) (int, error) {
	debugPrinter(2, string(raw))
	return len(raw), nil
}

func init() {
	pterm.ThemeDefault.SuccessMessageStyle = *pterm.NewStyle(pterm.FgLightGreen)
	// Customize default error.
	pterm.Success.Prefix = pterm.Prefix{
		Text:  " ✔",
		Style: pterm.NewStyle(pterm.FgLightGreen),
	}
	pterm.Error.Prefix = pterm.Prefix{
		Text:  "    ERROR:",
		Style: pterm.NewStyle(pterm.BgLightRed, pterm.FgBlack),
	}
	pterm.Info.Prefix = pterm.Prefix{
		Text: " •",
	}

	failOutput = pterm.PrefixPrinter{
		MessageStyle: pterm.NewStyle(pterm.FgRed),
		Prefix: pterm.Prefix{
			Style: pterm.NewStyle(pterm.FgRed),
			Text:  " ✗",
		},
	}

	SetDefaultOutput(os.Stderr)
}

// SetDefaultOutput sets the default output for all messages.
// Deprecated. Use pterm.SetDefaultOutput instead when https://github.com/pterm/pterm/issues/701 is fixed.
func SetDefaultOutput(w io.Writer) {
	pterm.SetDefaultOutput(w)
	pterm.Info = *(pterm.Info.WithWriter(w))
	pterm.Warning = *(pterm.Warning.WithWriter(w))
	pterm.Success = *(pterm.Success.WithWriter(w))
	pterm.Error = *(pterm.Error.WithWriter(w))
	pterm.Fatal = *(pterm.Fatal.WithWriter(w))
	pterm.Debug = *(pterm.Debug.WithWriter(w))
	pterm.Description = *(pterm.Description.WithWriter(w))
}

// UseLogFile writes output to stderr and a logFile.
func UseLogFile(inputLogFile *os.File) {
	// Prepend the log filename with a timestamp.
	ts := time.Now().Format("2006-01-02-15-04-05")

	var err error
	if inputLogFile != nil {
		// Use the existing log file if logFile is set
		LogWriter = io.MultiWriter(inputLogFile)
		SetDefaultOutput(LogWriter)
	} else {
		// Try to create a temp log file if one hasn't been made already
		if logFile, err = os.CreateTemp("", fmt.Sprintf("lula-%s-*.log", ts)); err != nil {
			WarnErr(err, "Error saving a log file to a temporary directory")
		} else {
			useLogFile = true
			LogWriter = io.MultiWriter(os.Stderr, logFile)
			SetDefaultOutput(LogWriter)
			message := fmt.Sprintf("Saving log file to %s", logFile.Name())
			Note(message)
		}
	}
}

// UseBuffer writes output to a buffer
func UseBuffer(buf *bytes.Buffer) {
	LogWriter = io.MultiWriter(buf)
	SetDefaultOutput(LogWriter)
}

// SetLogLevel sets the log level.
func SetLogLevel(lvl LogLevel) {
	logLevel = lvl
	if logLevel >= DebugLevel {
		pterm.EnableDebugMessages()
	}
}

// GetLogLevel returns the current log level.
func GetLogLevel() LogLevel {
	return logLevel
}

// DisableColor disables color in output
func DisableColor() {
	pterm.DisableColor()
}

// Debug prints a debug message.
func Debug(payload ...any) {
	debugPrinter(2, payload...)
}

// Debugf prints a debug message with a given format.
func Debugf(format string, a ...any) {
	message := fmt.Sprintf(format, a...)
	debugPrinter(2, message)
}

// ErrorWebf prints an error message and returns a web response.
func ErrorWebf(err any, w http.ResponseWriter, format string, a ...any) {
	debugPrinter(2, err)
	message := fmt.Sprintf(format, a...)
	Warn(message)
	http.Error(w, message, http.StatusInternalServerError)
}

// Warn prints a warning message.
func Warn(message string) {
	Warnf("%s", message)
}

// Warnf prints a warning message with a given format.
func Warnf(format string, a ...any) {
	message := Paragraphn(TermWidth-10, format, a...)
	pterm.Println()
	pterm.Warning.Println(message)
}

// WarnErr prints an error message as a warning.
func WarnErr(err any, message string) {
	debugPrinter(2, err)
	Warnf("%s", message)
}

// WarnErrf prints an error message as a warning with a given format.
func WarnErrf(err any, format string, a ...any) {
	debugPrinter(2, err)
	Warnf(format, a...)
}

// Fatal prints a fatal error message and exits with a 1.
func Fatal(err any, message string) {
	debugPrinter(2, err)
	errorPrinter(2).Println(message)
	debugPrinter(2, string(debug.Stack()))
	os.Exit(1)
}

// Fatalf prints a fatal error message and exits with a 1 with a given format.
func Fatalf(err any, format string, a ...any) {
	message := Paragraph(format, a...)
	Fatal(err, message)
}

// Info prints an info message.
func Info(message string) {
	Infof("%s", message)
}

// Infof prints an info message with a given format.
func Infof(format string, a ...any) {
	if logLevel > 0 {
		message := Paragraph(format, a...)
		pterm.Info.Println(message)
	}
}

// Detail prints detail message.
func Detail(message string) {
	Detailf("%s", message)
}

// Detailf prints a detail message preserving newlines
func Detailf(format string, a ...any) {
	if logLevel > 0 {
		message := fmt.Sprintf(format, a...)
		pterm.Info.Println(message)
	}
}

// Success prints a success message.
func Success(message string) {
	Successf("%s", message)
}

// Successf prints a success message with a given format.
func Successf(format string, a ...any) {
	message := Paragraph(format, a...)
	pterm.Success.Println(message)
}

// Fail prints a fail message.
func Fail(message string) {
	Failf("%s", message)
}

// Failf prints a fail message with a given format.
func Failf(format string, a ...any) {
	message := Paragraph(format, a...)
	failOutput.Println(message)
}

// Question prints a user prompt description message.
func Question(text string) {
	Questionf("%s", text)
}

// Questionf prints a user prompt description message with a given format.
func Questionf(format string, a ...any) {
	pterm.Println()
	message := Paragraph(format, a...)
	pterm.FgLightGreen.Println(message)
}

// Note prints a note message.
func Note(text string) {
	Notef("%s", text)
}

// Notef prints a note message  with a given format.
func Notef(format string, a ...any) {
	pterm.Println()
	message := Paragraphn(TermWidth-7, format, a...)
	notePrefix := pterm.PrefixPrinter{
		MessageStyle: &pterm.ThemeDefault.InfoMessageStyle,
		Prefix: pterm.Prefix{
			Style: &pterm.ThemeDefault.InfoPrefixStyle,
			Text:  "NOTE",
		},
	}
	notePrefix.Println(message)
}

// Title prints a title and an optional help description for that section
func Title(title string, help string) {
	titleFormatted := pterm.FgBlack.Sprint(pterm.BgWhite.Sprintf(" %s ", title))
	helpFormatted := pterm.FgGray.Sprint(help)
	pterm.Printfln("%s  %s", titleFormatted, helpFormatted)
}

// HeaderInfof prints a large header with a formatted message.
func HeaderInfof(format string, a ...any) {
	message := Truncate(fmt.Sprintf(format, a...), TermWidth, false)
	// Ensure the text is consistent for the header width
	padding := TermWidth - len(message)
	pterm.Println()
	pterm.DefaultHeader.
		WithBackgroundStyle(pterm.NewStyle(pterm.BgDarkGray)).
		WithTextStyle(pterm.NewStyle(pterm.FgLightWhite)).
		WithMargin(2).
		Printfln("%s", message+strings.Repeat(" ", padding))
}

// HorizontalRule prints a white horizontal rule to separate the terminal
func HorizontalRule() {
	pterm.Println()
	pterm.Println(RuleLine)
}

// JSONValue prints any value as JSON.
func JSONValue(value any) string {
	bytes, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		debugPrinter(2, fmt.Sprintf("ERROR marshalling json: %s", err.Error()))
	}
	return string(bytes)
}

func Printf(format string, a ...any) {
	pterm.Printf(format, a...)
}

// Paragraph formats text into a paragraph matching the TermWidth
func Paragraph(format string, a ...any) string {
	return Paragraphn(TermWidth, format, a...)
}

// Paragraphn formats text into an n column paragraph
func Paragraphn(n int, format string, a ...any) string {
	return pterm.DefaultParagraph.WithMaxWidth(n).Sprintf(format, a...)
}

// PrintDiff prints the differences between a and b with a as original and b as new
func PrintDiff(textA, textB string) {
	dmp := diffmatchpatch.New()

	diffs := dmp.DiffMain(textA, textB, true)

	diffs = dmp.DiffCleanupSemantic(diffs)

	pterm.Println(dmp.DiffPrettyText(diffs))
}

// Truncate truncates provided text to the requested length
func Truncate(text string, length int, invert bool) string {
	// Remove newlines and replace with semicolons
	textEscaped := strings.ReplaceAll(text, "\n", "; ")
	// Truncate the text if it is longer than length so it isn't too long.
	if len(textEscaped) > length {
		if invert {
			start := len(textEscaped) - length + 3
			textEscaped = "..." + textEscaped[start:]
		} else {
			end := length - 3
			textEscaped = textEscaped[:end] + "..."
		}
	}
	return textEscaped
}

// Table prints a padded table containing the specified header and data
// Note - columnSize should be an array of ints that add up to 100
func Table(header []string, data [][]string, columnSize []int) error {
	pterm.Println()
	termWidth := pterm.GetTerminalWidth() - 10 // Subtract 10 for padding

	if len(columnSize) != len(header) {
		Warn("The number of columns does not match the number of headers")
		columnSize = make([]int, len(header))
		for i := range columnSize {
			columnSize[i] = (len(header) / termWidth) * 100 // make them all equal
		}
	}

	table := pterm.TableData{
		header,
	}

	for _, row := range data {
		for i, cell := range row {
			row[i] = addLineBreaks(strings.Replace(cell, "\n", " ", -1), (columnSize[i]*termWidth)/100)
		}
		table = append(table, pterm.TableData{row}...)
	}

	return pterm.DefaultTable.WithHasHeader().WithData(table).WithRowSeparator("-").Render()
}

// Add line breaks for table

func addLineBreaks(input string, maxLineLength int) string {
	// words := splitWords(input) // Split the input into words, handling hyphens
	words := strings.Fields(input)
	var result strings.Builder // Use a strings.Builder for efficient string concatenation
	currentLineLength := 0

	for _, word := range words {
		if currentLineLength+len(word) > maxLineLength {
			// additionally split the word if it contains a hyphen
			firstPart, secondPart := splitHyphenedWords(word, currentLineLength, maxLineLength)
			if firstPart != "" {
				if currentLineLength > 0 {
					result.WriteString(" ")
				}
				result.WriteString(firstPart + "-")
			}
			result.WriteString("\n")
			currentLineLength = 0

			if secondPart != "" {
				word = secondPart
			}
		}
		if currentLineLength > 0 {
			result.WriteString(" ")
			currentLineLength++
		}
		result.WriteString(word)
		currentLineLength += len(word)
	}

	return result.String()
}

func splitHyphenedWords(input string, currentLength int, maxLength int) (firstPart string, secondPart string) {
	// get the indicies of all the hyphens
	hyphenIndicies := []int{}
	for i, char := range input {
		if char == '-' {
			hyphenIndicies = append(hyphenIndicies, i)
		}
	}

	if len(hyphenIndicies) != 0 {
		// starting from the last index, find the largest firstPart that fits within the maxLength
		for i := len(hyphenIndicies) - 1; i >= 0; i-- {
			hyphenIndex := hyphenIndicies[i]
			firstPart = input[:hyphenIndex]
			secondPart = input[hyphenIndex+1:]
			if len(firstPart)+currentLength <= maxLength {
				return firstPart, secondPart
			}
		}
	}
	return "", input
}

func debugPrinter(offset int, a ...any) {
	printer := pterm.Debug.WithShowLineNumber(logLevel > 2).WithLineNumberOffset(offset)
	now := time.Now().Format(time.RFC3339)
	// prepend to a
	a = append([]any{now, " - "}, a...)

	printer.Println(a...)

	// Always write to the log file
	if useLogFile {
		pterm.Debug.
			WithShowLineNumber(true).
			WithLineNumberOffset(offset).
			WithDebugger(false).
			WithWriter(logFile).
			Println(a...)
	}
}

func errorPrinter(offset int) *pterm.PrefixPrinter {
	return pterm.Error.WithShowLineNumber(logLevel > 2).WithLineNumberOffset(offset)
}

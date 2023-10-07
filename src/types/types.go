package types

// this is a temporary struct until a reporting model is selected
type ComplianceReport struct {
	UUID        string `json:"uuid" yaml:"uuid"`
	ControlId   string `json:"control-id" yaml:"control-id"`
	Description string `json:"description" yaml:"description"`
	Result      string `json:"result" yaml:"result"`
}

// The ReportObject keeps track of all pertinent information as it relates to relational data IE UUID's
// This will hopefully make transformation to the reporting model easier
type ReportObject struct {
	FilePaths  []string    `json:"file-paths" yaml:"file-paths"`
	UUID       string      `json:"uuid" yaml:"uuid"`
	Components []Component `json:"components" yaml:"components"`
}

type Component struct {
	UUID                   string                  `json:"uuid" yaml:"uuid"`
	ControlImplementations []ControlImplementation `json:"control-implementations" yaml:"control-implementations"`
}

type ControlImplementation struct {
	UUID            string           `json:"uuid" yaml:"uuid"`
	ImplementedReqs []ImplementedReq `json:"implemented-reqs" yaml:"implemented-reqs"`
}

type ImplementedReq struct {
	UUID    string   `json:"uuid" yaml:"uuid"`
	Results []Result `json:"results" yaml:"results"`
}

// native type for conversion to targeted report format
type Result struct {
	UUID        string `json:"uuid" yaml:"uuid"`
	ControlId   string `json:"control-id" yaml:"control-id"`
	Description string `json:"description" yaml:"description"`
	Passing     int    `json:"passing" yaml:"passing"`
	Failing     int    `json:"failing" yaml:"failing"`
	Result      string `json:"result" yaml:"result"`
}

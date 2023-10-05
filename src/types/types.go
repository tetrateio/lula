package types

// native type for conversion to targeted report format
type Result struct {
	UUID        string `json:"uuid" yaml:"uuid"`
	ControlId   string `json:"control-id" yaml:"control-id"`
	Description string `json:"description" yaml:"description"`
	Passing     int    `json:"passing" yaml:"passing"`
	Failing     int    `json:"failing" yaml:"failing"`
	Result      string `json:"result" yaml:"result"`
}

// this is a temporary struct until a reporting model is selected
type ComplianceReport struct {
	UUID        string `json:"uuid" yaml:"uuid"`
	ControlId   string `json:"control-id" yaml:"control-id"`
	Description string `json:"description" yaml:"description"`
	Result      string `json:"result" yaml:"result"`
}

// This is the object that keeps track of all pertinent relational data
type ResultObject struct {
	FilePaths             []string    `json:"file-paths" yaml:"file-paths"`
	ComponentDefinitionId string      `json:"component-definition-id" yaml:"component-definition-id"`
	Components            []Component `json:"components" yaml:"components"`
}

type Component struct {
	UUID            string           `json:"uuid" yaml:"uuid"`
	ImplementedReqs []ImplementedReq `json:"implemented-reqs" yaml:"implemented-reqs"`
}

type ImplementedReq struct {
	UUID    string   `json:"uuid" yaml:"uuid"`
	Results []Result `json:"results" yaml:"results"`
}

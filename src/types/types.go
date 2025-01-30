package types

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/defenseunicorns/lula/src/pkg/message"
)

// Define base errors for validations
var (
	ErrExecutionNotAllowed = errors.New("execution not allowed")
	ErrDomainGetResources  = errors.New("domain GetResources error")
	ErrProviderEvaluate    = errors.New("provider Evaluate error")
)

type LulaValidationType string

const (
	LulaValidationTypeNormal  LulaValidationType = "Lula Validation"
	DefaultLulaValidationType LulaValidationType = LulaValidationTypeNormal
)

type LulaValidation struct {
	// Name of the Validation
	Name string

	// UUID of the validation - tied to the component-definition.backmatter
	UUID string

	// Provider is the provider that is evaluating the validation
	Provider *Provider

	// Domain is the domain that provides the evidence for the validation
	Domain *Domain

	// DomainResources is the set of resources that the domain is providing
	DomainResources *DomainResources

	// LulaValidationType is the type of validation that is being performed
	LulaValidationType LulaValidationType

	// Evaluated is a boolean that represents if the validation has been evaluated
	Evaluated bool

	// ValidationTestData is a slice of test data corresponding to the lula validation
	ValidationTestData []*LulaValidationTestData

	// Result is the result of the validation
	Result *Result
}

// CreateFailingLulaValidation creates a placeholder LulaValidation object that is always failing
func CreateFailingLulaValidation(name string) *LulaValidation {
	return &LulaValidation{
		Name:      name,
		Evaluated: true,
		Result:    &Result{Failing: 1},
	}
}

// CreatePassingLulaValidation creates a placeholder LulaValidation object that is always passing
func CreatePassingLulaValidation(name string) *LulaValidation {
	return &LulaValidation{
		Name:      name,
		Evaluated: true,
		Result:    &Result{Passing: 1},
	}
}

// Lula Validation Options settings
type lulaValidationOptions struct {
	staticResources  DomainResources
	executionAllowed bool
	isInteractive    bool
	onlyResources    bool
	spinner          *message.Spinner
}

type LulaValidationOption func(*lulaValidationOptions)

// WithStaticResources sets the static resources for the LulaValidation object
func WithStaticResources(resources DomainResources) LulaValidationOption {
	return func(opts *lulaValidationOptions) {
		opts.staticResources = resources
	}
}

// ExecutionAllowed sets the value of the executionAllowed field in the LulaValidation object
func ExecutionAllowed(executionAllowed bool) LulaValidationOption {
	return func(opts *lulaValidationOptions) {
		opts.executionAllowed = executionAllowed
	}
}

// Interactive is a function that returns a boolean indicating if the validation should be interactive
func Interactive(isInteractive bool) LulaValidationOption {
	return func(opts *lulaValidationOptions) {
		opts.isInteractive = isInteractive
	}
}

// WithSpinner returns a LulaValidationOption that sets the spinner for the LulaValidation object
func WithSpinner(spinner *message.Spinner) LulaValidationOption {
	return func(opts *lulaValidationOptions) {
		opts.spinner = spinner
	}
}

// RequireExecutionConfirmation is a function that returns a boolean indicating if the validation requires confirmation before execution
func GetResourcesOnly(onlyResources bool) LulaValidationOption {
	return func(opts *lulaValidationOptions) {
		opts.onlyResources = onlyResources
	}
}

// Perform the validation, and store the result in the LulaValidation struct
func (v *LulaValidation) Validate(ctx context.Context, opts ...LulaValidationOption) error {
	if !v.Evaluated {
		var result Result
		var err error
		var resources DomainResources

		// Update the validation
		v.DomainResources = &resources
		v.Result = &result
		v.Evaluated = true

		// Set Validation config from options passed
		config := &lulaValidationOptions{
			staticResources:  nil,
			executionAllowed: false,
			isInteractive:    false,
			onlyResources:    false,
			spinner:          nil,
		}
		for _, opt := range opts {
			opt(config)
		}

		// Check if confirmation needed before execution
		if config.staticResources == nil {
			if (*v.Domain).IsExecutable() && !config.executionAllowed {
				if config.isInteractive {
					// Run confirmation user prompt
					if confirm := message.PromptForConfirmation(config.spinner); !confirm {
						return fmt.Errorf("%w: requested execution denied", ErrExecutionNotAllowed)
					}
				} else {
					return fmt.Errorf("%w: non-interactive execution not allowed", ErrExecutionNotAllowed)
				}
			}
		}

		// Get the resources
		if config.staticResources != nil {
			resources = config.staticResources
		} else {
			resources, err = (*v.Domain).GetResources(ctx)
			if err != nil {
				return fmt.Errorf("%w: %v", ErrDomainGetResources, err)
			}
			if config.onlyResources {
				return nil
			}
		}

		// Perform the evaluation using the provider
		result, err = (*v.Provider).Evaluate(ctx, resources)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrProviderEvaluate, err)
		}
	}
	return nil
}

// RunTests executes any tests defined in the validation and returns a report of the results
func (v *LulaValidation) RunTests(ctx context.Context, saveResources bool) (*LulaValidationTestReport, error) {
	if v.DomainResources == nil {
		return nil, fmt.Errorf("domain resources are nil, tests cannot be run")
	}

	// For each test, apply the transforms to the domain resources and run validate using those resources
	if len(v.ValidationTestData) != 0 {
		testReport := NewLulaValidationTestReport(v.Name)
		for _, d := range v.ValidationTestData {
			// Only execute test if it has not been executed yet
			if d.Test != nil && d.Result == nil {
				// Create a fresh copy of the resources and validation to run each test on
				testResources := deepCopyMap(*v.DomainResources)
				testValidation := &LulaValidation{
					Provider: v.Provider,
				}

				// Execute the test
				testResult, err := d.ExecuteTest(ctx, testValidation, testResources, saveResources)
				if err != nil {
					return nil, err
				}
				testReport.AddTestResult(testResult)
			} else if d.Result != nil {
				testReport.AddTestResult(d.Result)
			}
		}
		return testReport, nil
	}

	return nil, nil
}

// Check if the validation requires confirmation before possible execution code is run
func (v *LulaValidation) RequireExecutionConfirmation() (confirm bool) {
	return !(*v.Domain).IsExecutable()
}

// Return domain resources as a json []byte
func (v *LulaValidation) GetDomainResourcesAsJSON() []byte {
	if v.DomainResources == nil {
		return []byte("{}")
	}
	jsonData, err := json.MarshalIndent(v.DomainResources, "", "  ")
	if err != nil {
		message.Debugf("Error marshalling domain resources to JSON: %v", err)
		jsonData = []byte(`{"Error": "Error marshalling to JSON"}`)
	}
	return jsonData
}

type Domain interface {
	GetResources(context.Context) (DomainResources, error)
	IsExecutable() bool
}

type Provider interface {
	Evaluate(context.Context, DomainResources) (Result, error)
}

// native type for conversion to targeted report format
type Result struct {
	UUID         string            `json:"uuid" yaml:"uuid"`
	ControlId    string            `json:"control-id" yaml:"control-id"`
	Description  string            `json:"description" yaml:"description"`
	Passing      int               `json:"passing" yaml:"passing"`
	Failing      int               `json:"failing" yaml:"failing"`
	State        string            `json:"state" yaml:"state"`
	Observations map[string]string `json:"observations" yaml:"observations"`
}

func deepCopyMap(input map[string]interface{}) map[string]interface{} {
	if input == nil {
		return nil
	}

	// Create a new map to hold the copy
	copy := make(map[string]interface{})

	for key, value := range input {
		// Check the type of the value and copy accordingly
		switch v := value.(type) {
		case map[string]interface{}:
			// If the value is a map, recursively deep copy it
			copy[key] = deepCopyMap(v)
		case []interface{}:
			// If the value is a slice, deep copy each element
			copy[key] = deepCopySlice(v)
		default:
			// For other types (e.g., strings, ints), just assign directly
			copy[key] = v
		}
	}

	return copy
}

// Helper function to deep copy a slice of interface{}
func deepCopySlice(input []interface{}) []interface{} {
	if input == nil {
		return nil
	}

	// Create a new slice to hold the copy
	copy := make([]interface{}, len(input))

	for i, value := range input {
		// Check the type of the value and copy accordingly
		switch v := value.(type) {
		case map[string]interface{}:
			// If the value is a map, recursively deep copy it
			copy[i] = deepCopyMap(v)
		case []interface{}:
			// If the value is a slice, deep copy each element
			copy[i] = deepCopySlice(v)
		default:
			// For other types (e.g., strings, ints), just assign directly
			copy[i] = v
		}
	}

	return copy
}

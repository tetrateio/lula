package test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/defenseunicorns/go-oscal/src/pkg/files"
	"github.com/defenseunicorns/go-oscal/src/pkg/revision"
	oscalValidation "github.com/defenseunicorns/go-oscal/src/pkg/validation"
	"github.com/defenseunicorns/go-oscal/src/pkg/versioning"
	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-3"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"

	"github.com/defenseunicorns/lula/src/pkg/common"
	"github.com/defenseunicorns/lula/src/pkg/common/network"
	"github.com/defenseunicorns/lula/src/pkg/common/oscal"
	"github.com/defenseunicorns/lula/src/pkg/common/validation"
	"github.com/defenseunicorns/lula/src/pkg/message"
	"github.com/defenseunicorns/lula/src/test/util"
	"github.com/defenseunicorns/lula/src/types"
)

func TestPodLabelValidation(t *testing.T) {
	const ckTestPodLabel contextKey = "test-pod-label"
	featureTrueValidation := features.New("Check Pod Validation - Success").
		Setup(func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
			// Set the work directory to the directory containing the oscal component and the modules to import
			return context.WithValue(ctx, types.LulaValidationWorkDir, "./scenarios/pod-label")
		}).
		Setup(func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
			pod, err := util.GetPod("./scenarios/pod-label/pod.pass.yaml")
			if err != nil {
				t.Fatal(err)
			}
			if err = config.Client().Resources().Create(ctx, pod); err != nil {
				t.Fatal(err)
			}
			err = wait.For(conditions.New(config.Client().Resources()).PodConditionMatch(pod, corev1.PodReady, corev1.ConditionTrue), wait.WithTimeout(time.Minute*1))
			if err != nil {
				t.Fatal(err)
			}
			return context.WithValue(ctx, ckTestPodLabel, pod)
		}).
		Assess("Validate pod label", func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
			oscalPath := "./scenarios/pod-label/oscal-component.yaml"
			return validatePodLabelPass(ctx, t, oscalPath)
		}).
		Assess("Validate pod label (Kyverno)", func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
			oscalPath := "./scenarios/pod-label/oscal-component-kyverno.yaml"
			return validatePodLabelPass(ctx, t, oscalPath)
		}).
		Assess("Validate pod label (save-resources=remote)", func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
			oscalPath := "./scenarios/pod-label/oscal-component.yaml"
			return validateSaveResources(ctx, t, oscalPath)
		}).
		Teardown(func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
			pod := ctx.Value(ckTestPodLabel).(*corev1.Pod)
			if err := config.Client().Resources().Delete(ctx, pod); err != nil {
				t.Fatal(err)
			}

			err := wait.For(conditions.New(config.Client().Resources()).ResourceDeleted(pod), wait.WithTimeout(time.Minute*1))
			if err != nil {
				t.Fatal(err)
			}

			err = os.Remove("sar-test.yaml")
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).Feature()

	featureFalseValidation := features.New("Check Pod Validation - Failure").
		Setup(func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
			// Set the work directory to the directory containing the oscal component and the modules to import
			return context.WithValue(ctx, types.LulaValidationWorkDir, "./scenarios/pod-label")
		}).
		Setup(func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
			pod, err := util.GetPod("./scenarios/pod-label/pod.fail.yaml")
			if err != nil {
				t.Fatal(err)
			}
			if err = config.Client().Resources().Create(ctx, pod); err != nil {
				t.Fatal(err)
			}
			err = wait.For(conditions.New(config.Client().Resources()).PodConditionMatch(pod, corev1.PodReady, corev1.ConditionTrue), wait.WithTimeout(time.Minute*5))
			if err != nil {
				t.Fatal(err)
			}
			return context.WithValue(ctx, ckTestPodLabel, pod)
		}).
		Assess("Validate pod label", func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
			oscalPath := "./scenarios/pod-label/oscal-component.yaml"
			validatePodLabelFail(ctx, t, oscalPath)
			return ctx
		}).
		Assess("Validate pod label (Kyverno)", func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
			oscalPath := "./scenarios/pod-label/oscal-component-kyverno.yaml"
			validatePodLabelFail(ctx, t, oscalPath)
			return ctx
		}).
		Teardown(func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
			pod := ctx.Value(ckTestPodLabel).(*corev1.Pod)
			if err := config.Client().Resources().Delete(ctx, pod); err != nil {
				t.Fatal(err)
			}
			err := wait.For(conditions.New(config.Client().Resources()).ResourceDeleted(pod), wait.WithTimeout(time.Minute*1))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).Feature()

	featureBadValidation := features.New("Check Graceful Failure - check all not-satisfied and matching error").
		Setup(func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
			// Set the work directory to the directory containing the oscal component and the modules to import
			return context.WithValue(ctx, types.LulaValidationWorkDir, "./scenarios/pod-label")
		}).
		Setup(func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
			pod, err := util.GetPod("./scenarios/pod-label/pod.pass.yaml")
			if err != nil {
				t.Fatal(err)
			}
			if err = config.Client().Resources().Create(ctx, pod); err != nil {
				t.Fatal(err)
			}
			err = wait.For(conditions.New(config.Client().Resources()).PodConditionMatch(pod, corev1.PodReady, corev1.ConditionTrue), wait.WithTimeout(time.Minute*5))
			if err != nil {
				t.Fatal(err)
			}
			return context.WithValue(ctx, ckTestPodLabel, pod)
		}).
		Assess("All not-satisfied", func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
			oscalPath := "./scenarios/pod-label/oscal-component-all-bad.yaml"
			findings, observations := validatePodLabelFail(ctx, t, oscalPath)
			observationRemarksMap := generateObservationRemarksMap(*observations)

			for _, f := range *findings {
				// related observations should have len = 1
				relatedObs := *f.RelatedObservations
				if f.RelatedObservations == nil || len(relatedObs) != 1 {
					t.Fatal("RelatedObservations should have len = 1")
				}
				remarks, found := observationRemarksMap[relatedObs[0].ObservationUuid]
				if !found {
					t.Fatal("RelatedObservation not found in map")
				}

				switch f.Target.TargetId {
				case "ID-1":
					if !strings.Contains(remarks, common.ErrInvalidDomain.Error()) {
						t.Fatal("ID-1 - Remarks should contain ErrInvalidDomain")
					}
				case "ID-1.1":
					if !strings.Contains(remarks, common.ErrInvalidProvider.Error()) {
						t.Fatal("ID-1.1 - Remarks should contain ErrInvalidProvider")
					}
				case "ID-2":
					if !strings.Contains(remarks, common.ErrInvalidSchema.Error()) {
						t.Fatal("ID-2 - Remarks should contain ErrInvalidSchema")
					}
				case "ID-3":
					if !strings.Contains(remarks, common.ErrInvalidYaml.Error()) {
						t.Fatal("ID-3 - Remarks should contain ErrInvalidYaml")
					}
				case "ID-3.1":
					if !strings.Contains(remarks, common.ErrInvalidYaml.Error()) {
						t.Fatal("ID-3.1 - Remarks should contain ErrInvalidYaml")
					}
				case "ID-4":
					if !strings.Contains(remarks, types.ErrProviderEvaluate.Error()) {
						t.Fatal("ID-4 - Remarks should contain ErrProviderEvaluate")
					}
				case "ID-5":
					if !strings.Contains(remarks, types.ErrDomainGetResources.Error()) {
						t.Fatal("ID-5 - Remarks should contain ErrDomainGetResources")
					}
				case "ID-5.1":
					if !strings.Contains(remarks, types.ErrDomainGetResources.Error()) {
						t.Fatal("ID-5.1 - Remarks should contain ErrDomainGetResources")
					}
				case "ID-5.2":
					if !strings.Contains(remarks, types.ErrDomainGetResources.Error()) {
						t.Fatal("ID-5.2 - Remarks should contain ErrDomainGetResources")
					}
				case "ID-6":
					if !strings.Contains(remarks, types.ErrExecutionNotAllowed.Error()) {
						t.Fatal("ID-6 - Remarks should contain ErrExecutionNotAllowed")
					}
				}
			}
			return ctx
		}).
		Teardown(func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
			pod := ctx.Value(ckTestPodLabel).(*corev1.Pod)
			if err := config.Client().Resources().Delete(ctx, pod); err != nil {
				t.Fatal(err)
			}
			err := wait.For(conditions.New(config.Client().Resources()).ResourceDeleted(pod), wait.WithTimeout(time.Minute*1))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).Feature()

	testEnv.Test(t, featureTrueValidation, featureFalseValidation, featureBadValidation)
}

func validatePodLabelPass(ctx context.Context, t *testing.T, oscalPath string) context.Context {
	message.NoProgress = true

	tempDir := t.TempDir()

	// Upgrade the component definition to latest oscal version
	revisionOptions := revision.RevisionOptions{
		InputFile:  oscalPath,
		OutputFile: tempDir + "/oscal-component-upgraded.yaml",
		Version:    versioning.GetLatestSupportedVersion(),
	}
	revisionResponse, err := revision.RevisionCommand(&revisionOptions)
	if err != nil {
		t.Fatal("Failed to upgrade component definition with: ", err)
	}
	// Write the upgraded component definition to a temp file
	err = files.WriteOutput(revisionResponse.RevisedBytes, revisionOptions.OutputFile)
	if err != nil {
		t.Fatal("Failed to write upgraded component definition with: ", err)
	}

	validator, err := validation.New()
	if err != nil {
		t.Errorf("error creating validation context: %v", err)
	}

	assessment, err := validator.ValidateOnPath(ctx, revisionOptions.OutputFile, "")
	if err != nil {
		t.Fatalf("Failed to validate oscal file: %s", revisionOptions.OutputFile)
	}

	if len(assessment.Model.Results) == 0 {
		t.Fatal("Expected greater than zero results")
	}

	result := assessment.Model.Results[0]

	if result.Findings == nil {
		t.Fatal("Expected findings to be not nil")
	}

	for _, finding := range *result.Findings {
		state := finding.Target.Status.State
		if state != "satisfied" {
			t.Fatal("State should be satisfied, but got :", state)
		}
	}

	// Test report generation
	report, err := oscal.GenerateAssessmentResults(assessment.Model.Results)
	if err != nil {
		t.Fatal("Failed generation of Assessment Results object with: ", err)
	}

	// Write the assessment results to file
	err = oscal.WriteOscalModelNew("sar-test.yaml", report)
	require.NoError(t, err)

	initialResultCount := len(report.Model.Results)

	//Perform the write operation again and read the file to ensure result was appended
	report, err = oscal.GenerateAssessmentResults(assessment.Model.Results)
	if err != nil {
		t.Fatal("Failed generation of Assessment Results object with: ", err)
	}

	// Get the UUID of the report results - there should only be one
	resultId := report.Model.Results[0].UUID

	// Write the assessment results to file
	err = oscal.WriteOscalModelNew("sar-test.yaml", report)
	require.NoError(t, err)

	data, err := os.ReadFile("sar-test.yaml")
	if err != nil {
		t.Fatal(err)
	}

	var tempAssessment oscal.AssessmentResults
	err = tempAssessment.NewModel(data)
	if err != nil {
		t.Fatal(err)
	}

	// The number of results in the file should be more than initially
	if len(tempAssessment.Model.Results) <= initialResultCount {
		t.Fatal("Failed to append results to existing report")
	}

	if resultId != tempAssessment.Model.Results[0].UUID {
		t.Fatal("Failed to prepend results to existing report")
	}

	validatorResponse, err := oscalValidation.ValidationCommand("sar-test.yaml")
	if err != nil || validatorResponse.JsonSchemaError != nil {
		t.Fatal("File failed linting")
	}

	return ctx
}

func validatePodLabelFail(ctx context.Context, t *testing.T, oscalPath string) (*[]oscalTypes.Finding, *[]oscalTypes.Observation) {
	message.NoProgress = true

	validator, err := validation.New(validation.WithAllowExecution(false, true))
	if err != nil {
		t.Fatalf("error creating validation context: %v", err)
	}

	assessment, err := validator.ValidateOnPath(ctx, oscalPath, "")
	if err != nil {
		t.Fatalf("Failed to validate oscal file: %s", oscalPath)
	}

	if len(assessment.Model.Results) == 0 {
		t.Fatal("Expected greater than zero results")
	}

	result := assessment.Model.Results[0]

	if result.Findings == nil {
		t.Fatal("Expected findings to be not nil")
	}

	for _, finding := range *result.Findings {
		state := finding.Target.Status.State
		if finding.Target.TargetId != "ID-3" {
			// This validation is an empty test that should always pass unless the underlying API changes
			if state != "not-satisfied" {
				t.Fatal("State should be not-satisfied, but got :", state)
			}
		}
	}
	return result.Findings, result.Observations
}

func generateObservationRemarksMap(observations []oscalTypes.Observation) map[string]string {
	observationMap := make(map[string]string, len(observations))

	for i := range observations {
		observation := &observations[i]
		relevantEvidence := strings.Builder{}
		for _, re := range *observation.RelevantEvidence {
			relevantEvidence.WriteString(re.Remarks)
		}
		observationMap[observation.UUID] = relevantEvidence.String()
	}

	return observationMap
}

func validateSaveResources(ctx context.Context, t *testing.T, oscalPath string) context.Context {
	message.NoProgress = true
	tempDir := t.TempDir()

	validator, err := validation.New(
		validation.WithOutputDir(tempDir),
		validation.WithSaveResources(true),
	)
	if err != nil {
		t.Errorf("error creating validation context: %v", err)
	}

	assessment, err := validator.ValidateOnPath(ctx, oscalPath, "")
	if err != nil {
		t.Fatalf("Failed to validate oscal file: %s", oscalPath)
	}

	if len(assessment.Model.Results) == 0 {
		t.Fatal("Expected greater than zero results")
	}

	result := assessment.Model.Results[0]

	// Check that remote files are created
	for _, o := range *result.Observations {
		if o.Links == nil {
			t.Fatal("Expected observation links, got nil")
		}
		if len(*o.Links) != 1 {
			t.Fatal("Expected 1 link, got ", len(*o.Links))
		}
		link := (*o.Links)[0]

		// The link is a relative path to assessment-results.yaml, so need to provide absolute path to the file
		dataBytes, err := network.Fetch(tempDir + strings.TrimPrefix(link.Href, "file://."))
		if err != nil {
			t.Fatal("Unable to fetch remote resource: ", err)
		}
		var data map[string]interface{}
		err = json.Unmarshal(dataBytes, &data)
		if err != nil {
			t.Fatal("Received invalid JSON: ", err)
		}
		// Check that resource data is as expected
		if !validaPodResourceData(data) {
			t.Fatal("Unexpected resource data found")
		}
	}

	// Write the assessment results to file
	err = oscal.WriteOscalModelNew(filepath.Join(tempDir, "assessment-results.yaml"), assessment)
	if err != nil {
		t.Fatal("error writing assessment results to file")
	}

	return ctx
}

func validaPodResourceData(data map[string]interface{}) bool {
	for k, v := range data {
		// Check for the expected fields
		if k == "podsvt" {
			vSlice := v.([]interface{})
			if vSlice[0].(map[string]interface{})["metadata"].(map[string]interface{})["name"] == "test-pod-label" {
				return true
			}
		}
		if k == "podvt" {
			if v.(map[string]interface{})["metadata"].(map[string]interface{})["name"] == "test-pod-label" {
				return true
			}
		}
		if k == "empty" {
			if len(v.([]interface{})) == 0 {
				return true
			}
		}
	}
	return false
}

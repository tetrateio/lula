package test

import (
	"context"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"

	"github.com/defenseunicorns/lula/src/pkg/common/composition"
	"github.com/defenseunicorns/lula/src/pkg/common/validation"
	"github.com/defenseunicorns/lula/src/test/util"
)

type compDefContextKey string

const componentDefinitionCompositionPodKey compDefContextKey = "component-definition-composition-pod"

func TestComponentDefinitionComposition(t *testing.T) {
	featureComponentDefinitionComposition := features.New("Check component definition composition").
		Setup(func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
			// Create the pod
			pod, err := util.GetPod("./scenarios/composition-component-definition/pod.pass.yaml")
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
			ctx = context.WithValue(ctx, componentDefinitionCompositionPodKey, pod)

			return ctx
		}).
		Assess("Validate local composition file", func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
			compDefPath := "../../test/unit/common/composition/component-definition-import-multi-compdef.yaml"

			validator, err := validation.New(validation.WithComposition(nil, compDefPath))
			if err != nil {
				t.Errorf("error creating validation context: %v", err)
			}

			assessment, err := validator.ValidateOnPath(context.Background(), compDefPath, "")
			if err != nil {
				t.Errorf("Error validating component definition: %v", err)
			}

			if assessment.Model.Results == nil {
				t.Fatal("Expected to have results")
			}

			results := assessment.Model.Results

			var expectedFindings, expectedObservations int
			expectedResults := len(results)

			for _, result := range results {
				if result.Findings == nil {
					t.Fatal("Expected to have findings")
				}
				// There should
				for _, finding := range *result.Findings {
					expectedFindings++
					// we expect to find two related observations for this control id
					if finding.Target.TargetId == "ID-1" {
						if finding.RelatedObservations == nil {
							t.Fatal("Expected related observations")
						}
						if len(*finding.RelatedObservations) < 2 {
							t.Errorf("Expected 2 related observations, found %v", len(*finding.RelatedObservations))
						}
					}

				}

				expectedObservations += len(*result.Observations)
			}

			if expectedFindings == 0 {
				t.Errorf("Expected to find findings")
			}

			if expectedObservations == 0 {
				t.Errorf("Expected to find observations")
			}

			// Compare validation results to a composed component definition
			composer, err := composition.New(composition.WithModelFromLocalPath(compDefPath))
			if err != nil {
				t.Errorf("error creating composition context: %v", err)
			}

			oscalModel, err := composer.ComposeFromPath(ctx, compDefPath)
			if err != nil {
				t.Error(err)
			}

			if oscalModel.ComponentDefinition == nil {
				t.Errorf("component definition is nil")
			}

			composeResults, err := validator.ValidateOnCompDef(context.Background(), oscalModel.ComponentDefinition, "")
			if err != nil {
				t.Error(err)
			}

			if len(composeResults) != expectedResults {
				t.Errorf("Expected %v results, got %v", expectedResults, len(composeResults))
			}

			var composeFindings, composeObservations int
			for _, result := range composeResults {
				composeFindings += len(*result.Findings)
				composeObservations += len(*result.Observations)
			}

			if composeFindings != expectedFindings {
				t.Errorf("Expected %d findings, got %d", expectedFindings, composeFindings)
			}

			if composeObservations != expectedObservations {
				t.Errorf("Expected %d observations, got %d", expectedObservations, composeObservations)
			}
			return ctx
		}).
		Teardown(func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {

			// Delete the pod
			pod := ctx.Value(componentDefinitionCompositionPodKey).(*corev1.Pod)
			if err := config.Client().Resources().Delete(ctx, pod); err != nil {
				t.Fatal(err)
			}
			return ctx
		}).Feature()

	testEnv.Test(t, featureComponentDefinitionComposition)
}

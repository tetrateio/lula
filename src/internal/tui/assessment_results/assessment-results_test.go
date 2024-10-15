package assessmentresults_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/defenseunicorns/lula/src/internal/testhelpers"
	assessmentresults "github.com/defenseunicorns/lula/src/internal/tui/assessment_results"
	"github.com/defenseunicorns/lula/src/pkg/common/result"
)

func TestGetResults(t *testing.T) {
	oscalModel := testhelpers.OscalFromPath(t, validAssessmentResultsMulti)
	results := assessmentresults.GetResults(oscalModel.AssessmentResults)

	require.Equal(t, 2, len(results))

	// Check summary data about each result - should be sorted deterministically by time (latest first)
	assert.Equal(t, 2, results[0].SummaryData.NumFindings)
	assert.Equal(t, 0, results[0].SummaryData.NumFindingsSatisfied)
	assert.Equal(t, 2, results[0].SummaryData.NumObservations)
	assert.Equal(t, 0, results[0].SummaryData.NumObservationsSatisfied)

	assert.Equal(t, 2, results[1].SummaryData.NumFindings)
	assert.Equal(t, 2, results[1].SummaryData.NumFindingsSatisfied)
	assert.Equal(t, 2, results[1].SummaryData.NumObservations)
	assert.Equal(t, 2, results[1].SummaryData.NumObservationsSatisfied)
}

func TestGetResultsComparison(t *testing.T) {

	t.Run("Simple Satisfied -> Not-Satisfied", func(t *testing.T) {
		oscalModel := testhelpers.OscalFromPath(t, validAssessmentResultsMulti)
		results := assessmentresults.GetResults(oscalModel.AssessmentResults)
		require.Equal(t, 2, len(results))

		// Not-Satisfied is new; Compared to Satisfied
		findingsRows, observationsRows := assessmentresults.GetResultComparison(results[0], results[1])
		require.Equal(t, 2, len(findingsRows))
		assert.Equal(t, result.SATISFIED_TO_NOT_SATISFIED, findingsRows[0].Data[assessmentresults.ColumnKeyStatusChange]) // Should be ID-1, was satisfied now not-satisfied
		assert.Equal(t, result.SATISFIED_TO_NOT_SATISFIED, findingsRows[1].Data[assessmentresults.ColumnKeyStatusChange]) // Should be ID-2, was satisfied now not-satisfied

		require.Equal(t, 2, len(observationsRows))
		assert.Equal(t, result.SATISFIED_TO_NOT_SATISFIED, observationsRows[0].Data[assessmentresults.ColumnKeyStatusChange]) // Linked to ID-1
		assert.Equal(t, result.SATISFIED_TO_NOT_SATISFIED, observationsRows[1].Data[assessmentresults.ColumnKeyStatusChange]) // Linked to ID-2
	})

	t.Run("Removed Finding", func(t *testing.T) {
		oscalModel := testhelpers.OscalFromPath(t, validAssessmentResultsRemovedFinding)
		results := assessmentresults.GetResults(oscalModel.AssessmentResults)
		require.Equal(t, 2, len(results))

		// Finding is removed, check both rows have the right status change
		findingsRows, observationsRows := assessmentresults.GetResultComparison(results[0], results[1])
		require.Equal(t, 2, len(findingsRows))
		assert.Equal(t, result.SATISFIED_TO_NOT_SATISFIED, findingsRows[0].Data[assessmentresults.ColumnKeyStatusChange]) // Should be ID-1
		assert.Equal(t, result.REMOVED, findingsRows[1].Data[assessmentresults.ColumnKeyStatusChange])                    // Should be ID-2, removed

		require.Equal(t, 2, len(observationsRows))
		assert.Equal(t, result.SATISFIED_TO_NOT_SATISFIED, observationsRows[0].Data[assessmentresults.ColumnKeyStatusChange]) // Should be linked to ID-1
		assert.Equal(t, result.REMOVED, observationsRows[1].Data[assessmentresults.ColumnKeyStatusChange])                    // Should be linked to ID-2, removed
	})

	t.Run("Added Finding", func(t *testing.T) {
		oscalModel := testhelpers.OscalFromPath(t, validAssessmentResultsAddedFinding)
		results := assessmentresults.GetResults(oscalModel.AssessmentResults)
		require.Equal(t, 2, len(results))

		// Finding is added, check both rows have the right status change
		findingsRows, observationsRows := assessmentresults.GetResultComparison(results[0], results[1])
		require.Equal(t, 2, len(findingsRows))
		assert.Equal(t, result.SATISFIED_TO_NOT_SATISFIED, findingsRows[0].Data[assessmentresults.ColumnKeyStatusChange]) // Should be ID-1
		assert.Equal(t, result.NEW, findingsRows[1].Data[assessmentresults.ColumnKeyStatusChange])                        // Should be ID-2, added

		require.Equal(t, 2, len(observationsRows))
		assert.Equal(t, result.SATISFIED_TO_NOT_SATISFIED, observationsRows[0].Data[assessmentresults.ColumnKeyStatusChange]) // Should be linked to ID-1
		assert.Equal(t, result.NEW, observationsRows[1].Data[assessmentresults.ColumnKeyStatusChange])                        // Should be linked to ID-2, added
	})

	t.Run("Removed Observation", func(t *testing.T) {
		oscalModel := testhelpers.OscalFromPath(t, validAssessmentResultsRemovedObs)
		results := assessmentresults.GetResults(oscalModel.AssessmentResults)
		require.Equal(t, 2, len(results))

		// Finding is removed, check both rows have the right status change
		findingsRows, observationsRows := assessmentresults.GetResultComparison(results[0], results[1])
		require.Equal(t, 1, len(findingsRows))
		assert.Equal(t, result.SATISFIED_TO_NOT_SATISFIED, findingsRows[0].Data[assessmentresults.ColumnKeyStatusChange]) // Should be ID-1

		require.Equal(t, 2, len(observationsRows))
		assert.Equal(t, result.SATISFIED_TO_NOT_SATISFIED, observationsRows[0].Data[assessmentresults.ColumnKeyStatusChange]) // Should be linked to ID-1
		assert.Equal(t, result.REMOVED, observationsRows[1].Data[assessmentresults.ColumnKeyStatusChange])                    // Should be linked to ID-1, removed
	})

}

func TestGetReadableDesc(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		desc     string
		expected string
	}{
		{
			name:     "Test get readable desc",
			desc:     "[TEST]: 67456ae8-4505-4c93-b341-d977d90cb125 - istio-health-check",
			expected: "istio-health-check",
		},
		{
			name:     "Test get readable desc - no uuid",
			desc:     "test description",
			expected: "test description",
		},
		{
			name:     "Test get readable desc - no description",
			desc:     "[TEST]: 12345678-1234-1234-1234-123456789012",
			expected: "[TEST]: 12345678-1234-1234-1234-123456789012",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := assessmentresults.GetReadableObservationName(tt.desc)
			assert.Equal(t, tt.expected, got)
		})
	}
}

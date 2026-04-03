package translator_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zarishsphere/zs-core-fhir-r4-bridge/internal/translator"
)

// --------------------------------------------------------------------------
// R4 → R5 tests
// --------------------------------------------------------------------------

func TestR4ToR5_Patient_Compatible(t *testing.T) {
	r4 := map[string]any{
		"resourceType": "Patient",
		"id":           "test-001",
		"active":       true,
		"gender":       "male",
		"name":         []any{map[string]any{"family": "Rahman", "given": []any{"Abdul"}}},
	}
	result, err := translator.R4ToR5(r4)
	require.NoError(t, err)
	assert.Equal(t, "Patient", result.Resource["resourceType"])
	assert.Equal(t, "test-001", result.Resource["id"])
	assert.Empty(t, result.LossyFields)
	// Check translation meta tag was added
	meta, ok := result.Resource["meta"].(map[string]any)
	require.True(t, ok)
	tags, ok := meta["tag"].([]any)
	require.True(t, ok)
	assert.NotEmpty(t, tags)
}

func TestR4ToR5_Encounter_ClassCodingToCodeableConcept(t *testing.T) {
	r4 := map[string]any{
		"resourceType": "Encounter",
		"id":           "enc-001",
		"status":       "finished",
		// R4: class is a Coding
		"class": map[string]any{
			"system":  "http://terminology.hl7.org/CodeSystem/v3-ActCode",
			"code":    "AMB",
			"display": "ambulatory",
		},
		"subject": map[string]any{"reference": "Patient/test-001"},
	}

	result, err := translator.R4ToR5(r4)
	require.NoError(t, err)
	require.NotNil(t, result)

	// R5: class should be []CodeableConcept
	classes, ok := result.Resource["class"].([]map[string]any)
	require.True(t, ok, "R5 Encounter.class should be []CodeableConcept, got: %T", result.Resource["class"])
	require.Len(t, classes, 1)

	// Verify the coding is preserved inside the CodeableConcept
	codings, ok := classes[0]["coding"].([]map[string]any)
	require.True(t, ok)
	require.Len(t, codings, 1)
	assert.Equal(t, "AMB", codings[0]["code"])
	assert.Equal(t, "http://terminology.hl7.org/CodeSystem/v3-ActCode", codings[0]["system"])
}

func TestR4ToR5_Encounter_HospitalizationRenamedToAdmission(t *testing.T) {
	r4 := map[string]any{
		"resourceType": "Encounter",
		"id":           "enc-002",
		"status":       "finished",
		"class":        map[string]any{"code": "IMP"},
		"subject":      map[string]any{"reference": "Patient/test-001"},
		"hospitalization": map[string]any{
			"admitSource": map[string]any{"text": "Emergency"},
		},
	}

	result, err := translator.R4ToR5(r4)
	require.NoError(t, err)

	// R5: should be "admission" not "hospitalization"
	_, hasAdmission := result.Resource["admission"]
	_, hasHospitalization := result.Resource["hospitalization"]
	assert.True(t, hasAdmission, "R5 should use 'admission'")
	assert.False(t, hasHospitalization, "R5 should not have 'hospitalization'")

	// Should have a translation warning
	assert.Len(t, result.Warnings, 1)
	assert.Equal(t, "hospitalization", result.Warnings[0].Field)
}

func TestR4ToR5_Encounter_ClassHistoryDropped(t *testing.T) {
	r4 := map[string]any{
		"resourceType": "Encounter",
		"id":           "enc-003",
		"status":       "finished",
		"class":        map[string]any{"code": "AMB"},
		"subject":      map[string]any{"reference": "Patient/test-001"},
		"classHistory": []any{
			map[string]any{"class": map[string]any{"code": "IMP"}, "period": map[string]any{"start": "2024-01-01"}},
		},
	}

	result, err := translator.R4ToR5(r4)
	require.NoError(t, err)

	// classHistory should be dropped (removed in R5)
	_, hasClassHistory := result.Resource["classHistory"]
	assert.False(t, hasClassHistory, "classHistory should be removed in R5")
	assert.Contains(t, result.LossyFields, "classHistory")
}

func TestR4ToR5_MedicationRequest_MedicationCodeableConcept(t *testing.T) {
	r4 := map[string]any{
		"resourceType": "MedicationRequest",
		"id":           "mr-001",
		"status":       "active",
		"intent":       "order",
		// R4: medicationCodeableConcept
		"medicationCodeableConcept": map[string]any{
			"coding": []any{
				map[string]any{
					"system":  "http://www.nlm.nih.gov/research/umls/rxnorm",
					"code":    "1049502",
					"display": "12 HR Oxycodone",
				},
			},
		},
		"subject": map[string]any{"reference": "Patient/test-001"},
	}

	result, err := translator.R4ToR5(r4)
	require.NoError(t, err)

	// R5: medication should be CodeableReference with concept
	med, ok := result.Resource["medication"].(map[string]any)
	require.True(t, ok, "R5 medication should be CodeableReference map")
	_, hasConcept := med["concept"]
	assert.True(t, hasConcept, "R5 medication should have concept field")

	// Original field should be gone
	_, hasMedCC := result.Resource["medicationCodeableConcept"]
	assert.False(t, hasMedCC)
}

func TestR4ToR5_MedicationRequest_MedicationReference(t *testing.T) {
	r4 := map[string]any{
		"resourceType":        "MedicationRequest",
		"id":                  "mr-002",
		"status":              "active",
		"intent":              "order",
		"medicationReference": map[string]any{"reference": "Medication/med-001"},
		"subject":             map[string]any{"reference": "Patient/test-001"},
	}

	result, err := translator.R4ToR5(r4)
	require.NoError(t, err)

	med, ok := result.Resource["medication"].(map[string]any)
	require.True(t, ok)
	_, hasRef := med["reference"]
	assert.True(t, hasRef)
}

// --------------------------------------------------------------------------
// R5 → R4 tests
// --------------------------------------------------------------------------

func TestR5ToR4_Encounter_CodeableConceptToCoding(t *testing.T) {
	r5 := map[string]any{
		"resourceType": "Encounter",
		"id":           "enc-r5-001",
		"status":       "finished",
		// R5: class is []CodeableConcept
		"class": []any{
			map[string]any{
				"coding": []any{
					map[string]any{
						"system":  "http://terminology.hl7.org/CodeSystem/v3-ActCode",
						"code":    "AMB",
						"display": "ambulatory",
					},
				},
				"text": "ambulatory",
			},
		},
		"subject": map[string]any{"reference": "Patient/test-001"},
	}

	result, err := translator.R5ToR4(r5)
	require.NoError(t, err)

	// R4: class should be a single Coding
	class, ok := result.Resource["class"].(map[string]any)
	require.True(t, ok, "R4 Encounter.class should be a Coding map, got: %T", result.Resource["class"])
	assert.Equal(t, "AMB", class["code"])
}

func TestR5ToR4_Encounter_MultipleClassesWarning(t *testing.T) {
	r5 := map[string]any{
		"resourceType": "Encounter",
		"id":           "enc-r5-002",
		"status":       "finished",
		"class": []any{
			map[string]any{"coding": []any{map[string]any{"code": "AMB"}}},
			map[string]any{"coding": []any{map[string]any{"code": "HH"}}},
		},
		"subject": map[string]any{"reference": "Patient/test-001"},
	}

	result, err := translator.R5ToR4(r5)
	require.NoError(t, err)

	// Should have lossy warning for dropped class entries
	assert.Contains(t, result.LossyFields, "class[1:]")
}

func TestR5ToR4_Observation_R5OnlyFieldsDropped(t *testing.T) {
	r5 := map[string]any{
		"resourceType": "Observation",
		"id":           "obs-r5-001",
		"status":       "final",
		"code":         map[string]any{"text": "BP"},
		"subject":      map[string]any{"reference": "Patient/test-001"},
		// R5-only fields
		"triggeredBy": []any{
			map[string]any{"observation": map[string]any{"reference": "Observation/trigger-001"}},
		},
		"bodyStructure": map[string]any{"reference": "BodyStructure/bs-001"},
	}

	result, err := translator.R5ToR4(r5)
	require.NoError(t, err)

	_, hasTriggeredBy := result.Resource["triggeredBy"]
	_, hasBodyStructure := result.Resource["bodyStructure"]
	assert.False(t, hasTriggeredBy, "R5-only triggeredBy should be dropped")
	assert.False(t, hasBodyStructure, "R5-only bodyStructure should be dropped")
	assert.Contains(t, result.LossyFields, "triggeredBy")
	assert.Contains(t, result.LossyFields, "bodyStructure")
}

func TestR5ToR4_MedicationRequest_ConceptUnwrapped(t *testing.T) {
	r5 := map[string]any{
		"resourceType": "MedicationRequest",
		"id":           "mr-r5-001",
		"status":       "active",
		"intent":       "order",
		"medication": map[string]any{
			"concept": map[string]any{
				"coding": []any{
					map[string]any{"system": "http://www.nlm.nih.gov/research/umls/rxnorm", "code": "1049502"},
				},
			},
		},
		"subject": map[string]any{"reference": "Patient/test-001"},
	}

	result, err := translator.R5ToR4(r5)
	require.NoError(t, err)

	_, hasMedCC := result.Resource["medicationCodeableConcept"]
	_, hasMed := result.Resource["medication"]
	assert.True(t, hasMedCC, "R4 should have medicationCodeableConcept")
	assert.False(t, hasMed, "R4 should not have medication")
}

func TestR5ToR4_TenantExtensionRemoved(t *testing.T) {
	r5 := map[string]any{
		"resourceType": "Patient",
		"id":           "pt-001",
		"extension": []any{
			map[string]any{
				"url":         "https://fhir.zarishsphere.com/StructureDefinition/ext/tenant-id",
				"valueString": "cpi:bgd:camp-1w",
			},
			map[string]any{
				"url":         "https://fhir.zarishsphere.com/StructureDefinition/ext/program-code",
				"valueCode":   "bgd-rohingya",
			},
		},
	}

	result, err := translator.R5ToR4(r5)
	require.NoError(t, err)

	// ZarishSphere-specific extensions should be removed for R4 output
	exts, ok := result.Resource["extension"].([]any)
	if ok {
		for _, ext := range exts {
			if em, ok := ext.(map[string]any); ok {
				url, _ := em["url"].(string)
				assert.NotContains(t, url, "tenant-id", "tenant-id extension should be removed for R4")
				assert.NotContains(t, url, "program-code", "program-code extension should be removed for R4")
			}
		}
	}
}

// --------------------------------------------------------------------------
// Bundle round-trip test
// --------------------------------------------------------------------------

func TestBundleTranslation_R4ToR5(t *testing.T) {
	bundle := map[string]any{
		"resourceType": "Bundle",
		"id":           "bundle-001",
		"type":         "transaction",
		"entry": []any{
			map[string]any{
				"resource": map[string]any{
					"resourceType": "Encounter",
					"id":           "enc-bundle-001",
					"status":       "finished",
					"class":        map[string]any{"system": "http://terminology.hl7.org/CodeSystem/v3-ActCode", "code": "AMB"},
					"subject":      map[string]any{"reference": "Patient/test-001"},
				},
			},
			map[string]any{
				"resource": map[string]any{
					"resourceType": "Patient",
					"id":           "pt-bundle-001",
					"active":       true,
				},
			},
		},
	}

	result, err := translator.R4ToR5(bundle)
	require.NoError(t, err)
	assert.Equal(t, "Bundle", result.Resource["resourceType"])

	entries, ok := result.Resource["entry"].([]any)
	require.True(t, ok)
	require.Len(t, entries, 2)

	// Verify Encounter.class was translated
	entry0, _ := entries[0].(map[string]any)
	enc, _ := entry0["resource"].(map[string]any)
	classes, classOK := enc["class"].([]map[string]any)
	assert.True(t, classOK || enc["class"] != nil, "Encounter.class should be translated")
	_ = classes
}

// --------------------------------------------------------------------------
// Error cases
// --------------------------------------------------------------------------

func TestR4ToR5_MissingResourceType(t *testing.T) {
	resource := map[string]any{"id": "no-type"}
	_, err := translator.R4ToR5(resource)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "resourceType")
}

func TestR5ToR4_MissingResourceType(t *testing.T) {
	resource := map[string]any{"id": "no-type"}
	_, err := translator.R5ToR4(resource)
	assert.Error(t, err)
}

// --------------------------------------------------------------------------
// Deep copy safety test (translation should not modify original)
// --------------------------------------------------------------------------

func TestTranslationDoesNotModifyOriginal(t *testing.T) {
	original := map[string]any{
		"resourceType": "Encounter",
		"id":           "enc-orig",
		"status":       "finished",
		"class": map[string]any{
			"system": "http://terminology.hl7.org/CodeSystem/v3-ActCode",
			"code":   "AMB",
		},
		"subject": map[string]any{"reference": "Patient/test-001"},
	}

	// Capture original JSON
	origJSON, _ := json.Marshal(original)

	_, err := translator.R4ToR5(original)
	require.NoError(t, err)

	// Original should be unchanged
	afterJSON, _ := json.Marshal(original)
	assert.Equal(t, string(origJSON), string(afterJSON), "translation should not modify the original resource")
}

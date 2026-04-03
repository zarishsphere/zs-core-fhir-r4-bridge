// Package translator implements bidirectional FHIR R4 ↔ R5 resource translation.
//
// ZarishSphere is FHIR R5-native. Partner systems (DHIS2, OpenMRS, legacy HIS)
// predominantly use FHIR R4. This bridge translates between the two versions.
//
// Key structural differences translated:
//   - Encounter.class:        R4=Coding        → R5=[]CodeableConcept
//   - Encounter.hospitalization: R4 name       → R5=admission
//   - MedicationRequest.medication: R4=medication[x] → R5=CodeableReference
//   - Observation.triggeredBy: R5-only (dropped on R5→R4)
//   - Condition.participant:   R5-only (dropped on R5→R4)
//   - Bundle entries:          translated recursively
//
// All translated resources receive a meta.tag identifying translation origin.
// Lossy fields are documented in TranslationResult.LossyFields.
package translator

import (
	"fmt"

	"github.com/rs/zerolog/log"
	zsfhir "github.com/zarishsphere/zs-pkg-go-fhir/pkg/fhir"
)

const (
	VersionR4 = "4.0.1"
	VersionR5 = "5.0.0"
)

// TranslationWarning records a field that could not be fully translated.
type TranslationWarning struct {
	Field   string
	Message string
}

// TranslationResult holds the translated resource and any warnings.
type TranslationResult struct {
	Resource    map[string]any
	Warnings    []TranslationWarning
	LossyFields []string
}

func (r *TranslationResult) warn(field, msg string) {
	r.Warnings = append(r.Warnings, TranslationWarning{Field: field, Message: msg})
}

func (r *TranslationResult) lossy(field string) {
	r.LossyFields = append(r.LossyFields, field)
}

// R4ToR5 translates a FHIR R4 resource map to FHIR R5.
func R4ToR5(r4 map[string]any) (*TranslationResult, error) {
	rt, ok := r4["resourceType"].(string)
	if !ok {
		return nil, fmt.Errorf("r4bridge: missing resourceType")
	}

	result := &TranslationResult{Resource: deepCopy(r4)}
	applyCommonR4ToR5(result)

	switch rt {
	case "Patient":
		translatePatientR4ToR5(result)
	case "Encounter":
		translateEncounterR4ToR5(result)
	case "Observation":
		translateObservationR4ToR5(result)
	case "Condition":
		translateConditionR4ToR5(result)
	case "MedicationRequest":
		translateMedicationRequestR4ToR5(result)
	case "Immunization":
		// R4 compatible
	case "Bundle":
		translateBundleR4ToR5(result)
	default:
		log.Debug().Str("rt", rt).Msg("r4bridge: generic R4→R5 transform")
	}

	markAsTranslated(result.Resource, VersionR4, VersionR5)
	if len(result.Warnings) > 0 {
		log.Warn().Int("warnings", len(result.Warnings)).Strs("lossy", result.LossyFields).Msg("r4bridge: R4→R5 warnings")
	}
	return result, nil
}

// R5ToR4 translates a FHIR R5 resource map to FHIR R4.
func R5ToR4(r5 map[string]any) (*TranslationResult, error) {
	rt, ok := r5["resourceType"].(string)
	if !ok {
		return nil, fmt.Errorf("r4bridge: missing resourceType")
	}

	result := &TranslationResult{Resource: deepCopy(r5)}
	applyCommonR5ToR4(result)

	switch rt {
	case "Patient":
		translatePatientR5ToR4(result)
	case "Encounter":
		translateEncounterR5ToR4(result)
	case "Observation":
		translateObservationR5ToR4(result)
	case "Condition":
		translateConditionR5ToR4(result)
	case "MedicationRequest":
		translateMedicationRequestR5ToR4(result)
	case "Immunization":
		translateImmunizationR5ToR4(result)
	case "Bundle":
		translateBundleR5ToR4(result)
	default:
		log.Debug().Str("rt", rt).Msg("r4bridge: generic R5→R4 transform")
	}

	markAsTranslated(result.Resource, VersionR5, VersionR4)
	return result, nil
}

// --------------------------------------------------------------------------
// Common
// --------------------------------------------------------------------------

func applyCommonR4ToR5(result *TranslationResult) {
	r := result.Resource
	if meta, ok := r["meta"].(map[string]any); ok {
		if _, ok := meta["lastUpdated"]; !ok {
			meta["lastUpdated"] = zsfhir.NowUTC()
		}
	} else {
		r["meta"] = map[string]any{"lastUpdated": zsfhir.NowUTC()}
	}
}

func applyCommonR5ToR4(result *TranslationResult) {
	r := result.Resource
	removeExtension(r, zsfhir.ExtTenantID)
	removeExtension(r, zsfhir.ExtProgramCode)
}

// --------------------------------------------------------------------------
// Patient
// --------------------------------------------------------------------------

func translatePatientR4ToR5(_ *TranslationResult)  {} // compatible
func translatePatientR5ToR4(result *TranslationResult) {
	removeExtensionSuffix(result.Resource, "patient-pronouns")
}

// --------------------------------------------------------------------------
// Encounter — biggest structural change R4↔R5
// --------------------------------------------------------------------------

func translateEncounterR4ToR5(result *TranslationResult) {
	r := result.Resource
	// class: Coding → []CodeableConcept
	if classR4, ok := r["class"].(map[string]any); ok {
		sys, _ := classR4["system"].(string)
		code, _ := classR4["code"].(string)
		disp, _ := classR4["display"].(string)
		r["class"] = []map[string]any{zsfhir.CodeableConcept(sys, code, disp, disp)}
	}
	// hospitalization → admission
	if hosp, ok := r["hospitalization"]; ok {
		r["admission"] = hosp
		delete(r, "hospitalization")
		result.warn("hospitalization", "renamed to 'admission' in R5")
	}
	// classHistory removed in R5
	if _, ok := r["classHistory"]; ok {
		delete(r, "classHistory")
		result.lossy("classHistory")
		result.warn("classHistory", "removed in R5 — data dropped")
	}
}

func translateEncounterR5ToR4(result *TranslationResult) {
	r := result.Resource
	// class: []CodeableConcept → Coding (take first coding of first entry)
	if classes, ok := r["class"].([]any); ok && len(classes) > 0 {
		if cc, ok := classes[0].(map[string]any); ok {
			if codings, ok := cc["coding"].([]any); ok && len(codings) > 0 {
				r["class"] = codings[0]
			}
		}
		if len(classes) > 1 {
			result.lossy("class[1:]")
			result.warn("class", fmt.Sprintf("%d class entries; R4 takes first only", len(classes)))
		}
	}
	// admission → hospitalization
	if adm, ok := r["admission"]; ok {
		r["hospitalization"] = adm
		delete(r, "admission")
	}
}

// --------------------------------------------------------------------------
// Observation
// --------------------------------------------------------------------------

func translateObservationR4ToR5(_ *TranslationResult) {} // compatible

func translateObservationR5ToR4(result *TranslationResult) {
	r := result.Resource
	for _, f := range []string{"triggeredBy", "bodyStructure"} {
		if _, ok := r[f]; ok {
			delete(r, f)
			result.lossy(f)
		}
	}
}

// --------------------------------------------------------------------------
// Condition
// --------------------------------------------------------------------------

func translateConditionR4ToR5(_ *TranslationResult) {} // compatible

func translateConditionR5ToR4(result *TranslationResult) {
	r := result.Resource
	for _, f := range []string{"participant", "bodyStructure"} {
		if _, ok := r[f]; ok {
			delete(r, f)
			result.lossy(f)
		}
	}
}

// --------------------------------------------------------------------------
// MedicationRequest
// --------------------------------------------------------------------------

func translateMedicationRequestR4ToR5(result *TranslationResult) {
	r := result.Resource
	if medCC, ok := r["medicationCodeableConcept"]; ok {
		r["medication"] = map[string]any{"concept": medCC}
		delete(r, "medicationCodeableConcept")
	} else if medRef, ok := r["medicationReference"]; ok {
		r["medication"] = map[string]any{"reference": medRef}
		delete(r, "medicationReference")
	}
}

func translateMedicationRequestR5ToR4(result *TranslationResult) {
	r := result.Resource
	if med, ok := r["medication"].(map[string]any); ok {
		if concept, ok := med["concept"]; ok {
			r["medicationCodeableConcept"] = concept
		} else if ref, ok := med["reference"]; ok {
			r["medicationReference"] = ref
		}
		delete(r, "medication")
	}
}

// --------------------------------------------------------------------------
// Immunization
// --------------------------------------------------------------------------

func translateImmunizationR5ToR4(result *TranslationResult) {
	r := result.Resource
	for _, f := range []string{"programEligibility", "supportingInformation"} {
		if _, ok := r[f]; ok {
			delete(r, f)
			result.lossy(f)
		}
	}
}

// --------------------------------------------------------------------------
// Bundle — translate contained resources recursively
// --------------------------------------------------------------------------

func translateBundleR4ToR5(result *TranslationResult) {
	translateBundleEntries(result, func(res map[string]any) (map[string]any, *TranslationResult, error) {
		return translateEntry(res, R4ToR5)
	})
}

func translateBundleR5ToR4(result *TranslationResult) {
	translateBundleEntries(result, func(res map[string]any) (map[string]any, *TranslationResult, error) {
		return translateEntry(res, R5ToR4)
	})
}

type translateFn func(map[string]any) (*TranslationResult, error)

func translateEntry(res map[string]any, fn translateFn) (map[string]any, *TranslationResult, error) {
	tr, err := fn(res)
	if err != nil {
		return res, nil, err
	}
	return tr.Resource, tr, nil
}

func translateBundleEntries(result *TranslationResult, fn func(map[string]any) (map[string]any, *TranslationResult, error)) {
	entries, ok := result.Resource["entry"].([]any)
	if !ok {
		return
	}
	for i, entry := range entries {
		em, ok := entry.(map[string]any)
		if !ok {
			continue
		}
		res, ok := em["resource"].(map[string]any)
		if !ok {
			continue
		}
		translated, subResult, err := fn(res)
		if err == nil {
			em["resource"] = translated
			if subResult != nil {
				result.Warnings = append(result.Warnings, subResult.Warnings...)
				result.LossyFields = append(result.LossyFields, subResult.LossyFields...)
			}
		}
		entries[i] = em
	}
}

// --------------------------------------------------------------------------
// Utilities
// --------------------------------------------------------------------------

func markAsTranslated(r map[string]any, from, to string) {
	meta, ok := r["meta"].(map[string]any)
	if !ok {
		meta = map[string]any{}
		r["meta"] = meta
	}
	tags, _ := meta["tag"].([]any)
	tags = append(tags, map[string]any{
		"system":  "https://fhir.zarishsphere.com/CodeSystem/translation-source",
		"code":    "fhir-" + from + "-to-" + to,
		"display": fmt.Sprintf("Translated from FHIR %s to %s by zs-core-fhir-r4-bridge", from, to),
	})
	meta["tag"] = tags
}

func removeExtension(r map[string]any, url string) {
	exts, ok := r["extension"].([]any)
	if !ok {
		return
	}
	out := exts[:0]
	for _, ext := range exts {
		if em, ok := ext.(map[string]any); ok && em["url"] == url {
			continue
		}
		out = append(out, ext)
	}
	r["extension"] = out
}

func removeExtensionSuffix(r map[string]any, suffix string) {
	exts, ok := r["extension"].([]any)
	if !ok {
		return
	}
	out := exts[:0]
	for _, ext := range exts {
		if em, ok := ext.(map[string]any); ok {
			if url, _ := em["url"].(string); len(url) >= len(suffix) && url[len(url)-len(suffix):] == suffix {
				continue
			}
		}
		out = append(out, ext)
	}
	r["extension"] = out
}

func deepCopy(src map[string]any) map[string]any {
	dst := make(map[string]any, len(src))
	for k, v := range src {
		switch val := v.(type) {
		case map[string]any:
			dst[k] = deepCopy(val)
		case []any:
			arr := make([]any, len(val))
			for i, item := range val {
				if m, ok := item.(map[string]any); ok {
					arr[i] = deepCopy(m)
				} else {
					arr[i] = item
				}
			}
			dst[k] = arr
		default:
			dst[k] = v
		}
	}
	return dst
}

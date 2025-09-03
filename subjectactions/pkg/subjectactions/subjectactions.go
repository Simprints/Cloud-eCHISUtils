package subjectactions

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Check is a simple utility function that checks the given input string is a valid subjectActions that can be parsed by SID.
// It returns a SubjectSpecification for the subject and a nil error if the input string is valid.
// It returns a non-nil error when the input string is invalid.
func Check(input string) (SubjectSpecification, error) {
	if len(input) == 0 {
		return SubjectSpecification{}, errors.New("input is empty")
	}

	var sidEvents SIDEvents
	if err := json.Unmarshal([]byte(input), &sidEvents); err != nil {
		return SubjectSpecification{}, errors.Join(errors.New("could not unmarshal input into SIDEvents"), err)
	}
	if len(sidEvents.Events) != 1 {
		return SubjectSpecification{}, fmt.Errorf("expect 1 event, got %d", len(sidEvents.Events))
	}
	if sidEvents.Events[0].Type != RequiredSIDEventType {
		return SubjectSpecification{}, fmt.Errorf("expect event type %s event, got %s", RequiredSIDEventType, sidEvents.Events[0].Type)
	}

	payload := sidEvents.Events[0].Payload
	if len(payload.SubjectID) == 0 {
		return SubjectSpecification{}, errors.New("empty subjectId")
	}
	if len(payload.ProjectID) == 0 {
		return SubjectSpecification{}, errors.New("empty projectId")
	}
	if payload.ModuleID.ClassName != RequiredStringValueClassName {
		return SubjectSpecification{}, fmt.Errorf("expect moduleId to have className %s but got %s", RequiredStringValueClassName, payload.ModuleID.ClassName)
	}
	if len(payload.ModuleID.Value) == 0 {
		return SubjectSpecification{}, errors.New("empty moduleId")
	}
	if payload.AttendantID.ClassName != RequiredStringValueClassName {
		return SubjectSpecification{}, fmt.Errorf("expect attendantId to have className %s but got %s", RequiredStringValueClassName, payload.AttendantID.ClassName)
	}
	if len(payload.AttendantID.Value) == 0 {
		return SubjectSpecification{}, errors.New("empty attendantId")
	}

	if len(payload.BiometricReferences) != 1 {
		return SubjectSpecification{}, fmt.Errorf("expect 1 biometric reference, but got %d", len(payload.BiometricReferences))
	}
	reference := payload.BiometricReferences[0]
	if reference.Type != RequiredSIDBiometricReferenceType {
		return SubjectSpecification{}, fmt.Errorf("expect biometric reference type %s, got %s", RequiredSIDBiometricReferenceType, reference.Type)
	}
	if reference.Format != RequiredSIDBiometricReferenceFormat {
		return SubjectSpecification{}, fmt.Errorf("expect biometric reference format %s, got %s", RequiredSIDBiometricReferenceFormat, reference.Format)
	}

	subjectSpecification := SubjectSpecification{
		SubjectID:           payload.SubjectID,
		ProjectID:           payload.ProjectID,
		TokenizedModuleID:   payload.ModuleID.Value,
		TokenizedAtendantID: payload.AttendantID.Value,
	}
	return subjectSpecification, nil
}

// SubjectSpecification contains details for a given subject.
type SubjectSpecification struct {
	// SubjectID is the ID of the subject.
	SubjectID string
	// ProjectID is the ID of the Simprints project.
	ProjectID string
	// TokenizedModuleID is the tokenized (encrypted) version of the Simprints module ID, but it has not been validated.
	TokenizedModuleID string
	// TokenizedAttendantID is the tokenized (encrypted) version of the Simprints attendant ID, but it has not been validated.
	TokenizedAtendantID string
}

// SID expects the following schema (where '*' denotes some array index)
// For this project, we expect only 1 EnrolmentRecordCreation event.
// For this project, we only expect ISO templates for the fingerprint modality.
//
// > events.*.id
// > events.*.type (must always be 'EnrolmentRecordCreation')
// > events.*.payload.projectId
// > events.*.payload.subjectId
// > events.*.payload.moduleId.className (should be 'TokenizableString.Tokenized')
// > events.*.payload.moduleId.value (this field should be tokenized)
// > events.*.payload.attendantId.className (should be 'TokenizableString.Tokenized')
// > events.*.payload.attendantId.value (this field should be tokenized)
// > events.*.payload.biometricReferences.*.format (must always be 'ISO_19794_2')
// > events.*.payload.biometricReferences.*.id
// > events.*.payload.biometricReferences.*.type (must always be 'FINGERPRINT_REFERENCE')
// > events.*.payload.biometricReferences.*.templates.*.finger
// > events.*.payload.biometricReferences.*.templates.*.quality
// > events.*.payload.biometricReferences.*.templates.*.template

const (
	RequiredSIDEventType                = "EnrolmentRecordCreation"
	RequiredStringValueClassName        = "TokenizableString.Tokenized"
	RequiredSIDBiometricReferenceType   = "FINGERPRINT_REFERENCE"
	RequiredSIDBiometricReferenceFormat = "ISO_19794_2"
)

type SIDEvents struct {
	Events []SIDEvent `json:"events"`
}

type SIDEvent struct {
	ID      string                            `json:"id"`
	Type    string                            `json:"type"`
	Payload SIDEnrolmentRecordCreationPayload `json:"payload"`
}

type SIDEnrolmentRecordCreationPayload struct {
	SubjectID           string                  `json:"subjectId"`
	ProjectID           string                  `json:"projectId"`
	ModuleID            StringValue             `json:"moduleId"`
	AttendantID         StringValue             `json:"attendantId"`
	BiometricReferences []SIDBiometricReference `json:"biometricReferences"`
}

type StringValue struct {
	ClassName string `json:"className"`
	Value     string `json:"value"`
}

type SIDBiometricReference struct {
	ID        string        `json:"id"`
	Templates []SIDTemplate `json:"templates"`
	Format    string        `json:"format"`
	Type      string        `json:"type"`
}

type SIDTemplate struct {
	Quality  float64 `json:"quality"`
	Template string  `json:"template"`
	Finger   string  `json:"finger"`
}

package subjectactions

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
)

// Option customises the validation performed by Check.
type Option func(*options)

type options struct {
	allowedFormats []string
}

// WithAllowedFormats restricts the biometric reference formats accepted by Check (e.g. "ISO_19794_2", "RANK_ONE_3_1").
// SID only matches against references whose format is the one configured for the project, so this should be set to the
// format(s) of the matcher(s) configured for the project. When not set, any non-empty format is accepted.
func WithAllowedFormats(formats ...string) Option {
	return func(o *options) {
		o.allowedFormats = formats
	}
}

// Check is a simple utility function that checks the given input string is a valid subjectActions that can be parsed by SID.
// It returns a SubjectSpecification for the subject and a nil error if the input string is valid.
// It returns a non-nil error when the input string is invalid.
func Check(input string, opts ...Option) (SubjectSpecification, error) {
	var o options
	for _, opt := range opts {
		opt(&o)
	}

	if len(input) == 0 {
		return SubjectSpecification{}, SubjectActionsError{Detail: "input is empty"}
	}

	var sidEvents SIDEvents
	if err := json.Unmarshal([]byte(input), &sidEvents); err != nil {
		return SubjectSpecification{}, errors.Join(SubjectActionsError{Detail: "could not unmarshal input into SIDEvents"}, err)
	}
	// SID treats an absent or null schemaVersion as "1.0", but rejects any other value whose major version is not "1" (including "").
	if sidEvents.SchemaVersion != nil {
		major, _, _ := strings.Cut(*sidEvents.SchemaVersion, ".")
		if major != supportedSchemaMajorVersion {
			return SubjectSpecification{}, SubjectActionsError{Field: "schemaVersion", Detail: errorDetailUnexpectedField(supportedSchemaMajorVersion+".x", fmt.Sprintf("%q", *sidEvents.SchemaVersion))}
		}
	}
	if len(sidEvents.Events) != 1 {
		return SubjectSpecification{}, SubjectActionsError{Field: "events", Detail: errorDetailUnexpectedLength(1, len(sidEvents.Events))}
	}
	if sidEvents.Events[0].Type != requiredSIDEventType {
		return SubjectSpecification{}, SubjectActionsError{Field: "events[0].type", Detail: errorDetailUnexpectedField(requiredSIDEventType, sidEvents.Events[0].Type)}
	}
	if len(sidEvents.Events[0].ID) == 0 {
		return SubjectSpecification{}, SubjectActionsError{Field: "events[0].id", Detail: errorDetailEmptyField}
	}

	payload := sidEvents.Events[0].Payload
	if len(payload.SubjectID) == 0 {
		return SubjectSpecification{}, SubjectActionsError{Field: "subjectId", Detail: errorDetailEmptyField}
	}
	if len(payload.ProjectID) == 0 {
		return SubjectSpecification{}, SubjectActionsError{Field: "projectId", Detail: errorDetailEmptyField}
	}
	if payload.ModuleID.ClassName != requiredStringValueClassName {
		return SubjectSpecification{}, SubjectActionsError{Field: "moduleId.className", Detail: errorDetailUnexpectedField(requiredStringValueClassName, payload.ModuleID.ClassName)}
	}
	if len(payload.ModuleID.Value) == 0 {
		return SubjectSpecification{}, SubjectActionsError{Field: "moduleId.value", Detail: errorDetailEmptyField}
	}
	if payload.AttendantID.ClassName != requiredStringValueClassName {
		return SubjectSpecification{}, SubjectActionsError{Field: "attendantId.className", Detail: errorDetailUnexpectedField(requiredStringValueClassName, payload.AttendantID.ClassName)}
	}
	if len(payload.AttendantID.Value) == 0 {
		return SubjectSpecification{}, SubjectActionsError{Field: "attendantId.value", Detail: errorDetailEmptyField}
	}

	// External credentials are optional, but every subject must have at least one biometric reference.
	if len(payload.BiometricReferences) == 0 {
		return SubjectSpecification{}, SubjectActionsError{Field: "biometricReferences", Detail: "expect at least one biometric reference"}
	}

	var biometricFormats []string
	for i, reference := range payload.BiometricReferences {
		if err := checkBiometricReference(fmt.Sprintf("biometricReferences[%d]", i), reference, o); err != nil {
			return SubjectSpecification{}, err
		}
		if !slices.Contains(biometricFormats, reference.Format) {
			biometricFormats = append(biometricFormats, reference.Format)
		}
	}

	var externalCredentialTypes []string
	for i, credential := range payload.ExternalCredentials {
		if err := checkExternalCredential(fmt.Sprintf("externalCredentials[%d]", i), credential, payload.SubjectID); err != nil {
			return SubjectSpecification{}, err
		}
		if !slices.Contains(externalCredentialTypes, credential.Type) {
			externalCredentialTypes = append(externalCredentialTypes, credential.Type)
		}
	}

	subjectSpecification := SubjectSpecification{
		SubjectID:               payload.SubjectID,
		ProjectID:               payload.ProjectID,
		TokenizedModuleID:       payload.ModuleID.Value,
		TokenizedAtendantID:     payload.AttendantID.Value,
		BiometricFormats:        biometricFormats,
		ExternalCredentialTypes: externalCredentialTypes,
	}
	return subjectSpecification, nil
}

func checkBiometricReference(field string, reference SIDBiometricReference, o options) error {
	if len(reference.ID) == 0 {
		return SubjectActionsError{Field: field + ".id", Detail: errorDetailEmptyField}
	}
	if !slices.Contains(allowedSIDBiometricReferenceTypes, reference.Type) {
		return SubjectActionsError{Field: field + ".type", Detail: errorDetailUnexpectedField(strings.Join(allowedSIDBiometricReferenceTypes, " or "), reference.Type)}
	}
	if len(reference.Format) == 0 {
		return SubjectActionsError{Field: field + ".format", Detail: errorDetailEmptyField}
	}
	if len(o.allowedFormats) > 0 && !slices.Contains(o.allowedFormats, reference.Format) {
		return SubjectActionsError{Field: field + ".format", Detail: errorDetailUnexpectedField(strings.Join(o.allowedFormats, " or "), reference.Format)}
	}
	if len(reference.Templates) == 0 {
		return SubjectActionsError{Field: field + ".templates", Detail: errorDetailEmptyField}
	}
	for i, template := range reference.Templates {
		templateField := fmt.Sprintf("%s.templates[%d]", field, i)
		if len(template.Template) == 0 {
			return SubjectActionsError{Field: templateField + ".template", Detail: errorDetailEmptyField}
		}
		if !isBase64(template.Template) {
			return SubjectActionsError{Field: templateField + ".template", Detail: "not valid base64"}
		}
		if reference.Type == sidFingerprintReferenceType && !slices.Contains(allowedSIDFingers, template.Finger) {
			return SubjectActionsError{Field: templateField + ".finger", Detail: errorDetailUnexpectedField("one of "+strings.Join(allowedSIDFingers, ", "), template.Finger)}
		}
	}
	return nil
}

func checkExternalCredential(field string, credential SIDExternalCredential, subjectID string) error {
	if len(credential.ID) == 0 {
		return SubjectActionsError{Field: field + ".id", Detail: errorDetailEmptyField}
	}
	if !slices.Contains(allowedSIDExternalCredentialTypes, credential.Type) {
		return SubjectActionsError{Field: field + ".type", Detail: errorDetailUnexpectedField("one of "+strings.Join(allowedSIDExternalCredentialTypes, ", "), credential.Type)}
	}
	if credential.SubjectID != subjectID {
		return SubjectActionsError{Field: field + ".subjectId", Detail: errorDetailUnexpectedField(subjectID, credential.SubjectID)}
	}
	// SID fails to load the whole payload (and can fail loading other subjects) when the credential value is not tokenized.
	if credential.Value.ClassName != requiredStringValueClassName {
		return SubjectActionsError{Field: field + ".value.className", Detail: errorDetailUnexpectedField(requiredStringValueClassName, credential.Value.ClassName)}
	}
	if len(credential.Value.Value) == 0 {
		return SubjectActionsError{Field: field + ".value.value", Detail: errorDetailEmptyField}
	}
	return nil
}

// isBase64 mirrors SID's decoding (android.util.Base64 with NO_WRAP), which accepts the standard alphabet with or without padding.
func isBase64(s string) bool {
	if _, err := base64.StdEncoding.DecodeString(s); err == nil {
		return true
	}
	_, err := base64.RawStdEncoding.DecodeString(s)
	return err == nil
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
	// BiometricFormats are the distinct formats of the biometric references attached to the subject (never empty when Check succeeds).
	BiometricFormats []string
	// ExternalCredentialTypes are the distinct types of the external credentials attached to the subject (nil if there are none).
	ExternalCredentialTypes []string
}

// SubjectActionsError stores details about the way in which a given input failed to pass `Check()`.
type SubjectActionsError struct {
	// Detail contains information about what was wrong.
	Detail string
	// Field contains information about which field had something wrong if it is non-empty.
	Field string
}

// Error makes SubjectActionsError implement error.
func (e SubjectActionsError) Error() string {
	if e.Field == "" {
		return e.Detail
	}
	return fmt.Sprintf("%s: %s", e.Field, e.Detail)
}

// SIDEvents represents a collection of events generated by Simprints ID.
type SIDEvents struct {
	// SchemaVersion is the version of the CoSync schema (e.g. "1.0").
	// It is nil when the field is absent or null, which SID treats as "1.0".
	SchemaVersion *string `json:"schemaVersion"`
	// Events are the enrolment record events.
	Events []SIDEvent `json:"events"`
}

// SIDEvent represents a single event generated by Simprints ID.
type SIDEvent struct {
	// ID is the unique ID of this event.
	ID string `json:"id"`
	// Type is the kind of event this is (in our case, this should always be EnrolmentRecordCreation).
	Type string `json:"type"`
	// Payload for an EnrolmentRecordCreation event is the data associated with the creation of an enrolment record in SID.
	Payload SIDEnrolmentRecordCreationPayload `json:"payload"`
}

// Look at https://simprints.gitbook.io/docs/development/simprints-for-developers/other-intergrations/commcare-integration/cosync for more documentation of SID <> CommCare integration.
// See the schema comment further down this file for the exact fields SID accepts and where in SID they are defined.
type SIDEnrolmentRecordCreationPayload struct {
	/// SubjectID is the ID of the person (a.k.a 'subject') this enrolment record belongs to, and is the same as `simprintsId` or `guid` in the integration.
	SubjectID string `json:"subjectId"`
	// ProjectID is the ID of the Simprints project this enrolment record belongs to, and is the same as `projectId` used in integration.
	ProjectID string `json:"projectId"`
	// ModuleID is the ID of the Simprints module this enrolment record belongs to, and is the same as `moduleId` used in integration.
	ModuleID StringValue `json:"moduleId"`
	// AttendantID is the ID of the attendant (a.k.a 'user') that enrolled this subject, and is the same as `userId` used in integration.
	AttendantID StringValue `json:"attendantId"`
	// BiometricReferences contains all biometric data attached to this subject.
	BiometricReferences []SIDBiometricReference `json:"biometricReferences"`
	// ExternalCredentials contains all external credentials (e.g. ID cards, QR codes) linked to this subject.
	ExternalCredentials []SIDExternalCredential `json:"externalCredentials"`
}

// StringValue is a struct that a potentially tokenized (encrypted) string.
type StringValue struct {
	// ClassName is "TokenizableString.Tokenized" when the string is tokenized and "TokenizableString.Raw" when not.
	ClassName string `json:"className"`
	// Value is the actual tokenized or raw string (depending on what ClassName is).
	Value string `json:"value"`
}

// SIDBiometricReference represents a particular kind of biometric data.
type SIDBiometricReference struct {
	// ID is the unique ID of this biometric reference.
	ID string `json:"id"`
	// Templates is a collection of biometric templates that make up this biometric reference.
	Templates []SIDTemplate `json:"templates"`
	// Format is the format of the biometric templates (e.g. "ISO_19794_2" or "NEC_1_5" for fingerprint, "RANK_ONE_3_1" for face).
	Format string `json:"format"`
	// Type is the biometric modality of this biometric reference ('FINGERPRINT_REFERENCE' or 'FACE_REFERENCE').
	Type string `json:"type"`
	// Metadata is optional additional information about this biometric reference.
	Metadata map[string]string `json:"metadata,omitempty"`
}

// SIDTemplate represents a single biometric template.
type SIDTemplate struct {
	// Quality is the biometric quality of the template.
	//
	// Deprecated: quality is no longer part of SID's schema and is ignored by SID if present.
	Quality float64 `json:"quality,omitempty"`
	// Template is the base64 encoded biometric template data.
	Template string `json:"template"`
	// Finger represents which finger this biometric template corresponds to (e.g. RIGHT_THUMB).
	// It is only required for fingerprint templates.
	Finger string `json:"finger,omitempty"`
}

// SIDExternalCredential represents an external credential (e.g. an ID card or QR code) linked to a subject.
type SIDExternalCredential struct {
	// ID is the unique ID of this external credential.
	ID string `json:"id"`
	// Value is the credential value, which must be tokenized.
	Value StringValue `json:"value"`
	// SubjectID is the ID of the subject this credential is linked to.
	SubjectID string `json:"subjectId"`
	// Type is the kind of credential (e.g. 'NHISCard').
	Type string `json:"type"`
}

// How to find the schema SID expects
//
// The source of truth is the CoSync V1 models in the Android-Simprints-ID repository (https://github.com/Simprints/Android-Simprints-ID).
// Paths below are relative to the repository root:
//
//   - infra/events/src/main/java/com/simprints/infra/events/event/cosync/CoSyncEnrolmentRecordEvents.kt
//     Top-level object and `schemaVersion` handling (a missing version means V1; an unknown major version fails to parse).
//   - infra/events/src/main/java/com/simprints/infra/events/event/cosync/v1/
//     One file per part of the schema: CoSyncEnrolmentRecordEvent.kt (event + payload), CoSyncBiometricReference.kt,
//     CoSyncTemplate.kt, CoSyncTemplateIdentifier.kt (finger values), CoSyncExternalCredential.kt,
//     CoSyncExternalCredentialType.kt (credential types) and CoSyncTokenizableString.kt (className/value strings).
//   - infra/serialization/src/main/java/com/simprints/infra/serialization/DefaultJson.kt
//     JSON settings used to parse subjectActions (unknown keys are ignored; polymorphic types use the "type" field).
//   - infra/enrolment-records/repository/src/main/java/com/simprints/infra/enrolment/records/repository/commcare/CommCareCandidateRecordDataSource.kt
//     Reads subjectActions when CommCare is the biometric data source (identification/verification directly from CommCare).
//   - infra/event-sync/src/main/java/com/simprints/infra/eventsync/event/commcare/CommCareEventDataSource.kt
//     Reads subjectActions when SID down-syncs CommCare cases into its local database.
//   - feature/client-api/src/main/java/com/simprints/feature/clientapi/usecases/GetEnrolmentCreationEventForRecordUseCase.kt
//     Builds the subjectActions that SID returns to CommCare, which is a good example of a valid payload.
//
// When those models change (e.g. a new modality, credential type or schemaVersion), this file should be updated to match.
//
// SID expects the following schema (where '*' denotes some array index)
// For this project, we expect only 1 EnrolmentRecordCreation event.
// For this project, we expect fingerprint and/or face biometric references and optionally, external credentials.
//
// Check is intentionally stricter than SID in a few places, to catch payloads that SID would parse but not be able to use:
// moduleId/attendantId must be tokenized, there must be at least one biometric reference (external credentials are optional), each
// biometric reference must have at least one valid base64 template, and external credentials must belong to the subject.
// Use WithAllowedFormats to also restrict biometric reference formats to the ones configured for the project.
//
// > schemaVersion (optional; absent or null is treated as '1.0'; an empty string or any major version other than '1' is rejected by SID - if for some reason the version changes, we would need to update this checker)
// > events.*.id
// > events.*.type (must always be 'EnrolmentRecordCreation')
// > events.*.payload.projectId
// > events.*.payload.subjectId
// > events.*.payload.moduleId.className (should be 'TokenizableString.Tokenized')
// > events.*.payload.moduleId.value (this field should be tokenized)
// > events.*.payload.attendantId.className (should be 'TokenizableString.Tokenized')
// > events.*.payload.attendantId.value (this field should be tokenized)
//
// > events.*.payload.biometricReferences (optional and may be empty in SID; Check requires at least one)
// > events.*.payload.biometricReferences.*.id
// > events.*.payload.biometricReferences.*.type (one of 'FINGERPRINT_REFERENCE' or 'FACE_REFERENCE')
// > events.*.payload.biometricReferences.*.format (must match the matcher configured for the project, e.g.
//   'ISO_19794_2' or 'NEC_1_5' for fingerprint, 'RANK_ONE_1_23' or 'RANK_ONE_3_1' for face; references in any
//   other format are ignored when matching)
// > events.*.payload.biometricReferences.*.metadata (optional, map of string to string)
// > events.*.payload.biometricReferences.*.templates.*.template (base64 encoded template)
// > events.*.payload.biometricReferences.*.templates.*.finger (FINGERPRINT_REFERENCE only; one of 'NONE',
//   'LEFT_THUMB', 'LEFT_INDEX_FINGER', 'LEFT_3RD_FINGER', 'LEFT_4TH_FINGER', 'LEFT_5TH_FINGER', 'RIGHT_THUMB',
//   'RIGHT_INDEX_FINGER', 'RIGHT_3RD_FINGER', 'RIGHT_4TH_FINGER', 'RIGHT_5TH_FINGER')
// > events.*.payload.biometricReferences.*.templates.*.quality (no longer part of the schema; ignored by SID if present)
//
// > events.*.payload.externalCredentials (optional, may be empty)
// > events.*.payload.externalCredentials.*.id
// > events.*.payload.externalCredentials.*.subjectId (should be the same as events.*.payload.subjectId)
// > events.*.payload.externalCredentials.*.type (one of 'NHISCard', 'GhanaIdCard', 'QRCode' or 'FaydaCard')
// > events.*.payload.externalCredentials.*.value.className (must be 'TokenizableString.Tokenized'; SID fails to load
//   the payload if it is 'TokenizableString.Raw', which can also break loading of other subjects)
// > events.*.payload.externalCredentials.*.value.value (this field must be tokenized)
//
// Note: when CommCare is the biometric data source, SID only uses biometricReferences read from subjectActions.
// External credentials are only searched for in SID's local database, so they take effect only once CommCare cases
// have been down-synced into SID.

const (
	supportedSchemaMajorVersion  = "1"
	requiredSIDEventType         = "EnrolmentRecordCreation"
	requiredStringValueClassName = "TokenizableString.Tokenized"
	sidFingerprintReferenceType  = "FINGERPRINT_REFERENCE"
	sidFaceReferenceType         = "FACE_REFERENCE"
)

var (
	allowedSIDBiometricReferenceTypes = []string{sidFingerprintReferenceType, sidFaceReferenceType}
	// allowedSIDFingers mirrors CoSyncTemplateIdentifier in SID.
	allowedSIDFingers = []string{
		"NONE",
		"RIGHT_5TH_FINGER", "RIGHT_4TH_FINGER", "RIGHT_3RD_FINGER", "RIGHT_INDEX_FINGER", "RIGHT_THUMB",
		"LEFT_THUMB", "LEFT_INDEX_FINGER", "LEFT_3RD_FINGER", "LEFT_4TH_FINGER", "LEFT_5TH_FINGER",
	}
	// allowedSIDExternalCredentialTypes mirrors CoSyncExternalCredentialType in SID.
	allowedSIDExternalCredentialTypes = []string{"NHISCard", "GhanaIdCard", "QRCode", "FaydaCard"}
)

const (
	errorDetailEmptyField = "empty field"
)

func errorDetailUnexpectedField(expected string, got string) string {
	return fmt.Sprintf("expect %s, got %s", expected, got)
}

func errorDetailUnexpectedLength(expected int, got int) string {
	return fmt.Sprintf("expect length %d, got %d", expected, got)
}

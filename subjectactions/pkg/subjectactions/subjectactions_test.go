package subjectactions

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const aTemplate = "dGVtcGxhdGU=" // base64 of "template"

func fingerprintReference() map[string]any {
	return map[string]any{
		"id":        "aFingerprintReferenceID",
		"type":      "FINGERPRINT_REFERENCE",
		"format":    "ISO_19794_2",
		"templates": []any{map[string]any{"finger": "RIGHT_THUMB", "quality": 0.8, "template": aTemplate}},
	}
}

func faceReference() map[string]any {
	return map[string]any{
		"id":        "aFaceReferenceID",
		"type":      "FACE_REFERENCE",
		"format":    "RANK_ONE_3_1",
		"templates": []any{map[string]any{"template": aTemplate}},
	}
}

func externalCredential() map[string]any {
	return map[string]any{
		"id":        "aCredentialID",
		"subjectId": "aSubjectID",
		"type":      "NHISCard",
		"value":     map[string]any{"className": "TokenizableString.Tokenized", "value": "aTokenizedCredential"},
	}
}

// subjectActions builds a valid subjectActions JSON string with a single fingerprint reference,
// then applies the given modifications to the top-level object, the event and the payload.
func subjectActions(t *testing.T, modify func(root, event, payload map[string]any)) string {
	t.Helper()
	payload := map[string]any{
		"subjectId":           "aSubjectID",
		"projectId":           "aProjectID",
		"moduleId":            map[string]any{"className": "TokenizableString.Tokenized", "value": "aTokenizedModuleID"},
		"attendantId":         map[string]any{"className": "TokenizableString.Tokenized", "value": "aTokenizedAttendantID"},
		"biometricReferences": []any{fingerprintReference()},
	}
	event := map[string]any{"id": "anID", "type": "EnrolmentRecordCreation", "payload": payload}
	root := map[string]any{"events": []any{event}}
	if modify != nil {
		modify(root, event, payload)
	}
	b, err := json.Marshal(root)
	require.NoError(t, err)
	return string(b)
}

func firstReference(payload map[string]any) map[string]any {
	return payload["biometricReferences"].([]any)[0].(map[string]any)
}

func firstTemplate(payload map[string]any) map[string]any {
	return firstReference(payload)["templates"].([]any)[0].(map[string]any)
}

func TestCheck(t *testing.T) {
	baseSpec := SubjectSpecification{
		SubjectID:           "aSubjectID",
		ProjectID:           "aProjectID",
		TokenizedModuleID:   "aTokenizedModuleID",
		TokenizedAtendantID: "aTokenizedAttendantID",
	}
	withSpec := func(formats, credentialTypes []string) SubjectSpecification {
		s := baseSpec
		s.BiometricFormats = formats
		s.ExternalCredentialTypes = credentialTypes
		return s
	}

	tcs := map[string]struct {
		input          string
		opts           []Option
		expectedErr    string
		expectedResult SubjectSpecification
	}{
		// Top level and event
		"empty string should return error": {
			input:       "",
			expectedErr: "input is empty",
		},
		"invalid json should return error": {
			input:       "invalid json",
			expectedErr: "could not unmarshal input into SIDEvents",
		},
		"empty events array should return error": {
			input:       `{"events": []}`,
			expectedErr: "events: expect length 1",
		},
		"multiple events should return error": {
			input:       `{"events": [{}, {}]}`,
			expectedErr: "events: expect length 1",
		},
		"wrong event type should return error": {
			input:       `{"events": [{"id": "anID", "type": "WrongEventType", "payload": {}}]}`,
			expectedErr: "events[0].type: expect EnrolmentRecordCreation, got WrongEventType",
		},
		"empty event id should return error": {
			input:       subjectActions(t, func(_, event, _ map[string]any) { event["id"] = "" }),
			expectedErr: "events[0].id: empty",
		},
		"absent schemaVersion should be accepted": {
			input:          subjectActions(t, nil),
			expectedResult: withSpec([]string{"ISO_19794_2"}, nil),
		},
		"null schemaVersion should be accepted": {
			input:          subjectActions(t, func(root, _, _ map[string]any) { root["schemaVersion"] = nil }),
			expectedResult: withSpec([]string{"ISO_19794_2"}, nil),
		},
		"schemaVersion 1 should be accepted": {
			input:          subjectActions(t, func(root, _, _ map[string]any) { root["schemaVersion"] = "1" }),
			expectedResult: withSpec([]string{"ISO_19794_2"}, nil),
		},
		"schemaVersion 1.0 should be accepted": {
			input:          subjectActions(t, func(root, _, _ map[string]any) { root["schemaVersion"] = "1.0" }),
			expectedResult: withSpec([]string{"ISO_19794_2"}, nil),
		},
		"schemaVersion 1.3 should be accepted": {
			input:          subjectActions(t, func(root, _, _ map[string]any) { root["schemaVersion"] = "1.3" }),
			expectedResult: withSpec([]string{"ISO_19794_2"}, nil),
		},
		"schemaVersion 2.0 should return error": {
			input:       subjectActions(t, func(root, _, _ map[string]any) { root["schemaVersion"] = "2.0" }),
			expectedErr: `schemaVersion: expect 1.x, got "2.0"`,
		},
		"schemaVersion 10.0 should return error": {
			input:       subjectActions(t, func(root, _, _ map[string]any) { root["schemaVersion"] = "10.0" }),
			expectedErr: `schemaVersion: expect 1.x, got "10.0"`,
		},
		"empty schemaVersion should return error": {
			input:       subjectActions(t, func(root, _, _ map[string]any) { root["schemaVersion"] = "" }),
			expectedErr: `schemaVersion: expect 1.x, got ""`,
		},

		// Subject fields
		"empty subjectId should return error": {
			input:       subjectActions(t, func(_, _, payload map[string]any) { payload["subjectId"] = "" }),
			expectedErr: "subjectId: empty",
		},
		"empty projectId should return error": {
			input:       subjectActions(t, func(_, _, payload map[string]any) { payload["projectId"] = "" }),
			expectedErr: "projectId: empty",
		},
		"wrong moduleId className should return error": {
			input: subjectActions(t, func(_, _, payload map[string]any) {
				payload["moduleId"] = map[string]any{"className": "wrong", "value": "aTokenizedModuleID"}
			}),
			expectedErr: "moduleId.className: expect TokenizableString.Tokenized, got wrong",
		},
		"empty moduleId should return error": {
			input: subjectActions(t, func(_, _, payload map[string]any) {
				payload["moduleId"] = map[string]any{"className": "TokenizableString.Tokenized", "value": ""}
			}),
			expectedErr: "moduleId.value: empty",
		},
		"wrong attendantID className should return error": {
			input: subjectActions(t, func(_, _, payload map[string]any) {
				payload["attendantId"] = map[string]any{"className": "wrong", "value": "aTokenizedAttendantID"}
			}),
			expectedErr: "attendantId.className: expect TokenizableString.Tokenized, got wrong",
		},
		"empty attendantId should return error": {
			input: subjectActions(t, func(_, _, payload map[string]any) {
				payload["attendantId"] = map[string]any{"className": "TokenizableString.Tokenized", "value": ""}
			}),
			expectedErr: "attendantId.value: empty",
		},
		"no biometric references and no external credentials should return error": {
			input:       subjectActions(t, func(_, _, payload map[string]any) { delete(payload, "biometricReferences") }),
			expectedErr: "biometricReferences: expect at least one biometric reference",
		},
		"empty biometric references should return error": {
			input:       subjectActions(t, func(_, _, payload map[string]any) { payload["biometricReferences"] = []any{} }),
			expectedErr: "biometricReferences: expect at least one biometric reference",
		},
		"null biometric references should return error": {
			input:       subjectActions(t, func(_, _, payload map[string]any) { payload["biometricReferences"] = nil }),
			expectedErr: "biometricReferences: expect at least one biometric reference",
		},

		// Biometric references
		"valid ISO fingerprint input should return the correct subject specification": {
			input:          subjectActions(t, nil),
			expectedResult: withSpec([]string{"ISO_19794_2"}, nil),
		},
		"valid NEC fingerprint input should be accepted": {
			input:          subjectActions(t, func(_, _, payload map[string]any) { firstReference(payload)["format"] = "NEC_1_5" }),
			expectedResult: withSpec([]string{"NEC_1_5"}, nil),
		},
		"fingerprint template without quality should be accepted": {
			input:          subjectActions(t, func(_, _, payload map[string]any) { delete(firstTemplate(payload), "quality") }),
			expectedResult: withSpec([]string{"ISO_19794_2"}, nil),
		},
		"fingerprint template without padding should be accepted": {
			input:          subjectActions(t, func(_, _, payload map[string]any) { firstTemplate(payload)["template"] = "dGVtcGxhdGU" }),
			expectedResult: withSpec([]string{"ISO_19794_2"}, nil),
		},
		"valid face input should be accepted": {
			input:          subjectActions(t, func(_, _, payload map[string]any) { payload["biometricReferences"] = []any{faceReference()} }),
			expectedResult: withSpec([]string{"RANK_ONE_3_1"}, nil),
		},
		"face reference with metadata should be accepted": {
			input: subjectActions(t, func(_, _, payload map[string]any) {
				face := faceReference()
				face["metadata"] = map[string]any{"key": "value"}
				payload["biometricReferences"] = []any{face}
			}),
			expectedResult: withSpec([]string{"RANK_ONE_3_1"}, nil),
		},
		"multiple fingerprint and face references should be accepted": {
			input: subjectActions(t, func(_, _, payload map[string]any) {
				second := fingerprintReference()
				second["id"] = "anotherFingerprintReferenceID"
				payload["biometricReferences"] = []any{fingerprintReference(), second, faceReference()}
			}),
			expectedResult: withSpec([]string{"ISO_19794_2", "RANK_ONE_3_1"}, nil),
		},
		"empty biometric reference id should return error": {
			input:       subjectActions(t, func(_, _, payload map[string]any) { firstReference(payload)["id"] = "" }),
			expectedErr: "biometricReferences[0].id: empty",
		},
		"wrong biometric reference type should return error": {
			input:       subjectActions(t, func(_, _, payload map[string]any) { firstReference(payload)["type"] = "WRONG_TYPE" }),
			expectedErr: "biometricReferences[0].type: expect FINGERPRINT_REFERENCE or FACE_REFERENCE, got WRONG_TYPE",
		},
		"empty biometric reference format should return error": {
			input:       subjectActions(t, func(_, _, payload map[string]any) { firstReference(payload)["format"] = "" }),
			expectedErr: "biometricReferences[0].format: empty",
		},
		"format not in allowed formats should return error": {
			input:       subjectActions(t, nil),
			opts:        []Option{WithAllowedFormats("RANK_ONE_3_1")},
			expectedErr: "biometricReferences[0].format: expect RANK_ONE_3_1, got ISO_19794_2",
		},
		"format in allowed formats should be accepted": {
			input: subjectActions(t, func(_, _, payload map[string]any) {
				payload["biometricReferences"] = []any{fingerprintReference(), faceReference()}
			}),
			opts:           []Option{WithAllowedFormats("ISO_19794_2", "RANK_ONE_3_1")},
			expectedResult: withSpec([]string{"ISO_19794_2", "RANK_ONE_3_1"}, nil),
		},
		"biometric reference without templates should return error": {
			input:       subjectActions(t, func(_, _, payload map[string]any) { firstReference(payload)["templates"] = []any{} }),
			expectedErr: "biometricReferences[0].templates: empty",
		},
		"empty template should return error": {
			input:       subjectActions(t, func(_, _, payload map[string]any) { firstTemplate(payload)["template"] = "" }),
			expectedErr: "biometricReferences[0].templates[0].template: empty",
		},
		"template that is not base64 should return error": {
			input:       subjectActions(t, func(_, _, payload map[string]any) { firstTemplate(payload)["template"] = "not base64!" }),
			expectedErr: "biometricReferences[0].templates[0].template: not valid base64",
		},
		"unknown finger should return error": {
			input:       subjectActions(t, func(_, _, payload map[string]any) { firstTemplate(payload)["finger"] = "LEFT_6TH_FINGER" }),
			expectedErr: "biometricReferences[0].templates[0].finger: expect one of",
		},
		"missing finger on fingerprint template should return error": {
			input:       subjectActions(t, func(_, _, payload map[string]any) { delete(firstTemplate(payload), "finger") }),
			expectedErr: "biometricReferences[0].templates[0].finger: expect one of",
		},

		// External credentials
		"external credential without biometric references should return error": {
			input: subjectActions(t, func(_, _, payload map[string]any) {
				delete(payload, "biometricReferences")
				payload["externalCredentials"] = []any{externalCredential()}
			}),
			expectedErr: "biometricReferences: expect at least one biometric reference",
		},
		"biometrics with one external credential should be accepted": {
			input: subjectActions(t, func(_, _, payload map[string]any) {
				payload["externalCredentials"] = []any{externalCredential()}
			}),
			expectedResult: withSpec([]string{"ISO_19794_2"}, []string{"NHISCard"}),
		},
		"biometrics with empty external credentials should be accepted": {
			input:          subjectActions(t, func(_, _, payload map[string]any) { payload["externalCredentials"] = []any{} }),
			expectedResult: withSpec([]string{"ISO_19794_2"}, nil),
		},
		"biometrics and external credentials should be accepted": {
			input: subjectActions(t, func(_, _, payload map[string]any) {
				second := externalCredential()
				second["id"] = "anotherCredentialID"
				second["type"] = "QRCode"
				payload["biometricReferences"] = []any{fingerprintReference(), faceReference()}
				payload["externalCredentials"] = []any{externalCredential(), second}
			}),
			expectedResult: withSpec([]string{"ISO_19794_2", "RANK_ONE_3_1"}, []string{"NHISCard", "QRCode"}),
		},
		"empty external credential id should return error": {
			input: subjectActions(t, func(_, _, payload map[string]any) {
				credential := externalCredential()
				credential["id"] = ""
				payload["externalCredentials"] = []any{credential}
			}),
			expectedErr: "externalCredentials[0].id: empty",
		},
		"unknown external credential type should return error": {
			input: subjectActions(t, func(_, _, payload map[string]any) {
				credential := externalCredential()
				credential["type"] = "Passport"
				payload["externalCredentials"] = []any{credential}
			}),
			expectedErr: "externalCredentials[0].type: expect one of NHISCard, GhanaIdCard, QRCode, FaydaCard, got Passport",
		},
		"external credential for another subject should return error": {
			input: subjectActions(t, func(_, _, payload map[string]any) {
				credential := externalCredential()
				credential["subjectId"] = "anotherSubjectID"
				payload["externalCredentials"] = []any{credential}
			}),
			expectedErr: "externalCredentials[0].subjectId: expect aSubjectID, got anotherSubjectID",
		},
		"raw external credential value should return error": {
			input: subjectActions(t, func(_, _, payload map[string]any) {
				credential := externalCredential()
				credential["value"] = map[string]any{"className": "TokenizableString.Raw", "value": "1234"}
				payload["externalCredentials"] = []any{credential}
			}),
			expectedErr: "externalCredentials[0].value.className: expect TokenizableString.Tokenized, got TokenizableString.Raw",
		},
		"empty external credential value should return error": {
			input: subjectActions(t, func(_, _, payload map[string]any) {
				credential := externalCredential()
				credential["value"] = map[string]any{"className": "TokenizableString.Tokenized", "value": ""}
				payload["externalCredentials"] = []any{credential}
			}),
			expectedErr: "externalCredentials[0].value.value: empty",
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got, err := Check(tc.input, tc.opts...)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedResult, got)
		})
	}
}

// TestCheckSIDGeneratedPayload checks a payload in the shape SID itself writes to CommCare
// (see GetEnrolmentCreationEventForRecordUseCase in Android-Simprints-ID).
func TestCheckSIDGeneratedPayload(t *testing.T) {
	input := `{"schemaVersion":"1.0","events":[{"type":"EnrolmentRecordCreation","id":"0b6a1f2c-1a2b-4c3d-9e8f-001122334455","payload":{"subjectId":"8e3c2a10-5b6d-4e7f-8a9b-aabbccddeeff","projectId":"aProjectID","moduleId":{"className":"TokenizableString.Tokenized","value":"AdyF9pnMuOlXmEQtYLZ5hD3x4vDAmE+n4Q=="},"attendantId":{"className":"TokenizableString.Tokenized","value":"AWuA3H0WGtHOfOzQ4u1pEgGdYJEc+P0="},"biometricReferences":[{"type":"FINGERPRINT_REFERENCE","id":"ref-1","templates":[{"template":"Rk1SACAyMAAAAAEIAAABPAFiAMUAxQEAAAAoJ0CNAC5/AEBxAEQ","finger":"LEFT_THUMB"},{"template":"Rk1SACAyMAAAAADkAAABPAFiAMUAxQEAAAAoIUCEAC2jAEBoAEw","finger":"LEFT_INDEX_FINGER"}],"format":"ISO_19794_2"},{"type":"FACE_REFERENCE","id":"ref-2","templates":[{"template":"AAECAwQFBgcICQ=="}],"format":"RANK_ONE_3_1"}],"externalCredentials":[{"id":"cred-1","value":{"className":"TokenizableString.Tokenized","value":"AQIDBAUGBwgJ"},"subjectId":"8e3c2a10-5b6d-4e7f-8a9b-aabbccddeeff","type":"GhanaIdCard"}]}}]}`
	got, err := Check(input, WithAllowedFormats("ISO_19794_2", "RANK_ONE_3_1"))
	require.NoError(t, err)
	assert.Equal(t, []string{"ISO_19794_2", "RANK_ONE_3_1"}, got.BiometricFormats)
	assert.Equal(t, []string{"GhanaIdCard"}, got.ExternalCredentialTypes)
}

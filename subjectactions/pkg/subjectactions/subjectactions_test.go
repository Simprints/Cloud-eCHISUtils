package subjectactions

import (
	"strings"
	"testing"
)

func TestCheck(t *testing.T) {
	tcs := []struct {
		name              string
		input             string
		expectedErrPrefix string
		expectedResult    SubjectSpecification
	}{
		{
			name:              "empty string should return error",
			input:             "",
			expectedErrPrefix: "input is empty",
		},
		{
			name:              "invalid json should return error",
			input:             "invalid json",
			expectedErrPrefix: "could not unmarshal",
		},
		{
			name:              "empty events array should return error",
			input:             `{"events": []}`,
			expectedErrPrefix: "expect 1 event",
		},
		{
			name:              "multiple events should return error",
			input:             `{"events": [{}, {}]}`,
			expectedErrPrefix: "expect 1 event",
		},
		{
			name:              "wrong event type should return error",
			input:             `{"events": [{"id": "anID", "type": "WrongEventType", "payload": {}}]}`,
			expectedErrPrefix: "expect event type EnrolmentRecordCreation",
		},
		{
			name:              "empty subjectId should return error",
			input:             `{"events": [{"id": "anID", "type": "EnrolmentRecordCreation", "payload": {"subjectId":"","projectId":"aProjectID","moduleId":{"className":"TokenizableString.Tokenized","value":"aModuleID"},"attendantId":{"className":"TokenizableString.Tokenized","value":"anAttendantID"},"biometricReferences":[{"id":"aBiometricReferenceID","templates":[{"finger":"RIGHT_THUMB","quality":0.8,"template":"aTemplate"}],"format":"ISO_19794_2","type":"FINGERPRINT_REFERENCE"}]}}]}`,
			expectedErrPrefix: "empty subjectId",
		},
		{
			name:              "empty projectId should return error",
			input:             `{"events": [{"id": "anID", "type": "EnrolmentRecordCreation", "payload": {"subjectId":"aSubjectID","projectId":"","moduleId":{"className":"TokenizableString.Tokenized","value":"aModuleID"},"attendantId":{"className":"TokenizableString.Tokenized","value":"anAttendantID"},"biometricReferences":[{"id":"aBiometricReferenceID","templates":[{"finger":"RIGHT_THUMB","quality":0.8,"template":"aTemplate"}],"format":"ISO_19794_2","type":"FINGERPRINT_REFERENCE"}]}}]}`,
			expectedErrPrefix: "empty projectId",
		},
		{
			name:              "wrong moduleId className should return error",
			input:             `{"events": [{"id": "anID", "type": "EnrolmentRecordCreation", "payload": {"subjectId":"aSubjectID","projectId":"aProjectID","moduleId":{"className":"wrong","value":"aModuleID"},"attendantId":{"className":"TokenizableString.Tokenized","value":"anAttendantID"},"biometricReferences":[{"id":"aBiometricReferenceID","templates":[{"finger":"RIGHT_THUMB","quality":0.8,"template":"aTemplate"}],"format":"ISO_19794_2","type":"FINGERPRINT_REFERENCE"}]}}]}`,
			expectedErrPrefix: "expect moduleId to have className TokenizableString.Tokenized",
		},
		{
			name:              "empty moduleId should return error",
			input:             `{"events": [{"id": "anID", "type": "EnrolmentRecordCreation", "payload": {"subjectId":"aSubjectID","projectId":"aProjectID","moduleId":{"className":"TokenizableString.Tokenized","value":""},"attendantId":{"className":"TokenizableString.Tokenized","value":"anAttendantID"},"biometricReferences":[{"id":"aBiometricReferenceID","templates":[{"finger":"RIGHT_THUMB","quality":0.8,"template":"aTemplate"}],"format":"ISO_19794_2","type":"FINGERPRINT_REFERENCE"}]}}]}`,
			expectedErrPrefix: "empty moduleId",
		},
		{
			name:              "wrong attendantID className should return error",
			input:             `{"events": [{"id": "anID", "type": "EnrolmentRecordCreation", "payload": {"subjectId":"aSubjectID","projectId":"aProjectID","moduleId":{"className":"TokenizableString.Tokenized","value":"aModuleID"},"attendantId":{"className":"wrong","value":"anAttendantID"},"biometricReferences":[{"id":"aBiometricReferenceID","templates":[{"finger":"RIGHT_THUMB","quality":0.8,"template":"aTemplate"}],"format":"ISO_19794_2","type":"FINGERPRINT_REFERENCE"}]}}]}`,
			expectedErrPrefix: "expect attendantId to have className TokenizableString.Tokenized",
		},
		{
			name:              "empty attendantId should return error",
			input:             `{"events": [{"id": "anID", "type": "EnrolmentRecordCreation", "payload": {"subjectId":"aSubjectID","projectId":"aProjectID","moduleId":{"className":"TokenizableString.Tokenized","value":"aModuleID"},"attendantId":{"className":"TokenizableString.Tokenized","value":""},"biometricReferences":[{"id":"aBiometricReferenceID","templates":[{"finger":"RIGHT_THUMB","quality":0.8,"template":"aTemplate"}],"format":"ISO_19794_2","type":"FINGERPRINT_REFERENCE"}]}}]}`,
			expectedErrPrefix: "empty attendantId",
		},
		{
			name:              "empty biometric references should return error",
			input:             `{"events": [{"id": "anID", "type": "EnrolmentRecordCreation", "payload": {"subjectId":"aSubjectID","projectId":"aProjectID","moduleId":{"className":"TokenizableString.Tokenized","value":"aModuleID"},"attendantId":{"className":"TokenizableString.Tokenized","value":"anAttendantID"},"biometricReferences":[{"id":"aBiometricReferenceID","templates":[{"finger":"RIGHT_THUMB","quality":0.8,"template":"aTemplate"}],"format":"ISO_19794_2","type":"FINGERPRINT_REFERENCE"},{"id":"anotherBiometricReferenceID","templates":[{"finger":"LEFT_THUMB","quality":0.9,"template":"anotherTemplate"}],"format":"ISO_19794_2","type":"FINGERPRINT_REFERENCE"}]}}]}`,
			expectedErrPrefix: "expect 1 biometric reference",
		},
		{
			name:              "multiple biometric references should return error",
			input:             `{"events": [{"id": "anID", "type": "EnrolmentRecordCreation", "payload": {"subjectId":"aSubjectID","projectId":"aProjectID","moduleId":{"className":"TokenizableString.Tokenized","value":"aModuleID"},"attendantId":{"className":"TokenizableString.Tokenized","value":"anAttendantID"},"biometricReferences":[{"id":"aBiometricReferenceID","templates":[{"finger":"RIGHT_THUMB","quality":0.8,"template":"aTemplate"}],"format":"ISO_19794_2","type":"FINGERPRINT_REFERENCE"},{"id":"anotherBiometricReferenceID","templates":[{"finger":"LEFT_THUMB","quality":0.9,"template":"anotherTemplate"}],"format":"ISO_19794_2","type":"FINGERPRINT_REFERENCE"}]}}]}`,
			expectedErrPrefix: "expect 1 biometric reference",
		},
		{
			name:              "wrong biometric reference type should return error",
			input:             `{"events": [{"id": "anID", "type": "EnrolmentRecordCreation", "payload": {"subjectId":"aSubjectID","projectId":"aProjectID","moduleId":{"className":"TokenizableString.Tokenized","value":"aModuleID"},"attendantId":{"className":"TokenizableString.Tokenized","value":"anAttendantID"},"biometricReferences":[{"id":"aBiometricReferenceID","templates":[{"finger":"RIGHT_THUMB","quality":0.8,"template":"aTemplate"}],"format":"ISO_19794_2","type":"WRONG_TYPE"}]}}]}`,
			expectedErrPrefix: "expect biometric reference type FINGERPRINT_REFERENCE",
		},
		{
			name:              "wrong biometric reference format should return error",
			input:             `{"events": [{"id": "anID", "type": "EnrolmentRecordCreation", "payload": {"subjectId":"aSubjectID","projectId":"aProjectID","moduleId":{"className":"TokenizableString.Tokenized","value":"aModuleID"},"attendantId":{"className":"TokenizableString.Tokenized","value":"anAttendantID"},"biometricReferences":[{"id":"aBiometricReferenceID","templates":[{"finger":"RIGHT_THUMB","quality":0.8,"template":"aTemplate"}],"format":"WRONG_FORMAT","type":"FINGERPRINT_REFERENCE"}]}}]}`,
			expectedErrPrefix: "expect biometric reference format ISO_19794_2",
		},
		{
			name:              "valid input should return the correct subject specification",
			input:             `{"events": [{"id": "anID", "type": "EnrolmentRecordCreation", "payload": {"subjectId":"aSubjectID","projectId":"aProjectID","moduleId":{"className":"TokenizableString.Tokenized","value":"aModuleID"},"attendantId":{"className":"TokenizableString.Tokenized","value":"anAttendantID"},"biometricReferences":[{"id":"aBiometricReferenceID","templates":[{"finger":"RIGHT_THUMB","quality":0.8,"template":"aTemplate"}],"format":"ISO_19794_2","type":"FINGERPRINT_REFERENCE"}]}}]}`,
			expectedErrPrefix: "",
			expectedResult:    SubjectSpecification{SubjectID: "aSubjectID", ProjectID: "aProjectID", TokenizedModuleID: "aModuleID", TokenizedAtendantID: "anAttendantID"},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := Check(tc.input)
			if err != nil {
				if !strings.HasPrefix(err.Error(), tc.expectedErrPrefix) {
					t.Errorf("error got '%s', expectedErrPrefix '%v'", err.Error(), tc.expectedErrPrefix)
				}
				return
			}
			if tc.expectedErrPrefix != "" {
				t.Errorf("error got nil, expectedErrPrefix %v", tc.expectedErrPrefix)
			}
			if got != tc.expectedResult {
				t.Errorf("result got %v, expected %v", got, tc.expectedResult)
			}
		})
	}
}

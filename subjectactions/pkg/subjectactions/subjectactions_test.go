package subjectactions

import (
	"fmt"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheck(t *testing.T) {
	tcs := map[string]struct {
		input          string
		expectedErr    string
		expectedResult SubjectSpecification
	}{
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
		"empty subjectId should return error": {
			input:       `{"events": [{"id": "anID", "type": "EnrolmentRecordCreation", "payload": {"subjectId":"","projectId":"aProjectID","moduleId":{"className":"TokenizableString.Tokenized","value":"aTokenizedModuleID"},"attendantId":{"className":"TokenizableString.Tokenized","value":"aTokenizedAttendantID"},"biometricReferences":[{"id":"aBiometricReferenceID","templates":[{"finger":"RIGHT_THUMB","quality":0.8,"template":"aTemplate"}],"format":"ISO_19794_2","type":"FINGERPRINT_REFERENCE"}]}}]}`,
			expectedErr: "subjectId: empty",
		},
		"empty projectId should return error": {
			input:       `{"events": [{"id": "anID", "type": "EnrolmentRecordCreation", "payload": {"subjectId":"aSubjectID","projectId":"","moduleId":{"className":"TokenizableString.Tokenized","value":"aTokenizedModuleID"},"attendantId":{"className":"TokenizableString.Tokenized","value":"aTokenizedAttendantID"},"biometricReferences":[{"id":"aBiometricReferenceID","templates":[{"finger":"RIGHT_THUMB","quality":0.8,"template":"aTemplate"}],"format":"ISO_19794_2","type":"FINGERPRINT_REFERENCE"}]}}]}`,
			expectedErr: "projectId: empty",
		},
		"wrong moduleId className should return error": {
			input:       `{"events": [{"id": "anID", "type": "EnrolmentRecordCreation", "payload": {"subjectId":"aSubjectID","projectId":"aProjectID","moduleId":{"className":"wrong","value":"aTokenizedModuleID"},"attendantId":{"className":"TokenizableString.Tokenized","value":"aTokenizedAttendantID"},"biometricReferences":[{"id":"aBiometricReferenceID","templates":[{"finger":"RIGHT_THUMB","quality":0.8,"template":"aTemplate"}],"format":"ISO_19794_2","type":"FINGERPRINT_REFERENCE"}]}}]}`,
			expectedErr: "moduleId.className: expect TokenizableString.Tokenized, got wrong",
		},
		"empty moduleId should return error": {
			input:       `{"events": [{"id": "anID", "type": "EnrolmentRecordCreation", "payload": {"subjectId":"aSubjectID","projectId":"aProjectID","moduleId":{"className":"TokenizableString.Tokenized","value":""},"attendantId":{"className":"TokenizableString.Tokenized","value":"aTokenizedAttendantID"},"biometricReferences":[{"id":"aBiometricReferenceID","templates":[{"finger":"RIGHT_THUMB","quality":0.8,"template":"aTemplate"}],"format":"ISO_19794_2","type":"FINGERPRINT_REFERENCE"}]}}]}`,
			expectedErr: "moduleId.value: empty",
		},
		"wrong attendantID className should return error": {
			input:       `{"events": [{"id": "anID", "type": "EnrolmentRecordCreation", "payload": {"subjectId":"aSubjectID","projectId":"aProjectID","moduleId":{"className":"TokenizableString.Tokenized","value":"aTokenizedModuleID"},"attendantId":{"className":"wrong","value":"aTokenizedAttendantID"},"biometricReferences":[{"id":"aBiometricReferenceID","templates":[{"finger":"RIGHT_THUMB","quality":0.8,"template":"aTemplate"}],"format":"ISO_19794_2","type":"FINGERPRINT_REFERENCE"}]}}]}`,
			expectedErr: "attendantId.className: expect TokenizableString.Tokenized, got wrong",
		},
		"empty attendantId should return error": {
			input:       `{"events": [{"id": "anID", "type": "EnrolmentRecordCreation", "payload": {"subjectId":"aSubjectID","projectId":"aProjectID","moduleId":{"className":"TokenizableString.Tokenized","value":"aTokenizedModuleID"},"attendantId":{"className":"TokenizableString.Tokenized","value":""},"biometricReferences":[{"id":"aBiometricReferenceID","templates":[{"finger":"RIGHT_THUMB","quality":0.8,"template":"aTemplate"}],"format":"ISO_19794_2","type":"FINGERPRINT_REFERENCE"}]}}]}`,
			expectedErr: "attendantId.value: empty",
		},
		"empty biometric references should return error": {
			input:       `{"events": [{"id": "anID", "type": "EnrolmentRecordCreation", "payload": {"subjectId":"aSubjectID","projectId":"aProjectID","moduleId":{"className":"TokenizableString.Tokenized","value":"aTokenizedModuleID"},"attendantId":{"className":"TokenizableString.Tokenized","value":"aTokenizedAttendantID"},"biometricReferences":[{"id":"aBiometricReferenceID","templates":[{"finger":"RIGHT_THUMB","quality":0.8,"template":"aTemplate"}],"format":"ISO_19794_2","type":"FINGERPRINT_REFERENCE"},{"id":"anotherBiometricReferenceID","templates":[{"finger":"LEFT_THUMB","quality":0.9,"template":"anotherTemplate"}],"format":"ISO_19794_2","type":"FINGERPRINT_REFERENCE"}]}}]}`,
			expectedErr: "biometricReferences: expect length 1",
		},
		"multiple biometric references should return error": {
			input:       `{"events": [{"id": "anID", "type": "EnrolmentRecordCreation", "payload": {"subjectId":"aSubjectID","projectId":"aProjectID","moduleId":{"className":"TokenizableString.Tokenized","value":"aTokenizedModuleID"},"attendantId":{"className":"TokenizableString.Tokenized","value":"aTokenizedAttendantID"},"biometricReferences":[{"id":"aBiometricReferenceID","templates":[{"finger":"RIGHT_THUMB","quality":0.8,"template":"aTemplate"}],"format":"ISO_19794_2","type":"FINGERPRINT_REFERENCE"},{"id":"anotherBiometricReferenceID","templates":[{"finger":"LEFT_THUMB","quality":0.9,"template":"anotherTemplate"}],"format":"ISO_19794_2","type":"FINGERPRINT_REFERENCE"}]}}]}`,
			expectedErr: "biometricReferences: expect length 1",
		},
		"wrong biometric reference type should return error": {
			input:       `{"events": [{"id": "anID", "type": "EnrolmentRecordCreation", "payload": {"subjectId":"aSubjectID","projectId":"aProjectID","moduleId":{"className":"TokenizableString.Tokenized","value":"aTokenizedModuleID"},"attendantId":{"className":"TokenizableString.Tokenized","value":"aTokenizedAttendantID"},"biometricReferences":[{"id":"aBiometricReferenceID","templates":[{"finger":"RIGHT_THUMB","quality":0.8,"template":"aTemplate"}],"format":"ISO_19794_2","type":"WRONG_TYPE"}]}}]}`,
			expectedErr: "biometricReferences[0].type: expect FINGERPRINT_REFERENCE, got WRONG_TYPE",
		},
		"wrong biometric reference format should return error": {
			input:       `{"events": [{"id": "anID", "type": "EnrolmentRecordCreation", "payload": {"subjectId":"aSubjectID","projectId":"aProjectID","moduleId":{"className":"TokenizableString.Tokenized","value":"aTokenizedModuleID"},"attendantId":{"className":"TokenizableString.Tokenized","value":"aTokenizedAttendantID"},"biometricReferences":[{"id":"aBiometricReferenceID","templates":[{"finger":"RIGHT_THUMB","quality":0.8,"template":"aTemplate"}],"format":"WRONG_FORMAT","type":"FINGERPRINT_REFERENCE"}]}}]}`,
			expectedErr: "biometricReferences[0].format: expect ISO_19794_2, got WRONG_FORMAT",
		},
		"valid input should return the correct subject specification": {
			input:          `{"events": [{"id": "anID", "type": "EnrolmentRecordCreation", "payload": {"subjectId":"aSubjectID","projectId":"aProjectID","moduleId":{"className":"TokenizableString.Tokenized","value":"aTokenizedModuleID"},"attendantId":{"className":"TokenizableString.Tokenized","value":"aTokenizedAttendantID"},"biometricReferences":[{"id":"aBiometricReferenceID","templates":[{"finger":"RIGHT_THUMB","quality":0.8,"template":"aTemplate"}],"format":"ISO_19794_2","type":"FINGERPRINT_REFERENCE"}]}}]}`,
			expectedErr:    "",
			expectedResult: SubjectSpecification{SubjectID: "aSubjectID", ProjectID: "aProjectID", TokenizedModuleID: "aTokenizedModuleID", TokenizedAtendantID: "aTokenizedAttendantID"},
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got, err := Check(tc.input)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedResult, got)
		})
	}
}

func ExampleCheck() {
	input := `{"events": [{"id": "anID", "type": "EnrolmentRecordCreation", "payload": {"subjectId":"aSubjectID","projectId":"aProjectID","moduleId":{"className":"TokenizableString.Tokenized","value":"aTokenizedModuleID"},"attendantId":{"className":"TokenizableString.Tokenized","value":"aTokenizedAttendantID"},"biometricReferences":[{"id":"aBiometricReferenceID","templates":[{"finger":"RIGHT_THUMB","quality":0.8,"template":"aTemplate"}],"format":"ISO_19794_2","type":"FINGERPRINT_REFERENCE"}]}}]}`
	got, err := Check(input)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Subject ID: %s, Project ID: %s, (Tokenized) Module ID: %s, (Tokenized) Attendant ID: %s\n", got.SubjectID, got.ProjectID, got.TokenizedModuleID, got.TokenizedAtendantID)
	// Output:
	// Subject ID: aSubjectID, Project ID: aProjectID, (Tokenized) Module ID: aTokenizedModuleID, (Tokenized) Attendant ID: aTokenizedAttendantID
}

func ExampleCheck_error() {
	input := ""
	_, err := Check(input)
	if err != nil {
		fmt.Println("Error:", err)
	}
	// Output:
	// Error: input is empty
}

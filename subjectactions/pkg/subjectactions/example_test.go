package subjectactions

import (
	"fmt"
	"log"
)

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

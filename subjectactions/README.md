# subjectactions

## Library

### Install the Go library

Run the following command

```shell
go get github.com/simprints/cloud-echisutils/subjectactions
```

### Test

Run `go test ./...`

### Check `subjectActions` is valid

The `subjectActions` field [passed by Simprints ID to CommCare](https://simprints.gitbook.io/docs/development/simprints-for-developers/other-intergrations/commcare-integration/cosync#saving-the-biometric-data) can be validated by the code provided in this package.

Look at the documentation of `subjectactions.Check(input string)` for more details.

## CLI

### Build the CLI

Run the following command

```shell
go build -o subject-actions-checker github.com/simprints/cloud-echisutils/subjectactions/cmd
```

### Usage of the CLI

To run the CLI, you need to pass the subject actions that you want to validate either as a file:
```shell
# Unix based
./subject-actions-checker --file path/to/subject_actions.json

# Windows
.\subject-actions-checker.exe --file path/to/subject_actions.json

```

or as a string:
```shell
# Unix based
./subject-actions-checker --input="THE_SUBJECT_ACTIONS_STRING"

# Windows, make sure to escape double quotes correctly depending on whether you are using Powershell or CMD
.\subject-actions-checker.exe --input="THE_SUBJECT_ACTIONS_STRING"
```

For example:
```shell
# Unix based
./subject-actions-checker --input='{"events": [{"id": "anID", "type": "EnrolmentRecordCreation", "payload": {"subjectId":"aSubjectID","projectId":"aProjectID","moduleId":{"className":"TokenizableString.Tokenized","value":"aTokenizedModuleID"},"attendantId":{"className":"TokenizableString.Tokenized","value":"aTokenizedAttendantID"},"biometricReferences":[{"id":"aBiometricReferenceID","templates":[{"finger":"RIGHT_THUMB","quality":0.8,"template":"aTemplate"}],"format":"ISO_19794_2","type":"FINGERPRINT_REFERENCE"}]}}]}'
```

If the subject actions is valid, the CLI will print the subject specification as a JSON to the standard output and exit with
a code `0`.
If the subject actions is not valid, the CLI will print an error message to the standard output and exit with a code `1`.

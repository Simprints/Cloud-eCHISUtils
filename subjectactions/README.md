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
./subject-actions-checker -file path/to/subject_actions.json
```

or as a string:
```shell
./subject-actions-checker --input="THE_SUBJECT_ACTIONS_STRING"
```

The CLI will exit with a code `0` if the subject action is valid, or with a code `1` if it's not valid. 
In addition, it will print a message to the standard output if the subject action is invalid.
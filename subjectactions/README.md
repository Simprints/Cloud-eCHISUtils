# subjectactions

## Install the Go library

Run the following command

```shell
go get github.com/simprints/cloud-echisutils/subjectactions
```

## Test

Run `go test ./...`

## Check `subjectActions` is valid

The `subjectActions` field [passed by Simprints ID to CommCare](https://simprints.gitbook.io/docs/development/simprints-for-developers/other-intergrations/commcare-integration/cosync#saving-the-biometric-data) can be validated by the code provided in this package.

Look at the documentation of `subjectactions.Check(input string)` for more details.

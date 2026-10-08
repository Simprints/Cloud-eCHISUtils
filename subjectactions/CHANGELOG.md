# subjectactions' changelog

## 0.2.0 (October 8th, 2026)

IMPROVEMENTS:

- Accept face biometric references (`FACE_REFERENCE`) and multiple biometric references, in any format.
- Accept and validate external credentials (`externalCredentials`); credential values must be tokenized.
- Accept and validate the optional `schemaVersion` (absent, null, or major version `1`; an empty string is rejected, as in SID).
  We also reject higher versions because if the schema of `subjectActions` ever changes, we want to update this library.
- Validate event IDs, biometric reference IDs, base64 templates and finger identifiers.
- Require at least one biometric reference, and optionally, an external credential.
- Add the `WithAllowedFormats` option and the `--formats` CLI flag to restrict biometric reference formats.
- Add `BiometricFormats` and `ExternalCredentialTypes` to `SubjectSpecification`.
- Deprecate `SIDTemplate.Quality`, which is no longer part of SID's schema.
- Document where SID's schema is defined.

## 0.1.0 (September 9th, 2025)

Initial release

# Local contributor instructions

- Keep source-specific parsing, validation and normalization in this module.
- Do not invent canonical event identity when the source does not provide it.
- Preserve raw source evidence and report source-scoped limitations explicitly.
- Do not add live network access to ordinary normalization or tests.
- Run `go vet ./...` and `go test -count=1 ./...` before handing off changes.
- The adapter may depend on released core contracts, but must not depend on the CLI.

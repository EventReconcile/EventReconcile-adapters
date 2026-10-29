# EventReconcile-adapters

The first adapter target is SoroTrail NDJSON. The adapter validates and normalizes records into either canonical or source-scoped identity mode. It deliberately refuses to substitute SoroTrail's `op_index` for `event_index`, because those are not equivalent identity fields.

The current documented SoroTrail event shape includes `ledger`, `tx_hash`, `tx_index`, `op_index`, `topics` and `value`, but does not expose `event_index`. The adapter preserves its native `id` as a source-scoped identity rather than inventing an event index. Same-source analysis can therefore remain useful; cross-source canonical assertions remain `INCONCLUSIVE` without `event_index`. See ADR-002.

The loader accepts only `type: "contract"` records from one contract per export, preserves each raw NDJSON line for evidence, and enforces bounded file, line and record sizes. It performs no network access.

For the offline cross-repository workflow, build a core input document with:

```powershell
go run ./cmd/sorotrail-normalize --reference ref.ndjson --observed observed.ndjson --output input.json
eventreconcile --input input.json --output report.json
```

The normalizer depends only on the released core contracts; the CLI consumes the generated core input and does not import this adapter.

# Agent Instructions

## Project

Address Quality is an Indonesian address validation system.

The core validation flow is conceptually:

Normalize -> Evidence -> Candidates -> Hierarchy -> Score

Repository code is the source of truth for the actual implementation. Inspect it before making changes.

## Core Engineering Principle

> Never Improve Blindly.

Changes must be driven by evidence, tests, benchmarks, or a clearly identified defect.

Do not optimize, refactor, or redesign unrelated code while completing a scoped task.

## Agent Working Rules

Before editing:

1. Read the relevant implementation.
2. Read nearby tests.
3. Check existing documentation and configuration.
4. Understand the request flow and existing conventions.
5. Identify assumptions that cannot be established from repository evidence.

Prefer the smallest change that satisfies the task.

Do not introduce infrastructure or abstractions merely because they are common industry practice.

Do not silently reinterpret requirements. If repository evidence contradicts the task, document the contradiction.

## Address Quality Correctness

Treat validation behavior as sensitive.

Do not change any of the following unless the task explicitly requires it:

- normalization behavior
- evidence extraction semantics
- candidate generation
- hierarchy validation
- scoring
- confidence calculation
- status classification
- postal-code behavior
- public API contracts

Preserve the principle that evidence and candidate resolution are separate concerns.

Do not use performance work as an excuse to modify matching behavior.

When touching validation code, add or update tests that demonstrate behavior before and after the change.

## Agentic V1 Accuracy Workflow

### Scope

- Active accuracy development targets V1 only.
- V0 is frozen and has no active development. Do not modify V0 code, tests, benchmark scripts, benchmark results, reports, commands, or other V0-specific assets.
- Do not run `make benchmark-v0` unless the user explicitly requests it.
- If a V1 change appears to require modifying V0 behavior or assets, stop and ask the user for direction.

### Workflow

For every change that can affect V1 address parsing, normalization, matching, accuracy, or validation API output:

1. Before editing, run `make benchmark` to establish a fresh baseline.
2. Immediately run `make benchmark-page` and supply a unique `benchmark_build` label to record the baseline in `tests/api/page/metadata.json`.
3. Inspect `tests/api/page/full-benchmark.html` locally to identify failure patterns, then prioritize broadly applicable fixes. Do not overfit individual benchmark rows.
4. Implement a focused fix.
5. Run the relevant unit and API tests.
6. Run `make benchmark` again with the same API target, source, API version, configuration, and dataset.
7. Immediately run `make benchmark-page` with a new unique `benchmark_build` label so the iteration is appended to `tests/api/page/metadata.json`.
8. Compare the baseline and iteration metadata entries for overall exact matches across province, city, district, and subdistrict, and for province, city, district, and subdistrict accuracy.
9. Continue diagnosing, implementing, testing, and running the paired benchmark and benchmark-page commands until overall exact-match accuracy improves with no degradation in any hierarchy metric, or until a genuine blocker requires user input.

Every V1 benchmark run must be followed by `make benchmark-page`, including the baseline and every attempted iteration. Do not stop merely because the first attempted fix fails to improve accuracy. Never claim an accuracy improvement without benchmark evidence. Changes that cannot affect V1 address accuracy do not require this benchmark cycle.

### Benchmark and API Test Authority

- Refer to `tests/api/` and its README files for V1 benchmark and API-test commands, inputs, outputs, and configuration.
- The canonical V1 benchmark command is `make benchmark`.
- The canonical V1 benchmark runner is `tests/api/benchmark-test.js`.
- V1 benchmark results are written under `tests/api/benchmark/`.
- After every benchmark run, run `make benchmark-page`; a benchmark run is not complete until its metadata entry has been generated.
- Use a unique `benchmark_build` label for each run so `tests/api/page/build.js` appends a new entry instead of replacing an entry with the same release and label. Use labels such as `agent-v1-<timestamp>-baseline` and `agent-v1-<timestamp>-iteration-<n>`.
- Use `tests/api/page/metadata.json` as the authoritative source for accuracy comparisons. Identify entries by both `benchmark_build` and `benchmark_source`; do not compare unrelated historical entries merely because they are adjacent.
- Read `metrics.overall_accuracy` and every item in `metrics.hierarchy` from the selected metadata entries when calculating deltas.
- Ignore `tests/api/page/benchmark.html` for accuracy analysis. It is a generated publishing artifact, not the comparison source.
- Use `tests/api/page/full-benchmark.html` only for local diagnosis of failures, raw addresses, resolved values, and expected values. It contains confidential input and must never be published or exposed.
- Leave generated V1 benchmark results, `metadata.json`, and page outputs uncommitted for user review.
- If the API service or required dependencies are unavailable, diagnose safe in-scope causes and report the blocker. Never fabricate or estimate benchmark results.

### Protected Benchmark Assets

- Never modify `tests/api/benchmark-test.js` or related benchmark evaluation, scoring, or reporting logic to improve reported accuracy without explicit user approval.
- Never modify `tests/api/cases/address-tagged.csv` without explicit user approval.
- Never edit metrics in `tests/api/page/metadata.json` manually; it must be updated only by the required V1 `make benchmark-page` workflow.
- Before making any protected change, stop and ask the user for approval.
- An improvement caused by changing the evaluator or ground-truth dataset is not an implementation accuracy improvement.

### Degradation Reporting

- Report every negative benchmark delta, even if another metric or overall accuracy improves.
- For each degradation, provide the metric name, baseline value, post-change value, and delta.
- A V1 accuracy task is complete only when overall exact-match accuracy improves and province, city, district, and subdistrict accuracy do not decrease.
- In the final report, include the baseline and final `benchmark_build` labels, their `benchmark_source` files, before-and-after metrics, all degradations, relevant tests, remaining failures, and blockers.

## Agentic V1 Performance Workflow

Performance work covers aq-chore-53 (instrumentation), aq-chore-54 (k6 tests), and aq-chore-52 (Performance Report Audit Page). Unless a task explicitly changes the plan, use this order:

1. Repository/performance audit.
2. Performance measurement contract.
3. aq-chore-53 instrumentation.
4. aq-chore-54 k6 benchmark.
5. Benchmark artifact integration.
6. aq-chore-52 Performance Report.
7. End-to-end verification.
8. Continuous observability stack as a separate future task.

The dependency direction is performance contract -> instrumentation -> k6 benchmark -> machine-readable benchmark artifact -> Performance Report. The report consumes measurements; do not hard-code values that can come from generated artifacts.

### Measurement ownership and integrity

- k6 is authoritative for client-observed HTTP latency, p50/p90/p95/p99, throughput, request count, HTTP errors, status-code failures, and test concurrency.
- Application instrumentation explains server-side time. Runtime measurements provide diagnostic context only where they can be collected reliably. Do not duplicate metrics without a diagnostic reason.
- Missing data is not zero. Represent unavailable measurements explicitly, such as `not_measured` or null plus a reason. Never fabricate measurements.
- Do not invent SLOs, thresholds, or pass/fail criteria. Use documented targets if present; otherwise use TBD. Existing script thresholds are not automatically service SLOs.
- Preserve available timestamp, Git revision, application version, scenario, duration, concurrency/load, dataset identity/count, target environment, and runtime/environment characteristics in benchmark artifacts. Do not present runs as directly comparable when relevant workload or environment differs materially.
- The performance workload for this task is the five addresses hard-coded in `tests/api/load-test.js`. Use those five only unless the user changes scope. The private/uncommitted accuracy dataset and historical accuracy result counts (including 106) are separate; do not substitute them into performance testing. Describe the five-address workload as limited and do not claim it represents production traffic.
- Never automatically run stress tests against production. Stress tests require an explicitly selected non-production target and safeguards.

### Instrumentation and artifacts

Instrumentation must have bounded cardinality, avoid raw addresses, API keys, and sensitive data in labels, remain safe under concurrent requests, have low overhead, and avoid changing validation results. Instrument meaningful stages, not every function. Do not expose benchmark/debug endpoints publicly by accident; inspect deployment architecture and restrict any required snapshot access.

The benchmark pipeline must not require Grafana, Prometheus, or an OpenTelemetry Collector. Future observability may reuse transport-neutral application instrumentation, but do not install, configure, or deploy that stack unless explicitly requested. Controlled benchmarking answers “what happened during this experiment?”; continuous observability answers “what is happening over time?”.

Performance runs should produce a versioned, machine-readable artifact, preferably JSON, capable of representing schema version, metadata, environment, workload, k6, application, runtime, missing measurements, configured thresholds, and observations. Follow the contract in `agents/`. The report must consume the artifact, not raw k6 console output.

## Testing

Every implementation stage must be verifiable independently. Before declaring a task complete:

1. Build the affected code.
2. Run relevant unit/integration tests.
3. Run required formatting/linting.
4. Exercise changed behavior.
5. Verify generated artifacts where applicable.

For performance work, verify the complete path when applicable: Address Quality -> k6 workload -> client/application/runtime measurements -> benchmark artifact -> Performance Report. A file existing is not evidence that the feature works.

## Documentation

Document decisions future agents need to understand: what is measured, why, boundaries, limitations, reproduction steps, and assumptions. Avoid documentation that merely restates code. Task-produced Markdown artifacts for other agents belong under `agents/`.

## Security and scope control

Do not expose benchmark/debug endpoints publicly by accident. If diagnostic endpoints are needed, inspect deployment architecture and choose an appropriate restriction such as disabled by default, environment-gated, localhost-only, or explicit benchmark mode. Do not commit secrets, API keys, credentials, or sensitive request payloads.

Stop and request human input before changing the public API contract, validation/scoring semantics, adding a major persistent dependency, introducing substantial paid infrastructure, changing production deployment architecture, defining new SLOs or business performance targets, or running destructive/stress workloads against production. Ordinary implementation decisions within an approved design do not require approval.

## Agent Handoff

For multi-agent work, leave the repository understandable to the next agent. At the end of a task report what changed, files changed, tests/verification performed, generated artifacts, assumptions, unresolved issues, and recommended next step. Do not claim completion when verification failed.

## Git and Working Tree Safety

- Do not commit, amend, rebase, push, or otherwise publish changes unless the user explicitly requests it.
- Leave implementation changes and generated V1 accuracy benchmark artifacts uncommitted for user review.
- Preserve unrelated working-tree changes and do not overwrite or revert them.

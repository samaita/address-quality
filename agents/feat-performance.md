Address Quality Performance — Stage 0 & 1 Review

You are the architecture/review agent for the Address Quality performance engineering work.

Context

This review is the prerequisite for:

● aq-chore-53 — Setup Performance Preparation Metric Telemetry

● aq-chore-54 — Setup k6 Performance Test

● aq-chore-52 — Build Performance Report Audit Page

A separate future task will handle the full observability stack such as Prometheus, Grafana, and potentially OpenTelemetry infrastructure. The current machine does NOT have Grafana, Prometheus, or an OpenTelemetry Collector installed.

Do NOT install them in this task.

The goal of this task is to inspect the repository, understand the existing system, and define a trustworthy performance measurement contract before another agent starts implementation.

The implementation agent will likely be MiMo 2.6 Pro.

Core Principle

> Never Improve Blindly.

Do not optimize anything in this task.

Do not modify Address Quality matching, scoring, candidate generation, evidence handling, or API behavior.

Your job is to determine what should be measured and how the later implementation should be structured.




Stage 0 — Repository Audit

Inspect the repository thoroughly before proposing architecture.

Find and understand:

1. The /address-quality/v1/validate request flow.

2. HTTP server/framework setup.

3. Validation pipeline and meaningful internal stages.

4. Persistence/database access involved in validation.

5. External dependencies involved in validation.

6. Existing logging.

7. Existing metrics or tracing.

8. Existing benchmark/load-test code.

9. Existing k6 files or scripts.

10. Existing performance documentation.

11. Existing report/page architecture.

12. Deployment/container configuration.

13. Runtime constraints.

14. Existing test datasets that could represent Address Quality workloads.

Do not assume architecture from filenames alone. Read the relevant implementation.

Create:

docs/performance/00-current-state.md

Current-State Matrix

For each measurement below classify it as:

● AVAILABLE

● PARTIAL

● MISSING

● NOT_APPLICABLE

Evaluate:

● request count

● HTTP request latency

● latency distribution

● p50

● p90

● p95

● p99

● throughput / requests per second

● error rate

● HTTP status distribution

● concurrency

● CPU

● memory

● Go runtime metrics

● GC metrics

● goroutine count

● database/persistence latency

● external dependency latency

● internal Address Quality stage latency

For every AVAILABLE or PARTIAL item, identify exactly where the measurement comes from.

For every MISSING item, state whether it is actually required for the baseline.

Do not recommend instrumentation merely because it is commonly used.




Identify Measurement Boundaries

Map the actual validation flow.

Look for meaningful boundaries such as:

● normalization

● evidence extraction

● candidate retrieval

● hierarchy validation

● scoring

● persistence

● external lookups

These names are examples only.

Use the repository’s real architecture and terminology.

Do NOT propose timing every function.

For each potential boundary answer:

1. What question would measuring this answer?

2. Would the measurement help diagnose a regression?

3. Is it cheap enough to collect?

4. Is it stable enough to become part of the performance contract?

Reject metrics that do not provide useful diagnostic evidence.




Stage 1 — Performance Contract

Create:

docs/performance/01-performance-contract.md

This document becomes the contract used by the implementation agent.

Define three measurement domains.

A. k6 / Client-Observed Measurements

Determine what k6 should own.

At minimum evaluate:

● request count

● HTTP latency

● p50

● p90

● p95

● p99

● throughput

● HTTP error rate

● status-code failures

● VUs/concurrency

k6 should remain authoritative for client-observed HTTP performance.

B. Application Measurements

Determine which internal measurements Address Quality itself should expose for benchmark diagnostics.

Examples may include major validation-stage latency and request/error counters.

Only include metrics justified by the repository architecture.

Avoid duplication with k6 unless the server-side measurement answers a different question.

C. Runtime / Environment Measurements

Determine what can reliably be collected without installing a monitoring platform.

Evaluate:

● Go version

● process uptime

● goroutines

● heap

● allocations

● GC

● process memory

● CPU

If something cannot be measured reliably with the available environment, mark it:

NOT MEASURED

Never represent unavailable data as zero.




Metric Definition

For every metric accepted into the contract document:

● canonical name

● definition

● unit

● measurement source

● reason it exists

● expected collection mechanism

● labels/dimensions if applicable

● cardinality risk

● baseline-report requirement: REQUIRED / OPTIONAL

● limitations

Explicitly identify measurements that should NOT be implemented.

Do not invent SLOs or thresholds.

If the repository already defines targets, cite the repository source.

Otherwise use:

TBD




Benchmark Architecture Recommendation

Based on repository evidence, propose the smallest architecture that can support:

k6 workload

     |

     v

Address Quality API

     |

     +--> lightweight application/runtime measurements

     |

     v

machine-readable benchmark artifact

     |

     v

Performance Report

The benchmark pipeline must NOT require:

● Grafana

● Prometheus

● OpenTelemetry Collector

However, design the instrumentation so a future observability task can reasonably connect Prometheus/OpenTelemetry without rewriting the Address Quality core.

Do not implement that future stack now.




Benchmark Artifact Contract

Recommend a versioned machine-readable result schema that can eventually contain:

schema_version

metadata

environment

workload

k6

application

runtime

missing_measurements

thresholds

observations

This is conceptual, not mandatory structure.

Derive the actual recommendation from repository evidence.

The schema must support future comparison between:

● initial baseline

● previous release

● current release

● before optimization

● after optimization

Missing values must remain explicitly unavailable rather than becoming 0.

The report must not parse raw k6 console output.




Workload Review

Inspect existing datasets.

Recommend which data should drive:

1. smoke test

2. baseline test

3. load test

4. stress test

Do not claim that a tiny fixture represents production traffic.

Identify limitations in workload representativeness.

Stress testing must never target production automatically.




Performance Report Requirements

Do NOT build the report.

Define what aq-chore-52 will eventually consume and display.

At minimum evaluate whether the future report should contain:

● Executive Summary

● Environment

● Workload/Test Configuration

● Latency

● Throughput

● Reliability

● Resource Usage

● Internal Diagnostics

● Before/After Comparison

● Methodology / Audit Information

Define safeguards against misleading comparisons.

In particular, determine which environment/workload properties must match before two benchmark runs can reasonably be compared.




Future Observability Boundary

A separate future task will deploy the continuous observability stack.

Define the boundary clearly.

Benchmark/performance work answers:

> What happened during this controlled experiment?

Continuous observability answers:

> What is happening to the running system over time?

Recommend which instrumentation from this performance work should later be reusable by:

● Prometheus

● Grafana

● OpenTelemetry

But do not install, configure, or implement those systems now.




Deliverables

Produce exactly these primary artifacts:

1. docs/performance/00-current-state.md

2. docs/performance/01-performance-contract.md

Optionally create a third document only if necessary:

docs/performance/02-implementation-handoff.md

The handoff should contain a concrete ordered implementation plan for MiMo 2.6 Pro covering:

1. aq-chore-53

2. aq-chore-54

3. benchmark artifact assembly

4. aq-chore-52

5. end-to-end verification

Keep implementation tasks small enough to verify independently.




Final Review

At the end report:

Repository Findings

What actually exists today.

Architectural Decisions

What you recommend and why.

Rejected Complexity

What you intentionally chose NOT to add.

Risks

Potential sources of misleading measurements.

Open Questions

Only questions that genuinely require human judgment.

Implementation Handoff

Exact next steps for the coding agent.




Hard Constraints

Do NOT:

● optimize application performance

● change validation behavior

● change scoring

● change public API contracts

● install Grafana

● install Prometheus

● deploy OpenTelemetry Collector

● invent SLOs

● invent benchmark results

● fabricate production characteristics

● add unrelated refactoring

● treat missing measurements as zero

● build the Performance Report yet

You are reviewing and defining the measurement contract.

Prefer repository evidence over assumptions.

If repository evidence contradicts this prompt, document the contradiction instead of silently forcing the requested architecture.
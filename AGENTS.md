AGENTS.md

Project

Address Quality is an Indonesian address validation system.

The core validation flow is conceptually:

Normalize -> Evidence -> Candidates -> Hierarchy -> Score

Repository code is the source of truth for the actual implementation. Inspect it before making changes.

Core Engineering Principle

> Never Improve Blindly.

Changes must be driven by evidence, tests, benchmarks, or a clearly identified defect.

Do not optimize, refactor, or redesign unrelated code while completing a scoped task.

Agent Working Rules

Before editing:

1. Read the relevant implementation.

2. Read nearby tests.

3. Check existing documentation and configuration.

4. Understand the request flow and existing conventions.

5. Identify assumptions that cannot be established from repository evidence.

Prefer the smallest change that satisfies the task.

Do not introduce infrastructure or abstractions merely because they are common industry practice.

Do not silently reinterpret requirements. If repository evidence contradicts the task, document the contradiction.

Address Quality Correctness

Treat validation behavior as sensitive.

Do not change any of the following unless the task explicitly requires it:

● normalization behavior

● evidence extraction semantics

● candidate generation

● hierarchy validation

● scoring

● confidence calculation

● status classification

● postal-code behavior

● public API contracts

Preserve the principle that evidence and candidate resolution are separate concerns.

Do not use performance work as an excuse to modify matching behavior.

When touching validation code, add or update tests that demonstrate behavior before and after the change.

Performance Engineering

Performance work currently covers:

● aq-chore-53 — performance instrumentation

● aq-chore-54 — k6 performance tests

● aq-chore-52 — Performance Report Audit Page

The dependency direction is:

Performance contract

        |

        v

Instrumentation

        |

        v

k6 benchmark

        |

        v

Benchmark artifact

        |

        v

Performance Report

The report is a consumer of measurements. Do not hard-code benchmark numbers into the report when they can come from generated artifacts.

Measurement Ownership

k6 is authoritative for client-observed measurements such as:

● HTTP latency

● p50/p90/p95/p99

● throughput

● request count

● HTTP errors

● test concurrency

Application instrumentation exists to explain where server-side time is spent.

Runtime measurements provide additional diagnostic context where they can be collected reliably.

Do not duplicate metrics without a diagnostic reason.

Missing Measurements

Missing data is not zero.

Represent unavailable measurements explicitly, for example:

not_measured

or the equivalent defined by the benchmark schema.

Never fabricate measurements.

Performance Targets

Do not invent SLOs, thresholds, or pass/fail criteria.

Use existing documented targets when they exist.

Otherwise mark them as TBD.

Benchmark Integrity

A benchmark result must preserve enough metadata to understand how it was produced.

Where available, capture:

● timestamp

● git revision

● application version

● scenario

● duration

● concurrency/load

● dataset identity/count

● target environment

● runtime/environment characteristics

Do not present two runs as directly comparable when their relevant workload or environment differs materially.

Do not automatically run stress tests against production.

Observability Boundary

The current development machine may not have:

● Prometheus

● Grafana

● OpenTelemetry Collector

The benchmark pipeline must not depend on those systems.

Continuous observability is a separate concern from controlled performance benchmarking.

Conceptually:

                    Address Quality

                         |

                  Instrumentation

                    /          \

                   /            \

          Benchmarking       Observability

               |                  |

              k6          Prometheus / OTel

               |                  |

        benchmark artifact      Grafana

               |

        Performance Report

Future observability infrastructure should be able to reuse application instrumentation where practical, but do not introduce or deploy that infrastructure unless the active task explicitly requests it.

Grafana is visualization for operational monitoring. It is not the canonical storage format for benchmark evidence.

Instrumentation

Instrumentation must:

● have bounded cardinality

● avoid raw addresses in metric labels

● avoid API keys and sensitive request data

● remain safe under concurrent requests

● avoid changing validation results

● have reasonably low overhead

Instrument meaningful stages, not every function.

Choose measurement boundaries because they answer useful diagnostic questions.

Benchmark Artifacts

Performance runs should produce a versioned, machine-readable artifact.

Prefer a stable structured format such as JSON.

The artifact should be capable of representing:

● schema version

● metadata

● environment

● workload

● k6 measurements

● application measurements

● runtime measurements

● missing measurements

● configured thresholds

● observations

The exact schema must follow the implementation contract defined in the repository.

The Performance Report should consume this artifact rather than parse console output.

Testing

Every implementation stage must be verifiable independently.

Before declaring a task complete:

1. build the affected code

2. run relevant unit/integration tests

3. run formatting/linting required by the repository

4. exercise the changed behavior

5. verify generated artifacts where applicable

A file existing is not evidence that the feature works.

For performance work, verify the complete path when applicable:

Address Quality

      |

      v

k6 workload

      |

      +--> client measurements

      |

      +--> application/runtime measurements

      |

      v

benchmark artifact

      |

      v

Performance Report

Documentation

Document decisions that future agents need to understand.

Prefer explaining:

● what is measured

● why it is measured

● measurement boundaries

● limitations

● reproduction steps

● assumptions

Avoid documentation that merely restates code.

Security

Do not expose benchmark/debug endpoints publicly by accident.

If diagnostic endpoints are required, inspect the deployment architecture and choose an appropriate restriction such as:

● disabled by default

● environment gated

● localhost only

● explicit benchmark mode

Do not commit secrets, API keys, credentials, or sensitive request payloads.

Scope Control

Stop and request human input before:

● changing the public API contract

● changing address validation/scoring semantics

● adding a major persistent dependency

● introducing substantial paid infrastructure

● changing production deployment architecture

● defining new SLOs or business performance targets

● running destructive or stress workloads against production

Ordinary implementation decisions within an approved design do not require approval.

Agent Handoff

For multi-agent work, leave the repository understandable to the next agent.

At the end of a task report:

1. what changed

2. files changed

3. tests/verification performed

4. generated artifacts

5. assumptions

6. unresolved issues

7. recommended next step

Do not claim completion when verification failed.

Current Performance Task Order

Unless a task explicitly changes the plan, use:

1. repository/performance audit

2. performance measurement contract

3. aq-chore-53 instrumentation

4. aq-chore-54 k6 benchmark

5. benchmark artifact integration

6. aq-chore-52 Performance Report

7. end-to-end verification

8. continuous observability stack as a separate future task

Keep each stage independently reviewable and avoid unrelated refactoring.
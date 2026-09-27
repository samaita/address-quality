#!/usr/bin/env node
/*
 * Builds the performance page (tests/api/page/performance.html).
 *
 * The accuracy page (build.js -> benchmark.html) is the model, but performance
 * data is a different shape, so this has its own metadata and its own HTML:
 *
 *   performance-metadata.json   append-only history, one entry per perf run
 *   performance-template.html   page shell with the __AQ_DATA__ marker
 *   performance.html            generated output
 *
 * Reads the machine-readable benchmark artifact produced by aq-chore-54
 * (agents/aq-chore-54-performance-benchmark.json by default). Values are never
 * hand-copied: whatever the artifact does not measure is carried through as
 * not_measured with its reason.
 *
 * BEFORE / AFTER
 * --------------
 * Each metadata entry carries BOTH sides explicitly:
 *
 *   { performance_build, role, before: <snapshot|null>, after: <snapshot> }
 *
 * `before` is an embedded snapshot of the baseline entry's metrics, addressed
 * by its performance_build label. It is deliberately NOT taken from the
 * previous array element. AGENTS.md forbids comparing runs merely because they
 * are adjacent in the file, and adjacency silently compares unrelated runs as
 * soon as a second workload or scenario is recorded. So the baseline must be
 * named: `--before <label>`. Refusing to guess is the point.
 *
 * Usage:
 *   node tests/api/page/performance-build.js <performance_build> \
 *        [--before <label>] [--source <artifact.json>]
 *
 * The first entry has no baseline: it is recorded with role "baseline" and
 * before: null. Every later entry must name its baseline or it is rejected.
 */
'use strict';

const fs = require('fs');
const path = require('path');

const PAGE_DIR = __dirname;
const REPO_DIR = path.join(PAGE_DIR, '..', '..', '..');
const META_FILE = path.join(PAGE_DIR, 'performance-metadata.json');
const TEMPLATE = path.join(PAGE_DIR, 'performance-template.html');
const OUTPUT = path.join(PAGE_DIR, 'performance.html');
const RELEASE_FILE = path.join(PAGE_DIR, 'release.json');
const DEFAULT_ARTIFACT = path.join(REPO_DIR, 'agents', 'aq-chore-54-performance-benchmark.json');

function readJson(filePath) {
  return JSON.parse(fs.readFileSync(filePath, 'utf-8'));
}

function relArtifact(filePath) {
  return path.relative(REPO_DIR, path.resolve(filePath)).split(path.sep).join('/');
}

/** Every {status:"not_measured"} in the artifact, with its path and reason. */
function collectNotMeasured(node, trail, out) {
  if (!node || typeof node !== 'object') return out;
  if (!Array.isArray(node) && node.status === 'not_measured') {
    out.push({ path: trail || '(root)', reason: node.reason || null, value: node.value });
    return out;
  }
  if (Array.isArray(node)) {
    node.forEach((v, i) => collectNotMeasured(v, `${trail}[${i}]`, out));
    return out;
  }
  for (const [k, v] of Object.entries(node)) {
    collectNotMeasured(v, trail ? `${trail}.${k}` : k, out);
  }
  return out;
}

/**
 * The one metrics shape both sides share. Only what the artifact actually
 * measures; a stage that never ran is absent, never a zero.
 */
function metricsFrom(artifact) {
  const m = (artifact && artifact.measurements) || {};
  const client = m.client_observed || {};
  const warm = m.warm_steady_state || {};
  const cold = m.cold_start || {};

  const stages = Object.entries(warm.server_spans || {}).map(([stage, v]) => ({
    stage,
    count: v.count,
    sum_ms: v.sum_ms,
    share_of_stage_time_pct: v.share_of_stage_time_pct,
    calls_per_request: v.calls_per_request,
    span_status: v.span_status,
    http_status_class: v.http_status_class,
  }));

  return {
    requests: client.requests != null ? client.requests : null,
    throughput_rps: client.throughput_rps != null ? client.throughput_rps : null,
    http_errors: client.http_errors != null ? client.http_errors : null,
    http_error_rate: client.http_error_rate != null ? client.http_error_rate : null,
    thresholds_ok: client.thresholds_ok || null,
    http_req_duration_ms: client.http_req_duration_ms || null,
    http_req_waiting_ms: client.http_req_waiting_ms || null,
    iteration_duration_ms: client.iteration_duration_ms || null,
    cold_start_ms: cold.client_observed_time_total_s != null
      ? Math.round(cold.client_observed_time_total_s * 1000000) / 1000
      : null,
    cold_stages: cold.server_spans || null,
    stage_time_coverage_pct: m.stage_time_coverage_pct != null ? m.stage_time_coverage_pct : null,
    stages,
    runtime_characteristics: m.runtime_characteristics || null,
  };
}

/** One side of the comparison, carrying its own context so unlike runs can be
 *  spotted instead of silently differenced. */
function snapshotFrom(artifact, artifactSource, performanceBuild) {
  const p = artifact.provenance || {};
  const t = artifact.target_environment || {};
  const w = artifact.workload || {};
  const s = artifact.scenario || {};
  return {
    performance_build: performanceBuild,
    artifact_source: artifactSource,
    git_commit: p.git_revision || null,
    generated_at: artifact.generated_at || null,
    target_environment: {
      kind: t.kind || null,
      base_url: t.base_url || null,
      host: t.host || null,
      kernel: t.kernel || null,
      cpus: t.cpus != null ? t.cpus : null,
      production: t.production != null ? t.production : null,
    },
    workload: { id: w.id || null, count: w.count != null ? w.count : null, source: w.source || null },
    scenario: {
      kind: s.kind || null,
      max_vus: s.max_vus != null ? s.max_vus : null,
      configured_duration_s: s.configured_duration_s != null ? s.configured_duration_s : null,
      client_iterations: s.client_iterations != null ? s.client_iterations : null,
    },
    metrics: metricsFrom(artifact),
  };
}

function comparability(before, after) {
  if (!before) return { comparable: true, reasons: [] };
  const reasons = [];
  // Compare scenario INPUTS only. client_iterations is a result, and flagging
  // it would cry "not like-for-like" on every run until the warning is ignored.
  const diffs = (label, a, b, keys) => {
    const ks = keys || Object.keys(a || {});
    const changed = ks.filter((k) => (a || {})[k] !== (b || {})[k]);
    if (changed.length) reasons.push(`${label} differs: ${changed.join(', ')}`);
  };
  diffs('workload', after.workload, before.workload);
  diffs('scenario', after.scenario, before.scenario, ['kind', 'max_vus', 'configured_duration_s']);
  diffs('target_environment', after.target_environment, before.target_environment);
  return { comparable: reasons.length === 0, reasons };
}

function renderTemplate(payload) {
  const json = JSON.stringify(payload)
    .replace(/</g, '\\u003c')
    .replace(/\u2028/g, '\\u2028')
    .replace(/\u2029/g, '\\u2029');
  let html = fs.readFileSync(TEMPLATE, 'utf-8');
  if (!html.includes('__AQ_DATA__')) {
    console.error('performance-template.html is missing the __AQ_DATA__ marker.');
    process.exit(1);
  }
  return html.replace('__AQ_DATA__', json);
}

function parseArgs(argv) {
  const args = { label: null, before: null, source: null };
  const rest = argv.slice(2);
  for (let i = 0; i < rest.length; i++) {
    const a = rest[i];
    if (a === '--before') args.before = (rest[++i] || '').trim() || null;
    else if (a === '--source') args.source = (rest[++i] || '').trim() || null;
    else if (!args.label) args.label = a.trim();
  }
  return args;
}

function readMetadataEntries() {
  if (!fs.existsSync(META_FILE)) return [];
  const trimmed = fs.readFileSync(META_FILE, 'utf-8').trim();
  if (!trimmed) return [];
  const existing = JSON.parse(trimmed);
  return Array.isArray(existing) ? existing : [existing];
}

function main() {
  const args = parseArgs(process.argv);
  if (!args.label) {
    console.error(
      'performance_build is required: `node tests/api/page/performance-build.js <performance_build> ' +
        '[--before <label>] [--source <artifact.json>]` (or `make performance-page`, which prompts).'
    );
    process.exit(1);
  }

  const artifactPath = args.source || DEFAULT_ARTIFACT;
  if (!fs.existsSync(artifactPath)) {
    console.error(`benchmark artifact not found: ${artifactPath}`);
    process.exit(1);
  }
  const artifact = readJson(artifactPath);
  const artifactSource = relArtifact(artifactPath);

  let releaseConfig = {};
  if (fs.existsSync(RELEASE_FILE)) releaseConfig = readJson(RELEASE_FILE);
  else console.warn('release.json not found; using empty release config.');

  const after = snapshotFrom(artifact, artifactSource, args.label);

  // Resolve the baseline by label. Never by adjacency.
  const entries = readMetadataEntries();
  let before = null;
  let role = 'baseline';
  const existingIdx = entries.findIndex((e) => e.performance_build === args.label);
  if (args.before) {
    const base = entries.find((e) => e.performance_build === args.before);
    if (!base) {
      console.error(`--before "${args.before}" matches no entry in ${path.basename(META_FILE)}.`);
      console.error('Known performance_build labels:');
      entries.forEach((e) => console.error(`  ${e.performance_build} (${e.role || 'comparison'})`));
      process.exit(1);
    }
    before = base.after || null;
    role = 'comparison';
  } else if (existingIdx >= 0) {
    // Re-running a label means replacing that entry in place. Keep the
    // baseline it was recorded against instead of demanding it again.
    before = entries[existingIdx].before || null;
    role = entries[existingIdx].role || (before ? 'comparison' : 'baseline');
  } else if (entries.length) {
    console.error('Existing entries are present, so a baseline must be named explicitly.');
    console.error('Pass --before <performance_build>. Adjacency is not evidence (AGENTS.md).');
    console.error('Known performance_build labels:');
    entries.forEach((e) => console.error(`  ${e.performance_build} (${e.role || 'comparison'})`));
    process.exit(1);
  }

  const entry = {
    release: releaseConfig.release || null,
    build: releaseConfig.build || null,
    performance_build: args.label,
    artifact_source: artifactSource,
    git_commit: (artifact.provenance || {}).git_revision || null,
    generated_at: artifact.generated_at || null,
    role,
    before,
    after,
  };

  const key = args.label;
  const idx = entries.findIndex((e) => e.performance_build === key);
  if (idx >= 0) entries[idx] = entry;
  else entries.push(entry);

  fs.writeFileSync(META_FILE, JSON.stringify(entries, null, 2) + '\n');

  const payload = {
    entry,
    entries: entries.map((e) => ({
      performance_build: e.performance_build,
      role: e.role || null,
      generated_at: e.generated_at || null,
      git_commit: e.git_commit || null,
    })),
    comparability: comparability(before, after),
    not_measured: collectNotMeasured(artifact, 'artifact', []),
    limitations: artifact.limitations || [],
    stage_time_coverage_pct: after.metrics.stage_time_coverage_pct,
  };

  fs.writeFileSync(OUTPUT, renderTemplate(payload));

  console.log(`Wrote ${OUTPUT}`);
  console.log(`  artifact:   ${artifactSource}`);
  console.log(`  metadata:   ${entries.length} run(s) in ${path.basename(META_FILE)}`);
  console.log(
    `  before:     ${before ? `${before.performance_build} (${before.artifact_source})` : 'none (this entry is the baseline)'}`
  );
  console.log(`  after:      ${args.label}`);
  if (!payload.comparability.comparable) {
    console.warn('  WARNING: before/after context differs, so the delta is not a like-for-like comparison:');
    payload.comparability.reasons.forEach((r) => console.warn(`    ${r}`));
  }
}

if (require.main === module) {
  main();
}

module.exports = { metricsFrom, snapshotFrom, collectNotMeasured, comparability };

#!/usr/bin/env node
/*
 * Builds the embeddable benchmark page (tests/api/page/benchmark.html).
 *
 * Reads:
 *   - tests/api/page/release.json      (manually managed release/build/env config)
 *   - tests/api/page/metadata.json     (auto-appended array, one entry per test run)
 *   - latest tests/api/benchmark/*_benchmark_v1_*.json
 *   - latest tests/api/result/*_load-test_*.json
 *
 * Computes run metrics, appends (or replaces, keyed by release+build) a new
 * entry to metadata.json, and exposes the previous entry as "before" so the
 * page can render a before/after comparison.
 *
 * Writes two self-contained HTML files with all data inlined:
 *   - benchmark.html       trimmed (public): raw_address only for the three
 *                          Top-Challenge example records, everything else trimmed.
 *   - full-benchmark.html  untrimmed (local/gitignored): keeps raw_address for
 *                          every record, renders the top-100 failed matches
 *                          ranked by confidence and shows the performance run.
 */
'use strict';

const fs = require('fs');
const path = require('path');

const PAGE_DIR = __dirname;
const BENCH_DIR = path.join(PAGE_DIR, '..', 'benchmark');
const RESULT_DIR = path.join(PAGE_DIR, '..', 'result');
const TEMPLATE = path.join(PAGE_DIR, 'template.html');
const OUTPUT = path.join(PAGE_DIR, 'benchmark.html');
const FULL_OUTPUT = path.join(PAGE_DIR, 'full-benchmark.html');
const META_FILE = path.join(PAGE_DIR, 'metadata.json');
const RELEASE_FILE = path.join(PAGE_DIR, 'release.json');

function latestFile(dir, pattern) {
  if (!fs.existsSync(dir)) return null;
  const files = fs.readdirSync(dir)
    .filter((f) => pattern.test(f))
    .sort();
  return files.length ? files[files.length - 1] : null;
}

function readJson(filePath) {
  return JSON.parse(fs.readFileSync(filePath, 'utf-8'));
}

function pct(part, total) {
  return total ? Math.round((part / total) * 1000) / 10 : 0;
}

function isRoadLike(raw) {
  return /(?:^|[\s,])(?:jln?\.?|jalan)[\s.]/i.test(String(raw || ''));
}

function trimRecord(record) {
  const q = record.quality;
  const c = record.comparison || {};
  const actual = {
    province: c.actual_province,
    city: c.actual_city,
    district: c.actual_district,
    subdistrict: c.actual_subdistrict,
  };
  const same = {
    p: c.same_province,
    c: c.same_city,
    d: c.same_district,
    s: c.same_subdistrict,
  };
  if (!q) {
    return {
      error: true,
      id: null,
      status: null,
      confidence: null,
      formatted: null,
      location: null,
      missing: [],
      conflicts: [],
      ambiguous: [],
      actual,
      same,
      note: c.note || null,
    };
  }
  const a = q.assessment || {};
  return {
    error: false,
    id: q.address_id,
    status: q.status,
    confidence: q.confidence,
    formatted: q.formatted_address,
    location: q.location || null,
    missing: a.missing || [],
    conflicts: a.conflicts || [],
    ambiguous: a.ambiguous || [],
    actual,
    same,
    note: c.note || null,
  };
}

function exampleFrom(candidates, preferNote) {
  if (!candidates.length) return null;
  let best = null;
  for (const c of candidates) {
    if (preferNote && c.note === preferNote) {
      best = c;
      break;
    }
    if (!best || (c.confidence || 0) > (best.confidence || 0)) best = c;
  }
  return {
    raw: best.raw_address,
    id: best.quality && best.quality.address_id,
    formatted: best.quality && best.quality.formatted_address,
    status: best.quality && best.quality.status,
    confidence: best.quality && best.quality.confidence,
    actual: {
      province: best.comparison.actual_province,
      city: best.comparison.actual_city,
      district: best.comparison.actual_district,
      subdistrict: best.comparison.actual_subdistrict,
    },
  };
}

function buildChallenges(rows, trimmed) {
  const total = rows.length;
  const sameAll = (t) => t.same.p && t.same.c && t.same.d && t.same.s;
  const correct = trimmed.filter(sameAll);

  const villageIdx = [];
  const cityIdx = [];
  const roadIdx = [];
  trimmed.forEach((t, i) => {
    if (t.same.p && t.same.c && t.same.d && !t.same.s) villageIdx.push(i);
    if (!t.same.c) cityIdx.push(i);
    if ((isRoadLike(rows[i].raw_address) && !t.same.c) || t.note === 'INSUFFICIENT_DATA_ROAD') roadIdx.push(i);
  });
  const byIdx = (arr) => arr.map((i) => rows[i]);

  const villageEx = exampleFrom(byIdx(villageIdx));
  const cityEx = exampleFrom(byIdx(cityIdx));
  const roadEx = exampleFrom(byIdx(roadIdx), 'INSUFFICIENT_DATA_ROAD');

  const challenges = [
    {
      id: 'village',
      title: 'Village (kelurahan / desa) resolution',
      count: villageIdx.length,
      why: `In ${villageIdx.length} of ${total} records (${pct(villageIdx.length, total)}%), the village (kelurahan / desa) was resolved incorrectly or left empty even though province, city and district were correct.`,
      expected: villageEx
        ? `Resolve the correct village for the matched district (ground truth: ${villageEx.actual.subdistrict}).`
        : 'Resolve the correct village for the matched district.',
      current: villageEx
        ? `Resolved to ${villageEx.formatted || 'an empty location'} with ${Math.round((villageEx.confidence || 0) * 100)}% confidence (status ${villageEx.status}).`
        : 'No example available for this run.',
      example: villageEx,
    },
    {
      id: 'city',
      title: 'Ambiguous city / regency names',
      count: cityIdx.length,
      why: `In ${cityIdx.length} of ${total} records (${pct(cityIdx.length, total)}%), the address resolved to the wrong city or regency. Many city and regency names are shared across provinces (e.g. Bandung, Sukasari, Sukarasa), so a short or incomplete input can match a wrong location with high confidence.`,
      expected: cityEx
        ? `Resolve to the correct city / regency (ground truth: ${cityEx.actual.city}).`
        : 'Resolve to the correct city / regency.',
      current: cityEx
        ? `Resolved to ${cityEx.formatted || 'an empty location'} with ${Math.round((cityEx.confidence || 0) * 100)}% confidence (status ${cityEx.status}).`
        : 'No example available for this run.',
      example: cityEx,
    },
    {
      id: 'road',
      title: 'Road-level data is not resolved',
      count: roadIdx.length,
      why: `In ${roadIdx.length} of ${total} records (${pct(roadIdx.length, total)}%), a road-based address could not be resolved to the correct location. Street-level data is not part of the reference dataset, so road names are recognized but not validated (the benchmark output records INSUFFICIENT_DATA_ROAD for these cases).`,
      expected: roadEx
        ? `Use road / street context to pin the correct location (ground truth: ${roadEx.actual.city} / ${roadEx.actual.district}).`
        : 'Use road / street context to pin the correct location.',
      current: roadEx
        ? `Resolved to ${roadEx.formatted || 'an empty location'} with ${Math.round((roadEx.confidence || 0) * 100)}% confidence (status ${roadEx.status}).`
        : 'No example available for this run.',
      example: roadEx,
    },
  ];

  return { total, correct: correct.length, challenges };
}

function datasetVersionOf(rows) {
  const datasetRecord = rows.find(
    (r) => r.quality && r.quality.metadata && r.quality.metadata.location_version
  );
  return datasetRecord
    ? {
        version: datasetRecord.quality.metadata.location_version,
        source: datasetRecord.quality.metadata.location_source,
      }
    : { version: null, source: null };
}

function computeMetrics(rows) {
  const total = rows.length;
  const hierarchy = [
    { k: 'same_province', level: 'Province' },
    { k: 'same_city', level: 'City' },
    { k: 'same_district', level: 'District (Kecamatan)' },
    { k: 'same_subdistrict', level: 'Subdistrict (Kelurahan)' },
  ].map((L) => {
    const correct = rows.filter((r) => r.comparison && r.comparison[L.k] === true).length;
    return { level: L.level, correct, pct: pct(correct, total) };
  });

  const exact = rows.filter((r) =>
    r.comparison && r.comparison.same_province && r.comparison.same_city &&
    r.comparison.same_district && r.comparison.same_subdistrict
  ).length;
  hierarchy.push({ level: 'All four levels (exact)', correct: exact, pct: pct(exact, total) });

  const confs = rows
    .filter((r) => r.quality && r.quality.confidence !== null && r.quality.confidence !== undefined)
    .map((r) => r.quality.confidence);
  const avgConf = confs.length ? confs.reduce((a, c) => a + c, 0) / confs.length : null;

  return {
    overall_accuracy: pct(exact, total),
    dataset_size: total,
    average_confidence: avgConf === null ? null : Math.round(avgConf * 1000) / 10,
    dataset_version: null,
    hierarchy,
  };
}

function buildEntry(releaseConfig, benchmarkFile, perfFile, benchmarkRows) {
  const perf = perfFile ? readJson(path.join(RESULT_DIR, perfFile)) : null;
  const dataset = datasetVersionOf(benchmarkRows);
  const runner = releaseConfig.runner || {};
  const server = releaseConfig.server || {};

  const metrics = computeMetrics(benchmarkRows);
  metrics.dataset_version = dataset.version;

  return {
    release: releaseConfig.release || null,
    build: releaseConfig.build || null,
    git_commit: releaseConfig.git_commit || null,
    benchmark_timestamp: benchmarkFile ? benchmarkFile.slice(0, 10) : null,
    dataset_version: dataset.version,
    dataset_source: dataset.source,
    dataset_generation_timestamp: null,
    load_test_timestamp: perf ? perf.generated_at || null : null,
    benchmark_source: benchmarkFile,
    benchmark_runner: [runner.cpu, runner.memory, runner.os].filter(Boolean).join(' · ') || null,
    runner,
    server,
    metrics,
  };
}

function buildPayload(meta, before, benchmarkFile, perfFile, benchmarkRows) {
  const perf = perfFile ? readJson(path.join(RESULT_DIR, perfFile)) : null;
  const trimmed = benchmarkRows.map(trimRecord);
  const { total, correct, challenges } = buildChallenges(benchmarkRows, trimmed);

  return {
    meta,
    before,
    benchmark: {
      source: benchmarkFile,
      timestamp: benchmarkFile ? benchmarkFile.slice(0, 10) : null,
      dataset_size: total,
      dataset_version: meta.dataset_version,
      dataset_source: meta.dataset_source,
      exact_matches: correct,
      records: trimmed,
      challenges,
    },
    performance: perf
      ? Object.assign({ source: perfFile }, perf)
      : { source: null },
  };
}

function buildFullPayload(meta, before, benchmarkFile, perfFile, benchmarkRows) {
  const payload = buildPayload(meta, before, benchmarkFile, perfFile, benchmarkRows);
  const records = payload.benchmark.records;

  benchmarkRows.forEach((row, i) => {
    records[i].raw = row.raw_address;
  });

  const failures = [];
  benchmarkRows.forEach((row) => {
    const q = row.quality;
    const c = row.comparison || {};
    if (!q) return;
    if (c.same_province && c.same_city && c.same_district && c.same_subdistrict) return;
    const a = q.assessment || {};
    failures.push({
      raw: row.raw_address,
      id: q.address_id,
      status: q.status,
      confidence: q.confidence,
      formatted: q.formatted_address,
      actual: {
        province: c.actual_province,
        city: c.actual_city,
        district: c.actual_district,
        subdistrict: c.actual_subdistrict,
      },
      same: {
        p: c.same_province,
        c: c.same_city,
        d: c.same_district,
        s: c.same_subdistrict,
      },
      missing: a.missing || [],
      conflicts: a.conflicts || [],
      ambiguous: a.ambiguous || [],
      note: c.note || null,
    });
  });

  failures.sort(
    (a, b) => (b.confidence === null || b.confidence === undefined ? -1 : b.confidence) -
             (a.confidence === null || a.confidence === undefined ? -1 : a.confidence)
  );

  payload.benchmark.full = true;
  payload.benchmark.failure_total = failures.length;
  payload.benchmark.failures = failures.slice(0, 100);

  return payload;
}

function renderTemplate(payload) {
  const json = JSON.stringify(payload)
    .replace(/</g, '\\u003c')
    .replace(/\u2028/g, '\\u2028')
    .replace(/\u2029/g, '\\u2029');

  let html = fs.readFileSync(TEMPLATE, 'utf-8');
  if (!html.includes('__AQ_DATA__')) {
    console.error('template.html is missing the __AQ_DATA__ marker.');
    process.exit(1);
  }
  html = html.replace('__AQ_DATA__', json);
  return html;
}

function entryKey(entry) {
  return `${entry.release || ''}|${entry.build || ''}`;
}

function upsertEntry(entries, entry) {
  const key = entryKey(entry);
  const idx = entries.findIndex((e) => entryKey(e) === key);
  if (idx >= 0) {
    const before = idx > 0 ? entries[idx - 1] : null;
    entries[idx] = entry;
    return { entries, before };
  }
  const before = entries.length ? entries[entries.length - 1] : null;
  entries.push(entry);
  return { entries, before };
}

function readMetadataEntries() {
  if (!fs.existsSync(META_FILE)) return [];
  const existing = readJson(META_FILE);
  return Array.isArray(existing) ? existing : [existing];
}

function main() {
  let releaseConfig = {};
  if (fs.existsSync(RELEASE_FILE)) {
    releaseConfig = readJson(RELEASE_FILE);
  } else {
    console.warn('release.json not found; using empty release config.');
  }

  const benchmarkFile = latestFile(BENCH_DIR, /_benchmark_v\d+_\d{4}\.json$/);
  if (!benchmarkFile) {
    console.error(
      'No benchmark JSON found in tests/api/benchmark (expected *_benchmark_v1_0000.json).'
    );
    process.exit(1);
  }
  const perfFile = latestFile(RESULT_DIR, /_load-test_\d{4}\.json$/);
  if (!perfFile) {
    console.warn('No load-test JSON found in tests/api/result; performance section will show N/A.');
  }

  const benchmarkRows = readJson(path.join(BENCH_DIR, benchmarkFile));
  const entry = buildEntry(releaseConfig, benchmarkFile, perfFile, benchmarkRows);

  const { entries, before } = upsertEntry(readMetadataEntries(), entry);
  fs.writeFileSync(META_FILE, JSON.stringify(entries, null, 2) + '\n');

  const payload = buildPayload(entry, before, benchmarkFile, perfFile, benchmarkRows);
  const fullPayload = buildFullPayload(entry, before, benchmarkFile, perfFile, benchmarkRows);

  const html = renderTemplate(payload);
  const fullHtml = renderTemplate(fullPayload);

  fs.mkdirSync(PAGE_DIR, { recursive: true });
  fs.writeFileSync(OUTPUT, html);
  fs.writeFileSync(FULL_OUTPUT, fullHtml);
  console.log(`Wrote ${OUTPUT}`);
  console.log(`Wrote ${FULL_OUTPUT} (${fullPayload.benchmark.failure_total} failures shown, raw addresses included)`);
  console.log(
    `  benchmark:  ${benchmarkFile} (${benchmarkRows.length} records)`
  );
  console.log(
    `  performance: ${perfFile || 'N/A'} (${perfFile ? 'load-test summary' : 'missing'})`
  );
  console.log(
    `  metadata:   ${entries.length} test run(s) in metadata.json${before ? ' (before = ' + entryKey(before) + ')' : ' (no previous run)'}`
  );
}

main();

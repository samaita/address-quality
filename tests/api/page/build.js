#!/usr/bin/env node
/*
 * Builds the embeddable benchmark page (tests/api/page/benchmark.html).
 *
 * Reads:
 *   - tests/api/page/metadata.json     (manually managed release/build/env metadata)
 *   - latest tests/api/benchmark/*_benchmark_v1_*.json
 *   - latest tests/api/result/*_load-test_*.json
 *
 * Writes a single self-contained HTML file with all data inlined.
 * Privacy rule: raw_address is confidential. It is embedded ONLY for the
 * three Top-Challenge example records. All other records are trimmed.
 */
'use strict';

const fs = require('fs');
const path = require('path');

const PAGE_DIR = __dirname;
const BENCH_DIR = path.join(PAGE_DIR, '..', 'benchmark');
const RESULT_DIR = path.join(PAGE_DIR, '..', 'result');
const TEMPLATE = path.join(PAGE_DIR, 'template.html');
const OUTPUT = path.join(PAGE_DIR, 'benchmark.html');
const META_FILE = path.join(PAGE_DIR, 'metadata.json');

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

function buildPayload(meta, benchmarkFile, perfFile, benchmarkRows) {
  const perf = perfFile ? readJson(path.join(RESULT_DIR, perfFile)) : null;

  const datasetRecord = benchmarkRows.find(
    (r) => r.quality && r.quality.metadata && r.quality.metadata.location_version
  );
  const datasetVersion = datasetRecord
    ? datasetRecord.quality.metadata.location_version
    : null;
  const datasetSource = datasetRecord
    ? datasetRecord.quality.metadata.location_source
    : null;

  const trimmed = benchmarkRows.map(trimRecord);
  const { total, correct, challenges } = buildChallenges(benchmarkRows, trimmed);

  return {
    meta,
    benchmark: {
      source: benchmarkFile,
      timestamp: benchmarkFile ? benchmarkFile.slice(0, 10) : null,
      dataset_size: total,
      dataset_version: datasetVersion,
      dataset_source: datasetSource,
      exact_matches: correct,
      records: trimmed,
      challenges,
    },
    performance: perf
      ? Object.assign({ source: perfFile }, perf)
      : { source: null },
  };
}

function main() {
  let meta = {};
  if (fs.existsSync(META_FILE)) {
    meta = readJson(META_FILE);
  } else {
    console.warn('metadata.json not found; using empty metadata.');
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
  const payload = buildPayload(meta, benchmarkFile, perfFile, benchmarkRows);

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

  fs.mkdirSync(PAGE_DIR, { recursive: true });
  fs.writeFileSync(OUTPUT, html);
  console.log(`Wrote ${OUTPUT}`);
  console.log(
    `  benchmark:  ${benchmarkFile} (${benchmarkRows.length} records)`
  );
  console.log(
    `  performance: ${perfFile || 'N/A'} (${perfFile ? 'load-test summary' : 'missing'})`
  );
}

main();

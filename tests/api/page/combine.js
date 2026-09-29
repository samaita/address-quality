#!/usr/bin/env node
'use strict';

const fs = require('node:fs');
const path = require('node:path');
const { renderTemplate } = require('./build');
const { payloadFor, readJson, relArtifact } = require('./performance-build');

const PAGE_DIR = __dirname;
const REPO_DIR = path.join(PAGE_DIR, '..', '..', '..');
const TEMPLATE = path.join(PAGE_DIR, 'performance-template.html');
const OUTPUT = path.join(PAGE_DIR, 'performance.html');
const COMBINED_OUTPUT = path.join(PAGE_DIR, 'benchmark.html');
const FULL_OUTPUT = path.join(PAGE_DIR, 'full-benchmark.html');
const PERFORMANCE_META = path.join(PAGE_DIR, 'performance-metadata.json');
function escapeScriptJson(value) {
  return JSON.stringify(value).replace(/</g, '\\u003c').replace(/\u2028/g, '\\u2028').replace(/\u2029/g, '\\u2029');
}

function renderPerformance(payload) {
  let html = fs.readFileSync(TEMPLATE, 'utf8');
  if (!html.includes('__AQ_DATA__')) throw new Error('performance-template.html missing __AQ_DATA__ marker');
  return html.replace('const AQ_DATA = __AQ_DATA__;', `const AQ_DATA = ${escapeScriptJson(payload)};`);
}

function combine(accuracyPayload, performancePayload) {
  return { ...accuracyPayload, performance_report: performancePayload };
}

function main() {
  const [accuracyPath = path.join(PAGE_DIR, 'accuracy-data.json'), fullAccuracyPath = path.join(PAGE_DIR, 'full-accuracy-data.json')] = process.argv.slice(2);
  if (!fs.existsSync(PERFORMANCE_META)) throw new Error('performance-metadata.json not found; run make performance-page first');
  const history = readJson(PERFORMANCE_META);
  const entry = history[history.length - 1];
  if (!entry) throw new Error('performance-metadata.json has no performance runs');
  const artifactPath = path.resolve(REPO_DIR, entry.artifact_source);
  const artifact = readJson(artifactPath);
  const report = payloadFor(entry, history, artifact);
  const accuracy = readJson(accuracyPath);
  const fullAccuracy = readJson(fullAccuracyPath);

  fs.writeFileSync(OUTPUT, renderPerformance(report));
  fs.writeFileSync(COMBINED_OUTPUT, renderTemplate(combine(accuracy, report)));
  fs.writeFileSync(FULL_OUTPUT, renderTemplate(combine(fullAccuracy, report)));
  console.log(`Wrote ${OUTPUT}, ${COMBINED_OUTPUT}, and ${FULL_OUTPUT} (performance source ${relArtifact(artifactPath)})`);
}

if (require.main === module) main();
module.exports = { combine, escapeScriptJson };

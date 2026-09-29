#!/usr/bin/env node
'use strict';

const fs = require('node:fs');
const path = require('node:path');
const { renderTemplate } = require('./build');
const { payloadFor, readJson, relArtifact } = require('./performance-build');

const PAGE_DIR = __dirname;
const REPO_DIR = path.join(PAGE_DIR, '..', '..', '..');
const COMBINED_OUTPUT = path.join(PAGE_DIR, 'benchmark.html');
const FULL_OUTPUT = path.join(PAGE_DIR, 'full-benchmark.html');
const PERFORMANCE_META = path.join(PAGE_DIR, 'performance-metadata.json');

// The performance report is inlined as data (performance_report); template.html
// renders the performance section from it directly. No nested HTML document is
// embedded, so there is nothing to render separately here.
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

  fs.writeFileSync(COMBINED_OUTPUT, renderTemplate(combine(accuracy, report)));
  fs.writeFileSync(FULL_OUTPUT, renderTemplate(combine(fullAccuracy, report)));
  console.log(`Wrote ${COMBINED_OUTPUT} and ${FULL_OUTPUT} (performance source ${relArtifact(artifactPath)})`);
}

if (require.main === module) main();
module.exports = { combine };

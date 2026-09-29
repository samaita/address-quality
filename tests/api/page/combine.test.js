'use strict';

const assert = require('node:assert/strict');
const test = require('node:test');
const { combine } = require('./combine');

test('combines independent accuracy and performance payloads', () => {
  const accuracy = { benchmark: { source: 'accuracy.json' }, meta: { release: 'v1' } };
  const performance = { entry: { performance_build: 'run-1' }, limitations: ['limited workload'] };
  const combined = combine(accuracy, performance);
  assert.deepEqual(combined.benchmark, accuracy.benchmark);
  assert.equal(combined.performance_report.entry.performance_build, 'run-1');
  // The performance section is rendered inline from performance_report; no
  // nested HTML document is embedded.
  assert.equal(combined.performance_page, undefined);
});

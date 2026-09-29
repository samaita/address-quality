'use strict';

const assert = require('node:assert/strict');
const test = require('node:test');
const { combine, escapeScriptJson } = require('./combine');

test('combines independent accuracy and performance payloads', () => {
  const accuracy = { benchmark: { source: 'accuracy.json' }, meta: { release: 'v1' } };
  const performance = { entry: { performance_build: 'run-1' }, limitations: ['limited workload'] };
  const combined = combine(accuracy, performance);
  assert.deepEqual(combined.benchmark, accuracy.benchmark);
  assert.equal(combined.performance_report.entry.performance_build, 'run-1');
  assert.equal(combined.performance_page, undefined);
});

test('escapes script-closing input before embedding performance data', () => {
  assert.equal(escapeScriptJson('</script>'), '"\\u003c/script>"');
});

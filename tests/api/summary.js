import { textSummary } from 'https://jslib.k6.io/k6-summary/0.0.1/index.js';

function flattenChecks(group, out) {
  for (const check of group.checks || []) {
    out.push({ name: check.name, passes: check.passes, fails: check.fails, group: group.path || '' });
  }
  for (const sub of group.groups || []) {
    flattenChecks(sub, out);
  }
  return out;
}

function buildReport(data, testName) {
  const metrics = {};
  for (const [name, metric] of Object.entries(data.metrics)) {
    metrics[name] = {
      type: metric.type,
      contains: metric.contains,
      values: metric.values,
      thresholds: metric.thresholds || undefined,
    };
  }

  const thresholds = {};
  for (const [name, metric] of Object.entries(data.metrics)) {
    if (metric.thresholds) {
      thresholds[name] = {
        ok: Object.values(metric.thresholds).every((t) => t.ok),
        thresholds: metric.thresholds,
      };
    }
  }

  return {
    test: testName,
    generated_at: new Date().toISOString(),
    test_run_duration_ms: data.state.testRunDurationMs,
    thresholds,
    checks: flattenChecks(data.root_group, []),
    metrics,
  };
}

export function makeHandleSummary(testName) {
  return function handleSummary(data) {
    const output = {
      stdout: textSummary(data, { indent: '  ', enableColors: false }),
    };
    if (__ENV.RESULT_JSON) {
      output[__ENV.RESULT_JSON] = JSON.stringify(buildReport(data, testName), null, 2);
    }
    return output;
  };
}

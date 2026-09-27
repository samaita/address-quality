import http from 'k6/http';
import { check, sleep } from 'k6';
import { SharedArray } from 'k6/data';
import { Trend } from 'k6/metrics';

import { makeHandleSummary } from './summary.js';

const { resolveBaseUrl, resolveApiKey } = require('./config.js');

const BASE_URL = resolveBaseUrl();
const API_KEY = resolveApiKey();
const TARGET_VUS = parseInt(__ENV.K6_VUS) || 10;

console.log(`BASE_URL: ${BASE_URL.url} (from ${BASE_URL.source})`);
console.log(`API_KEY: ${API_KEY.apiKey ? 'set' : 'blank'} (from ${API_KEY.source})`);

const workload = JSON.parse(open('./performance-workload.json'));

const addresses = new SharedArray('addresses', function () {
  return workload.addresses;
});

console.log(`workload: ${workload.id} (${addresses.length} addresses)`);

export const options = {
  stages: [
    { duration: '10s', target: TARGET_VUS },
    { duration: '30s', target: TARGET_VUS },
    { duration: '10s', target: 0 },
  ],
  thresholds: {
    http_req_failed: ['rate<0.01'],
    checks: ['rate==1.0'],
  },
  summaryTrendStats: ['avg', 'min', 'med', 'max', 'p(90)', 'p(95)', 'p(99)'],
};

const successDuration = new Trend('success_duration');

export const handleSummary = makeHandleSummary('load-test');

export default function () {
  const idx = (__VU - 1 + __ITER) % addresses.length;
  const payload = JSON.stringify({ address: addresses[idx] });
  const headers = { 'Content-Type': 'application/json' };
  if (API_KEY.apiKey) headers['X-API-Key'] = API_KEY.apiKey;

  const res = http.post(`${BASE_URL.url}/v1/validate`, payload, { headers });

  if (res.status !== 429) {
    successDuration.add(res.timings.duration);
  }

  check(res, {
    'status is 200': (r) => r.status === 200,
    'request_id is present': (r) => r.status !== 200 || r.json().request_id !== '',
  });
}

import http from 'k6/http';
import { check, group } from 'k6';

const BASE_URL = __ENV.K6_BASE_URL || 'http://localhost:7300';

export const options = {
  vus: 1,
  iterations: 1,
  thresholds: {
    checks: ['rate==1.0'],
  },
};

export default function () {
  group('GET /health', function () {
    const res = http.get(`${BASE_URL}/health`);
    check(res, {
      'status is 200': (r) => r.status === 200,
      'status field is ok': (r) => r.json().status === 'ok',
      'database field is ok': (r) => r.json().database === 'ok',
      'timestamp is present': (r) => r.json().timestamp !== '',
    });
  });

  group('POST /v1/validate - valid address', function () {
    const payload = JSON.stringify({ address: 'Jl. Merdeka No.1, Jakarta Pusat 10110' });
    const headers = { 'Content-Type': 'application/json' };
    const res = http.post(`${BASE_URL}/v1/validate`, payload, { headers });
    check(res, {
      'status is 200': (r) => r.status === 200,
      'request_id is present': (r) => r.json().request_id !== '',
      'address_id is present': (r) => r.json().quality.address_id !== '',
      'confidence is 0.0': (r) => r.json().quality.confidence === 0.0,
      'raw_input matches': (r) => r.json().quality.raw_input === 'Jl. Merdeka No.1, Jakarta Pusat 10110',
    });
  });

  group('POST /v1/validate - HTML sanitization', function () {
    const payload = JSON.stringify({ address: '<script>alert(1)</script>Jl. Sudirman' });
    const headers = { 'Content-Type': 'application/json' };
    const res = http.post(`${BASE_URL}/v1/validate`, payload, { headers });
    check(res, {
      'status is 200': (r) => r.status === 200,
      'script tags stripped': (r) => !r.json().quality.normalized_input.includes('<script>'),
      'text preserved': (r) => r.json().quality.normalized_input.includes('Jl. Sudirman'),
    });
  });

  group('POST /v1/validate - empty address', function () {
    const payload = JSON.stringify({ address: '' });
    const headers = { 'Content-Type': 'application/json' };
    const res = http.post(`${BASE_URL}/v1/validate`, payload, { headers });
    check(res, {
      'status is 400': (r) => r.status === 400,
      'error field is present': (r) => r.json().error !== '',
    });
  });

  group('POST /v1/validate - malformed body', function () {
    const headers = { 'Content-Type': 'application/json' };
    const res = http.post(`${BASE_URL}/v1/validate`, 'not json', { headers });
    check(res, {
      'status is 400': (r) => r.status === 400,
      'error field is present': (r) => r.json().error !== '',
    });
  });
}

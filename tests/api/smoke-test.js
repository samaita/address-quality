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
    const body = res.json();
    check(res, {
      'status is 200': (r) => r.status === 200,
      'request_id is present': (r) => body.request_id !== '',
      'timestamp is present': (r) => body.timestamp !== '',
      'data.address_id is present': (r) => body.data.address_id !== '',
      'data.status is known': (r) => ['VALID','INCOMPLETE','AMBIGUOUS','CONFLICT','UNKNOWN'].includes(body.data.status),
      'data.confidence is a number': (r) => typeof body.data.confidence === 'number',
      'data.raw_input matches': (r) => body.data.raw_input === 'Jl. Merdeka No.1, Jakarta Pusat 10110',
      'data.normalized_input is present': (r) => body.data.normalized_input !== '',
      'data.formatted_address is present': (r) => body.data.formatted_address !== '',
      'data.location has fields': (r) =>
        'province' in body.data.location &&
        'city' in body.data.location &&
        'district' in body.data.location &&
        'sub_district' in body.data.location &&
        'postal_code' in body.data.location,
      'data.assessment has matched/missing/conflicts/ambiguous': (r) =>
        Array.isArray(body.data.assessment.matched) &&
        Array.isArray(body.data.assessment.missing) &&
        (body.data.assessment.conflicts === null || Array.isArray(body.data.assessment.conflicts)) &&
        Array.isArray(body.data.assessment.ambiguous),
      'data.resolution has strategy/candidate_count/candidates': (r) =>
        Array.isArray(body.data.resolution.strategy) &&
        typeof body.data.resolution.candidate_count === 'number' &&
        Array.isArray(body.data.resolution.candidates),
      'data.metadata has location_source and location_version': (r) =>
        body.data.metadata.location_source !== '' &&
        body.data.metadata.location_version !== '',
    });
  });

  group('POST /v1/validate - HTML sanitization', function () {
    const payload = JSON.stringify({ address: '<script>alert(1)</script>Jl. Sudirman' });
    const headers = { 'Content-Type': 'application/json' };
    const res = http.post(`${BASE_URL}/v1/validate`, payload, { headers });
    const body = res.json();
    check(res, {
      'status is 200': (r) => r.status === 200,
      'script tags stripped': (r) => !body.data.normalized_input.includes('<script>'),
      'text preserved': (r) => body.data.normalized_input.toLowerCase().includes('sudirman'),
      'data.status is present': (r) => body.data.status !== '',
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

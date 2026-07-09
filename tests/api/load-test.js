import http from 'k6/http';
import { check, sleep } from 'k6';
import { SharedArray } from 'k6/data';
import { Trend } from 'k6/metrics';

const BASE_URL = __ENV.K6_BASE_URL || 'http://localhost:7300';
const TARGET_VUS = parseInt(__ENV.K6_VUS) || 10;

const addresses = new SharedArray('addresses', function () {
  return [
    'Jl. Merdeka No.1, Jakarta Pusat 10110',
    'Jl. Siliwangi No.1, Bogor 16119',
    'Gang Mawar No.5, Bekasi Timur, Kota Bekasi 17111',
    'Dekat Masjid Al-Ikhlas, Kapuas Tengah, Kalimantan Tengah',
    'Perumahan Citra Garden Blok A2 No.3, Surabaya 60231',
  ];
});

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
};

const successDuration = new Trend('success_duration');

export default function () {
  const idx = (__VU - 1 + __ITER) % addresses.length;
  const payload = JSON.stringify({ address: addresses[idx] });
  const headers = { 'Content-Type': 'application/json' };

  const res = http.post(`${BASE_URL}/v1/validate`, payload, { headers });

  if (res.status !== 429) {
    successDuration.add(res.timings.duration);
  }

  check(res, {
    'status is 200': (r) => r.status === 200,
    'request_id is present': (r) => r.status !== 200 || r.json().request_id !== '',
  });

  sleep(0.5);
}

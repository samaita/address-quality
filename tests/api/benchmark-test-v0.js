const fs = require('fs');
const path = require('path');
const http = require('http');

const { resolveBaseUrl, resolveApiKey } = require('./config.js');

const BASE_URL = resolveBaseUrl();
const API_KEY = resolveApiKey();
const API_VERSION = process.env.API_VERSION || 'v0';
const OUTPUT_DIR = 'tests/api/benchmark-v0';

const MISS_DELAY_MS = 5000;

console.log(`BASE_URL: ${BASE_URL.url} (from ${BASE_URL.source})`);
console.log(`API_KEY: ${API_KEY.apiKey ? 'set' : 'blank'} (from ${API_KEY.source})`);
console.log(`CACHE miss back-off: ${MISS_DELAY_MS / 1000}s per miss`);

let INPUT_FILE = 'tests/api/cases/address-tagged.csv';
if (process.env.INPUT_FILE) INPUT_FILE = process.env.INPUT_FILE;

const csvArg = process.argv.find(a => a.startsWith('--csv='));
if (csvArg) INPUT_FILE = csvArg.split('=')[1];

const sourceArg = process.argv.find(a => a.startsWith('--source='));
const SOURCE = sourceArg ? sourceArg.split('=')[1] : '';

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

function parseCsvLine(line, delimiter) {
  const result = [];
  let current = '';
  let inQuotes = false;
  for (let i = 0; i < line.length; i++) {
    const ch = line[i];
    if (inQuotes) {
      if (ch === '"') {
        if (i + 1 < line.length && line[i + 1] === '"') {
          current += '"';
          i++;
        } else {
          inQuotes = false;
        }
      } else {
        current += ch;
      }
    } else {
      if (ch === '"') {
        inQuotes = true;
      } else if (ch === delimiter) {
        result.push(current);
        current = '';
      } else {
        current += ch;
      }
    }
  }
  result.push(current);
  return result;
}

function readInputCsv(filePath) {
  const content = fs.readFileSync(filePath, 'utf-8');
  const rawLines = content.split(/\r?\n/);

  const joinedLines = [];
  let buffer = '';
  let inQuotes = false;
  for (const line of rawLines) {
    if (!inQuotes && line.trim() === '' && buffer === '') continue;
    if (buffer === '') {
      buffer = line;
    } else {
      buffer += '\n' + line;
    }
    for (const ch of line) {
      if (ch === '"') inQuotes = !inQuotes;
    }
    if (!inQuotes) {
      joinedLines.push(buffer);
      buffer = '';
    }
  }
  if (buffer.trim() !== '') joinedLines.push(buffer);

  if (joinedLines.length < 2) {
    console.error('Input CSV must have at least a header row + 1 data row');
    process.exit(1);
  }

  const delimiter = joinedLines[0].includes(';') ? ';' : ',';
  const rows = [];
  for (let i = 1; i < joinedLines.length; i++) {
    const cols = parseCsvLine(joinedLines[i], delimiter);
    if (cols.length < 3) {
      console.warn(`Skipping row ${i + 1}: only ${cols.length} columns found`);
      continue;
    }
    rows.push({
      address: cols[2].trim(),
      actualProvince: (cols[3] || '').trim(),
      actualCity: (cols[4] || '').trim(),
      actualDistrict: (cols[5] || '').trim(),
      actualSubdistrict: (cols[6] || '').trim(),
    });
  }
  return rows;
}

function determineOutputFile() {
  const now = new Date();
  const yyyy = now.getFullYear();
  const mm = String(now.getMonth() + 1).padStart(2, '0');
  const dd = String(now.getDate()).padStart(2, '0');
  const dateStr = `${yyyy}-${mm}-${dd}`;
  const prefix = `${dateStr}_benchmark_${API_VERSION}_`;

  if (!fs.existsSync(OUTPUT_DIR)) {
    fs.mkdirSync(OUTPUT_DIR, { recursive: true });
  }

  let maxSerial = -1;
  const files = fs.readdirSync(OUTPUT_DIR);
  for (const f of files) {
    if (f.startsWith(prefix) && f.endsWith('.json')) {
      const serialPart = f.slice(prefix.length, -4);
      const serial = parseInt(serialPart, 10);
      if (!isNaN(serial) && serial > maxSerial) {
        maxSerial = serial;
      }
    }
  }

  const nextSerial = String(maxSerial + 1).padStart(4, '0');
  return path.join(OUTPUT_DIR, `${prefix}${nextSerial}.json`);
}

function postRequest(address) {
  return new Promise((resolve) => {
    const url = new URL(`/${API_VERSION}/validate`, BASE_URL.url);
    const body = JSON.stringify({ address });
    const headers = {
      'Content-Type': 'application/json',
      'Content-Length': Buffer.byteLength(body),
    };
    if (API_KEY.apiKey) headers['X-API-Key'] = API_KEY.apiKey;
    const options = {
      hostname: url.hostname,
      port: url.port || 80,
      path: url.pathname,
      method: 'POST',
      headers,
    };

    const req = http.request(options, (res) => {
      let data = '';
      res.on('data', (chunk) => { data += chunk; });
      res.on('end', () => {
        try {
          resolve({ ok: true, statusCode: res.statusCode, data: JSON.parse(data) });
        } catch {
          resolve({ ok: false, error: `HTTP ${res.statusCode}` });
        }
      });
    });

    req.on('error', (err) => {
      resolve({ ok: false, error: err.message });
    });

    req.write(body);
    req.end();
  });
}

function normalizeName(value) {
  if (!value) return '';
  return String(value)
    .toLowerCase()
    .replace(/^(kecamatan|kec|kelurahan|kel|desa|provinsi)[\s.]*/, '')
    .replace(/[\s.]+$/, '')
    .trim();
}

function parseGoogleLocation(payload) {
  if (!payload || !payload.results || !payload.results.length) return null;
  const first = payload.results[0];
  const components = {};
  for (const c of first.addressComponents || []) {
    for (const type of c.types || []) {
      if (!components[type]) components[type] = c.longText || c.shortText || '';
    }
  }
  return {
    province: components.administrative_area_level_1 || null,
    city: components.administrative_area_level_2 || null,
    district: components.administrative_area_level_3 || null,
    sub_district: components.administrative_area_level_4 || null,
    formatted: first.formattedAddress || null,
    latitude: first.location && first.location.latitude !== undefined ? first.location.latitude : null,
    longitude: first.location && first.location.longitude !== undefined ? first.location.longitude : null,
    granularity: first.granularity || null,
    place_id: first.placeId || null,
  };
}

async function main() {
  console.log(`Reading input: ${INPUT_FILE}`);
  const rows = readInputCsv(INPUT_FILE);
  console.log(`Loaded ${rows.length} address(es)`);

  const outputFile = determineOutputFile();
  console.log(`Output: ${outputFile}`);

  const outRows = [];
  let succeeded = 0;
  let failed = 0;
  let hits = 0;
  let misses = 0;

  for (let i = 0; i < rows.length; i++) {
    const row = rows[i];
    process.stdout.write(`[${i + 1}/${rows.length}] ${row.address.slice(0, 60)}... `);

    const escapedAddress = row.address.replace(/\n/g, ' ');
    const result = await postRequest(escapedAddress);

    if (result.ok) {
      const cacheStatus = result.data.cache_status || null;
      const loc = parseGoogleLocation(result.data.data);

      if (loc) {
        const outputProvince = (loc.province || '').trim();
        const outputCity = (loc.city || '').trim();
        const outputDistrict = (loc.district || '').trim();
        const outputSubdistrict = (loc.sub_district || '').trim();

        const sameProvince = normalizeName(outputProvince) === normalizeName(row.actualProvince);
        const sameCity = normalizeName(outputCity) === normalizeName(row.actualCity);
        const sameDistrict = normalizeName(outputDistrict) === normalizeName(row.actualDistrict);
        const sameSubdistrict = normalizeName(outputSubdistrict) === normalizeName(row.actualSubdistrict);

        outRows.push({
          source: SOURCE,
          raw_address: row.address,
          quality: {
            address_id: null,
            status: 'FOUND',
            confidence: null,
            formatted_address: loc.formatted,
            location: {
              province: outputProvince,
              city: outputCity,
              district: outputDistrict,
              sub_district: outputSubdistrict,
              latitude: loc.latitude,
              longitude: loc.longitude,
              granularity: loc.granularity,
            },
            assessment: { missing: [], conflicts: [], ambiguous: [] },
            cache_status: cacheStatus,
            place_id: loc.place_id,
          },
          comparison: {
            actual_province: row.actualProvince,
            actual_city: row.actualCity,
            actual_district: row.actualDistrict,
            actual_subdistrict: row.actualSubdistrict,
            same_province: sameProvince,
            same_city: sameCity,
            same_district: sameDistrict,
            same_subdistrict: sameSubdistrict,
          },
        });
        succeeded++;
        console.log(`OK (${result.statusCode}, ${cacheStatus || '?'})`);
      } else {
        outRows.push({
          source: SOURCE,
          raw_address: row.address,
          quality: {
            address_id: null,
            status: result.statusCode === 404 ? 'NOT_FOUND' : 'FAILED',
            confidence: null,
            formatted_address: null,
            location: null,
            assessment: { missing: [], conflicts: [], ambiguous: [] },
            cache_status: cacheStatus,
            place_id: null,
          },
          comparison: {
            actual_province: row.actualProvince,
            actual_city: row.actualCity,
            actual_district: row.actualDistrict,
            actual_subdistrict: row.actualSubdistrict,
            same_province: false,
            same_city: false,
            same_district: false,
            same_subdistrict: false,
          },
        });
        succeeded++;
        console.log(`NOT_FOUND (${result.statusCode}, ${cacheStatus || '?'})`);
      }

      if (cacheStatus === 'MISS') {
        misses++;
        process.stdout.write(`  cache MISS — pausing ${MISS_DELAY_MS / 1000}s... `);
        await sleep(MISS_DELAY_MS);
        console.log('resumed');
      } else if (cacheStatus === 'HIT') {
        hits++;
      }
    } else {
      outRows.push({
        source: SOURCE,
        raw_address: row.address,
        quality: null,
        comparison: {
          actual_province: row.actualProvince,
          actual_city: row.actualCity,
          actual_district: row.actualDistrict,
          actual_subdistrict: row.actualSubdistrict,
          same_province: false,
          same_city: false,
          same_district: false,
          same_subdistrict: false,
        },
      });
      failed++;
      console.log(`FAIL (${result.error})`);
    }
  }

  fs.writeFileSync(outputFile, JSON.stringify(outRows, null, 2));
  console.log(`\nDone. ${succeeded} succeeded, ${failed} failed (${hits} cache HIT, ${misses} cache MISS).`);
  console.log(`Results written to ${outputFile}`);
}

main().catch((err) => {
  console.error('Fatal error:', err);
  process.exit(1);
});

const fs = require('fs');
const path = require('path');
const http = require('http');

const BASE_URL = process.env.BASE_URL || 'http://localhost:7300';
const API_VERSION = process.env.API_VERSION || 'v1';
const OUTPUT_DIR = 'tests/api/benchmark';

let INPUT_FILE = 'tests/api/cases/address-tagged.csv';
if (process.env.INPUT_FILE) INPUT_FILE = process.env.INPUT_FILE;

const csvArg = process.argv.find(a => a.startsWith('--csv='));
if (csvArg) INPUT_FILE = csvArg.split('=')[1];

const sourceArg = process.argv.find(a => a.startsWith('--source='));
const SOURCE = sourceArg ? sourceArg.split('=')[1] : '';

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

function postRequest(address, source) {
  return new Promise((resolve) => {
    const url = new URL(`/${API_VERSION}/validate`, BASE_URL);
    const body = JSON.stringify({ address, source_code: source });
    const options = {
      hostname: url.hostname,
      port: url.port || 80,
      path: url.pathname,
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Content-Length': Buffer.byteLength(body),
      },
    };

    const req = http.request(options, (res) => {
      let data = '';
      res.on('data', (chunk) => { data += chunk; });
      res.on('end', () => {
        if (res.statusCode === 200) {
          try {
            resolve({ ok: true, data: JSON.parse(data) });
          } catch {
            resolve({ ok: false, error: 'Invalid JSON response' });
          }
        } else {
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

async function main() {
  console.log(`Reading input: ${INPUT_FILE}`);
  const rows = readInputCsv(INPUT_FILE);
  console.log(`Loaded ${rows.length} address(es)`);

  const outputFile = determineOutputFile();
  console.log(`Output: ${outputFile}`);

  const outRows = [];
  let succeeded = 0;
  let failed = 0;

  for (let i = 0; i < rows.length; i++) {
    const row = rows[i];
    process.stdout.write(`[${i + 1}/${rows.length}] ${row.address.slice(0, 60)}... `);

    const escapedAddress = row.address.replace(/\n/g, ' ');
    const result = await postRequest(escapedAddress, SOURCE);
    if (result.ok) {
      const q = result.data.quality;
      const loc = q.location || {};
      const outputProvince = (loc.province || '').trim();
      const outputCity = (loc.city || '').trim();
      const outputDistrict = (loc.district || '').trim();
      const outputSubdistrict = (loc.sub_district || '').trim();

      const sameProvince = outputProvince.toLowerCase() === row.actualProvince.toLowerCase();
      const sameCity = outputCity.toLowerCase() === row.actualCity.toLowerCase();
      const sameDistrict = outputDistrict.toLowerCase() === row.actualDistrict.toLowerCase();
      const sameSubdistrict = outputSubdistrict.toLowerCase() === row.actualSubdistrict.toLowerCase();

      outRows.push({
        source: SOURCE,
        raw_address: row.address,
        quality: result.data.quality,
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
      console.log('OK');
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
  console.log(`\nDone. ${succeeded} succeeded, ${failed} failed.`);
  console.log(`Results written to ${outputFile}`);
}

main().catch((err) => {
  console.error('Fatal error:', err);
  process.exit(1);
});

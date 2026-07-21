const fs = require('fs');
const path = require('path');
const http = require('http');

const INPUT_FILE = process.env.INPUT_FILE || 'tests/api/cases/address-tagged.csv';
const BASE_URL = process.env.BASE_URL || 'http://localhost:7300';
const API_VERSION = process.env.API_VERSION || 'v1';
const OUTPUT_DIR = 'tests/api/benchmark';

const sourceArg = process.argv.find(a => a.startsWith('--source='));
const SOURCE = sourceArg ? sourceArg.split('=')[1] : '';

function csvEscape(value) {
  if (value == null) return '';
  const str = String(value);
  if (str.includes(',') || str.includes('"') || str.includes('\n')) {
    return '"' + str.replace(/"/g, '""') + '"';
  }
  return str;
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
  const lines = content.split(/\r?\n/).filter(line => line.trim() !== '');
  if (lines.length < 2) {
    console.error('Input CSV must have at least a header row + 1 data row');
    process.exit(1);
  }

  const delimiter = lines[0].includes(';') ? ';' : ',';
  const rows = [];
  for (let i = 1; i < lines.length; i++) {
    const cols = parseCsvLine(lines[i], delimiter);
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
    if (f.startsWith(prefix) && f.endsWith('.csv')) {
      const serialPart = f.slice(prefix.length, -4);
      const serial = parseInt(serialPart, 10);
      if (!isNaN(serial) && serial > maxSerial) {
        maxSerial = serial;
      }
    }
  }

  const nextSerial = String(maxSerial + 1).padStart(4, '0');
  return path.join(OUTPUT_DIR, `${prefix}${nextSerial}.csv`);
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

  const header = [
    'Source',
    'Address',
    'FormattedOutput',
    'ActualProvince',
    'ActualCity',
    'ActualDistrict',
    'ActualSubdistrict',
    'OutputProvince',
    'OutputCity',
    'OutputDistrict',
    'OutputSubDistrict',
    'PostalCode',
    'Confidence',
    'NormalizedInput',
    'LocationVersion',
    'LocationSource',
    'AddressID',
    'SameProvince',
    'SameCity',
    'SameDistrict',
    'SameSubdistrict',
  ];

  const outLines = [header.map(csvEscape).join(',')];
  let succeeded = 0;
  let failed = 0;

  for (let i = 0; i < rows.length; i++) {
    const row = rows[i];
    process.stdout.write(`[${i + 1}/${rows.length}] ${row.address.slice(0, 60)}... `);

    const result = await postRequest(row.address, SOURCE);
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

      outLines.push([
        '',
        row.address,
        q.formatted_output || '',
        row.actualProvince,
        row.actualCity,
        row.actualDistrict,
        row.actualSubdistrict,
        outputProvince,
        outputCity,
        outputDistrict,
        outputSubdistrict,
        loc.postal_code || '',
        q.confidence !== undefined ? q.confidence : '',
        q.normalized_input || '',
        q.location_version || '',
        q.location_source || '',
        q.address_id || '',
        sameProvince,
        sameCity,
        sameDistrict,
        sameSubdistrict,
      ].map(csvEscape).join(','));
      succeeded++;
      console.log('OK');
    } else {
      outLines.push([
        '',
        row.address,
        '',
        row.actualProvince,
        row.actualCity,
        row.actualDistrict,
        row.actualSubdistrict,
        '', '', '', '',
        '', '', '', '', '', '',
        false, false, false, false,
      ].map(csvEscape).join(','));
      failed++;
      console.log(`FAIL (${result.error})`);
    }
  }

  fs.writeFileSync(outputFile, outLines.join('\n') + '\n', 'utf-8');
  console.log(`\nDone. ${succeeded} succeeded, ${failed} failed.`);
  console.log(`Results written to ${outputFile}`);
}

main().catch((err) => {
  console.error('Fatal error:', err);
  process.exit(1);
});

// SPDX-License-Identifier: BUSL-1.1
// Copyright (c) 2026 Samaita
// Resolves BASE_URL and API_KEY for test scripts (k6 and Node), with precedence:
// env var -> /etc/address-quality/.env.prod -> root .env -> default.
const isK6 = typeof __ENV !== 'undefined';

function envVar(name) {
  if (isK6) {
    return (__ENV[name] || '').trim();
  }
  return (process.env[name] || '').trim();
}

function readFile(filePath) {
  if (isK6) {
    try {
      return open(filePath);
    } catch (e) {
      return null;
    }
  }
  try {
    return require('fs').readFileSync(filePath, 'utf-8');
  } catch (e) {
    return null;
  }
}

function valueForKey(content, key) {
  if (!content) return null;
  const re = new RegExp(`^\\s*${key}\\s*=\\s*(.*)\\s*$`, 'm');
  const m = content.match(re);
  return m ? m[1].trim() : null;
}

function valueFromFile(filePath, keys) {
  const content = readFile(filePath);
  if (!content) return null;
  for (const key of keys) {
    const value = valueForKey(content, key);
    if (value) return value;
  }
  return null;
}

const ROOT_ENV_PATHS = isK6
  ? ['../../.env', '.env']
  : ['.env', require('path').join(__dirname, '..', '..', '.env')];

function resolveFromEnv(envKeys, fileKeys, fallback) {
  for (const key of envKeys) {
    const value = envVar(key);
    if (value) return { value, source: `${key} env` };
  }
  const prod = valueFromFile('/etc/address-quality/.env.prod', fileKeys);
  if (prod) return { value: prod, source: '/etc/address-quality/.env.prod' };
  for (const p of ROOT_ENV_PATHS) {
    const local = valueFromFile(p, fileKeys);
    if (local) return { value: local, source: p };
  }
  return { value: fallback, source: 'default' };
}

function resolveBaseUrl() {
  const resolved = resolveFromEnv(
    ['K6_BASE_URL', 'BASE_URL'],
    ['K6_BASE_URL', 'BASE_URL'],
    'http://localhost:7300'
  );
  return { url: resolved.value, source: resolved.source };
}

function resolveApiKey() {
  const resolved = resolveFromEnv(['API_KEY'], ['API_KEY'], '');
  return { apiKey: resolved.value, source: resolved.source };
}

module.exports = { resolveBaseUrl, resolveApiKey };

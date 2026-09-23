'use strict';

const assert = require('node:assert/strict');
const test = require('node:test');

const { formatActualAddress, parseMetadataEntries } = require('./build.js');

test('treats empty metadata as an empty run history', () => {
  assert.deepEqual(parseMetadataEntries(''), []);
  assert.deepEqual(parseMetadataEntries('  \n'), []);
});

test('formats tagged hierarchy from smallest to largest', () => {
  assert.equal(
    formatActualAddress({
      subdistrict: 'Margahayu Tengah',
      district: 'Margahayu',
      city: 'Kabupaten Bandung',
      province: 'Jawa Barat',
    }),
    'Margahayu Tengah, Margahayu, Kabupaten Bandung, Jawa Barat'
  );
});

test('normalizes administrative prefixes while preserving city type', () => {
  assert.equal(
    formatActualAddress({
      subdistrict: 'Kel. Braga',
      district: 'Kec. Sumur Bandung',
      city: 'Kota Bandung',
      province: 'Provinsi Jawa Barat',
    }),
    'Braga, Sumur Bandung, Kota Bandung, Jawa Barat'
  );
  assert.equal(
    formatActualAddress({
      subdistrict: 'Kelurahan Cikeruh',
      district: 'Kecamatan Jatinangor',
      city: 'Kabupaten Sumedang',
      province: 'Jawa Barat',
    }),
    'Cikeruh, Jatinangor, Kabupaten Sumedang, Jawa Barat'
  );
  assert.equal(
    formatActualAddress({
      subdistrict: 'Desa Sayang',
      district: 'Kec Jatinangor',
      city: 'Kabupaten Sumedang',
      province: 'Jawa Barat',
    }),
    'Sayang, Jatinangor, Kabupaten Sumedang, Jawa Barat'
  );
});

test('omits missing levels without extra separators', () => {
  assert.equal(
    formatActualAddress({ city: 'Kota Bandung', province: 'Jawa Barat' }),
    'Kota Bandung, Jawa Barat'
  );
});

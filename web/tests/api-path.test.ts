import assert from 'node:assert/strict';
import test from 'node:test';
import { API_PREFIX, apiPath } from '../src/services/api-path.js';

test('keeps future service paths under the versioned same-origin prefix', () => {
  assert.equal(API_PREFIX, '/api/v1');
  assert.equal(apiPath(), '/api/v1');
  assert.equal(apiPath('events'), '/api/v1/events');
  assert.equal(apiPath('events', '00000000-0000-4000-8000-000000000001'),
    '/api/v1/events/00000000-0000-4000-8000-000000000001');
});

test('preserves decimal uint64 strings without number conversion', () => {
  assert.equal(apiPath('hosts', '18446744073709551615'),
    '/api/v1/hosts/18446744073709551615');
});

test('rejects paths, URLs, traversal, queries and encoded separators as segments', () => {
  for (const segment of ['', '.', '..', '../events', '/events', '//example.test',
    'https://example.test', 'events?token=x', 'events#x', '%2e%2e', '%2F',
    'events\\other', ' events', 'events\n', 'events\u0000']) {
    assert.throws(() => apiPath(segment), TypeError, JSON.stringify(segment));
  }
});

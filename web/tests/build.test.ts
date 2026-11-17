import assert from 'node:assert/strict';
import { readFile, stat } from 'node:fs/promises';
import test from 'node:test';

const dist = new URL('../../dist/', import.meta.url);

test('production build emits an HTML entry and nonempty local JavaScript and CSS assets', async () => {
  const html = await readFile(new URL('index.html', dist), 'utf8');
  assert.match(html, /<html lang="en">/);
  assert.match(html, /<title>Argus — Phase 2 foundation<\/title>/);
  assert.match(html, /<div id="root"><\/div>/);
  assert.doesNotMatch(html, /\/src\/main\.tsx/);

  const assets = [...html.matchAll(/(?:src|href)="(\/assets\/[^"?#]+\.(?:js|css))"/g)]
    .map((match) => match[1]);
  assert.ok(assets.some((asset) => asset.endsWith('.js')), 'missing bundled JavaScript');
  assert.ok(assets.some((asset) => asset.endsWith('.css')), 'missing bundled CSS');
  for (const asset of assets) {
    const url = new URL(asset.slice(1), dist);
    assert.ok(url.href.startsWith(dist.href), 'asset must remain inside dist');
    assert.ok((await stat(url)).size > 0, `empty build asset: ${asset}`);
  }
});

import assert from 'node:assert/strict';
import test from 'node:test';
import { renderToStaticMarkup } from 'react-dom/server';
import { App } from '../src/App.js';

test('renders an accessible static foundation notice, not a monitoring dashboard', () => {
  const html = renderToStaticMarkup(<App />);
  assert.match(html, /<main aria-labelledby="foundation-title">/);
  assert.match(html, /<h1 id="foundation-title">Foundation only — monitoring not implemented<\/h1>/);
  assert.match(html, /does not collect events, contact an API, or display monitoring data/);
  assert.match(html, /separately approved later phases/);
  assert.doesNotMatch(html, /<(?:table|form|button|iframe)\b/);
});

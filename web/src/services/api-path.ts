export const API_PREFIX = '/api/v1';

export function apiPath(...segments: readonly string[]): string {
  for (const segment of segments) {
    if (!/^[A-Za-z0-9_-]+$/.test(segment)) {
      throw new TypeError('API path segments must contain only letters, digits, underscores, or hyphens');
    }
  }
  return [API_PREFIX, ...segments].join('/');
}

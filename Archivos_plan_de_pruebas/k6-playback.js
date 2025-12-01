import http from 'k6/http';
import { check, sleep } from 'k6';
import { Trend, Counter } from 'k6/metrics';

/**
 * k6 script: Playback (stream start) test
 *
 * Env vars (required):
 *  - BASE_URL: e.g. https://api.mi-plataforma.com
 *  - TOKEN: Bearer token (or leave empty if not needed)
 *  - VIDEO_IDS: comma-separated list of IDs to pick from
 *
 * Optional:
 *  - START_BYTES: how many bytes to fetch to simulate startup (default 1048576 ~1MB)
 *  - TIMEOUT_MS: request timeout (default 60000 ms)
 */

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const TOKEN = __ENV.TOKEN || '';
const TIMEOUT_MS = Number(__ENV.TIMEOUT_MS || 60000);
const START_BYTES = Number(__ENV.START_BYTES || 1048576);
const VIDEO_IDS = (__ENV.VIDEO_IDS || '').split(',').map(s => s.trim()).filter(Boolean);

export const options = {
  scenarios: {
    ramp_playback: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '1m', target: 20 },
        { duration: '3m', target: 100 },
        { duration: '3m', target: 250 },
        { duration: '3m', target: 500 },
        { duration: '2m', target: 0 },
      ],
      gracefulRampDown: '30s',
    },
  },
  thresholds: {
    'http_req_duration{type:startup}': ['avg<1500', 'p(95)<3000'],
    'checks': ['rate>0.99'],
  },
};

const startTrend = new Trend('startup_duration');
const startupBytes = new Counter('startup_bytes');

function pickId() {
  if (VIDEO_IDS.length === 0) return null;
  const idx = Math.floor(Math.random() * VIDEO_IDS.length);
  return VIDEO_IDS[idx];
}

export default function () {
  const headers = {
    'Authorization': TOKEN ? `Bearer ${TOKEN}` : '',
    'Range': `bytes=0-${START_BYTES - 1}`, // request first chunk to simulate playback start
  };

  const id = pickId();
  if (!id) {
    check(null, { 'no video id configured': () => false });
    sleep(1);
    return;
  }

  const res = http.get(`${BASE_URL}/api/videos/${id}/stream`, {
    headers,
    tags: { type: 'startup' },
    timeout: `${TIMEOUT_MS}ms`,
  });

  check(res, {
    'partial content (206) or OK': (r) => r.status === 206 || r.status === 200,
    'received >= requested bytes': (r) => {
      const cl = Number(r.headers['Content-Length'] || 0);
      return cl >= START_BYTES || (r.body && r.body.length >= Math.min(START_BYTES, (r.body.length || 0)));
    },
  });

  startTrend.add(res.timings.duration);

  const len = Number(res.headers['Content-Length'] || 0);
  if (len > 0) startupBytes.add(len);
  else if (res.body && res.body.length) startupBytes.add(res.body.length);

  sleep(1);
}

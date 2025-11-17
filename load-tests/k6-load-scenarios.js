import http from 'k6/http';
import { Trend, Counter } from 'k6/metrics';
import { check, sleep } from 'k6';
import { SharedArray } from 'k6/data';
import { b64encode } from 'k6/encoding';

function boolEnv(name, fallback = false) {
  const raw = (__ENV[name] || '').trim().toLowerCase();
  if (!raw) {
    return fallback;
  }
  return ['1', 'true', 'yes', 'y', 'on'].includes(raw);
}

function parseJSONEnv(name, fallback) {
  const raw = __ENV[name];
  if (!raw) {
    return fallback;
  }
  try {
    const parsed = JSON.parse(raw);
    return parsed;
  } catch (err) {
    console.error(`WARN ${name} tiene un JSON inválido (${err.message}). Se usará el valor por defecto.`);
    return fallback;
  }
}

const API_BASE_URL = (__ENV.API_BASE_URL || 'http://localhost:8080/api').replace(/\/$/, '');
const FRONT_BASE_URL = (__ENV.FRONT_BASE_URL || 'http://localhost:3000').replace(/\/$/, '');
const PUBLIC_LIMIT = Number(__ENV.PUBLIC_LIMIT || 20);
const JWT_TOKEN = __ENV.JWT_TOKEN || '';
const VIDEO_FILE = __ENV.VIDEO_FILE || 'assets/intro.mp4';
const ENABLE_UPLOADS = boolEnv('ENABLE_UPLOADS', false);
const ENABLE_VOTES = boolEnv('ENABLE_VOTES', false);
const CHECK_ASSETS = boolEnv('CHECK_ASSETS', false);

const PUBLIC_STAGES = parseJSONEnv(
  'PUBLIC_STAGES',
  [
    { target: Number(__ENV.PUBLIC_START_RATE || 5), duration: '1m' },
    { target: Number(__ENV.PUBLIC_PEAK_RATE || 30), duration: '4m' },
    { target: Number(__ENV.PUBLIC_END_RATE || 5), duration: '1m' },
  ],
);

const scenarios = {};

// Solo agregar browse_public si tiene rate > 0
if (Number(__ENV.PUBLIC_PEAK_RATE || 30) > 0) {
  scenarios.browse_public = {
    executor: 'ramping-arrival-rate',
    startRate: Number(__ENV.PUBLIC_START_RATE || 5),
    stages: PUBLIC_STAGES,
    timeUnit: '1s',
    preAllocatedVUs: Number(__ENV.PUBLIC_MIN_VUS || 10),
    maxVUs: Number(__ENV.PUBLIC_MAX_VUS || 150),
    exec: 'browsePublic',
  };
}

// Solo agregar browse_frontend si tiene rate > 0
if (Number(__ENV.FRONT_RATE || 8) > 0) {
  scenarios.browse_frontend = {
    executor: 'constant-arrival-rate',
    rate: Number(__ENV.FRONT_RATE || 8),
    duration: __ENV.FRONT_DURATION || '5m',
    timeUnit: '1s',
    preAllocatedVUs: Number(__ENV.FRONT_MIN_VUS || 5),
    maxVUs: Number(__ENV.FRONT_MAX_VUS || 80),
    exec: 'browseLanding',
    startTime: __ENV.FRONT_START_TIME || '30s',
  };
}

// Cargar el archivo de video como binario
let videoBinary = null;
if (ENABLE_UPLOADS) {
  try {
    videoBinary = open(VIDEO_FILE, 'b');
    const size = videoBinary ? (videoBinary.byteLength || videoBinary.length || 0) : 0;
    console.log(`✓ Video cargado exitosamente: ${VIDEO_FILE}`);
    console.log(`✓ Tamaño del video: ${size} bytes`);
  } catch (err) {
    console.error(`✗ ERROR cargando video ${VIDEO_FILE}: ${err.message}`);
    console.error(`Asegúrate de que el archivo existe en la ruta especificada`);
  }
}

if (ENABLE_UPLOADS && videoBinary) {
  scenarios.upload_videos = {
    executor: 'constant-arrival-rate',
    rate: Number(__ENV.UPLOAD_RATE || 2),
    duration: __ENV.UPLOAD_DURATION || '5m',
    timeUnit: '1s',
    preAllocatedVUs: Number(__ENV.UPLOAD_MIN_VUS || 5),
    maxVUs: Number(__ENV.UPLOAD_MAX_VUS || 50),
    exec: 'uploadVideos',
    startTime: __ENV.UPLOAD_START_TIME || '1m',
  };
}

if (ENABLE_VOTES && JWT_TOKEN) {
  scenarios.vote_cycle = {
    executor: 'ramping-arrival-rate',
    startRate: Number(__ENV.VOTE_START_RATE || 2),
    stages: parseJSONEnv(
      'VOTE_STAGES',
      [
        { target: Number(__ENV.VOTE_START_RATE || 2), duration: '1m' },
        { target: Number(__ENV.VOTE_PEAK_RATE || 10), duration: '3m' },
        { target: Number(__ENV.VOTE_END_RATE || 3), duration: '1m' },
      ],
    ),
    timeUnit: '1s',
    preAllocatedVUs: Number(__ENV.VOTE_MIN_VUS || 5),
    maxVUs: Number(__ENV.VOTE_MAX_VUS || 60),
    exec: 'voteCycle',
    startTime: __ENV.VOTE_START_TIME || '90s',
  };
}

export const options = {
  scenarios,
  thresholds: {
    http_req_failed: ['rate<0.02'],
    'http_req_duration{scenario:browse_public}': ['p(95)<800'],
    'http_req_duration{scenario:browse_frontend}': ['p(95)<500'],
    'http_req_duration{scenario:upload_videos}': ['p(95)<4000'],
    'http_req_duration{scenario:vote_cycle}': ['p(95)<1200'],
    's3_asset_duration': ['p(95)<1500'],
  },
};

const s3AssetDuration = new Trend('s3_asset_duration');
const uploadDuration = new Trend('upload_duration');
const voteFailures = new Counter('vote_failures');

function randomPause() {
  const min = Number(__ENV.PAUSE_MIN || 0.7);
  const max = Number(__ENV.PAUSE_MAX || 2);
  return Math.random() * (max - min) + min;
}

function buildAuthHeaders() {
  if (!JWT_TOKEN) {
    return {};
  }
  return { Authorization: `Bearer ${JWT_TOKEN}` };
}

function fetchVideoList(limit = PUBLIC_LIMIT) {
  const res = http.get(`${API_BASE_URL}/public/videos?limit=${limit}&offset=0`, {
    tags: { endpoint: 'GET /public/videos' },
  });
  check(res, {
    'public videos 200': (r) => r.status === 200,
  });
  if (res.status !== 200) {
    return [];
  }
  const data = res.json();
  return Array.isArray(data) ? data : [];
}

function fetchRandomVideoId() {
  const list = fetchVideoList();
  if (list.length === 0) {
    return null;
  }
  const randomIndex = Math.floor(Math.random() * list.length);
  return list[randomIndex].video_id;
}

export function browsePublic() {
  const videos = fetchVideoList();
  if (CHECK_ASSETS && videos.length > 0) {
    const sample = videos[Math.floor(Math.random() * videos.length)];
    if (sample && sample.processed_url) {
      const res = http.get(sample.processed_url, {
        tags: { endpoint: 'GET processed_url' },
        timeout: __ENV.ASSET_TIMEOUT || '5s',
      });
      s3AssetDuration.add(res.timings.duration);
      check(res, {
        'processed asset 200': (r) => r.status === 200,
      });
    }
  }
  sleep(randomPause());
}

export function browseLanding() {
  const res = http.get(`${FRONT_BASE_URL}/`, {
    tags: { endpoint: 'GET / (front)' },
  });
  check(res, {
    'landing page 200': (r) => r.status === 200,
  });
  sleep(randomPause());
}

export function uploadVideos() {
  if (!videoBinary) {
    console.warn('WARN No se encontró el archivo de video definido por VIDEO_FILE. Se omite el escenario.');
    sleep(1);
    return;
  }
  
  const title = `load-test-${Date.now()}-${Math.random().toString(16).slice(2, 8)}`;
  console.log(`[UPLOAD] Intentando subir video: ${title}`);
  console.log(`[UPLOAD] URL: ${API_BASE_URL}/videos`);
  const fileSize = videoBinary ? (videoBinary.byteLength || videoBinary.length || 0) : 0;
  console.log(`[UPLOAD] Token presente: ${JWT_TOKEN ? 'SÍ' : 'NO'}`);
  console.log(`[UPLOAD] Tamaño del archivo: ${fileSize} bytes`);
  
  const payload = {
    title: title,
    video_file: http.file(videoBinary, 'load-test.mp4', 'video/mp4'),
  };
  const params = {
    headers: {
      ...buildAuthHeaders(),
    },
    tags: { endpoint: 'POST /videos' },
  };
  
  const res = http.post(`${API_BASE_URL}/videos`, payload, params);
  uploadDuration.add(res.timings.duration);
  
  console.log(`[UPLOAD] Status Code: ${res.status}`);
  console.log(`[UPLOAD] Response Body: ${res.body}`);
  console.log(`[UPLOAD] Response Headers: ${JSON.stringify(res.headers)}`);
  
  const success = res.status === 201 || res.status === 202;
  if (!success) {
    console.error(`[UPLOAD] ERROR - Expected 201/202, got ${res.status}`);
    console.error(`[UPLOAD] ERROR Body: ${res.body}`);
  } else {
    console.log(`[UPLOAD] ✓ Video subido exitosamente`);
  }
  
  check(res, {
    'upload accepted (201/202)': (r) => r.status === 201 || r.status === 202,
  });
  sleep(randomPause());
}

export function voteCycle() {
  const videoId = fetchRandomVideoId();
  if (!videoId) {
    sleep(1);
    return;
  }
  const headers = {
    ...buildAuthHeaders(),
  };
  const voteRes = http.post(`${API_BASE_URL}/public/videos/${videoId}/vote`, null, {
    headers,
    tags: { endpoint: 'POST /public/videos/:id/vote' },
  });
  const okVote = voteRes.status === 200 || voteRes.status === 400;
  if (!okVote) {
    voteFailures.add(1);
  }
  check(voteRes, {
    'vote request ok (200/400 expected)': () => okVote,
  });
  sleep(Number(__ENV.VOTE_PAUSE || 0.5));
  const unvoteRes = http.del(`${API_BASE_URL}/public/videos/${videoId}/vote`, null, {
    headers,
    tags: { endpoint: 'DELETE /public/videos/:id/vote' },
  });
  const okUnvote = unvoteRes.status === 200 || unvoteRes.status === 400;
  if (!okUnvote) {
    voteFailures.add(1);
  }
  check(unvoteRes, {
    'unvote request ok (200/400 expected)': () => okUnvote,
  });
  sleep(randomPause());
}


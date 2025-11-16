import { check, sleep } from 'k6';
import http from 'k6/http';

const baseURL = __ENV.BASE_URL || 'http://localhost:8080';

export const options = {
  stages: [
    { duration: '10s', target: 5 },
    { duration: '30s', target: 5 },
    { duration: '10s', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<300'],
    http_req_failed: ['rate<0.01'],
  },
};

export function setup() {
  console.log('Creating single test team...');

  const runId = __ENV.RUN_ID || Date.now();
  const teamPayload = JSON.stringify({
    team_name: `k6-test-team-${runId}`,
    members: [
      { user_id: "test-author", username: "TestAuthor", is_active: true },
      { user_id: "test-rev-1", username: "TestRev1", is_active: true },
      { user_id: "test-rev-2", username: "TestRev2", is_active: true },
    ],
  });

  const teamRes = http.post(`${baseURL}/team/add`, teamPayload, {
    headers: { 'Content-Type': 'application/json' },
  });
  
  console.log(`Team creation result: ${teamRes.status}`);
  
  const teamCreated = teamRes.status === 201 ||
    (teamRes.status === 400 && teamRes.json()?.error?.code === 'TEAM_EXISTS');

  if (!teamCreated) {
    throw new Error(`Failed to create test team, status=${teamRes.status}`);
  }

  return {
    teamCreated: true,
    teamName: `k6-test-team-${runId}`
  };
}

export default function (data) {
  const timestamp = Date.now();
  const vuId = __VU;
  const iter = __ITER;
  
  const prId = `PR-${vuId}-${iter}-${timestamp}`;
  
  const prRes = http.post(`${baseURL}/pullRequest/create`, JSON.stringify({
    pull_request_id: prId,
    pull_request_name: `Load Test PR ${vuId}-${iter}`,
    author_id: "test-author",
  }), { headers: { 'Content-Type': 'application/json' } });
  
  check(prRes, { 'PR created': (r) => r.status === 201 });

  const reviewRes = http.get(`${baseURL}/users/getReview?user_id=test-rev-1`);
  check(reviewRes, { 'assignments retrieved': (r) => r.status === 200 });

  sleep(0.1);
}

export function teardown() {
  console.log('Simple load test completed');
}

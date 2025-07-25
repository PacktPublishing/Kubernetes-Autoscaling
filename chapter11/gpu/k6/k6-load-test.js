import http from 'k6/http';
import { check, sleep } from 'k6';
import { open } from 'k6/fs';

const testImageData = open('/data/test-image.png', 'b');

export const options = {
  scenarios: {
    load_test: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '30s', target: 50 },
        { duration: '1m',  target: 100 },
        { duration: '1m',  target: 200 },
        { duration: '1m',  target: 200 },
        { duration: '30s', target: 50 },
        { duration: '30s', target: 0 },
      ],
      gracefulStop: '10s',
    },
  },
  thresholds: {
    http_req_duration: ['p(95)<7000'],
    checks: ['rate>0.90'],
  },
};

export default function () {
  const formData = {
    image: http.file(testImageData, 'test-image.png', 'image/png'),
  };

  const response = http.post('http://gpu-inference-service.default.svc.cluster.local/predict', formData);

  check(response, {
    'prediction status is 200': (r) => r.status === 200,
    'response body contains JSON': (r) => {
      try {
        const json = JSON.parse(r.body);
        return json.predicted_class !== undefined || json.error !== undefined;
      } catch {
        return false;
      }
    }
  });

  sleep(1);
}

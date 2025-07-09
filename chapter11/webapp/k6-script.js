import http from 'k6/http';
import { sleep } from 'k6';

export const options = {
  stages: [
    { duration: '120s', target: 30 },
    { duration: '120s', target: 45 },
    { duration: '90s', target: 60 },
    { duration: '90s', target: 75 },
    { duration: '90', target: 90 },
    { duration: '30s', target: 100 },
    { duration: '15s', target: 0 },
  ]
};

const ITERATIONS = 500000;  // Fixed number of iterations for consistency

export default function () {
  const url = `http://montecarlo-pi.default.svc.cluster.local/monte-carlo-pi?iterations=${ITERATIONS}`;
  http.get(url);
  sleep(1);
}
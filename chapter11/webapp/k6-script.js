import http from 'k6/http';
import { sleep } from 'k6';

export const options = {
  stages: [
    { duration: '120s', target: 300 },
    { duration: '120s', target: 400 },
    { duration: '90s', target: 500 },
    { duration: '90s', target: 700 },
    { duration: '90', target: 850 },
    { duration: '30s', target: 500 },
    { duration: '15s', target: 200 },
  ]
};

const ITERATIONS = 500000;  // Fixed number of iterations for consistency

export default function () {
  const url = `http://montecarlo-pi.default.svc.cluster.local/monte-carlo-pi?iterations=${ITERATIONS}`;
  http.get(url);
  sleep(1);
}
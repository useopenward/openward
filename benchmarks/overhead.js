// benchmarks/overhead.js
import http from 'k6/http';
import { Trend } from 'k6/metrics';



const directLatency  = new Trend('direct_latency',  true);
const proxiedLatency = new Trend('proxied_latency', true);

const API_KEY = 'd17fa4bb-1184-4c22-bfd7-240bf3a20c9d';

export const options = {
  scenarios: {
    direct: {
      executor: 'constant-vus',
      vus: 500,
      duration: '30s',
      exec: 'direct',
    },
    proxied: {
      executor: 'constant-vus',
      vus: 500,
      duration: '30s',
      exec: 'proxied',
    },
  },
};

export function direct() {
  const res = http.get('http://localhost:9999/get');
  directLatency.add(res.timings.duration);
}

export function proxied() {
  const res = http.get('http://localhost:8080/get', {
    headers: { 'X-API-Key': API_KEY },
  });
  proxiedLatency.add(res.timings.duration);
}

import http from 'k6/http';
import { check } from 'k6';

export const options = {
  vus: 20,
  duration: '10s',
};

export default function () {
  const response = http.get('http://localhost:8080/home');

  check(response, {
    'status is 200': (r) => r.status === 200,
  });
}

// run the load test using the following command:
// k6 run loadtest.js

//custom command
// k6 run --vus 20 --duration 10s loadtest.js
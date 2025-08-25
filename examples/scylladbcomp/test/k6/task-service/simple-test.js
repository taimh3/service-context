import http from 'k6/http';
import { check } from 'k6';

export const options = {
    vus: 1,
    duration: '10s',
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export default function () {
    // Simple test to verify k6 setup
    const response = http.get(`${BASE_URL}/ping`);

    check(response, {
        'status is 200': (r) => r.status === 200,
    });
}

import http from 'k6/http';
import { check } from 'k6';

// Simple test configuration
export const options = {
    vus: 1,
    duration: '10s',
};

// Base URL - can be overridden via environment variable
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

// Simple test data
const SIMPLE_PERSON = {
    first_name: 'Test',
    last_name: 'User',
    email: ['test.user@example.com']
};

export default function () {
    console.log('🧪 Running simple Person API test...');

    // Test 1: Health check
    console.log('1. Testing health endpoint...');
    const healthResponse = http.get(`${BASE_URL}/ping`);
    check(healthResponse, {
        'Health check is 200': (r) => r.status === 200,
        'Health check has pong message': (r) => r.json('message') === 'pong',
    });

    // Test 2: Create a person
    console.log('2. Testing person creation...');
    const createResponse = http.post(
        `${BASE_URL}/v1/scylla/persons`,
        JSON.stringify(SIMPLE_PERSON),
        {
            headers: { 'Content-Type': 'application/json' },
        }
    );

    check(createResponse, {
        'Create person is 200': (r) => r.status === 200,
        'Create response has data': (r) => r.json('data') !== undefined,
    });

    // Test 3: List persons
    console.log('3. Testing person listing...');
    const listResponse = http.get(`${BASE_URL}/v1/scylla/persons`);

    check(listResponse, {
        'List persons is 200': (r) => r.status === 200,
        'List response has data array': (r) => Array.isArray(r.json('data')),
    });

    // Test 4: List persons with filter
    console.log('4. Testing person listing with filter...');
    const filterResponse = http.get(`${BASE_URL}/v1/scylla/persons?first_name=Test`);

    check(filterResponse, {
        'Filter persons is 200': (r) => r.status === 200,
        'Filter response has data array': (r) => Array.isArray(r.json('data')),
        'Filter response has proper structure': (r) => {
            try {
                const response = r.json();
                return response.filter !== undefined || response.data !== undefined;
            } catch (e) {
                return false;
            }
        },
    });

    console.log('✅ Simple test completed');
}

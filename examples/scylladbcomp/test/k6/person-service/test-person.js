import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');

// Test configuration
export const options = {
    stages: [
        { duration: '30s', target: 5 },   // Ramp up to 5 users over 30 seconds
        { duration: '1m', target: 10 },   // Stay at 10 users for 1 minute
        { duration: '30s', target: 0 },   // Ramp down to 0 users over 30 seconds
    ],
    thresholds: {
        http_req_duration: ['p(95)<500'], // 95% of requests should complete within 500ms
        errors: ['rate<0.1'],             // Error rate should be less than 10%
        http_req_failed: ['rate<0.1'],    // HTTP error rate should be less than 10%
    },
};

// Base URL - can be overridden via environment variable
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

// Test data
const TEST_PERSON = {
    first_name: 'John',
    last_name: 'Doe',
    email: ['john.doe@example.com', 'john@example.com']
};

const TEST_PERSON_2 = {
    first_name: 'Jane',
    last_name: 'Smith',
    email: ['jane.smith@example.com']
};

const TEST_PERSON_3 = {
    first_name: 'John',
    last_name: 'Smith',
    email: ['john.smith@example.com']
};

export default function () {
    // Test 1: Health Check
    testHealthCheck();

    // Test 2: Create persons for testing
    console.log('Creating test persons...');
    testCreatePerson(TEST_PERSON);
    testCreatePerson(TEST_PERSON_2);
    testCreatePerson(TEST_PERSON_3);

    // Test 3: List all persons
    testListAllPersons();

    // Test 4: List persons with first_name filter
    testListPersonsByFirstName('John');

    // Test 5: List persons with both first_name and last_name filter
    testListPersonsByFullName('John', 'Doe');

    // Test 6: Test edge cases and error scenarios
    testErrorScenarios();

    // Small delay between iterations
    sleep(0.1);
}

function testHealthCheck() {
    const response = http.get(`${BASE_URL}/ping`);

    const success = check(response, {
        'Health check status is 200': (r) => r.status === 200,
        'Health check response time < 200ms': (r) => r.timings.duration < 200,
        'Health check returns pong': (r) => r.json('message') === 'pong',
    });

    if (!success) {
        errorRate.add(1);
        console.error('Health check failed');
    }
}

function testCreatePerson(personData) {
    const params = {
        headers: {
            'Content-Type': 'application/json',
        },
    };

    const response = http.post(
        `${BASE_URL}/v1/scylla/persons`,
        JSON.stringify(personData),
        params
    );

    const success = check(response, {
        'Create person status is 200': (r) => r.status === 200,
        'Create person response time < 1000ms': (r) => r.timings.duration < 1000,
        'Create person returns success message': (r) => r.json('data') !== undefined,
        'Response contains success message': (r) => {
            const data = r.json('data');
            return typeof data === 'string' && data.includes('successfully');
        },
    });

    if (!success) {
        errorRate.add(1);
        console.error('Create person failed:', response.body);
    } else {
        console.log(`Person created: ${personData.first_name} ${personData.last_name}`);
    }

    return success;
}

function testListAllPersons() {
    const response = http.get(`${BASE_URL}/v1/scylla/persons`);

    const success = check(response, {
        'List all persons status is 200': (r) => r.status === 200,
        'List all persons response time < 1000ms': (r) => r.timings.duration < 1000,
        'List all persons returns array': (r) => Array.isArray(r.json('data')),
        // 'Response has proper structure': (r) => {
        //     const responseData = r.json();
        //     return responseData.data !== undefined && responseData.filter !== undefined;
        // },
    });

    if (!success) {
        errorRate.add(1);
        console.error('List all persons failed:', response.body);
    } else {
        const data = response.json();
        console.log(`Found ${data.data.length} persons`);
    }
}

function testListPersonsByFirstName(firstName) {
    const response = http.get(`${BASE_URL}/v1/scylla/persons?first_name=${firstName}`);

    const success = check(response, {
        'List persons by first name status is 200': (r) => r.status === 200,
        'List persons by first name response time < 1000ms': (r) => r.timings.duration < 1000,
        'List persons by first name returns array': (r) => Array.isArray(r.json('data')),
        // 'Filter applied correctly': (r) => {
        //     try {
        //         const data = r.json();
        //         return data.filter && data.filter.first_name === firstName;
        //     } catch (e) {
        //         return false;
        //     }
        // },
        // 'All returned persons have correct first_name': (r) => {
        //     try {
        //         const persons = r.json('data');
        //         return persons.every(person => person.first_name === firstName);
        //     } catch (e) {
        //         return true; // If we can't check, consider it passed to avoid false negatives
        //     }
        // },
    });

    if (!success) {
        errorRate.add(1);
        console.error(`List persons by first name '${firstName}' failed:`, response.body);
    } else {
        const data = response.json();
        console.log(`Found ${data.data.length} persons with first name '${firstName}'`);
    }
}

function testListPersonsByFullName(firstName, lastName) {
    const response = http.get(`${BASE_URL}/v1/scylla/persons?first_name=${firstName}&last_name=${lastName}`);

    const success = check(response, {
        'List persons by full name status is 200': (r) => r.status === 200,
        'List persons by full name response time < 1000ms': (r) => r.timings.duration < 1000,
        'List persons by full name returns array': (r) => Array.isArray(r.json('data')),
        // 'Filter applied correctly': (r) => {
        //     try {
        //         const data = r.json();
        //         return data.filter &&
        //             data.filter.first_name === firstName &&
        //             data.filter.last_name === lastName;
        //     } catch (e) {
        //         return false;
        //     }
        // },
        // 'All returned persons have correct full name': (r) => {
        //     try {
        //         const persons = r.json('data');
        //         return persons.every(person =>
        //             person.first_name === firstName && person.last_name === lastName
        //         );
        //     } catch (e) {
        //         return true; // If we can't check, consider it passed to avoid false negatives
        //     }
        // },
    });

    if (!success) {
        errorRate.add(1);
        console.error(`List persons by full name '${firstName} ${lastName}' failed:`, response.body);
    } else {
        const data = response.json();
        console.log(`Found ${data.data.length} persons with full name '${firstName} ${lastName}'`);
    }
}

function testErrorScenarios() {
    // Test 1: Create person with missing required fields
    const invalidPerson1 = {
        first_name: '', // Empty first name should fail validation
        last_name: 'Doe',
        email: ['test@example.com']
    };

    let response = http.post(
        `${BASE_URL}/v1/scylla/persons`,
        JSON.stringify(invalidPerson1),
        {
            headers: { 'Content-Type': 'application/json' },
        }
    );

    check(response, {
        'Empty first name returns 400': (r) => r.status === 400,
        // 'Error response has error object': (r) => r.json('error') !== undefined,
    });

    // Test 2: Create person with missing last name
    const invalidPerson2 = {
        first_name: 'John',
        last_name: '', // Empty last name should fail validation
        email: ['test@example.com']
    };

    response = http.post(
        `${BASE_URL}/v1/scylla/persons`,
        JSON.stringify(invalidPerson2),
        {
            headers: { 'Content-Type': 'application/json' },
        }
    );

    check(response, {
        'Empty last name returns 400': (r) => r.status === 400,
        // 'Error response has proper structure': (r) => {
        //     const error = r.json('error');
        //     return error && error.code !== undefined;
        // },
    });

    // Test 3: Create person with invalid email format
    const invalidPerson3 = {
        first_name: 'John',
        last_name: 'Doe',
        email: ['invalid-email-format'] // Invalid email format
    };

    response = http.post(
        `${BASE_URL}/v1/scylla/persons`,
        JSON.stringify(invalidPerson3),
        {
            headers: { 'Content-Type': 'application/json' },
        }
    );

    // Note: This might not fail depending on validation implementation
    check(response, {
        'Invalid email handled properly': (r) => r.status === 200 || r.status === 400,
    });

    // Test 4: Create person with malformed JSON
    response = http.post(
        `${BASE_URL}/v1/scylla/persons`,
        '{"first_name": "John", "last_name":}', // Malformed JSON
        {
            headers: { 'Content-Type': 'application/json' },
        }
    );

    check(response, {
        'Malformed JSON returns 400': (r) => r.status === 400,
    });

    // Test 5: Test with invalid query parameters
    response = http.get(`${BASE_URL}/v1/scylla/persons?invalid_param=test`);

    check(response, {
        'Invalid query params handled gracefully': (r) => r.status === 200,
        'Returns array even with invalid params': (r) => Array.isArray(r.json('data')),
    });
}

// Test scenarios for different load patterns
export function testHighLoad() {
    // Simulate high load scenario with multiple list requests
    for (let i = 0; i < 5; i++) {
        testListAllPersons();
        testListPersonsByFirstName('John');
        sleep(0.1);
    }
}

export function testCreateMultiplePersons() {
    // Test creating multiple persons simultaneously
    const testPersons = [
        {
            first_name: 'Alice',
            last_name: 'Johnson',
            email: ['alice.johnson@example.com']
        },
        {
            first_name: 'Bob',
            last_name: 'Wilson',
            email: ['bob.wilson@example.com']
        },
        {
            first_name: 'Charlie',
            last_name: 'Brown',
            email: ['charlie.brown@example.com']
        }
    ];

    testPersons.forEach((person, index) => {
        const response = http.post(
            `${BASE_URL}/v1/scylla/persons`,
            JSON.stringify(person),
            {
                headers: { 'Content-Type': 'application/json' },
            }
        );

        check(response, {
            [`Create person ${index + 1} (${person.first_name}) status is 200`]: (r) => r.status === 200,
            [`Person ${index + 1} response time < 1000ms`]: (r) => r.timings.duration < 1000,
        });
    });
}

// Performance testing scenarios
export function testFilterPerformance() {
    // Test performance of different filter combinations
    const filterTests = [
        { params: '', description: 'no filters' },
        { params: '?first_name=John', description: 'first_name filter only' },
        { params: '?first_name=John&last_name=Doe', description: 'both filters' },
        { params: '?first_name=NonExistent', description: 'non-existent first_name' },
        { params: '?first_name=John&last_name=NonExistent', description: 'partial match' },
    ];

    filterTests.forEach(test => {
        const response = http.get(`${BASE_URL}/v1/scylla/persons${test.params}`);

        check(response, {
            [`Filter test (${test.description}) status is 200`]: (r) => r.status === 200,
            [`Filter test (${test.description}) response time < 500ms`]: (r) => r.timings.duration < 500,
            [`Filter test (${test.description}) returns array`]: (r) => Array.isArray(r.json('data')),
        });
    });
}

// Data validation scenarios
export function testDataValidation() {
    const validationTests = [
        {
            data: { first_name: 'A'.repeat(101), last_name: 'Test' }, // Very long first name
            expectedStatus: [200, 400], // Might pass or fail depending on validation
            description: 'very long first name'
        },
        {
            data: { first_name: 'Test', last_name: 'B'.repeat(101) }, // Very long last name
            expectedStatus: [200, 400],
            description: 'very long last name'
        },
        {
            data: { first_name: 'Test', last_name: 'User', email: [] }, // Empty email array
            expectedStatus: [200],
            description: 'empty email array'
        },
        {
            data: { first_name: 'Test', last_name: 'User', email: null }, // Null email
            expectedStatus: [200, 400],
            description: 'null email'
        }
    ];

    validationTests.forEach(test => {
        const response = http.post(
            `${BASE_URL}/v1/scylla/persons`,
            JSON.stringify(test.data),
            {
                headers: { 'Content-Type': 'application/json' },
            }
        );

        check(response, {
            [`Validation test (${test.description}) returns expected status`]: (r) =>
                test.expectedStatus.includes(r.status),
        });
    });
}

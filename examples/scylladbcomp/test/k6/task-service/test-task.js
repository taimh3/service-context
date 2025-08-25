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
const TEST_TASK = {
    title: 'K6 Test Task',
    description: 'This is a test task created by k6 load testing',
    status: 'doing'
};

const UPDATED_TASK = {
    title: 'K6 Updated Test Task',
    description: 'This task has been updated by k6 load testing',
    status: 'done'
};

export default function () {
    // Test 1: Health Check
    testHealthCheck();

    // Test 2: Create a new task
    console.log('Creating a new task...');
    const taskId = testCreateTask();

    // Test 3: Get the created task
    if (taskId) {
        console.log(`Retrieving task with ID: ${taskId}`);
        testGetTask(taskId);

        // Test 4: Update the task
        testUpdateTask(taskId);

        // Test 5: List tasks with filters
        testListTasks();

        // Test 6: Delete the task
        testDeleteTask(taskId);
    }

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

function testCreateTask() {
    const params = {
        headers: {
            'Content-Type': 'application/json',
        },
    };

    const response = http.post(
        `${BASE_URL}/v1/scylla/tasks`,
        JSON.stringify(TEST_TASK),
        params
    );

    const success = check(response, {
        'Create task status is 200': (r) => r.status === 200,
        'Create task response time < 1000ms': (r) => r.timings.duration < 1000,
        'Create task returns success message': (r) => r.json('data') !== undefined,
    });

    if (!success) {
        errorRate.add(1);
        console.error('Create task failed:', response.body);
        return null;
    }

    // Extract task ID from response if available
    // Note: Adjust this based on actual API response format
    const responseData = response.json();
    if (responseData && responseData.data) {
        // If the response contains the created task ID, extract it
        // This might need adjustment based on actual API response
        console.log('Task created successfully:', responseData.data);

        return responseData.data;
    }

    // For testing purposes, we'll use a mock task ID
    // In real scenario, you'd extract this from the response
    return 'test_task_' + Math.floor(Math.random() * 10000);
}

function testGetTask(taskId) {
    const response = http.get(`${BASE_URL}/v1/scylla/tasks/${taskId}`);

    const success = check(response, {
        'Get task status is 200 or 404': (r) => r.status === 200 || r.status === 404,
        'Get task response time < 500ms': (r) => r.timings.duration < 500,
    });

    if (response.status === 200) {
        const additionalChecks = check(response, {
            'Get task returns task data': (r) => r.json('data') !== undefined,
            'Task has required fields': (r) => {
                const task = r.json('data');
                return task && task.id && task.title && task.status;
            },
        });

        if (!additionalChecks) {
            errorRate.add(1);
        }
    } else if (response.status === 404) {
        console.log(`Task ${taskId} not found (expected for test data)`);
    } else {
        errorRate.add(1);
        console.error('Get task failed with unexpected status:', response.status);
    }
}

function testUpdateTask(taskId) {
    const params = {
        headers: {
            'Content-Type': 'application/json',
        },
    };

    const response = http.patch(
        `${BASE_URL}/v1/scylla/tasks/${taskId}`,
        JSON.stringify(UPDATED_TASK),
        params
    );

    const success = check(response, {
        'Update task status is 200 or 404': (r) => r.status === 200 || r.status === 404,
        'Update task response time < 1000ms': (r) => r.timings.duration < 1000,
    });

    if (response.status === 200) {
        const additionalChecks = check(response, {
            'Update task returns success message': (r) => r.json('data') !== undefined,
        });

        if (!additionalChecks) {
            errorRate.add(1);
        }
    } else if (response.status === 404) {
        console.log(`Task ${taskId} not found for update (expected for test data)`);
    } else {
        errorRate.add(1);
        console.error('Update task failed with unexpected status:', response.status);
    }
}

function testListTasks() {
    // Test listing all tasks
    let response = http.get(`${BASE_URL}/v1/scylla/tasks`);

    let success = check(response, {
        'List all tasks status is 200': (r) => r.status === 200,
        'List all tasks response time < 1000ms': (r) => r.timings.duration < 1000,
        'List all tasks returns array': (r) => Array.isArray(r.json('data')),
    });

    if (!success) {
        errorRate.add(1);
        console.error('List all tasks failed:', response.body);
    }

    // Test listing tasks with status filter
    response = http.get(`${BASE_URL}/v1/scylla/tasks?status=doing`);

    success = check(response, {
        'List tasks with filter status is 200': (r) => r.status === 200,
        'List tasks with filter response time < 1000ms': (r) => r.timings.duration < 1000,
        'List tasks with filter returns array': (r) => Array.isArray(r.json('data')),
        // 'Filter applied correctly': (r) => {
        //     const data = r.json();
        //     return data.filter && data.filter.status === 'doing';
        // },
    });

    if (!success) {
        errorRate.add(1);
        console.error('List tasks with filter failed:', response.body);
    }

    // Test listing tasks with pagination
    response = http.get(`${BASE_URL}/v1/scylla/tasks?limit=5`);

    success = check(response, {
        'List tasks with pagination status is 200': (r) => r.status === 200,
        'List tasks with pagination response time < 1000ms': (r) => r.timings.duration < 1000,
        'Pagination metadata present': (r) => {
            const data = r.json();
            return data.paging && typeof data.paging.limit === 'number';
        },
    });

    if (!success) {
        errorRate.add(1);
        console.error('List tasks with pagination failed:', response.body);
    }
}

function testDeleteTask(taskId) {
    const response = http.del(`${BASE_URL}/v1/scylla/tasks/${taskId}`);

    const success = check(response, {
        'Delete task status is 200 or 404': (r) => r.status === 200 || r.status === 404,
        'Delete task response time < 1000ms': (r) => r.timings.duration < 1000,
    });

    if (response.status === 200) {
        const additionalChecks = check(response, {
            'Delete task returns success message': (r) => r.json('data') !== undefined,
        });

        if (!additionalChecks) {
            errorRate.add(1);
        }
    } else if (response.status === 404) {
        console.log(`Task ${taskId} not found for deletion (expected for test data)`);
    } else {
        errorRate.add(1);
        console.error('Delete task failed with unexpected status:', response.status);
    }
}

// Test scenarios for different load patterns
export function testHighLoad() {
    // Simulate high load scenario
    for (let i = 0; i < 5; i++) {
        testListTasks();
        sleep(0.1);
    }
}

export function testCreateMultipleTasks() {
    // Test creating multiple tasks simultaneously
    const tasks = [];
    for (let i = 0; i < 3; i++) {
        const taskData = {
            ...TEST_TASK,
            title: `${TEST_TASK.title} ${i}`,
        };

        const response = http.post(
            `${BASE_URL}/v1/scylla/tasks`,
            JSON.stringify(taskData),
            {
                headers: { 'Content-Type': 'application/json' },
            }
        );

        check(response, {
            [`Create task ${i} status is 200`]: (r) => r.status === 200,
        });
    }
}

// Error handling scenarios
export function testErrorScenarios() {
    // Test creating task with invalid data
    const invalidTask = {
        title: '', // Empty title should fail validation
        status: 'invalid_status'
    };

    const response = http.post(
        `${BASE_URL}/v1/scylla/tasks`,
        JSON.stringify(invalidTask),
        {
            headers: { 'Content-Type': 'application/json' },
        }
    );

    check(response, {
        'Invalid task creation returns 400': (r) => r.status === 400,
        'Error response has error object': (r) => r.json('error') !== undefined,
    });

    // Test getting non-existent task
    const nonExistentResponse = http.get(`${BASE_URL}/v1/scylla/tasks/99999999`);

    check(nonExistentResponse, {
        'Non-existent task returns 404': (r) => r.status === 404,
        'Not found error has proper structure': (r) => {
            const error = r.json('error');
            return error && error.code === 'NOT_FOUND';
        },
    });
}

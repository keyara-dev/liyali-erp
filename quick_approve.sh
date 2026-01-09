#!/bin/bash

# Quick approval script for testing automation
TASK_ID="b7cfdd47-9c7e-4279-8df1-6a5d3cf04c1e"

# Stage 1: Department Manager
echo "Approving Stage 1..."
MANAGER_TOKEN=$(curl -s -X POST "http://localhost:8080/api/v1/auth/login" -H "Content-Type: application/json" -d '{"email": "manager@demo.com", "password": "password"}' | jq -r '.data.accessToken')
curl -s -X POST "http://localhost:8080/api/v1/approvals/$TASK_ID/approve" -H "Authorization: Bearer $MANAGER_TOKEN" -H "Content-Type: application/json" -d '{"signature": "manager_approval", "comment": "Approved by department manager"}'

# Get next task
TASK_ID=$(psql -h localhost -U postgres -d liyali-dev-db -t -c "SELECT id FROM workflow_tasks WHERE entity_id = '82490e7c-a0a4-494e-aa4f-96e3acb2b8b4' AND status = 'pending' ORDER BY stage_number LIMIT 1;" | tr -d ' ')

# Stage 2: Finance
echo "Approving Stage 2..."
FINANCE_TOKEN=$(curl -s -X POST "http://localhost:8080/api/v1/auth/login" -H "Content-Type: application/json" -d '{"email": "finance@demo.com", "password": "password"}' | jq -r '.data.accessToken')
curl -s -X POST "http://localhost:8080/api/v1/approvals/$TASK_ID/approve" -H "Authorization: Bearer $FINANCE_TOKEN" -H "Content-Type: application/json" -d '{"signature": "finance_approval", "comment": "Approved by finance"}'

# Get next task
TASK_ID=$(psql -h localhost -U postgres -d liyali-dev-db -t -c "SELECT id FROM workflow_tasks WHERE entity_id = '82490e7c-a0a4-494e-aa4f-96e3acb2b8b4' AND status = 'pending' ORDER BY stage_number LIMIT 1;" | tr -d ' ')

# Stage 3: Final Approver
echo "Approving Stage 3..."
APPROVER_TOKEN=$(curl -s -X POST "http://localhost:8080/api/v1/auth/login" -H "Content-Type: application/json" -d '{"email": "approver@demo.com", "password": "password"}' | jq -r '.data.accessToken')
curl -s -X POST "http://localhost:8080/api/v1/approvals/$TASK_ID/approve" -H "Authorization: Bearer $APPROVER_TOKEN" -H "Content-Type: application/json" -d '{"signature": "final_approval", "comment": "Final approval granted"}'

echo "All stages approved!"
#!/usr/bin/env pwsh
# Test admin console login

Write-Host "Testing Admin Console Login..." -ForegroundColor Cyan
Write-Host ""

# Test 1: Check if admin console is accessible
Write-Host "1. Testing admin console accessibility..." -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri "http://localhost:3001/login" -Method GET -UseBasicParsing
    Write-Host "   ✓ Admin console is accessible (Status: $($response.StatusCode))" -ForegroundColor Green
} catch {
    Write-Host "   ✗ Admin console is not accessible: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

Write-Host ""

# Test 2: Check if backend API is accessible
Write-Host "2. Testing backend API accessibility..." -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri "http://localhost:8081/health" -Method GET -UseBasicParsing
    Write-Host "   ✓ Backend API is accessible (Status: $($response.StatusCode))" -ForegroundColor Green
} catch {
    Write-Host "   ✗ Backend API is not accessible: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

Write-Host ""

# Test 3: Test direct backend login
Write-Host "3. Testing direct backend login..." -ForegroundColor Yellow
$loginBody = @{
    email = "admin@liyali.com"
    password = "password"
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "http://localhost:8081/api/v1/auth/login" -Method POST -Body $loginBody -ContentType "application/json"
    
    if ($response.success) {
        Write-Host "   ✓ Backend login successful" -ForegroundColor Green
        Write-Host "   User: $($response.data.user.name)" -ForegroundColor Gray
        Write-Host "   Role: $($response.data.user.role)" -ForegroundColor Gray
        Write-Host "   Token: $($response.data.accessToken.Substring(0, 50))..." -ForegroundColor Gray
    } else {
        Write-Host "   ✗ Backend login failed: $($response.message)" -ForegroundColor Red
        exit 1
    }
} catch {
    Write-Host "   ✗ Backend login error: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "All tests passed! ✓" -ForegroundColor Green
Write-Host ""
Write-Host "Admin Console Credentials:" -ForegroundColor Cyan
Write-Host "  URL:      http://localhost:3001/login" -ForegroundColor White
Write-Host "  Email:    admin@liyali.com" -ForegroundColor White
Write-Host "  Password: password" -ForegroundColor White
Write-Host ""
Write-Host "You can now test the login in your browser!" -ForegroundColor Yellow

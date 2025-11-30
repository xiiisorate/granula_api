# Quick API Test Script
# Tests basic API functionality

$baseUrl = "http://localhost:8080/api/v1"
$timestamp = Get-Date -Format "yyyyMMddHHmmss"
$email = "test_integ_$timestamp@test.com"
$global:token = ""

Write-Host "============================================"
Write-Host "GRANULA API - Quick Integration Test"
Write-Host "============================================"
Write-Host ""

# Test 1: Health check
Write-Host "1. Testing Health Endpoint..."
try {
    $health = Invoke-WebRequest -Uri "http://localhost:8080/health" -UseBasicParsing
    Write-Host "   Status: $($health.StatusCode) - PASS" -ForegroundColor Green
} catch {
    Write-Host "   FAIL: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# Test 2: Registration
Write-Host "2. Testing User Registration..."
Write-Host "   Email: $email"
$regBody = @{
    email = $email
    password = "Test123456"
    name = "Test User"
} | ConvertTo-Json

try {
    $regResult = Invoke-WebRequest -Uri "$baseUrl/auth/register" -Method POST -Body $regBody -ContentType "application/json" -UseBasicParsing
    $regContent = $regResult.Content | ConvertFrom-Json
    if ($regContent.data.access_token) {
        $global:token = $regContent.data.access_token
        Write-Host "   Status: $($regResult.StatusCode) - PASS" -ForegroundColor Green
    } else {
        Write-Host "   FAIL: No token received" -ForegroundColor Red
    }
} catch {
    Write-Host "   FAIL: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 3: Get Profile
if ($global:token) {
    Write-Host "3. Testing Get Profile..."
    try {
        $headers = @{ Authorization = "Bearer $global:token" }
        $profileResult = Invoke-WebRequest -Uri "$baseUrl/users/me" -Headers $headers -UseBasicParsing
        Write-Host "   Status: $($profileResult.StatusCode) - PASS" -ForegroundColor Green
    } catch {
        Write-Host "   FAIL: $($_.Exception.Message)" -ForegroundColor Red
    }

    # Test 4: Create Workspace
    Write-Host "4. Testing Create Workspace..."
    $wsBody = @{
        name = "Test Workspace $timestamp"
        description = "Created during integration test"
    } | ConvertTo-Json

    try {
        $wsResult = Invoke-WebRequest -Uri "$baseUrl/workspaces" -Method POST -Headers $headers -Body $wsBody -ContentType "application/json" -UseBasicParsing
        $wsContent = $wsResult.Content | ConvertFrom-Json
        if ($wsContent.data.id) {
            $global:workspaceId = $wsContent.data.id
            Write-Host "   Status: $($wsResult.StatusCode) - PASS (ID: $($wsContent.data.id.Substring(0,8))...)" -ForegroundColor Green
        }
    } catch {
        Write-Host "   FAIL: $($_.Exception.Message)" -ForegroundColor Red
    }

    # Test 5: List Workspaces
    Write-Host "5. Testing List Workspaces..."
    try {
        $listResult = Invoke-WebRequest -Uri "$baseUrl/workspaces" -Headers $headers -UseBasicParsing
        $listContent = $listResult.Content | ConvertFrom-Json
        Write-Host "   Status: $($listResult.StatusCode) - PASS (Count: $($listContent.data.workspaces.Count))" -ForegroundColor Green
    } catch {
        Write-Host "   FAIL: $($_.Exception.Message)" -ForegroundColor Red
    }

    # Test 6: List Notifications
    Write-Host "6. Testing List Notifications..."
    try {
        $notifResult = Invoke-WebRequest -Uri "$baseUrl/notifications" -Headers $headers -UseBasicParsing
        Write-Host "   Status: $($notifResult.StatusCode) - PASS" -ForegroundColor Green
    } catch {
        Write-Host "   FAIL: $($_.Exception.Message)" -ForegroundColor Red
    }
}

Write-Host ""
Write-Host "============================================"
Write-Host "Quick test completed!"
Write-Host "============================================"


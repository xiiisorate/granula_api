# =============================================================================
# Granula API - Full E2E Test Suite
# =============================================================================
# Comprehensive end-to-end testing of all API endpoints
# =============================================================================

$ErrorActionPreference = "Continue"
$BaseUrl = "http://localhost:8080"
$TestResults = @()

function Write-TestHeader($name) {
    Write-Host "`n" -NoNewline
    Write-Host "=" * 60 -ForegroundColor Cyan
    Write-Host "  $name" -ForegroundColor Yellow
    Write-Host "=" * 60 -ForegroundColor Cyan
}

function Write-TestResult($name, $status, $details = "") {
    $icon = if ($status) { "✅" } else { "❌" }
    $color = if ($status) { "Green" } else { "Red" }
    Write-Host "$icon $name" -ForegroundColor $color
    if ($details) { Write-Host "   $details" -ForegroundColor Gray }
    $script:TestResults += @{Name=$name; Status=$status; Details=$details}
}

function Invoke-ApiTest {
    param(
        [string]$Method,
        [string]$Endpoint,
        [object]$Body = $null,
        [hashtable]$Headers = @{},
        [int]$ExpectedStatus = 200,
        [string]$TestName
    )
    
    try {
        $params = @{
            Uri = "$BaseUrl$Endpoint"
            Method = $Method
            ContentType = "application/json"
            UseBasicParsing = $true
        }
        
        if ($Body) {
            $params.Body = if ($Body -is [string]) { $Body } else { $Body | ConvertTo-Json -Depth 10 }
        }
        
        if ($Headers.Count -gt 0) {
            $params.Headers = $Headers
        }
        
        $response = Invoke-WebRequest @params -ErrorAction Stop
        $success = $response.StatusCode -eq $ExpectedStatus
        $content = $response.Content | ConvertFrom-Json -ErrorAction SilentlyContinue
        
        Write-TestResult $TestName $success "Status: $($response.StatusCode)"
        return @{Success=$success; Response=$content; StatusCode=$response.StatusCode}
    }
    catch {
        $statusCode = $_.Exception.Response.StatusCode.value__
        if ($statusCode -eq $ExpectedStatus) {
            Write-TestResult $TestName $true "Expected error status: $statusCode"
            return @{Success=$true; Response=$null; StatusCode=$statusCode}
        }
        Write-TestResult $TestName $false "Error: $($_.Exception.Message)"
        return @{Success=$false; Response=$null; StatusCode=$statusCode}
    }
}

# =============================================================================
# Test Variables
# =============================================================================
$TestEmail = "e2e_test_$(Get-Random)@granula.ru"
$TestPassword = "SecureE2EPass123!"
$TestName = "E2E Test User"
$AccessToken = ""
$RefreshToken = ""
$UserId = ""
$WorkspaceId = ""
$SceneId = ""
$NotificationId = ""

# =============================================================================
# 1. HEALTH CHECKS
# =============================================================================
Write-TestHeader "1. HEALTH CHECKS"

Invoke-ApiTest -Method "GET" -Endpoint "/health" -TestName "Health Check"
Invoke-ApiTest -Method "GET" -Endpoint "/ready" -TestName "Readiness Check"

# =============================================================================
# 2. AUTH ENDPOINTS
# =============================================================================
Write-TestHeader "2. AUTH ENDPOINTS"

# Register
$registerBody = @{email=$TestEmail; password=$TestPassword; name=$TestName}
$registerResult = Invoke-ApiTest -Method "POST" -Endpoint "/api/v1/auth/register" -Body $registerBody -ExpectedStatus 201 -TestName "Auth: Register new user"

if ($registerResult.Success -and $registerResult.Response) {
    $AccessToken = $registerResult.Response.data.access_token
    $RefreshToken = $registerResult.Response.data.refresh_token
    $UserId = $registerResult.Response.data.user_id
    Write-Host "   User ID: $UserId" -ForegroundColor Gray
}

# Register duplicate (should fail)
Invoke-ApiTest -Method "POST" -Endpoint "/api/v1/auth/register" -Body $registerBody -ExpectedStatus 409 -TestName "Auth: Register duplicate (expect 409)"

# Login
$loginBody = @{email=$TestEmail; password=$TestPassword}
$loginResult = Invoke-ApiTest -Method "POST" -Endpoint "/api/v1/auth/login" -Body $loginBody -TestName "Auth: Login"

if ($loginResult.Success -and $loginResult.Response) {
    $AccessToken = $loginResult.Response.data.access_token
    $RefreshToken = $loginResult.Response.data.refresh_token
}

# Login with wrong password
$wrongLoginBody = @{email=$TestEmail; password="WrongPassword123"}
Invoke-ApiTest -Method "POST" -Endpoint "/api/v1/auth/login" -Body $wrongLoginBody -ExpectedStatus 401 -TestName "Auth: Login wrong password (expect 401)"

# Refresh Token
$refreshBody = @{refresh_token=$RefreshToken}
$refreshResult = Invoke-ApiTest -Method "POST" -Endpoint "/api/v1/auth/refresh" -Body $refreshBody -TestName "Auth: Refresh token"

if ($refreshResult.Success -and $refreshResult.Response) {
    $AccessToken = $refreshResult.Response.data.access_token
    $RefreshToken = $refreshResult.Response.data.refresh_token
}

# Invalid refresh token
$invalidRefreshBody = @{refresh_token="invalid_token_here"}
Invoke-ApiTest -Method "POST" -Endpoint "/api/v1/auth/refresh" -Body $invalidRefreshBody -ExpectedStatus 401 -TestName "Auth: Invalid refresh token (expect 401)"

# =============================================================================
# 3. USER ENDPOINTS (Protected)
# =============================================================================
Write-TestHeader "3. USER ENDPOINTS"

$authHeaders = @{Authorization = "Bearer $AccessToken"}

# Get Profile
Invoke-ApiTest -Method "GET" -Endpoint "/api/v1/users/me" -Headers $authHeaders -TestName "User: Get profile"

# Update Profile
$updateBody = @{name="Updated E2E User"}
Invoke-ApiTest -Method "PATCH" -Endpoint "/api/v1/users/me" -Body $updateBody -Headers $authHeaders -TestName "User: Update profile"

# Change Password
$passwordBody = @{current_password=$TestPassword; new_password="NewSecurePass456!"}
Invoke-ApiTest -Method "PUT" -Endpoint "/api/v1/users/me/password" -Body $passwordBody -Headers $authHeaders -TestName "User: Change password"

# Update password for further tests
$TestPassword = "NewSecurePass456!"

# Unauthorized access
Invoke-ApiTest -Method "GET" -Endpoint "/api/v1/users/me" -ExpectedStatus 401 -TestName "User: Unauthorized access (expect 401)"

# =============================================================================
# 4. NOTIFICATION ENDPOINTS (Protected)
# =============================================================================
Write-TestHeader "4. NOTIFICATION ENDPOINTS"

# Get Notifications
Invoke-ApiTest -Method "GET" -Endpoint "/api/v1/notifications" -Headers $authHeaders -TestName "Notifications: List"

# Get Unread Count
Invoke-ApiTest -Method "GET" -Endpoint "/api/v1/notifications/count" -Headers $authHeaders -TestName "Notifications: Unread count"

# Mark All As Read
Invoke-ApiTest -Method "POST" -Endpoint "/api/v1/notifications/read-all" -Headers $authHeaders -TestName "Notifications: Mark all read"

# Delete All Read
Invoke-ApiTest -Method "DELETE" -Endpoint "/api/v1/notifications" -Headers $authHeaders -TestName "Notifications: Delete all read"

# =============================================================================
# 5. WORKSPACE ENDPOINTS (Protected)
# =============================================================================
Write-TestHeader "5. WORKSPACE ENDPOINTS"

# List Workspaces
Invoke-ApiTest -Method "GET" -Endpoint "/api/v1/workspaces" -Headers $authHeaders -TestName "Workspaces: List"

# Create Workspace
$workspaceBody = @{name="E2E Test Workspace"; description="Test workspace for E2E"}
$workspaceResult = Invoke-ApiTest -Method "POST" -Endpoint "/api/v1/workspaces" -Body $workspaceBody -Headers $authHeaders -ExpectedStatus 201 -TestName "Workspaces: Create"

# Get Workspace
Invoke-ApiTest -Method "GET" -Endpoint "/api/v1/workspaces/test-id" -Headers $authHeaders -TestName "Workspaces: Get by ID"

# Update Workspace
Invoke-ApiTest -Method "PATCH" -Endpoint "/api/v1/workspaces/test-id" -Body @{name="Updated"} -Headers $authHeaders -TestName "Workspaces: Update"

# Delete Workspace
Invoke-ApiTest -Method "DELETE" -Endpoint "/api/v1/workspaces/test-id" -Headers $authHeaders -TestName "Workspaces: Delete"

# =============================================================================
# 6. SCENE ENDPOINTS (Protected)
# =============================================================================
Write-TestHeader "6. SCENE ENDPOINTS"

# List Scenes
Invoke-ApiTest -Method "GET" -Endpoint "/api/v1/scenes" -Headers $authHeaders -TestName "Scenes: List"

# Create Scene
Invoke-ApiTest -Method "POST" -Endpoint "/api/v1/scenes" -Body @{name="Test Scene"} -Headers $authHeaders -ExpectedStatus 201 -TestName "Scenes: Create"

# Get Scene
Invoke-ApiTest -Method "GET" -Endpoint "/api/v1/scenes/test-id" -Headers $authHeaders -TestName "Scenes: Get by ID"

# Update Scene
Invoke-ApiTest -Method "PATCH" -Endpoint "/api/v1/scenes/test-id" -Body @{name="Updated Scene"} -Headers $authHeaders -TestName "Scenes: Update"

# Delete Scene
Invoke-ApiTest -Method "DELETE" -Endpoint "/api/v1/scenes/test-id" -Headers $authHeaders -TestName "Scenes: Delete"

# =============================================================================
# 7. AI ENDPOINTS (Protected)
# =============================================================================
Write-TestHeader "7. AI ENDPOINTS"

# AI Recognize
Invoke-ApiTest -Method "POST" -Endpoint "/api/v1/ai/recognize" -Body @{image_url="test"} -Headers $authHeaders -TestName "AI: Recognize floor plan"

# AI Generate
Invoke-ApiTest -Method "POST" -Endpoint "/api/v1/ai/generate" -Body @{prompt="test"} -Headers $authHeaders -TestName "AI: Generate design"

# AI Chat
Invoke-ApiTest -Method "POST" -Endpoint "/api/v1/ai/chat" -Body @{message="test"} -Headers $authHeaders -TestName "AI: Chat"

# =============================================================================
# 8. LOGOUT
# =============================================================================
Write-TestHeader "8. LOGOUT"

# Logout
$logoutBody = @{refresh_token=$RefreshToken}
Invoke-ApiTest -Method "POST" -Endpoint "/api/v1/auth/logout" -Body $logoutBody -Headers $authHeaders -TestName "Auth: Logout"

# Logout All (re-login first)
$loginBody = @{email=$TestEmail; password=$TestPassword}
$loginResult = Invoke-ApiTest -Method "POST" -Endpoint "/api/v1/auth/login" -Body $loginBody -TestName "Auth: Re-login for logout-all"

if ($loginResult.Success -and $loginResult.Response) {
    $AccessToken = $loginResult.Response.data.access_token
    $authHeaders = @{Authorization = "Bearer $AccessToken"}
}

Invoke-ApiTest -Method "POST" -Endpoint "/api/v1/auth/logout-all" -Headers $authHeaders -TestName "Auth: Logout all devices"

# =============================================================================
# 9. DELETE ACCOUNT
# =============================================================================
Write-TestHeader "9. CLEANUP"

# Re-login before delete
$loginResult = Invoke-ApiTest -Method "POST" -Endpoint "/api/v1/auth/login" -Body $loginBody -TestName "Auth: Re-login for delete"

if ($loginResult.Success -and $loginResult.Response) {
    $AccessToken = $loginResult.Response.data.access_token
    $authHeaders = @{Authorization = "Bearer $AccessToken"}
}

# Delete Account
Invoke-ApiTest -Method "DELETE" -Endpoint "/api/v1/users/me" -Headers $authHeaders -TestName "User: Delete account"

# =============================================================================
# TEST SUMMARY
# =============================================================================
Write-TestHeader "TEST SUMMARY"

$passed = ($TestResults | Where-Object { $_.Status -eq $true }).Count
$failed = ($TestResults | Where-Object { $_.Status -eq $false }).Count
$total = $TestResults.Count

Write-Host "`nTotal Tests: $total" -ForegroundColor White
Write-Host "Passed: $passed" -ForegroundColor Green
Write-Host "Failed: $failed" -ForegroundColor $(if ($failed -gt 0) { "Red" } else { "Green" })
Write-Host "`nSuccess Rate: $([math]::Round($passed/$total*100, 1))%" -ForegroundColor $(if ($failed -gt 0) { "Yellow" } else { "Green" })

if ($failed -gt 0) {
    Write-Host "`nFailed Tests:" -ForegroundColor Red
    $TestResults | Where-Object { $_.Status -eq $false } | ForEach-Object {
        Write-Host "  - $($_.Name): $($_.Details)" -ForegroundColor Red
    }
}

Write-Host "`n" -NoNewline


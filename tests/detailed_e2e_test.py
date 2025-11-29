#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
=============================================================================
Granula API - Detailed E2E Test Suite
=============================================================================
Comprehensive end-to-end testing with full request/response logging
=============================================================================
"""

import json
import random
import string
import time
from datetime import datetime
from dataclasses import dataclass
from typing import Optional, Dict, Any
import requests

# =============================================================================
# Configuration
# =============================================================================

BASE_URL = "http://localhost:8080"
TIMEOUT = 30

# Colors for terminal output
class Colors:
    HEADER = '\033[95m'
    BLUE = '\033[94m'
    CYAN = '\033[96m'
    GREEN = '\033[92m'
    YELLOW = '\033[93m'
    RED = '\033[91m'
    WHITE = '\033[97m'
    ENDC = '\033[0m'
    BOLD = '\033[1m'
    UNDERLINE = '\033[4m'
    GRAY = '\033[90m'

# =============================================================================
# Helper Classes
# =============================================================================

@dataclass
class TestResult:
    name: str
    passed: bool
    expected_status: int
    actual_status: int
    request_method: str
    request_url: str
    request_headers: Dict[str, str]
    request_body: Optional[Dict]
    response_body: Optional[Dict]
    response_time_ms: float
    error: Optional[str] = None

class TestContext:
    """Stores test context between tests"""
    def __init__(self):
        self.email = f"e2e_test_{''.join(random.choices(string.digits, k=10))}@granula.ru"
        self.password = "SecureE2EPass123!"
        self.name = "E2E Test User"
        self.access_token = ""
        self.refresh_token = ""
        self.user_id = ""

# =============================================================================
# Output Functions
# =============================================================================

def print_header(text: str):
    """Print section header"""
    print(f"\n{Colors.CYAN}{'=' * 80}{Colors.ENDC}")
    print(f"{Colors.CYAN}{Colors.BOLD}  {text}{Colors.ENDC}")
    print(f"{Colors.CYAN}{'=' * 80}{Colors.ENDC}\n")

def print_test_start(name: str, method: str, url: str):
    """Print test start"""
    print(f"{Colors.YELLOW}▶ {name}{Colors.ENDC}")
    print(f"  {Colors.GRAY}├─ Method: {Colors.BLUE}{method}{Colors.ENDC}")
    print(f"  {Colors.GRAY}├─ URL: {Colors.BLUE}{url}{Colors.ENDC}")

def print_request(headers: Dict, body: Optional[Dict]):
    """Print request details"""
    print(f"  {Colors.GRAY}├─ Headers:{Colors.ENDC}")
    for key, value in headers.items():
        # Mask authorization token
        if key.lower() == "authorization" and len(value) > 30:
            value = value[:20] + "..." + value[-10:]
        print(f"  {Colors.GRAY}│   {key}: {Colors.CYAN}{value}{Colors.ENDC}")
    
    if body:
        print(f"  {Colors.GRAY}├─ Request Body:{Colors.ENDC}")
        body_str = json.dumps(body, indent=4, ensure_ascii=False)
        for line in body_str.split('\n'):
            # Mask passwords
            if '"password"' in line.lower():
                line = line.split(':')[0] + ': "********"'
            print(f"  {Colors.GRAY}│   {Colors.WHITE}{line}{Colors.ENDC}")

def print_response(status: int, expected: int, body: Optional[Dict], time_ms: float):
    """Print response details"""
    status_color = Colors.GREEN if status == expected else Colors.RED
    print(f"  {Colors.GRAY}├─ Status: {status_color}{status}{Colors.ENDC} (expected: {expected})")
    print(f"  {Colors.GRAY}├─ Time: {Colors.CYAN}{time_ms:.2f}ms{Colors.ENDC}")
    
    if body:
        print(f"  {Colors.GRAY}├─ Response Body:{Colors.ENDC}")
        body_str = json.dumps(body, indent=4, ensure_ascii=False)
        for line in body_str.split('\n')[:20]:  # Limit to 20 lines
            # Mask tokens
            if 'token' in line.lower() and ':' in line:
                parts = line.split(':')
                if len(parts) > 1 and len(parts[1]) > 30:
                    line = parts[0] + ': "' + parts[1].strip(' "')[:15] + '..."'
            print(f"  {Colors.GRAY}│   {Colors.WHITE}{line}{Colors.ENDC}")
        if len(body_str.split('\n')) > 20:
            print(f"  {Colors.GRAY}│   ... (truncated){Colors.ENDC}")

def print_test_result(passed: bool, error: Optional[str] = None):
    """Print test result"""
    if passed:
        print(f"  {Colors.GRAY}└─ Result: {Colors.GREEN}✓ PASSED{Colors.ENDC}\n")
    else:
        print(f"  {Colors.GRAY}└─ Result: {Colors.RED}✗ FAILED{Colors.ENDC}")
        if error:
            print(f"      {Colors.RED}Error: {error}{Colors.ENDC}\n")
        else:
            print()

# =============================================================================
# Test Functions
# =============================================================================

def make_request(
    method: str,
    endpoint: str,
    headers: Dict[str, str] = None,
    body: Dict = None,
    expected_status: int = 200,
    test_name: str = ""
) -> TestResult:
    """Make HTTP request and return result"""
    
    url = f"{BASE_URL}{endpoint}"
    headers = headers or {"Content-Type": "application/json"}
    
    print_test_start(test_name, method, url)
    print_request(headers, body)
    
    start_time = time.time()
    
    try:
        response = requests.request(
            method=method,
            url=url,
            headers=headers,
            json=body,
            timeout=TIMEOUT
        )
        
        time_ms = (time.time() - start_time) * 1000
        
        try:
            response_body = response.json()
        except:
            response_body = {"raw": response.text[:500]}
        
        print_response(response.status_code, expected_status, response_body, time_ms)
        
        passed = response.status_code == expected_status
        print_test_result(passed)
        
        return TestResult(
            name=test_name,
            passed=passed,
            expected_status=expected_status,
            actual_status=response.status_code,
            request_method=method,
            request_url=url,
            request_headers=headers,
            request_body=body,
            response_body=response_body,
            response_time_ms=time_ms
        )
        
    except Exception as e:
        time_ms = (time.time() - start_time) * 1000
        error_msg = str(e)
        print(f"  {Colors.GRAY}├─ Status: {Colors.RED}ERROR{Colors.ENDC}")
        print_test_result(False, error_msg)
        
        return TestResult(
            name=test_name,
            passed=False,
            expected_status=expected_status,
            actual_status=0,
            request_method=method,
            request_url=url,
            request_headers=headers,
            request_body=body,
            response_body=None,
            response_time_ms=time_ms,
            error=error_msg
        )

def get_auth_headers(token: str) -> Dict[str, str]:
    """Get headers with authorization"""
    return {
        "Content-Type": "application/json",
        "Authorization": f"Bearer {token}"
    }

# =============================================================================
# Test Suite
# =============================================================================

def run_tests():
    """Run all E2E tests"""
    
    results: list[TestResult] = []
    ctx = TestContext()
    
    print(f"\n{Colors.BOLD}{Colors.CYAN}")
    print("╔══════════════════════════════════════════════════════════════════════════════╗")
    print("║                     GRANULA API - DETAILED E2E TESTS                         ║")
    print("║                                                                              ║")
    print(f"║  Started: {datetime.now().strftime('%Y-%m-%d %H:%M:%S'):68}║")
    print(f"║  Base URL: {BASE_URL:67}║")
    print("╚══════════════════════════════════════════════════════════════════════════════╝")
    print(f"{Colors.ENDC}")
    
    # =========================================================================
    # 1. Health Checks
    # =========================================================================
    print_header("1. HEALTH CHECKS")
    
    results.append(make_request(
        "GET", "/health",
        expected_status=200,
        test_name="Health Check"
    ))
    
    results.append(make_request(
        "GET", "/ready",
        expected_status=200,
        test_name="Readiness Check"
    ))
    
    # =========================================================================
    # 2. Authentication - Registration
    # =========================================================================
    print_header("2. AUTHENTICATION - REGISTRATION")
    
    result = make_request(
        "POST", "/api/v1/auth/register",
        body={"email": ctx.email, "password": ctx.password, "name": ctx.name},
        expected_status=201,
        test_name="Register New User"
    )
    results.append(result)
    
    if result.passed and result.response_body:
        data = result.response_body.get("data", {})
        ctx.access_token = data.get("access_token", "")
        ctx.refresh_token = data.get("refresh_token", "")
        ctx.user_id = data.get("user_id", "")
        print(f"  {Colors.CYAN}📝 Saved: user_id={ctx.user_id[:8]}...{Colors.ENDC}\n")
    
    # Test duplicate registration
    results.append(make_request(
        "POST", "/api/v1/auth/register",
        body={"email": ctx.email, "password": ctx.password, "name": ctx.name},
        expected_status=409,
        test_name="Register Duplicate Email (expect 409)"
    ))
    
    # =========================================================================
    # 3. Authentication - Login
    # =========================================================================
    print_header("3. AUTHENTICATION - LOGIN")
    
    result = make_request(
        "POST", "/api/v1/auth/login",
        body={"email": ctx.email, "password": ctx.password},
        expected_status=200,
        test_name="Login with Valid Credentials"
    )
    results.append(result)
    
    if result.passed and result.response_body:
        data = result.response_body.get("data", {})
        ctx.access_token = data.get("access_token", ctx.access_token)
        ctx.refresh_token = data.get("refresh_token", ctx.refresh_token)
    
    results.append(make_request(
        "POST", "/api/v1/auth/login",
        body={"email": ctx.email, "password": "WrongPassword123!"},
        expected_status=401,
        test_name="Login with Wrong Password (expect 401)"
    ))
    
    # =========================================================================
    # 4. Authentication - Token Refresh
    # =========================================================================
    print_header("4. AUTHENTICATION - TOKEN REFRESH")
    
    result = make_request(
        "POST", "/api/v1/auth/refresh",
        body={"refresh_token": ctx.refresh_token},
        expected_status=200,
        test_name="Refresh Token"
    )
    results.append(result)
    
    if result.passed and result.response_body:
        data = result.response_body.get("data", {})
        ctx.access_token = data.get("access_token", ctx.access_token)
        ctx.refresh_token = data.get("refresh_token", ctx.refresh_token)
    
    results.append(make_request(
        "POST", "/api/v1/auth/refresh",
        body={"refresh_token": "invalid_token_here"},
        expected_status=401,
        test_name="Refresh with Invalid Token (expect 401)"
    ))
    
    # =========================================================================
    # 5. User Profile
    # =========================================================================
    print_header("5. USER PROFILE")
    
    results.append(make_request(
        "GET", "/api/v1/users/me",
        headers=get_auth_headers(ctx.access_token),
        expected_status=200,
        test_name="Get User Profile"
    ))
    
    results.append(make_request(
        "PATCH", "/api/v1/users/me",
        headers=get_auth_headers(ctx.access_token),
        body={"name": "Updated E2E User"},
        expected_status=200,
        test_name="Update User Profile"
    ))
    
    results.append(make_request(
        "GET", "/api/v1/users/me",
        expected_status=401,
        test_name="Get Profile Without Auth (expect 401)"
    ))
    
    # =========================================================================
    # 6. Password Change
    # =========================================================================
    print_header("6. PASSWORD CHANGE")
    
    new_password = "NewSecurePass456!"
    result = make_request(
        "PUT", "/api/v1/users/me/password",
        headers=get_auth_headers(ctx.access_token),
        body={"current_password": ctx.password, "new_password": new_password},
        expected_status=200,
        test_name="Change Password"
    )
    results.append(result)
    
    if result.passed:
        ctx.password = new_password
        print(f"  {Colors.CYAN}📝 Password updated{Colors.ENDC}\n")
    
    # Verify new password works
    result = make_request(
        "POST", "/api/v1/auth/login",
        body={"email": ctx.email, "password": ctx.password},
        expected_status=200,
        test_name="Login with New Password"
    )
    results.append(result)
    
    if result.passed and result.response_body:
        data = result.response_body.get("data", {})
        ctx.access_token = data.get("access_token", ctx.access_token)
        ctx.refresh_token = data.get("refresh_token", ctx.refresh_token)
    
    # =========================================================================
    # 7. Notifications
    # =========================================================================
    print_header("7. NOTIFICATIONS")
    
    results.append(make_request(
        "GET", "/api/v1/notifications",
        headers=get_auth_headers(ctx.access_token),
        expected_status=200,
        test_name="List Notifications"
    ))
    
    results.append(make_request(
        "GET", "/api/v1/notifications/count",
        headers=get_auth_headers(ctx.access_token),
        expected_status=200,
        test_name="Get Unread Count"
    ))
    
    results.append(make_request(
        "POST", "/api/v1/notifications/read-all",
        headers=get_auth_headers(ctx.access_token),
        expected_status=200,
        test_name="Mark All as Read"
    ))
    
    # =========================================================================
    # 8. Workspaces (Placeholders)
    # =========================================================================
    print_header("8. WORKSPACES")
    
    results.append(make_request(
        "GET", "/api/v1/workspaces",
        headers=get_auth_headers(ctx.access_token),
        expected_status=200,
        test_name="List Workspaces"
    ))
    
    results.append(make_request(
        "POST", "/api/v1/workspaces",
        headers=get_auth_headers(ctx.access_token),
        body={"name": "Test Workspace", "description": "E2E Test"},
        expected_status=201,
        test_name="Create Workspace"
    ))
    
    # =========================================================================
    # 9. AI Endpoints (Placeholders)
    # =========================================================================
    print_header("9. AI ENDPOINTS")
    
    results.append(make_request(
        "POST", "/api/v1/ai/recognize",
        headers=get_auth_headers(ctx.access_token),
        body={"image_url": "https://example.com/floorplan.jpg"},
        expected_status=200,
        test_name="AI Recognize Floor Plan"
    ))
    
    results.append(make_request(
        "POST", "/api/v1/ai/chat",
        headers=get_auth_headers(ctx.access_token),
        body={"message": "Как сделать перепланировку?"},
        expected_status=200,
        test_name="AI Chat"
    ))
    
    # =========================================================================
    # 10. Logout
    # =========================================================================
    print_header("10. LOGOUT")
    
    results.append(make_request(
        "POST", "/api/v1/auth/logout",
        headers=get_auth_headers(ctx.access_token),
        body={"refresh_token": ctx.refresh_token},
        expected_status=200,
        test_name="Logout"
    ))
    
    # Re-login for final tests
    result = make_request(
        "POST", "/api/v1/auth/login",
        body={"email": ctx.email, "password": ctx.password},
        expected_status=200,
        test_name="Re-login After Logout"
    )
    results.append(result)
    
    if result.passed and result.response_body:
        data = result.response_body.get("data", {})
        ctx.access_token = data.get("access_token", ctx.access_token)
    
    results.append(make_request(
        "POST", "/api/v1/auth/logout-all",
        headers=get_auth_headers(ctx.access_token),
        expected_status=200,
        test_name="Logout from All Devices"
    ))
    
    # =========================================================================
    # 11. Cleanup - Delete Account
    # =========================================================================
    print_header("11. CLEANUP")
    
    # Re-login for deletion
    result = make_request(
        "POST", "/api/v1/auth/login",
        body={"email": ctx.email, "password": ctx.password},
        expected_status=200,
        test_name="Re-login for Account Deletion"
    )
    results.append(result)
    
    if result.passed and result.response_body:
        data = result.response_body.get("data", {})
        ctx.access_token = data.get("access_token", ctx.access_token)
    
    results.append(make_request(
        "DELETE", "/api/v1/users/me",
        headers=get_auth_headers(ctx.access_token),
        expected_status=200,
        test_name="Delete User Account"
    ))
    
    # =========================================================================
    # Summary
    # =========================================================================
    print_summary(results)
    
    return results

def print_summary(results: list[TestResult]):
    """Print test summary"""
    
    passed = sum(1 for r in results if r.passed)
    failed = sum(1 for r in results if not r.passed)
    total = len(results)
    success_rate = (passed / total * 100) if total > 0 else 0
    avg_time = sum(r.response_time_ms for r in results) / total if total > 0 else 0
    
    print(f"\n{Colors.BOLD}{Colors.CYAN}")
    print("╔══════════════════════════════════════════════════════════════════════════════╗")
    print("║                              TEST SUMMARY                                    ║")
    print("╠══════════════════════════════════════════════════════════════════════════════╣")
    print(f"║  Total Tests: {total:63}║")
    print(f"║  {Colors.GREEN}Passed: {passed:66}{Colors.CYAN}║")
    print(f"║  {Colors.RED}Failed: {failed:66}{Colors.CYAN}║")
    print(f"║  Success Rate: {success_rate:.1f}%{' ' * (60 - len(f'{success_rate:.1f}'))}║")
    print(f"║  Avg Response Time: {avg_time:.2f}ms{' ' * (55 - len(f'{avg_time:.2f}'))}║")
    print("╠══════════════════════════════════════════════════════════════════════════════╣")
    
    if failed > 0:
        print("║  Failed Tests:                                                               ║")
        for r in results:
            if not r.passed:
                name = r.name[:50]
                error = (r.error or f"Got {r.actual_status}")[:25]
                print(f"║    {Colors.RED}✗ {name}{' ' * (50 - len(name))} - {error}{Colors.CYAN}{' ' * (20 - len(error))}║")
    else:
        print(f"║  {Colors.GREEN}All tests passed! 🎉{Colors.CYAN}{' ' * 56}║")
    
    print("╚══════════════════════════════════════════════════════════════════════════════╝")
    print(f"{Colors.ENDC}")

# =============================================================================
# Main
# =============================================================================

if __name__ == "__main__":
    try:
        results = run_tests()
        failed = sum(1 for r in results if not r.passed)
        exit(0 if failed == 0 else 1)
    except KeyboardInterrupt:
        print(f"\n{Colors.YELLOW}Tests interrupted by user{Colors.ENDC}")
        exit(1)
    except Exception as e:
        print(f"\n{Colors.RED}Fatal error: {e}{Colors.ENDC}")
        exit(1)


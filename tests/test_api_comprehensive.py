#!/usr/bin/env python3
"""
Comprehensive API Test Suite for Granula API
=============================================
Tests all API endpoints step-by-step with detailed logging.

Usage:
    python test_api_comprehensive.py

Requirements:
    pip install requests colorama
"""

import requests
import json
import base64
import time
import logging
import sys
from pathlib import Path
from datetime import datetime
from typing import Optional, Dict, Any, Tuple

# Try to import colorama for colored output
try:
    from colorama import init, Fore, Style
    init()
    HAS_COLORS = True
except ImportError:
    HAS_COLORS = False
    class Fore:
        GREEN = RED = YELLOW = CYAN = MAGENTA = BLUE = WHITE = RESET = ""
    class Style:
        BRIGHT = RESET_ALL = ""


# =============================================================================
# Configuration
# =============================================================================
API_BASE_URL = "https://api.granula.raitokyokai.tech/api/v1"
IMAGES_DIR = Path(__file__).parent.parent / "Квартиры"
LOG_FILE = Path(__file__).parent / f"test_results_{datetime.now().strftime('%Y%m%d_%H%M%S')}.log"

# Test user credentials (will be created if not exists)
# Using timestamp to ensure unique email for each test run
import time as _time
_timestamp = int(_time.time())
TEST_USER_EMAIL = f"test_api_{_timestamp}@granula.ru"
TEST_USER_PASSWORD = "TestPassword123!"
TEST_USER_NAME = "API Test User"


# =============================================================================
# Logging Setup
# =============================================================================
def setup_logging():
    """Setup logging to both file and console."""
    # Create formatter
    formatter = logging.Formatter(
        '%(asctime)s | %(levelname)-8s | %(message)s',
        datefmt='%Y-%m-%d %H:%M:%S'
    )
    
    # File handler
    file_handler = logging.FileHandler(LOG_FILE, encoding='utf-8')
    file_handler.setFormatter(formatter)
    file_handler.setLevel(logging.DEBUG)
    
    # Console handler
    console_handler = logging.StreamHandler(sys.stdout)
    console_handler.setFormatter(formatter)
    console_handler.setLevel(logging.INFO)
    
    # Root logger
    logger = logging.getLogger()
    logger.setLevel(logging.DEBUG)
    logger.addHandler(file_handler)
    logger.addHandler(console_handler)
    
    return logger


logger = setup_logging()


# =============================================================================
# Helper Functions
# =============================================================================
def print_header(text: str):
    """Print a section header."""
    separator = "=" * 70
    logger.info("")
    logger.info(separator)
    logger.info(f"  {text}")
    logger.info(separator)


def print_step(step_num: int, text: str):
    """Print a test step."""
    if HAS_COLORS:
        print(f"\n{Fore.CYAN}[Step {step_num}]{Style.RESET_ALL} {text}")
    logger.info(f"[Step {step_num}] {text}")


def print_success(text: str):
    """Print success message."""
    if HAS_COLORS:
        print(f"  {Fore.GREEN}✓ {text}{Style.RESET_ALL}")
    logger.info(f"✓ SUCCESS: {text}")


def print_error(text: str):
    """Print error message."""
    if HAS_COLORS:
        print(f"  {Fore.RED}✗ {text}{Style.RESET_ALL}")
    logger.error(f"✗ FAILED: {text}")


def print_info(text: str):
    """Print info message."""
    if HAS_COLORS:
        print(f"  {Fore.YELLOW}→ {text}{Style.RESET_ALL}")
    logger.info(f"→ {text}")


def print_response(response: requests.Response, show_body: bool = True):
    """Print response details."""
    logger.debug(f"Response Status: {response.status_code}")
    logger.debug(f"Response Headers: {dict(response.headers)}")
    
    if show_body:
        try:
            body = response.json()
            logger.debug(f"Response Body: {json.dumps(body, indent=2, ensure_ascii=False)}")
            return body
        except:
            logger.debug(f"Response Text: {response.text[:500]}")
            return None
    return None


def load_image_as_base64(image_path: Path) -> Tuple[str, str]:
    """Load image and convert to base64."""
    with open(image_path, 'rb') as f:
        image_data = f.read()
    
    # Determine MIME type
    suffix = image_path.suffix.lower()
    mime_types = {
        '.jpg': 'image/jpeg',
        '.jpeg': 'image/jpeg',
        '.png': 'image/png',
        '.gif': 'image/gif',
        '.webp': 'image/webp'
    }
    mime_type = mime_types.get(suffix, 'image/jpeg')
    
    base64_data = base64.b64encode(image_data).decode('utf-8')
    return base64_data, mime_type


# =============================================================================
# API Client Class
# =============================================================================
class GranulaAPIClient:
    """Client for Granula API testing."""
    
    def __init__(self, base_url: str):
        self.base_url = base_url
        self.session = requests.Session()
        self.access_token: Optional[str] = None
        self.refresh_token: Optional[str] = None
        self.user_id: Optional[str] = None
        
        # Test data storage
        self.workspace_id: Optional[str] = None
        self.scene_id: Optional[str] = None
        self.branch_id: Optional[str] = None
        self.floor_plan_id: Optional[str] = None
        self.recognition_job_id: Optional[str] = None
        self.generation_job_id: Optional[str] = None
        self.chat_message_id: Optional[str] = None
        self.chat_context_id: Optional[str] = None
        
    def _headers(self, with_auth: bool = True) -> Dict[str, str]:
        """Get request headers."""
        headers = {
            "Content-Type": "application/json",
            "Accept": "application/json"
        }
        if with_auth and self.access_token:
            headers["Authorization"] = f"Bearer {self.access_token}"
        return headers
    
    def _request(self, method: str, endpoint: str, **kwargs) -> requests.Response:
        """Make API request."""
        url = f"{self.base_url}{endpoint}"
        logger.debug(f"Request: {method} {url}")
        if 'json' in kwargs:
            logger.debug(f"Request Body: {json.dumps(kwargs['json'], indent=2, ensure_ascii=False)}")
        
        response = self.session.request(method, url, **kwargs)
        return response

    # =========================================================================
    # Auth Endpoints
    # =========================================================================
    def register(self, email: str, password: str, name: str) -> Tuple[bool, Dict]:
        """Register a new user."""
        response = self._request(
            "POST", "/auth/register",
            headers=self._headers(with_auth=False),
            json={"email": email, "password": password, "name": name}
        )
        body = print_response(response)
        
        if response.status_code == 201:
            if body and "data" in body:
                self.access_token = body["data"].get("access_token")
                self.refresh_token = body["data"].get("refresh_token")
                self.user_id = body["data"].get("user_id")
            return True, body
        return False, body
    
    def login(self, email: str, password: str) -> Tuple[bool, Dict]:
        """Login user."""
        response = self._request(
            "POST", "/auth/login",
            headers=self._headers(with_auth=False),
            json={"email": email, "password": password}
        )
        body = print_response(response)
        
        if response.status_code == 200:
            if body and "data" in body:
                self.access_token = body["data"].get("access_token")
                self.refresh_token = body["data"].get("refresh_token")
                self.user_id = body["data"].get("user_id")
            return True, body
        return False, body
    
    def refresh_tokens(self) -> Tuple[bool, Dict]:
        """Refresh access token."""
        response = self._request(
            "POST", "/auth/refresh",
            headers=self._headers(with_auth=False),
            json={"refresh_token": self.refresh_token}
        )
        body = print_response(response)
        
        if response.status_code == 200:
            if body and "data" in body:
                self.access_token = body["data"].get("access_token")
                self.refresh_token = body["data"].get("refresh_token")
            return True, body
        return False, body
    
    def logout(self) -> Tuple[bool, Dict]:
        """Logout user."""
        response = self._request(
            "POST", "/auth/logout",
            headers=self._headers(),
            json={"refresh_token": self.refresh_token}
        )
        body = print_response(response)
        return response.status_code == 200, body
    
    def logout_all(self) -> Tuple[bool, Dict]:
        """Logout from all sessions."""
        response = self._request(
            "POST", "/auth/logout-all",
            headers=self._headers()
        )
        body = print_response(response)
        return response.status_code == 200, body

    # =========================================================================
    # User Endpoints
    # =========================================================================
    def get_me(self) -> Tuple[bool, Dict]:
        """Get current user profile."""
        response = self._request(
            "GET", "/users/me",
            headers=self._headers()
        )
        body = print_response(response)
        return response.status_code == 200, body
    
    def update_profile(self, name: str = None, phone: str = None) -> Tuple[bool, Dict]:
        """Update user profile."""
        data = {}
        if name:
            data["name"] = name
        if phone:
            data["phone"] = phone
            
        response = self._request(
            "PATCH", "/users/me",
            headers=self._headers(),
            json=data
        )
        body = print_response(response)
        return response.status_code == 200, body

    # =========================================================================
    # Workspace Endpoints
    # =========================================================================
    def create_workspace(self, name: str, description: str = "") -> Tuple[bool, Dict]:
        """Create a new workspace."""
        response = self._request(
            "POST", "/workspaces",
            headers=self._headers(),
            json={"name": name, "description": description}
        )
        body = print_response(response)
        
        if response.status_code == 201 and body and "data" in body:
            self.workspace_id = body["data"].get("id")
        return response.status_code == 201, body
    
    def list_workspaces(self) -> Tuple[bool, Dict]:
        """List user workspaces."""
        response = self._request(
            "GET", "/workspaces",
            headers=self._headers()
        )
        body = print_response(response)
        return response.status_code == 200, body
    
    def get_workspace(self, workspace_id: str) -> Tuple[bool, Dict]:
        """Get workspace by ID."""
        response = self._request(
            "GET", f"/workspaces/{workspace_id}",
            headers=self._headers()
        )
        body = print_response(response)
        return response.status_code == 200, body

    # =========================================================================
    # Floor Plan Endpoints
    # =========================================================================
    def upload_floor_plan(self, workspace_id: str, image_path: Path, name: str = "Test Plan") -> Tuple[bool, Dict]:
        """Upload a floor plan image."""
        with open(image_path, 'rb') as f:
            files = {
                'file': (image_path.name, f, 'image/jpeg')
            }
            data = {
                'workspace_id': workspace_id,
                'name': name
            }
            
            headers = {"Authorization": f"Bearer {self.access_token}"}
            response = self._request(
                "POST", "/floor-plans",
                headers=headers,
                files=files,
                data=data
            )
        
        body = print_response(response)
        
        if response.status_code in [200, 201] and body and "data" in body:
            self.floor_plan_id = body["data"].get("id")
        return response.status_code in [200, 201], body
    
    def list_floor_plans(self, workspace_id: str = None) -> Tuple[bool, Dict]:
        """List floor plans."""
        params = {}
        if workspace_id:
            params["workspace_id"] = workspace_id
            
        response = self._request(
            "GET", "/floor-plans",
            headers=self._headers(),
            params=params
        )
        body = print_response(response)
        return response.status_code == 200, body

    # =========================================================================
    # AI Recognition Endpoints
    # =========================================================================
    def recognize_floor_plan(self, image_path: Path, floor_plan_id: str = None) -> Tuple[bool, Dict]:
        """Start floor plan recognition."""
        base64_data, mime_type = load_image_as_base64(image_path)
        
        response = self._request(
            "POST", "/ai/recognize",
            headers=self._headers(),
            json={
                "floor_plan_id": floor_plan_id or "test-recognition",
                "image_base64": base64_data,
                "image_type": mime_type,
                "options": {
                    "detect_load_bearing": True,
                    "detect_wet_zones": True,
                    "detect_furniture": False
                }
            }
        )
        body = print_response(response)
        
        if response.status_code == 200 and body and "data" in body:
            self.recognition_job_id = body["data"].get("job_id")
        return response.status_code == 200, body
    
    def get_recognition_status(self, job_id: str) -> Tuple[bool, Dict]:
        """Get recognition job status."""
        response = self._request(
            "GET", f"/ai/recognize/{job_id}/status",
            headers=self._headers()
        )
        body = print_response(response)
        return response.status_code == 200, body

    # =========================================================================
    # AI Chat Endpoints
    # =========================================================================
    def send_chat_message(self, message: str, scene_id: str = None, branch_id: str = None, context_id: str = None) -> Tuple[bool, Dict]:
        """Send a chat message to AI."""
        data = {"message": message}
        if scene_id:
            data["scene_id"] = scene_id
        if branch_id:
            data["branch_id"] = branch_id
        if context_id:
            data["context_id"] = context_id
            
        response = self._request(
            "POST", "/ai/chat",
            headers=self._headers(),
            json=data
        )
        body = print_response(response)
        
        if response.status_code == 200 and body and "data" in body:
            self.chat_message_id = body["data"].get("message_id")
            self.chat_context_id = body["data"].get("context_id")
        return response.status_code == 200, body
    
    def get_chat_history(self, scene_id: str = None, limit: int = 50) -> Tuple[bool, Dict]:
        """Get chat history."""
        params = {"limit": limit}
        if scene_id:
            params["scene_id"] = scene_id
            
        response = self._request(
            "GET", "/ai/chat/history",
            headers=self._headers(),
            params=params
        )
        body = print_response(response)
        return response.status_code == 200, body
    
    def clear_chat_history(self, scene_id: str = None) -> Tuple[bool, Dict]:
        """Clear chat history."""
        data = {}
        if scene_id:
            data["scene_id"] = scene_id
            
        response = self._request(
            "DELETE", "/ai/chat/history",
            headers=self._headers(),
            json=data
        )
        body = print_response(response)
        return response.status_code == 200, body

    # =========================================================================
    # AI Generation Endpoints
    # =========================================================================
    def generate_variants(self, scene_id: str, prompt: str, variants_count: int = 3) -> Tuple[bool, Dict]:
        """Generate layout variants."""
        response = self._request(
            "POST", "/ai/generate",
            headers=self._headers(),
            json={
                "scene_id": scene_id,
                "prompt": prompt,
                "variants_count": variants_count,
                "preserve_load_bearing": True,
                "check_compliance": True
            }
        )
        body = print_response(response)
        
        if response.status_code == 200 and body and "data" in body:
            self.generation_job_id = body["data"].get("job_id")
        return response.status_code == 200, body
    
    def get_generation_status(self, job_id: str) -> Tuple[bool, Dict]:
        """Get generation job status."""
        response = self._request(
            "GET", f"/ai/generate/{job_id}/status",
            headers=self._headers()
        )
        body = print_response(response)
        return response.status_code == 200, body

    # =========================================================================
    # AI Context Endpoints
    # =========================================================================
    def get_ai_context(self, scene_id: str, branch_id: str = None) -> Tuple[bool, Dict]:
        """Get AI context for a scene."""
        params = {"scene_id": scene_id}
        if branch_id:
            params["branch_id"] = branch_id
            
        response = self._request(
            "GET", "/ai/context",
            headers=self._headers(),
            params=params
        )
        body = print_response(response)
        return response.status_code == 200, body
    
    def update_ai_context(self, scene_id: str, branch_id: str = None, force: bool = False) -> Tuple[bool, Dict]:
        """Update AI context."""
        data = {"scene_id": scene_id, "force": force}
        if branch_id:
            data["branch_id"] = branch_id
            
        response = self._request(
            "POST", "/ai/context",
            headers=self._headers(),
            json=data
        )
        body = print_response(response)
        return response.status_code == 200, body

    # =========================================================================
    # Compliance Endpoints
    # =========================================================================
    def check_compliance(self, scene_id: str, branch_id: str = None) -> Tuple[bool, Dict]:
        """Check compliance for a scene."""
        data = {"scene_id": scene_id}
        if branch_id:
            data["branch_id"] = branch_id
            
        response = self._request(
            "POST", "/compliance/check",
            headers=self._headers(),
            json=data
        )
        body = print_response(response)
        return response.status_code == 200, body

    # =========================================================================
    # Branch Endpoints
    # =========================================================================
    def list_branches(self, scene_id: str) -> Tuple[bool, Dict]:
        """List branches for a scene."""
        response = self._request(
            "GET", f"/branches",
            headers=self._headers(),
            params={"scene_id": scene_id}
        )
        body = print_response(response)
        return response.status_code == 200, body
    
    def create_branch(self, scene_id: str, name: str, source_branch_id: str = None) -> Tuple[bool, Dict]:
        """Create a new branch."""
        data = {"scene_id": scene_id, "name": name}
        if source_branch_id:
            data["source_branch_id"] = source_branch_id
            
        response = self._request(
            "POST", "/branches",
            headers=self._headers(),
            json=data
        )
        body = print_response(response)
        
        if response.status_code == 201 and body and "data" in body:
            self.branch_id = body["data"].get("id")
        return response.status_code == 201, body

    # =========================================================================
    # Scene Endpoints
    # =========================================================================
    def get_scene(self, scene_id: str) -> Tuple[bool, Dict]:
        """Get scene by ID."""
        response = self._request(
            "GET", f"/scenes/{scene_id}",
            headers=self._headers()
        )
        body = print_response(response)
        return response.status_code == 200, body


# =============================================================================
# Test Scenarios
# =============================================================================
class APITestSuite:
    """Test suite for Granula API."""
    
    def __init__(self):
        self.client = GranulaAPIClient(API_BASE_URL)
        self.results = {
            "passed": 0,
            "failed": 0,
            "skipped": 0,
            "tests": []
        }
        
    def record_result(self, test_name: str, passed: bool, details: str = ""):
        """Record test result."""
        status = "PASSED" if passed else "FAILED"
        self.results["tests"].append({
            "name": test_name,
            "status": status,
            "details": details
        })
        if passed:
            self.results["passed"] += 1
        else:
            self.results["failed"] += 1
            
    def run_all_tests(self):
        """Run all test scenarios."""
        print_header("GRANULA API COMPREHENSIVE TEST SUITE")
        logger.info(f"API URL: {API_BASE_URL}")
        logger.info(f"Log file: {LOG_FILE}")
        logger.info(f"Images dir: {IMAGES_DIR}")
        
        try:
            # Phase 1: Authentication
            self.test_auth_flow()
            
            # Phase 2: User Profile
            self.test_user_profile()
            
            # Phase 3: Workspaces
            self.test_workspaces()
            
            # Phase 4: Floor Plans
            self.test_floor_plans()
            
            # Phase 5: AI Recognition
            self.test_ai_recognition()
            
            # Phase 6: AI Chat
            self.test_ai_chat()
            
            # Phase 7: AI Context
            self.test_ai_context()
            
            # Phase 8: AI Generation (if scene exists)
            self.test_ai_generation()
            
        except Exception as e:
            logger.exception(f"Test suite failed with exception: {e}")
            
        finally:
            self.print_summary()
            
    def test_auth_flow(self):
        """Test authentication flow."""
        print_header("PHASE 1: AUTHENTICATION")
        
        # Step 1: Try to login first (user might exist)
        print_step(1, "Attempting login with test user")
        success, body = self.client.login(TEST_USER_EMAIL, TEST_USER_PASSWORD)
        
        if success:
            print_success(f"Login successful! User ID: {self.client.user_id}")
            self.record_result("Login", True, f"User ID: {self.client.user_id}")
        else:
            print_info("Login failed, trying to register new user")
            
            # Step 2: Register new user
            print_step(2, "Registering new test user")
            success, body = self.client.register(TEST_USER_EMAIL, TEST_USER_PASSWORD, TEST_USER_NAME)
            
            if success:
                print_success(f"Registration successful! User ID: {self.client.user_id}")
                self.record_result("Register", True, f"User ID: {self.client.user_id}")
            else:
                print_error(f"Registration failed: {body}")
                self.record_result("Register", False, str(body))
                return
        
        # Step 3: Test token refresh
        print_step(3, "Testing token refresh")
        if self.client.refresh_token:
            success, body = self.client.refresh_tokens()
            if success:
                print_success("Token refresh successful")
                self.record_result("Token Refresh", True)
            else:
                print_error(f"Token refresh failed: {body}")
                self.record_result("Token Refresh", False, str(body))
        else:
            print_info("No refresh token available, skipping")
            self.results["skipped"] += 1
            
    def test_user_profile(self):
        """Test user profile endpoints."""
        print_header("PHASE 2: USER PROFILE")
        
        if not self.client.access_token:
            print_error("No access token, skipping user tests")
            self.results["skipped"] += 1
            return
            
        # Step 1: Get current profile
        print_step(1, "Getting current user profile")
        success, body = self.client.get_me()
        
        if success:
            user_data = body.get("data", {})
            print_success(f"Got profile: {user_data.get('email', 'N/A')}")
            print_info(f"  Name: {user_data.get('name', 'N/A')}")
            print_info(f"  Role: {user_data.get('role', 'N/A')}")
            self.record_result("Get Profile", True)
        else:
            print_error(f"Failed to get profile: {body}")
            self.record_result("Get Profile", False, str(body))
            
        # Step 2: Update profile
        print_step(2, "Updating user profile")
        new_name = f"Test User {datetime.now().strftime('%H%M%S')}"
        success, body = self.client.update_profile(name=new_name)
        
        if success:
            print_success(f"Profile updated with name: {new_name}")
            self.record_result("Update Profile", True)
        else:
            print_error(f"Failed to update profile: {body}")
            self.record_result("Update Profile", False, str(body))
            
    def test_workspaces(self):
        """Test workspace endpoints."""
        print_header("PHASE 3: WORKSPACES")
        
        if not self.client.access_token:
            print_error("No access token, skipping workspace tests")
            self.results["skipped"] += 1
            return
            
        # Step 1: List existing workspaces
        print_step(1, "Listing existing workspaces")
        success, body = self.client.list_workspaces()
        
        if success:
            workspaces = body.get("data", {}).get("workspaces", [])
            print_success(f"Found {len(workspaces)} workspace(s)")
            self.record_result("List Workspaces", True)
            
            # Use existing workspace if available
            if workspaces:
                self.client.workspace_id = workspaces[0].get("id")
                print_info(f"Using existing workspace: {self.client.workspace_id}")
        else:
            print_error(f"Failed to list workspaces: {body}")
            self.record_result("List Workspaces", False, str(body))
            
        # Step 2: Create new workspace
        print_step(2, "Creating new workspace")
        ws_name = f"Test Workspace {datetime.now().strftime('%H%M%S')}"
        success, body = self.client.create_workspace(ws_name, "Created by API test")
        
        if success:
            print_success(f"Workspace created: {self.client.workspace_id}")
            self.record_result("Create Workspace", True)
        else:
            print_error(f"Failed to create workspace: {body}")
            self.record_result("Create Workspace", False, str(body))
            
        # Step 3: Get workspace details
        if self.client.workspace_id:
            print_step(3, "Getting workspace details")
            success, body = self.client.get_workspace(self.client.workspace_id)
            
            if success:
                ws = body.get("data", {})
                print_success(f"Got workspace: {ws.get('name', 'N/A')}")
                self.record_result("Get Workspace", True)
            else:
                print_error(f"Failed to get workspace: {body}")
                self.record_result("Get Workspace", False, str(body))
                
    def test_floor_plans(self):
        """Test floor plan endpoints."""
        print_header("PHASE 4: FLOOR PLANS")
        
        if not self.client.access_token or not self.client.workspace_id:
            print_error("No access token or workspace, skipping floor plan tests")
            self.results["skipped"] += 1
            return
            
        # Find a test image
        if not IMAGES_DIR.exists():
            print_error(f"Images directory not found: {IMAGES_DIR}")
            self.results["skipped"] += 1
            return
            
        images = list(IMAGES_DIR.glob("*.jpg")) + list(IMAGES_DIR.glob("*.png"))
        if not images:
            print_error("No images found for testing")
            self.results["skipped"] += 1
            return
            
        test_image = images[0]
        print_info(f"Using test image: {test_image.name}")
        
        # Step 1: Upload floor plan
        print_step(1, "Uploading floor plan")
        success, body = self.client.upload_floor_plan(
            self.client.workspace_id,
            test_image,
            f"Test Plan {datetime.now().strftime('%H%M%S')}"
        )
        
        if success:
            print_success(f"Floor plan uploaded: {self.client.floor_plan_id}")
            self.record_result("Upload Floor Plan", True)
        else:
            print_error(f"Failed to upload floor plan: {body}")
            self.record_result("Upload Floor Plan", False, str(body))
            
        # Step 2: List floor plans
        print_step(2, "Listing floor plans")
        success, body = self.client.list_floor_plans(self.client.workspace_id)
        
        if success:
            plans = body.get("data", {}).get("floor_plans", [])
            print_success(f"Found {len(plans)} floor plan(s)")
            self.record_result("List Floor Plans", True)
        else:
            print_error(f"Failed to list floor plans: {body}")
            self.record_result("List Floor Plans", False, str(body))
            
    def test_ai_recognition(self):
        """Test AI recognition endpoints."""
        print_header("PHASE 5: AI RECOGNITION")
        
        if not self.client.access_token:
            print_error("No access token, skipping AI recognition tests")
            self.results["skipped"] += 1
            return
            
        # Find a test image
        images = list(IMAGES_DIR.glob("*.jpg")) + list(IMAGES_DIR.glob("*.png"))
        if not images:
            print_error("No images found for testing")
            self.results["skipped"] += 1
            return
            
        test_image = images[0]
        print_info(f"Using test image: {test_image.name} ({test_image.stat().st_size / 1024:.1f} KB)")
        
        # Step 1: Start recognition
        print_step(1, "Starting floor plan recognition")
        success, body = self.client.recognize_floor_plan(test_image, self.client.floor_plan_id)
        
        if success:
            print_success(f"Recognition started! Job ID: {self.client.recognition_job_id}")
            print_info(f"Status: {body.get('data', {}).get('status', 'unknown')}")
            self.record_result("Start Recognition", True)
        else:
            print_error(f"Failed to start recognition: {body}")
            self.record_result("Start Recognition", False, str(body))
            return
            
        # Step 2: Poll for completion
        print_step(2, "Polling recognition status")
        if self.client.recognition_job_id:
            max_attempts = 30
            for attempt in range(max_attempts):
                time.sleep(2)
                success, body = self.client.get_recognition_status(self.client.recognition_job_id)
                
                if success:
                    status = body.get("data", {}).get("status", "unknown")
                    progress = body.get("data", {}).get("progress", 0)
                    print_info(f"Attempt {attempt + 1}: Status={status}, Progress={progress}%")
                    
                    if status == "completed":
                        print_success("Recognition completed!")
                        scene_data = body.get("data", {}).get("scene", {})
                        if scene_data:
                            print_info(f"  Walls: {len(scene_data.get('walls', []))}")
                            print_info(f"  Rooms: {len(scene_data.get('rooms', []))}")
                            print_info(f"  Openings: {len(scene_data.get('openings', []))}")
                        self.record_result("Recognition Completion", True)
                        break
                    elif status == "failed":
                        print_error(f"Recognition failed: {body.get('data', {}).get('error', 'unknown')}")
                        self.record_result("Recognition Completion", False, "Job failed")
                        break
                else:
                    print_error(f"Failed to get status: {body}")
                    
            else:
                print_error("Recognition timed out")
                self.record_result("Recognition Completion", False, "Timeout")
                
    def test_ai_chat(self):
        """Test AI chat endpoints."""
        print_header("PHASE 6: AI CHAT")
        
        if not self.client.access_token:
            print_error("No access token, skipping AI chat tests")
            self.results["skipped"] += 1
            return
            
        # Step 1: Send first message
        print_step(1, "Sending chat message to AI")
        message = "Привет! Расскажи, какие основные правила перепланировки квартир в России?"
        success, body = self.client.send_chat_message(message)
        
        if success:
            response_text = body.get("data", {}).get("response", "")
            print_success("Chat message sent and response received!")
            print_info(f"  Message ID: {self.client.chat_message_id}")
            print_info(f"  Context ID: {self.client.chat_context_id}")
            print_info(f"  Response length: {len(response_text)} chars")
            print_info(f"  Response preview: {response_text[:200]}...")
            
            token_usage = body.get("data", {}).get("token_usage", {})
            if token_usage:
                print_info(f"  Tokens: prompt={token_usage.get('prompt_tokens', 0)}, completion={token_usage.get('completion_tokens', 0)}")
                
            self.record_result("Chat Message", True)
        else:
            print_error(f"Failed to send chat message: {body}")
            self.record_result("Chat Message", False, str(body))
            
        # Step 2: Send follow-up message (continuing context)
        print_step(2, "Sending follow-up message")
        if self.client.chat_context_id:
            follow_up = "Можно ли снести стену между кухней и гостиной?"
            success, body = self.client.send_chat_message(follow_up, context_id=self.client.chat_context_id)
            
            if success:
                response_text = body.get("data", {}).get("response", "")
                print_success("Follow-up response received!")
                print_info(f"  Response preview: {response_text[:200]}...")
                
                # Check for actions
                actions = body.get("data", {}).get("actions", [])
                if actions:
                    print_info(f"  AI suggested {len(actions)} action(s)")
                    
                self.record_result("Chat Follow-up", True)
            else:
                print_error(f"Failed to send follow-up: {body}")
                self.record_result("Chat Follow-up", False, str(body))
        else:
            print_info("No context ID, skipping follow-up")
            
        # Step 3: Get chat history
        print_step(3, "Getting chat history")
        success, body = self.client.get_chat_history()
        
        if success:
            messages = body.get("data", {}).get("messages", [])
            print_success(f"Retrieved {len(messages)} message(s) from history")
            self.record_result("Chat History", True)
        else:
            print_error(f"Failed to get chat history: {body}")
            self.record_result("Chat History", False, str(body))
            
    def test_ai_context(self):
        """Test AI context endpoints."""
        print_header("PHASE 7: AI CONTEXT")
        
        if not self.client.access_token:
            print_error("No access token, skipping AI context tests")
            self.results["skipped"] += 1
            return
            
        # Use a fake scene ID for testing (real scene would come from recognition)
        test_scene_id = "test-scene-123"
        
        # Step 1: Get AI context
        print_step(1, "Getting AI context")
        success, body = self.client.get_ai_context(test_scene_id)
        
        if success:
            context_data = body.get("data", {})
            print_success("AI context retrieved!")
            print_info(f"  Context ID: {context_data.get('context_id', 'N/A')}")
            print_info(f"  Context size: {context_data.get('context_size', 0)} tokens")
            self.record_result("Get AI Context", True)
        else:
            print_error(f"Failed to get AI context: {body}")
            self.record_result("Get AI Context", False, str(body))
            
        # Step 2: Update AI context
        print_step(2, "Updating AI context")
        success, body = self.client.update_ai_context(test_scene_id, force=True)
        
        if success:
            context_data = body.get("data", {})
            print_success("AI context updated!")
            print_info(f"  Updated: {context_data.get('updated', False)}")
            print_info(f"  New size: {context_data.get('context_size', 0)} tokens")
            self.record_result("Update AI Context", True)
        else:
            print_error(f"Failed to update AI context: {body}")
            self.record_result("Update AI Context", False, str(body))
            
    def test_ai_generation(self):
        """Test AI generation endpoints."""
        print_header("PHASE 8: AI GENERATION")
        
        if not self.client.access_token:
            print_error("No access token, skipping AI generation tests")
            self.results["skipped"] += 1
            return
            
        # Use a test scene ID
        test_scene_id = "test-scene-gen-123"
        
        # Step 1: Start generation
        print_step(1, "Starting variant generation")
        prompt = "Предложи варианты объединения кухни с гостиной"
        success, body = self.client.generate_variants(test_scene_id, prompt, variants_count=2)
        
        if success:
            print_success(f"Generation started! Job ID: {self.client.generation_job_id}")
            print_info(f"Status: {body.get('data', {}).get('status', 'unknown')}")
            self.record_result("Start Generation", True)
        else:
            print_error(f"Failed to start generation: {body}")
            self.record_result("Start Generation", False, str(body))
            return
            
        # Step 2: Poll for completion
        print_step(2, "Polling generation status")
        if self.client.generation_job_id:
            max_attempts = 30
            for attempt in range(max_attempts):
                time.sleep(2)
                success, body = self.client.get_generation_status(self.client.generation_job_id)
                
                if success:
                    status = body.get("data", {}).get("status", "unknown")
                    progress = body.get("data", {}).get("progress", 0)
                    print_info(f"Attempt {attempt + 1}: Status={status}, Progress={progress}%")
                    
                    if status == "completed":
                        print_success("Generation completed!")
                        variants = body.get("data", {}).get("variants", [])
                        print_info(f"  Generated {len(variants)} variant(s)")
                        for i, v in enumerate(variants):
                            print_info(f"    [{i+1}] {v.get('name', 'N/A')}: {v.get('description', 'N/A')[:50]}...")
                        self.record_result("Generation Completion", True)
                        break
                    elif status == "failed":
                        print_error(f"Generation failed: {body.get('data', {}).get('error', 'unknown')}")
                        self.record_result("Generation Completion", False, "Job failed")
                        break
                else:
                    print_error(f"Failed to get status: {body}")
                    
            else:
                print_error("Generation timed out")
                self.record_result("Generation Completion", False, "Timeout")
                
    def print_summary(self):
        """Print test summary."""
        print_header("TEST SUMMARY")
        
        total = self.results["passed"] + self.results["failed"] + self.results["skipped"]
        
        logger.info(f"Total tests: {total}")
        logger.info(f"  Passed:  {self.results['passed']}")
        logger.info(f"  Failed:  {self.results['failed']}")
        logger.info(f"  Skipped: {self.results['skipped']}")
        
        if HAS_COLORS:
            print(f"\n{Style.BRIGHT}Results:{Style.RESET_ALL}")
            print(f"  {Fore.GREEN}Passed:  {self.results['passed']}{Style.RESET_ALL}")
            print(f"  {Fore.RED}Failed:  {self.results['failed']}{Style.RESET_ALL}")
            print(f"  {Fore.YELLOW}Skipped: {self.results['skipped']}{Style.RESET_ALL}")
            
        if self.results["failed"] > 0:
            logger.info("\nFailed tests:")
            for test in self.results["tests"]:
                if test["status"] == "FAILED":
                    logger.info(f"  - {test['name']}: {test['details']}")
                    
        logger.info(f"\nFull log saved to: {LOG_FILE}")


# =============================================================================
# Main Entry Point
# =============================================================================
if __name__ == "__main__":
    print(f"\n{'='*70}")
    print("  GRANULA API COMPREHENSIVE TEST SUITE")
    print(f"{'='*70}")
    print(f"  Started at: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
    print(f"  API URL: {API_BASE_URL}")
    print(f"{'='*70}\n")
    
    suite = APITestSuite()
    suite.run_all_tests()


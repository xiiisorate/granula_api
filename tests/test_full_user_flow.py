#!/usr/bin/env python3
"""
Granula API - Full User Flow Test
==================================
Полный тест пользовательского сценария от регистрации до заявки эксперту.

Тестирует ВСЕ эндпоинты из Swagger в правильной последовательности.

Usage:
    python test_full_user_flow.py

Requirements:
    pip install requests colorama
"""

import requests
import json
import base64
import time
import logging
import sys
import os
from pathlib import Path
from datetime import datetime
from typing import Optional, Dict, Any, Tuple, List

# Configuration
API_BASE_URL = "https://api.granula.raitokyokai.tech/api/v1"
IMAGES_DIR = Path(__file__).parent.parent / "Квартиры"
LOG_FILE = Path(__file__).parent / f"full_flow_{datetime.now().strftime('%Y%m%d_%H%M%S')}.log"

# Unique test user for this run
TIMESTAMP = int(time.time())
TEST_EMAIL = f"fulltest_{TIMESTAMP}@granula.ru"
TEST_PASSWORD = "SecurePassword123!"
TEST_NAME = "Full Test User"


class Colors:
    """ANSI color codes for terminal output."""
    GREEN = '\033[92m'
    RED = '\033[91m'
    YELLOW = '\033[93m'
    CYAN = '\033[96m'
    MAGENTA = '\033[95m'
    BLUE = '\033[94m'
    WHITE = '\033[97m'
    BOLD = '\033[1m'
    RESET = '\033[0m'


def setup_logging():
    """Setup logging to file and console."""
    formatter = logging.Formatter(
        '%(asctime)s | %(levelname)-8s | %(message)s',
        datefmt='%Y-%m-%d %H:%M:%S'
    )
    
    file_handler = logging.FileHandler(LOG_FILE, encoding='utf-8')
    file_handler.setFormatter(formatter)
    file_handler.setLevel(logging.DEBUG)
    
    console_handler = logging.StreamHandler(sys.stdout)
    console_handler.setFormatter(formatter)
    console_handler.setLevel(logging.INFO)
    
    logger = logging.getLogger()
    logger.setLevel(logging.DEBUG)
    logger.addHandler(file_handler)
    logger.addHandler(console_handler)
    
    return logger


logger = setup_logging()


def print_header(text: str, level: int = 1):
    """Print section header."""
    if level == 1:
        sep = "═" * 70
        print(f"\n{Colors.CYAN}{Colors.BOLD}{sep}{Colors.RESET}")
        print(f"{Colors.CYAN}{Colors.BOLD}  {text}{Colors.RESET}")
        print(f"{Colors.CYAN}{Colors.BOLD}{sep}{Colors.RESET}\n")
    else:
        sep = "─" * 50
        print(f"\n{Colors.BLUE}{sep}{Colors.RESET}")
        print(f"{Colors.BLUE}  {text}{Colors.RESET}")
        print(f"{Colors.BLUE}{sep}{Colors.RESET}\n")
    logger.info(f"{'=' * 70}")
    logger.info(f"  {text}")
    logger.info(f"{'=' * 70}")


def print_step(step: str, description: str):
    """Print test step."""
    print(f"{Colors.MAGENTA}[{step}]{Colors.RESET} {description}")
    logger.info(f"[{step}] {description}")


def print_success(text: str):
    """Print success message."""
    print(f"  {Colors.GREEN}✓ {text}{Colors.RESET}")
    logger.info(f"✓ {text}")


def print_error(text: str):
    """Print error message."""
    print(f"  {Colors.RED}✗ {text}{Colors.RESET}")
    logger.error(f"✗ {text}")


def print_info(text: str):
    """Print info message."""
    print(f"  {Colors.YELLOW}→ {text}{Colors.RESET}")
    logger.info(f"→ {text}")


def print_response_preview(response: requests.Response, max_len: int = 500):
    """Print response preview."""
    try:
        data = response.json()
        text = json.dumps(data, ensure_ascii=False, indent=2)
        if len(text) > max_len:
            text = text[:max_len] + "\n... (truncated)"
        logger.debug(f"Response: {text}")
        return data
    except:
        logger.debug(f"Response text: {response.text[:max_len]}")
        return None


class GranulaFullTest:
    """Full user flow test for Granula API."""
    
    def __init__(self):
        self.session = requests.Session()
        self.base_url = API_BASE_URL
        
        # Auth tokens
        self.access_token: Optional[str] = None
        self.refresh_token: Optional[str] = None
        self.user_id: Optional[str] = None
        
        # Created resources
        self.workspace_id: Optional[str] = None
        self.floor_plan_id: Optional[str] = None
        self.scene_id: Optional[str] = None
        self.branch_id: Optional[str] = None
        self.recognition_job_id: Optional[str] = None
        self.generation_job_id: Optional[str] = None
        self.chat_context_id: Optional[str] = None
        self.request_id: Optional[str] = None
        
        # Test results
        self.results = {"passed": 0, "failed": 0, "skipped": 0, "tests": []}
        
    def _headers(self, with_auth: bool = True) -> Dict[str, str]:
        """Get request headers."""
        headers = {"Content-Type": "application/json", "Accept": "application/json"}
        if with_auth and self.access_token:
            headers["Authorization"] = f"Bearer {self.access_token}"
        return headers
    
    def _request(self, method: str, endpoint: str, **kwargs) -> requests.Response:
        """Make API request with logging."""
        url = f"{self.base_url}{endpoint}"
        logger.debug(f"Request: {method} {url}")
        if 'json' in kwargs:
            logger.debug(f"Body: {json.dumps(kwargs['json'], ensure_ascii=False)[:500]}")
        return self.session.request(method, url, **kwargs)
    
    def record(self, name: str, passed: bool, details: str = ""):
        """Record test result."""
        self.results["tests"].append({"name": name, "passed": passed, "details": details})
        if passed:
            self.results["passed"] += 1
            print_success(f"{name}: OK")
        else:
            self.results["failed"] += 1
            print_error(f"{name}: {details}")
            
    def load_test_image(self) -> Tuple[Optional[str], Optional[str], Optional[Path]]:
        """Load a test image from Квартиры folder."""
        if not IMAGES_DIR.exists():
            return None, None, None
        images = list(IMAGES_DIR.glob("*.jpg")) + list(IMAGES_DIR.glob("*.png"))
        if not images:
            return None, None, None
        img_path = images[0]
        with open(img_path, 'rb') as f:
            data = base64.b64encode(f.read()).decode('utf-8')
        mime = "image/jpeg" if img_path.suffix.lower() in ['.jpg', '.jpeg'] else "image/png"
        return data, mime, img_path

    # =========================================================================
    # PHASE 1: AUTHENTICATION
    # =========================================================================
    def test_auth_register(self) -> bool:
        """Test user registration."""
        print_step("1.1", "POST /auth/register - Register new user")
        
        resp = self._request("POST", "/auth/register", headers=self._headers(False), json={
            "email": TEST_EMAIL,
            "password": TEST_PASSWORD,
            "name": TEST_NAME
        })
        data = print_response_preview(resp)
        
        if resp.status_code == 201 and data and "data" in data:
            self.access_token = data["data"].get("access_token")
            self.refresh_token = data["data"].get("refresh_token")
            self.user_id = data["data"].get("user_id")
            self.record("Register", True, f"User ID: {self.user_id}")
            return True
        else:
            self.record("Register", False, f"Status: {resp.status_code}")
            return False
            
    def test_auth_login(self) -> bool:
        """Test login."""
        print_step("1.2", "POST /auth/login - Login user")
        
        resp = self._request("POST", "/auth/login", headers=self._headers(False), json={
            "email": TEST_EMAIL,
            "password": TEST_PASSWORD
        })
        data = print_response_preview(resp)
        
        if resp.status_code == 200 and data and "data" in data:
            self.access_token = data["data"].get("access_token")
            self.refresh_token = data["data"].get("refresh_token")
            self.record("Login", True)
            return True
        else:
            self.record("Login", False, f"Status: {resp.status_code}")
            return False
            
    def test_auth_refresh(self) -> bool:
        """Test token refresh."""
        print_step("1.3", "POST /auth/refresh - Refresh token")
        
        if not self.refresh_token:
            self.results["skipped"] += 1
            print_info("Skipped: no refresh token")
            return False
            
        resp = self._request("POST", "/auth/refresh", headers=self._headers(False), json={
            "refresh_token": self.refresh_token
        })
        data = print_response_preview(resp)
        
        if resp.status_code == 200 and data and "data" in data:
            self.access_token = data["data"].get("access_token")
            self.refresh_token = data["data"].get("refresh_token")
            self.record("Token Refresh", True)
            return True
        else:
            self.record("Token Refresh", False, f"Status: {resp.status_code}")
            return False

    # =========================================================================
    # PHASE 2: USER PROFILE
    # =========================================================================
    def test_user_get_me(self) -> bool:
        """Test get current user profile."""
        print_step("2.1", "GET /users/me - Get profile")
        
        resp = self._request("GET", "/users/me", headers=self._headers())
        data = print_response_preview(resp)
        
        if resp.status_code == 200 and data:
            user = data.get("data", {})
            print_info(f"Email: {user.get('email')}, Name: {user.get('name')}")
            self.record("Get Profile", True)
            return True
        else:
            self.record("Get Profile", False, f"Status: {resp.status_code}")
            return False
            
    def test_user_update(self) -> bool:
        """Test update user profile."""
        print_step("2.2", "PATCH /users/me - Update profile")
        
        new_name = f"Updated User {TIMESTAMP}"
        resp = self._request("PATCH", "/users/me", headers=self._headers(), json={
            "name": new_name,
            "phone": "+7 999 123 4567"
        })
        data = print_response_preview(resp)
        
        if resp.status_code == 200:
            self.record("Update Profile", True, f"New name: {new_name}")
            return True
        else:
            self.record("Update Profile", False, f"Status: {resp.status_code}")
            return False

    # =========================================================================
    # PHASE 3: WORKSPACES
    # =========================================================================
    def test_workspace_create(self) -> bool:
        """Test workspace creation."""
        print_step("3.1", "POST /workspaces - Create workspace")
        
        resp = self._request("POST", "/workspaces", headers=self._headers(), json={
            "name": f"Test Apartment {TIMESTAMP}",
            "description": "Test workspace for API testing",
            "address": "г. Москва, ул. Тестовая, д. 1",
            "total_area": 65.5,
            "rooms_count": 2
        })
        data = print_response_preview(resp)
        
        if resp.status_code == 201 and data and "data" in data:
            self.workspace_id = data["data"].get("id")
            print_info(f"Workspace ID: {self.workspace_id}")
            self.record("Create Workspace", True)
            return True
        else:
            self.record("Create Workspace", False, f"Status: {resp.status_code}, Body: {data}")
            return False
            
    def test_workspace_list(self) -> bool:
        """Test list workspaces."""
        print_step("3.2", "GET /workspaces - List workspaces")
        
        resp = self._request("GET", "/workspaces", headers=self._headers())
        data = print_response_preview(resp)
        
        if resp.status_code == 200:
            count = len(data.get("data", []))
            print_info(f"Found {count} workspace(s)")
            self.record("List Workspaces", True)
            return True
        else:
            self.record("List Workspaces", False, f"Status: {resp.status_code}")
            return False
            
    def test_workspace_get(self) -> bool:
        """Test get workspace."""
        print_step("3.3", "GET /workspaces/{id} - Get workspace")
        
        if not self.workspace_id:
            self.results["skipped"] += 1
            print_info("Skipped: no workspace")
            return False
            
        resp = self._request("GET", f"/workspaces/{self.workspace_id}", headers=self._headers())
        data = print_response_preview(resp)
        
        if resp.status_code == 200:
            ws = data.get("data", {})
            print_info(f"Name: {ws.get('name')}")
            self.record("Get Workspace", True)
            return True
        else:
            self.record("Get Workspace", False, f"Status: {resp.status_code}")
            return False
            
    def test_workspace_update(self) -> bool:
        """Test update workspace."""
        print_step("3.4", "PATCH /workspaces/{id} - Update workspace")
        
        if not self.workspace_id:
            self.results["skipped"] += 1
            return False
            
        resp = self._request("PATCH", f"/workspaces/{self.workspace_id}", headers=self._headers(), json={
            "description": "Updated description",
            "rooms_count": 3
        })
        data = print_response_preview(resp)
        
        if resp.status_code == 200:
            self.record("Update Workspace", True)
            return True
        else:
            self.record("Update Workspace", False, f"Status: {resp.status_code}")
            return False

    # =========================================================================
    # PHASE 4: FLOOR PLANS
    # =========================================================================
    def test_floorplan_upload(self) -> bool:
        """Test floor plan upload."""
        print_step("4.1", "POST /floor-plans - Upload floor plan")
        
        if not self.workspace_id:
            self.results["skipped"] += 1
            return False
            
        img_data, mime, img_path = self.load_test_image()
        if not img_data:
            self.results["skipped"] += 1
            print_info("Skipped: no test images")
            return False
            
        print_info(f"Using image: {img_path.name}")
        
        # Upload as multipart/form-data
        with open(img_path, 'rb') as f:
            files = {'file': (img_path.name, f, mime)}
            data = {'workspace_id': self.workspace_id, 'name': f'Test Plan {TIMESTAMP}'}
            headers = {"Authorization": f"Bearer {self.access_token}"}
            resp = self._request("POST", "/floor-plans", headers=headers, files=files, data=data)
        
        resp_data = print_response_preview(resp)
        
        if resp.status_code in [200, 201] and resp_data:
            self.floor_plan_id = resp_data.get("data", {}).get("id")
            print_info(f"Floor Plan ID: {self.floor_plan_id}")
            self.record("Upload Floor Plan", True)
            return True
        else:
            self.record("Upload Floor Plan", False, f"Status: {resp.status_code}")
            return False
            
    def test_floorplan_list(self) -> bool:
        """Test list floor plans."""
        print_step("4.2", "GET /floor-plans - List floor plans")
        
        if not self.workspace_id:
            self.results["skipped"] += 1
            print_info("Skipped: no workspace")
            return False
        
        # workspace_id is required query parameter
        resp = self._request("GET", "/floor-plans", headers=self._headers(), params={
            "workspace_id": self.workspace_id
        })
        data = print_response_preview(resp)
        
        if resp.status_code == 200:
            items = data.get("data", {}).get("items", [])
            print_info(f"Found {len(items)} floor plan(s)")
            self.record("List Floor Plans", True)
            return True
        else:
            self.record("List Floor Plans", False, f"Status: {resp.status_code}")
            return False

    # =========================================================================
    # PHASE 5: AI RECOGNITION
    # =========================================================================
    def test_ai_recognize(self) -> bool:
        """Test AI floor plan recognition."""
        print_step("5.1", "POST /ai/recognize - Start recognition")
        
        img_data, mime, img_path = self.load_test_image()
        if not img_data:
            self.results["skipped"] += 1
            return False
            
        print_info(f"Sending image {img_path.name} ({len(img_data) / 1024:.1f} KB base64)")
        
        resp = self._request("POST", "/ai/recognize", headers=self._headers(), json={
            "floor_plan_id": self.floor_plan_id or "test-recognition",
            "image_base64": img_data,
            "image_type": mime,
            "options": {
                "detect_load_bearing": True,
                "detect_wet_zones": True,
                "detect_furniture": False
            }
        })
        data = print_response_preview(resp)
        
        if resp.status_code == 200 and data:
            self.recognition_job_id = data.get("data", {}).get("job_id")
            print_info(f"Job ID: {self.recognition_job_id}")
            self.record("Start Recognition", True)
            return True
        else:
            self.record("Start Recognition", False, f"Status: {resp.status_code}")
            return False
            
    def test_ai_recognize_status(self) -> bool:
        """Test recognition status polling."""
        print_step("5.2", "GET /ai/recognize/{job_id}/status - Poll status")
        
        if not self.recognition_job_id:
            self.results["skipped"] += 1
            return False
            
        max_attempts = 30
        for i in range(max_attempts):
            time.sleep(2)
            resp = self._request("GET", f"/ai/recognize/{self.recognition_job_id}/status", headers=self._headers())
            data = print_response_preview(resp)
            
            if resp.status_code == 200:
                status = data.get("data", {}).get("status", "unknown")
                progress = data.get("data", {}).get("progress", 0)
                print_info(f"Attempt {i+1}: {status} ({progress}%)")
                
                if status == "completed":
                    # Try to get scene_id from result
                    scene_data = data.get("data", {}).get("scene", {})
                    if scene_data:
                        self.scene_id = scene_data.get("id") or f"scene-{self.recognition_job_id}"
                    self.record("Recognition Complete", True)
                    return True
                elif status == "failed":
                    error = data.get("data", {}).get("error", "unknown")
                    self.record("Recognition Complete", False, f"Job failed: {error}")
                    return False
                    
        self.record("Recognition Complete", False, "Timeout")
        return False

    # =========================================================================
    # PHASE 6: AI CHAT
    # =========================================================================
    def test_ai_chat_send(self) -> bool:
        """Test sending chat message."""
        print_step("6.1", "POST /ai/chat - Send message")
        
        resp = self._request("POST", "/ai/chat", headers=self._headers(), json={
            "message": "Привет! Какие есть ограничения при объединении кухни с гостиной?",
            "scene_id": self.scene_id
        })
        data = print_response_preview(resp)
        
        if resp.status_code == 200 and data:
            self.chat_context_id = data.get("data", {}).get("context_id")
            response_text = data.get("data", {}).get("response", "")
            print_info(f"Response: {response_text[:150]}...")
            print_info(f"Context ID: {self.chat_context_id}")
            self.record("Chat Message", True)
            return True
        else:
            self.record("Chat Message", False, f"Status: {resp.status_code}")
            return False
            
    def test_ai_chat_followup(self) -> bool:
        """Test follow-up chat message."""
        print_step("6.2", "POST /ai/chat - Follow-up message")
        
        resp = self._request("POST", "/ai/chat", headers=self._headers(), json={
            "message": "А если кухня газифицирована, какие есть варианты?",
            "scene_id": self.scene_id,
            "context_id": self.chat_context_id
        })
        data = print_response_preview(resp)
        
        if resp.status_code == 200 and data:
            response_text = data.get("data", {}).get("response", "")
            print_info(f"Response: {response_text[:150]}...")
            self.record("Chat Follow-up", True)
            return True
        else:
            self.record("Chat Follow-up", False, f"Status: {resp.status_code}")
            return False
            
    def test_ai_chat_history(self) -> bool:
        """Test getting chat history."""
        print_step("6.3", "GET /ai/chat/history - Get history")
        
        resp = self._request("GET", "/ai/chat/history", headers=self._headers(), params={"limit": 50})
        data = print_response_preview(resp)
        
        if resp.status_code == 200:
            count = len(data.get("data", {}).get("messages", []))
            print_info(f"Found {count} messages")
            self.record("Chat History", True)
            return True
        else:
            self.record("Chat History", False, f"Status: {resp.status_code}")
            return False

    # =========================================================================
    # PHASE 7: AI GENERATION
    # =========================================================================
    def test_ai_generate(self) -> bool:
        """Test AI variant generation."""
        print_step("7.1", "POST /ai/generate - Generate variants")
        
        scene_id = self.scene_id or "test-scene-gen"
        
        resp = self._request("POST", "/ai/generate", headers=self._headers(), json={
            "scene_id": scene_id,
            "prompt": "Предложи 2 варианта объединения кухни с гостиной с учетом норм СНиП",
            "variants_count": 2,
            "preserve_load_bearing": True,
            "check_compliance": True
        })
        data = print_response_preview(resp)
        
        if resp.status_code == 200 and data:
            self.generation_job_id = data.get("data", {}).get("job_id")
            print_info(f"Generation Job ID: {self.generation_job_id}")
            self.record("Start Generation", True)
            return True
        else:
            self.record("Start Generation", False, f"Status: {resp.status_code}")
            return False
            
    def test_ai_generate_status(self) -> bool:
        """Test generation status polling."""
        print_step("7.2", "GET /ai/generate/{job_id}/status - Poll status")
        
        if not self.generation_job_id:
            self.results["skipped"] += 1
            return False
            
        max_attempts = 30
        for i in range(max_attempts):
            time.sleep(2)
            resp = self._request("GET", f"/ai/generate/{self.generation_job_id}/status", headers=self._headers())
            data = print_response_preview(resp)
            
            if resp.status_code == 200:
                status = data.get("data", {}).get("status", "unknown")
                progress = data.get("data", {}).get("progress", 0)
                print_info(f"Attempt {i+1}: {status} ({progress}%)")
                
                if status == "completed":
                    variants = data.get("data", {}).get("variants", [])
                    print_info(f"Generated {len(variants)} variant(s)")
                    for idx, v in enumerate(variants):
                        print_info(f"  [{idx+1}] {v.get('name', 'N/A')}")
                    self.record("Generation Complete", True)
                    return True
                elif status == "failed":
                    error = data.get("data", {}).get("error", "unknown")
                    self.record("Generation Complete", False, f"Job failed: {error}")
                    return False
                    
        self.record("Generation Complete", False, "Timeout")
        return False

    # =========================================================================
    # PHASE 8: AI CONTEXT
    # =========================================================================
    def test_ai_context_get(self) -> bool:
        """Test getting AI context."""
        print_step("8.1", "GET /ai/context - Get context")
        
        scene_id = self.scene_id or "test-scene"
        resp = self._request("GET", "/ai/context", headers=self._headers(), params={"scene_id": scene_id})
        data = print_response_preview(resp)
        
        if resp.status_code == 200:
            ctx = data.get("data", {})
            print_info(f"Context ID: {ctx.get('context_id')}, Size: {ctx.get('context_size')} tokens")
            self.record("Get AI Context", True)
            return True
        else:
            self.record("Get AI Context", False, f"Status: {resp.status_code}")
            return False
            
    def test_ai_context_update(self) -> bool:
        """Test updating AI context."""
        print_step("8.2", "POST /ai/context - Update context")
        
        scene_id = self.scene_id or "test-scene"
        resp = self._request("POST", "/ai/context", headers=self._headers(), json={
            "scene_id": scene_id,
            "force": True
        })
        data = print_response_preview(resp)
        
        if resp.status_code == 200:
            self.record("Update AI Context", True)
            return True
        else:
            self.record("Update AI Context", False, f"Status: {resp.status_code}")
            return False

    # =========================================================================
    # PHASE 9: COMPLIANCE
    # =========================================================================
    def test_compliance_check(self) -> bool:
        """Test compliance check."""
        print_step("9.1", "POST /compliance/check - Check compliance")
        
        scene_id = self.scene_id or "test-scene"
        resp = self._request("POST", "/compliance/check", headers=self._headers(), json={
            "scene_id": scene_id
        })
        data = print_response_preview(resp)
        
        if resp.status_code == 200:
            result = data.get("data", {})
            compliant = result.get("compliant", False)
            violations = len(result.get("violations", []))
            print_info(f"Compliant: {compliant}, Violations: {violations}")
            self.record("Compliance Check", True)
            return True
        else:
            self.record("Compliance Check", False, f"Status: {resp.status_code}")
            return False
            
    def test_compliance_rules(self) -> bool:
        """Test getting compliance rules."""
        print_step("9.2", "GET /compliance/rules - Get rules")
        
        resp = self._request("GET", "/compliance/rules", headers=self._headers())
        data = print_response_preview(resp)
        
        if resp.status_code == 200:
            rules = data.get("data", {}).get("rules", [])
            print_info(f"Found {len(rules)} rule(s)")
            self.record("Get Compliance Rules", True)
            return True
        else:
            self.record("Get Compliance Rules", False, f"Status: {resp.status_code}")
            return False

    # =========================================================================
    # PHASE 10: BRANCHES
    # =========================================================================
    def test_branches_list(self) -> bool:
        """Test listing branches."""
        print_step("10.1", "GET /scenes/{scene_id}/branches - List branches")
        
        scene_id = self.scene_id or "test-scene"
        # Branches are under /scenes/:scene_id/branches
        resp = self._request("GET", f"/scenes/{scene_id}/branches", headers=self._headers())
        data = print_response_preview(resp)
        
        if resp.status_code == 200:
            branches = data.get("data", {}).get("branches", [])
            if isinstance(data.get("data"), list):
                branches = data.get("data", [])
            print_info(f"Found {len(branches)} branch(es)")
            self.record("List Branches", True)
            return True
        else:
            self.record("List Branches", False, f"Status: {resp.status_code}")
            return False
            
    def test_branches_create(self) -> bool:
        """Test creating branch."""
        print_step("10.2", "POST /scenes/{scene_id}/branches - Create branch")
        
        scene_id = self.scene_id or "test-scene"
        # Branches are under /scenes/:scene_id/branches
        resp = self._request("POST", f"/scenes/{scene_id}/branches", headers=self._headers(), json={
            "name": f"Test Branch {TIMESTAMP}",
            "description": "Test branch for API testing"
        })
        data = print_response_preview(resp)
        
        if resp.status_code in [200, 201]:
            self.branch_id = data.get("data", {}).get("id")
            print_info(f"Branch ID: {self.branch_id}")
            self.record("Create Branch", True)
            return True
        else:
            self.record("Create Branch", False, f"Status: {resp.status_code}")
            return False

    # =========================================================================
    # PHASE 11: REQUESTS (Заявки)
    # =========================================================================
    def test_requests_create(self) -> bool:
        """Test creating expert request."""
        print_step("11.1", "POST /requests - Create request")
        
        if not self.workspace_id:
            self.results["skipped"] += 1
            print_info("Skipped: no workspace")
            return False
        
        resp = self._request("POST", "/requests", headers=self._headers(), json={
            "workspace_id": self.workspace_id,
            "title": "Консультация по перепланировке кухни",
            "description": "Нужна консультация по объединению кухни с гостиной",
            "category": "consultation",
            "priority": "normal",
            "contact": {
                "name": "Тест Пользователь",
                "phone": "+7 999 123 4567",
                "email": TEST_EMAIL
            }
        })
        data = print_response_preview(resp)
        
        if resp.status_code in [200, 201]:
            self.request_id = data.get("data", {}).get("id")
            print_info(f"Request ID: {self.request_id}")
            self.record("Create Request", True)
            return True
        else:
            self.record("Create Request", False, f"Status: {resp.status_code}, Body: {data}")
            return False
            
    def test_requests_list(self) -> bool:
        """Test listing requests."""
        print_step("11.2", "GET /requests - List requests")
        
        resp = self._request("GET", "/requests", headers=self._headers())
        data = print_response_preview(resp)
        
        if resp.status_code == 200:
            requests = data.get("data", {}).get("requests", [])
            print_info(f"Found {len(requests)} request(s)")
            self.record("List Requests", True)
            return True
        else:
            self.record("List Requests", False, f"Status: {resp.status_code}")
            return False

    # =========================================================================
    # PHASE 12: NOTIFICATIONS
    # =========================================================================
    def test_notifications_list(self) -> bool:
        """Test listing notifications."""
        print_step("12.1", "GET /notifications - List notifications")
        
        resp = self._request("GET", "/notifications", headers=self._headers())
        data = print_response_preview(resp)
        
        if resp.status_code == 200:
            # Handle both list and dict response formats
            if isinstance(data, list):
                notifications = data
            elif isinstance(data, dict):
                notifications = data.get("data", {})
                if isinstance(notifications, dict):
                    notifications = notifications.get("notifications", [])
            else:
                notifications = []
            print_info(f"Found {len(notifications)} notification(s)")
            self.record("List Notifications", True)
            return True
        else:
            self.record("List Notifications", False, f"Status: {resp.status_code}")
            return False

    # =========================================================================
    # PHASE 13: LOGOUT
    # =========================================================================
    def test_auth_logout(self) -> bool:
        """Test logout."""
        print_step("13.1", "POST /auth/logout - Logout")
        
        resp = self._request("POST", "/auth/logout", headers=self._headers(), json={
            "refresh_token": self.refresh_token
        })
        data = print_response_preview(resp)
        
        if resp.status_code == 200:
            self.record("Logout", True)
            return True
        else:
            self.record("Logout", False, f"Status: {resp.status_code}")
            return False

    # =========================================================================
    # RUN ALL TESTS
    # =========================================================================
    def run(self):
        """Run all tests in sequence."""
        print_header("GRANULA API - FULL USER FLOW TEST")
        print_info(f"API: {API_BASE_URL}")
        print_info(f"Test User: {TEST_EMAIL}")
        print_info(f"Log: {LOG_FILE}")
        
        # Phase 1: Auth
        print_header("PHASE 1: AUTHENTICATION", 2)
        self.test_auth_register()
        if not self.access_token:
            self.test_auth_login()
        self.test_auth_refresh()
        
        if not self.access_token:
            print_error("Cannot continue without authentication")
            self.print_summary()
            return
            
        # Phase 2: User
        print_header("PHASE 2: USER PROFILE", 2)
        self.test_user_get_me()
        self.test_user_update()
        
        # Phase 3: Workspaces
        print_header("PHASE 3: WORKSPACES", 2)
        self.test_workspace_create()
        self.test_workspace_list()
        self.test_workspace_get()
        self.test_workspace_update()
        
        # Phase 4: Floor Plans
        print_header("PHASE 4: FLOOR PLANS", 2)
        self.test_floorplan_upload()
        self.test_floorplan_list()
        
        # Phase 5: AI Recognition
        print_header("PHASE 5: AI RECOGNITION", 2)
        self.test_ai_recognize()
        self.test_ai_recognize_status()
        
        # Phase 6: AI Chat
        print_header("PHASE 6: AI CHAT", 2)
        self.test_ai_chat_send()
        self.test_ai_chat_followup()
        self.test_ai_chat_history()
        
        # Phase 7: AI Generation
        print_header("PHASE 7: AI GENERATION", 2)
        self.test_ai_generate()
        self.test_ai_generate_status()
        
        # Phase 8: AI Context
        print_header("PHASE 8: AI CONTEXT", 2)
        self.test_ai_context_get()
        self.test_ai_context_update()
        
        # Phase 9: Compliance
        print_header("PHASE 9: COMPLIANCE", 2)
        self.test_compliance_check()
        self.test_compliance_rules()
        
        # Phase 10: Branches
        print_header("PHASE 10: BRANCHES", 2)
        self.test_branches_list()
        self.test_branches_create()
        
        # Phase 11: Requests
        print_header("PHASE 11: REQUESTS", 2)
        self.test_requests_create()
        self.test_requests_list()
        
        # Phase 12: Notifications
        print_header("PHASE 12: NOTIFICATIONS", 2)
        self.test_notifications_list()
        
        # Phase 13: Logout
        print_header("PHASE 13: LOGOUT", 2)
        self.test_auth_logout()
        
        # Summary
        self.print_summary()
        
    def print_summary(self):
        """Print test summary."""
        print_header("TEST SUMMARY")
        
        total = self.results["passed"] + self.results["failed"] + self.results["skipped"]
        
        print(f"\n{Colors.BOLD}Results:{Colors.RESET}")
        print(f"  {Colors.GREEN}Passed:  {self.results['passed']}{Colors.RESET}")
        print(f"  {Colors.RED}Failed:  {self.results['failed']}{Colors.RESET}")
        print(f"  {Colors.YELLOW}Skipped: {self.results['skipped']}{Colors.RESET}")
        print(f"  Total:   {total}")
        
        logger.info(f"Total: {total}, Passed: {self.results['passed']}, Failed: {self.results['failed']}, Skipped: {self.results['skipped']}")
        
        if self.results["failed"] > 0:
            print(f"\n{Colors.RED}Failed tests:{Colors.RESET}")
            for test in self.results["tests"]:
                if not test["passed"]:
                    print(f"  • {test['name']}: {test['details']}")
                    logger.error(f"FAILED: {test['name']}: {test['details']}")
                    
        print(f"\n{Colors.CYAN}Full log: {LOG_FILE}{Colors.RESET}")


if __name__ == "__main__":
    test = GranulaFullTest()
    test.run()


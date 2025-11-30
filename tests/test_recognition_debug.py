#!/usr/bin/env python3
"""
Детальный тест AI распознавания планировки с полным логированием.
Показывает все ответы API на каждом этапе.
Сохраняет все результаты в JSON файл.
"""

import requests
import json
import time
import base64
import os
import sys
from pathlib import Path
from datetime import datetime

# =============================================================================
# CONFIGURATION
# =============================================================================

API_BASE = "https://api.granula.raitokyokai.tech/api/v1"
# API_BASE = "http://localhost:3001/api/v1"  # Для локального теста

# Уникальный email для каждого запуска
TIMESTAMP = int(time.time())
TEST_EMAIL = f"debug_{TIMESTAMP}@test.ru"
TEST_PASSWORD = "TestPass123!"
TEST_NAME = "Debug Tester"

# Путь к папке с планами квартир
APARTMENTS_DIR = Path(__file__).parent.parent / "Квартиры"

# Результаты теста для сохранения
TEST_RESULTS = {
    "timestamp": datetime.now().isoformat(),
    "api_base": API_BASE,
    "test_email": TEST_EMAIL,
    "steps": [],
    "recognition_result": None,
    "scene_data": None,
    "summary": {
        "success": False,
        "errors": []
    }
}

# =============================================================================
# HELPERS
# =============================================================================

def log(msg: str, data=None):
    """Печать с timestamp"""
    ts = datetime.now().strftime("%H:%M:%S.%f")[:-3]
    print(f"\n[{ts}] {msg}")
    if data is not None:
        if isinstance(data, (dict, list)):
            print(json.dumps(data, indent=2, ensure_ascii=False, default=str))
        else:
            print(data)

def log_response(resp: requests.Response, label: str) -> dict:
    """Полный лог ответа"""
    print(f"\n{'='*60}")
    print(f"📡 {label}")
    print(f"{'='*60}")
    print(f"Status: {resp.status_code}")
    
    step_data = {
        "label": label,
        "status_code": resp.status_code,
        "url": resp.url,
        "body": None
    }
    
    try:
        data = resp.json()
        print(f"Body:\n{json.dumps(data, indent=2, ensure_ascii=False, default=str)}")
        step_data["body"] = data
        TEST_RESULTS["steps"].append(step_data)
        return data
    except:
        print(resp.text[:2000] if resp.text else "(empty)")
        step_data["body"] = resp.text[:500] if resp.text else None
        TEST_RESULTS["steps"].append(step_data)
        return None

def image_to_base64(filepath: str) -> tuple:
    """Конвертирует изображение в base64 data URL"""
    with open(filepath, "rb") as f:
        content = f.read()
    
    ext = Path(filepath).suffix.lower()
    mime_types = {
        ".jpg": "image/jpeg",
        ".jpeg": "image/jpeg",
        ".png": "image/png",
        ".gif": "image/gif",
        ".webp": "image/webp",
    }
    mime = mime_types.get(ext, "image/jpeg")
    
    encoded = base64.b64encode(content).decode("utf-8")
    return f"data:{mime};base64,{encoded}", mime

def save_results():
    """Сохраняет результаты теста в JSON"""
    result_file = Path(__file__).parent / f"test_results_{TIMESTAMP}.json"
    with open(result_file, "w", encoding="utf-8") as f:
        json.dump(TEST_RESULTS, f, indent=2, ensure_ascii=False, default=str)
    log(f"💾 Все результаты сохранены в: {result_file}")
    return result_file

# =============================================================================
# MAIN TEST
# =============================================================================

def main():
    log("🚀 НАЧИНАЕМ ТЕСТ РАСПОЗНАВАНИЯ")
    log(f"API: {API_BASE}")
    log(f"Email: {TEST_EMAIL}")
    
    # Найдём картинку для теста
    images = list(APARTMENTS_DIR.glob("*.jpg")) + list(APARTMENTS_DIR.glob("*.jpeg")) + list(APARTMENTS_DIR.glob("*.png"))
    if not images:
        log("❌ Нет изображений в папке Квартиры!")
        TEST_RESULTS["summary"]["errors"].append("No images found")
        save_results()
        sys.exit(1)
    
    test_image = images[0]
    log(f"📷 Тестовое изображение: {test_image.name}")
    log(f"   Размер: {test_image.stat().st_size / 1024:.1f} KB")
    TEST_RESULTS["test_image"] = str(test_image.name)
    
    # =========================================================================
    # STEP 1: Регистрация
    # =========================================================================
    log("📝 STEP 1: Регистрация пользователя...")
    
    resp = requests.post(f"{API_BASE}/auth/register", json={
        "email": TEST_EMAIL,
        "password": TEST_PASSWORD,
        "name": TEST_NAME
    })
    data = log_response(resp, "POST /auth/register")
    
    if resp.status_code not in [200, 201]:
        log("❌ Регистрация не удалась!")
        TEST_RESULTS["summary"]["errors"].append("Registration failed")
        save_results()
        sys.exit(1)
    
    token = data.get("data", {}).get("access_token")
    if not token:
        log("❌ Нет токена в ответе!")
        TEST_RESULTS["summary"]["errors"].append("No token in response")
        save_results()
        sys.exit(1)
    
    log(f"✅ Токен получен: {token[:50]}...")
    TEST_RESULTS["token"] = token[:50] + "..."
    
    headers = {"Authorization": f"Bearer {token}"}
    
    # =========================================================================
    # STEP 2: Создание воркспейса
    # =========================================================================
    log("🏠 STEP 2: Создание воркспейса...")
    
    resp = requests.post(f"{API_BASE}/workspaces", 
        headers=headers,
        json={
            "name": f"Debug Workspace {TIMESTAMP}",
            "description": "Тест распознавания",
            "address": "г. Тест, ул. Дебаг, д. 1",
            "total_area": 50.0,
            "rooms_count": 2
        }
    )
    data = log_response(resp, "POST /workspaces")
    
    workspace_id = None
    if resp.status_code in [200, 201] and data:
        workspace_id = data.get("data", {}).get("id")
    
    if not workspace_id:
        log("⚠️ Workspace не создан, продолжаем без него...")
        TEST_RESULTS["summary"]["errors"].append("Workspace creation failed")
    else:
        log(f"✅ Workspace ID: {workspace_id}")
        TEST_RESULTS["workspace_id"] = workspace_id
    
    # =========================================================================
    # STEP 3: Загрузка плана
    # =========================================================================
    log("📤 STEP 3: Загрузка плана квартиры...")
    
    floor_plan_id = None
    if workspace_id:
        with open(test_image, "rb") as f:
            files = {"file": (test_image.name, f, "image/jpeg")}
            form_data = {
                "workspace_id": workspace_id,
                "name": f"План {test_image.name}"
            }
            resp = requests.post(
                f"{API_BASE}/floor-plans",
                headers=headers,
                files=files,
                data=form_data
            )
        data = log_response(resp, "POST /floor-plans")
        
        if resp.status_code in [200, 201] and data:
            floor_plan_id = data.get("data", {}).get("id")
            log(f"✅ Floor Plan ID: {floor_plan_id}")
            TEST_RESULTS["floor_plan_id"] = floor_plan_id
    
    # =========================================================================
    # STEP 4: AI Распознавание
    # =========================================================================
    log("🤖 STEP 4: Запуск AI распознавания...")
    
    image_base64, mime_type = image_to_base64(str(test_image))
    log(f"   Base64 длина: {len(image_base64)} символов")
    log(f"   MIME тип: {mime_type}")
    
    recognize_payload = {
        "floor_plan_id": floor_plan_id or f"test-{TIMESTAMP}",
        "image_base64": image_base64,
        "image_type": mime_type,
        "options": {
            "detect_load_bearing": True,
            "detect_wet_zones": True,
            "detect_furniture": True
        }
    }
    
    log("📤 Отправляем запрос на /ai/recognize...")
    
    resp = requests.post(
        f"{API_BASE}/ai/recognize",
        headers={**headers, "Content-Type": "application/json"},
        json=recognize_payload
    )
    data = log_response(resp, "POST /ai/recognize")
    
    job_id = None
    if data:
        job_id = data.get("data", {}).get("job_id") or data.get("job_id")
    
    if not job_id:
        log("❌ Нет job_id в ответе!")
        TEST_RESULTS["summary"]["errors"].append("No job_id in recognize response")
        save_results()
        sys.exit(1)
    
    log(f"✅ Job ID: {job_id}")
    TEST_RESULTS["job_id"] = job_id
    
    # =========================================================================
    # STEP 5: Polling статуса
    # =========================================================================
    log("⏳ STEP 5: Polling статуса распознавания...")
    
    recognition_result = None
    max_attempts = 30
    
    for attempt in range(max_attempts):
        log(f"   Попытка {attempt + 1}/{max_attempts}...")
        
        resp = requests.get(
            f"{API_BASE}/ai/recognize/{job_id}/status",
            headers=headers
        )
        data = log_response(resp, f"GET /ai/recognize/{job_id}/status")
        
        if not data:
            log("❌ Пустой ответ!")
            break
        
        status_data = data.get("data", data)
        status = status_data.get("status")
        progress = status_data.get("progress", 0)
        
        log(f"   Status: {status}, Progress: {progress}%")
        
        if status == "completed":
            log("✅ РАСПОЗНАВАНИЕ ЗАВЕРШЕНО!")
            recognition_result = status_data.get("result")
            
            if recognition_result:
                log("📦 RESULT JSON:")
                log(None, recognition_result)
                TEST_RESULTS["recognition_result"] = recognition_result
                
                # Сохраним отдельно
                result_file = Path(__file__).parent / f"recognition_result_{TIMESTAMP}.json"
                with open(result_file, "w", encoding="utf-8") as f:
                    json.dump(recognition_result, f, indent=2, ensure_ascii=False)
                log(f"💾 Результат сохранён в: {result_file}")
            else:
                log("⚠️ Статус completed, но нет result!")
                TEST_RESULTS["summary"]["errors"].append("Completed but no result")
            break
        
        elif status == "failed":
            log("❌ РАСПОЗНАВАНИЕ FAILED!")
            error = status_data.get("error") or status_data.get("message")
            log(f"   Error: {error}")
            TEST_RESULTS["summary"]["errors"].append(f"Recognition failed: {error}")
            break
        
        elif status in ["processing", "pending", "queued"]:
            log(f"   Ждём 3 секунды...")
            time.sleep(3)
        
        else:
            log(f"⚠️ Неизвестный статус: {status}")
            time.sleep(3)
    
    else:
        log("⏰ Timeout!")
        TEST_RESULTS["summary"]["errors"].append("Recognition timeout")
    
    # =========================================================================
    # STEP 6: Создание сцены из результата распознавания
    # =========================================================================
    scene_id = None
    
    if workspace_id and floor_plan_id:
        log("🎮 STEP 6: Создание 3D сцены из результата...")
        
        resp = requests.post(
            f"{API_BASE}/workspaces/{workspace_id}/scenes",
            headers={**headers, "Content-Type": "application/json"},
            json={
                "name": f"Scene from {test_image.name}",
                "description": "Created from recognition result",
                "floor_plan_id": floor_plan_id
            }
        )
        data = log_response(resp, f"POST /workspaces/{workspace_id}/scenes")
        
        if resp.status_code in [200, 201] and data:
            scene_id = data.get("data", {}).get("id")
            if scene_id:
                log(f"✅ Scene ID: {scene_id}")
                TEST_RESULTS["scene_id"] = scene_id
        
        # =====================================================================
        # STEP 7: Получение сцены
        # ВАЖНО: GET /scenes/{scene_id} (НЕ /workspaces/{id}/scenes/{id}!)
        # =====================================================================
        if scene_id:
            log("🔍 STEP 7: Получение данных сцены...")
            
            # ПРАВИЛЬНЫЙ путь: /scenes/{scene_id}
            resp = requests.get(
                f"{API_BASE}/scenes/{scene_id}",
                headers=headers
            )
            data = log_response(resp, f"GET /scenes/{scene_id}")
            
            if resp.status_code == 200 and data:
                TEST_RESULTS["scene_data"] = data.get("data", data)
                log("✅ Данные сцены получены!")
            else:
                log(f"⚠️ Ошибка получения сцены: {resp.status_code}")
                TEST_RESULTS["summary"]["errors"].append(f"Get scene failed: {resp.status_code}")
    
    # =========================================================================
    # STEP 8: Тест AI Chat
    # =========================================================================
    log("💬 STEP 8: Тест AI Chat...")
    
    chat_payload = {
        "message": "Можно ли снести стену между кухней и гостиной?",
        "scene_id": scene_id or ""
    }
    
    resp = requests.post(
        f"{API_BASE}/ai/chat",
        headers={**headers, "Content-Type": "application/json"},
        json=chat_payload
    )
    data = log_response(resp, "POST /ai/chat")
    
    if resp.status_code == 200 and data:
        TEST_RESULTS["chat_response"] = data.get("data", {}).get("response", "")[:500]
        log("✅ AI Chat работает!")
    
    # =========================================================================
    # SUMMARY
    # =========================================================================
    TEST_RESULTS["summary"]["success"] = len(TEST_RESULTS["summary"]["errors"]) == 0
    
    log("\n" + "="*60)
    log("📊 ИТОГИ ТЕСТА")
    log("="*60)
    log(f"   Token: {'✅' if token else '❌'}")
    log(f"   Workspace: {'✅ ' + str(workspace_id)[:8] if workspace_id else '❌'}")
    log(f"   Floor Plan: {'✅ ' + str(floor_plan_id)[:8] if floor_plan_id else '❌'}")
    log(f"   Recognition: {'✅' if recognition_result else '❌'}")
    log(f"   Scene: {'✅ ' + str(scene_id)[:8] if scene_id else '❌'}")
    log(f"   Errors: {len(TEST_RESULTS['summary']['errors'])}")
    
    if TEST_RESULTS["summary"]["errors"]:
        log("   ❌ Ошибки:")
        for err in TEST_RESULTS["summary"]["errors"]:
            log(f"      - {err}")
    
    # Сохраняем все результаты
    result_file = save_results()
    
    log(f"\n🎉 Тест завершён! Результаты: {result_file}")

if __name__ == "__main__":
    main()

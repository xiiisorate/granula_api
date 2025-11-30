#!/usr/bin/env python3
"""
Детальный тест AI распознавания планировки с полным логированием.
Показывает все ответы API на каждом этапе.
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

def log_response(resp: requests.Response, label: str):
    """Полный лог ответа"""
    print(f"\n{'='*60}")
    print(f"📡 {label}")
    print(f"{'='*60}")
    print(f"Status: {resp.status_code}")
    print(f"Headers: {dict(resp.headers)}")
    print(f"\nBody:")
    try:
        data = resp.json()
        print(json.dumps(data, indent=2, ensure_ascii=False, default=str))
        return data
    except:
        print(resp.text[:2000] if resp.text else "(empty)")
        return None

def image_to_base64(filepath: str) -> str:
    """Конвертирует изображение в base64 data URL"""
    with open(filepath, "rb") as f:
        content = f.read()
    
    # Определяем MIME тип
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
        sys.exit(1)
    
    test_image = images[0]
    log(f"📷 Тестовое изображение: {test_image.name}")
    log(f"   Размер: {test_image.stat().st_size / 1024:.1f} KB")
    
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
        sys.exit(1)
    
    token = data.get("data", {}).get("access_token")
    if not token:
        log("❌ Нет токена в ответе!")
        sys.exit(1)
    
    log(f"✅ Токен получен: {token[:50]}...")
    
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
        log("❌ Workspace не создан! Пробуем получить существующий...")
        resp = requests.get(f"{API_BASE}/workspaces", headers=headers)
        data = log_response(resp, "GET /workspaces")
        workspaces = data.get("data", {}).get("workspaces", [])
        if workspaces:
            workspace_id = workspaces[0].get("id")
            log(f"✅ Используем существующий workspace: {workspace_id}")
    
    if not workspace_id:
        log("❌ Нет доступных workspaces!")
        # Продолжим без workspace для тестирования recognize
    else:
        log(f"✅ Workspace ID: {workspace_id}")
    
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
    
    # =========================================================================
    # STEP 4: AI Распознавание
    # =========================================================================
    log("🤖 STEP 4: Запуск AI распознавания...")
    
    # Конвертируем изображение в base64
    image_base64, mime_type = image_to_base64(str(test_image))
    log(f"   Base64 длина: {len(image_base64)} символов")
    log(f"   MIME тип: {mime_type}")
    
    recognize_payload = {
        "floor_plan_id": floor_plan_id or "test-floor-plan-id",
        "image_base64": image_base64,
        "image_type": mime_type,
        "options": {
            "detect_load_bearing": True,
            "detect_wet_zones": True,
            "detect_furniture": True
        }
    }
    
    log("📤 Отправляем запрос на /ai/recognize...")
    log(f"   Payload keys: {list(recognize_payload.keys())}")
    log(f"   image_base64 начало: {image_base64[:100]}...")
    
    resp = requests.post(
        f"{API_BASE}/ai/recognize",
        headers={**headers, "Content-Type": "application/json"},
        json=recognize_payload
    )
    data = log_response(resp, "POST /ai/recognize")
    
    if resp.status_code not in [200, 201, 202]:
        log("❌ Распознавание не запустилось!")
        # Пробуем без floor_plan_id
        log("🔄 Пробуем без floor_plan_id...")
        del recognize_payload["floor_plan_id"]
        resp = requests.post(
            f"{API_BASE}/ai/recognize",
            headers={**headers, "Content-Type": "application/json"},
            json=recognize_payload
        )
        data = log_response(resp, "POST /ai/recognize (без floor_plan_id)")
    
    job_id = None
    if data:
        job_id = data.get("data", {}).get("job_id")
        if not job_id:
            # Может быть в другом формате
            job_id = data.get("job_id")
    
    if not job_id:
        log("❌ Нет job_id в ответе!")
        log("Структура ответа:")
        log(f"  Keys: {list(data.keys()) if data else 'None'}")
        if data and "data" in data:
            log(f"  data keys: {list(data['data'].keys()) if isinstance(data.get('data'), dict) else type(data.get('data'))}")
    else:
        log(f"✅ Job ID: {job_id}")
    
    # =========================================================================
    # STEP 5: Polling статуса
    # =========================================================================
    if job_id:
        log("⏳ STEP 5: Polling статуса распознавания...")
        
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
            
            # Проверяем разные варианты структуры
            status_data = data.get("data", data)  # Может быть без обёртки
            status = status_data.get("status")
            progress = status_data.get("progress", 0)
            
            log(f"   Status: {status}, Progress: {progress}%")
            
            if status == "completed":
                log("✅ РАСПОЗНАВАНИЕ ЗАВЕРШЕНО!")
                result = status_data.get("result")
                if result:
                    log("📦 RESULT JSON:")
                    log(None, result)
                    
                    # Сохраним в файл для анализа
                    result_file = Path(__file__).parent / f"recognition_result_{TIMESTAMP}.json"
                    with open(result_file, "w", encoding="utf-8") as f:
                        json.dump(result, f, indent=2, ensure_ascii=False)
                    log(f"💾 Результат сохранён в: {result_file}")
                else:
                    log("⚠️ Статус completed, но нет result!")
                    log("Все ключи в status_data:")
                    log(None, list(status_data.keys()))
                break
            
            elif status == "failed":
                log("❌ РАСПОЗНАВАНИЕ FAILED!")
                error = status_data.get("error") or status_data.get("message")
                log(f"   Error: {error}")
                break
            
            elif status in ["processing", "pending", "queued"]:
                log(f"   Ждём 3 секунды...")
                time.sleep(3)
            
            else:
                log(f"⚠️ Неизвестный статус: {status}")
                time.sleep(3)
        
        else:
            log("⏰ Timeout! Распознавание не завершилось за отведённое время")
    
    # =========================================================================
    # STEP 6: Проверка сцены (если был workspace)
    # =========================================================================
    if workspace_id:
        log("🎮 STEP 6: Проверяем сцены в workspace...")
        
        resp = requests.get(
            f"{API_BASE}/workspaces/{workspace_id}/scenes",
            headers=headers
        )
        data = log_response(resp, f"GET /workspaces/{workspace_id}/scenes")
        
        scenes = []
        if data:
            scenes = data.get("data", {}).get("scenes", [])
            if not scenes:
                scenes = data.get("scenes", [])
        
        log(f"   Найдено сцен: {len(scenes)}")
        
        if scenes:
            scene_id = scenes[0].get("id")
            log(f"   Получаем детали сцены: {scene_id}")
            
            resp = requests.get(
                f"{API_BASE}/scenes/{scene_id}",
                headers=headers
            )
            data = log_response(resp, f"GET /scenes/{scene_id}")
        else:
            log("⚠️ Сцен нет. Пробуем создать...")
            
            if floor_plan_id:
                resp = requests.post(
                    f"{API_BASE}/workspaces/{workspace_id}/scenes",
                    headers={**headers, "Content-Type": "application/json"},
                    json={
                        "name": f"Test Scene {TIMESTAMP}",
                        "description": "Created for debug",
                        "floor_plan_id": floor_plan_id
                    }
                )
                data = log_response(resp, f"POST /workspaces/{workspace_id}/scenes")
                
                if resp.status_code in [200, 201] and data:
                    scene_id = data.get("data", {}).get("id")
                    if scene_id:
                        log(f"✅ Сцена создана: {scene_id}")
                        
                        # Получаем детали
                        resp = requests.get(
                            f"{API_BASE}/scenes/{scene_id}",
                            headers=headers
                        )
                        data = log_response(resp, f"GET /scenes/{scene_id}")
    
    # =========================================================================
    # STEP 7: Прямой тест AI Chat (для проверки контекста)
    # =========================================================================
    log("💬 STEP 7: Тест AI Chat...")
    
    chat_payload = {
        "message": "Привет! Это тест распознавания планировки.",
        "scene_id": ""  # Пустой для общего чата
    }
    
    resp = requests.post(
        f"{API_BASE}/ai/chat",
        headers={**headers, "Content-Type": "application/json"},
        json=chat_payload
    )
    data = log_response(resp, "POST /ai/chat")
    
    # =========================================================================
    # SUMMARY
    # =========================================================================
    log("\n" + "="*60)
    log("📊 ИТОГИ ТЕСТА")
    log("="*60)
    log(f"   Token: {'✅' if token else '❌'}")
    log(f"   Workspace: {'✅ ' + str(workspace_id)[:8] if workspace_id else '❌'}")
    log(f"   Floor Plan: {'✅ ' + str(floor_plan_id)[:8] if floor_plan_id else '❌'}")
    log(f"   Recognition Job: {'✅ ' + str(job_id)[:8] if job_id else '❌'}")

if __name__ == "__main__":
    main()


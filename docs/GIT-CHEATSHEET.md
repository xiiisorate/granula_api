# Git Cheatsheet для Granula

## 🎯 Быстрые команды

### Ежедневный workflow

```powershell
# Утром: получить изменения
git fetch origin
git pull origin dev/shared  # если работаете в shared

# Перед началом работы над новым сервисом
git checkout dev/shared
git pull origin dev/shared
git checkout -b dev/d1-МОЙ-СЕРВИС  # или dev/d2-...

# Во время работы: частые коммиты
git add .
git commit -m "feat(сервис): описание"
git push origin dev/d1-МОЙ-СЕРВИС

# Когда нужны изменения из shared
git fetch origin
git merge origin/dev/shared

# В конце дня: пуш всего
git push origin dev/d1-МОЙ-СЕРВИС
```

---

## 📝 Формат коммит-сообщений

```
<тип>(<область>): <описание>

Примеры:
feat(auth): implement user registration
feat(shared): add common proto types
fix(gateway): fix JWT validation
refactor(scene): simplify element updates
docs(readme): update quick start guide
chore(docker): update compose file
```

| Тип | Когда использовать |
|-----|-------------------|
| `feat` | Новая функциональность |
| `fix` | Исправление бага |
| `refactor` | Рефакторинг без изменения поведения |
| `docs` | Документация |
| `chore` | Служебные задачи (конфиги, зависимости) |
| `test` | Тесты |

---

## 👤 Developer 1 (Core) — Команды

### Создание веток

```powershell
# Auth Service
git checkout dev/shared && git pull
git checkout -b dev/d1-auth-service

# User Service
git checkout dev/shared && git pull
git checkout -b dev/d1-user-service

# Workspace Service
git checkout dev/shared && git pull
git checkout -b dev/d1-workspace-service

# Request Service
git checkout dev/shared && git pull
git checkout -b dev/d1-request-service

# Notification Service
git checkout dev/shared && git pull
git checkout -b dev/d1-notification-service

# API Gateway
git checkout dev/shared && git pull
git checkout -b dev/d1-api-gateway
```

### Типичный день

```powershell
# 1. Начало работы
git checkout dev/d1-auth-service
git fetch origin
git merge origin/dev/shared  # подтянуть изменения shared

# 2. Работа над фичей
# ... пишете код ...
git add auth-service/
git commit -m "feat(auth): implement login endpoint"

# 3. Еще фича
# ... пишете код ...
git add auth-service/
git commit -m "feat(auth): add JWT token generation"

# 4. Пуш
git push origin dev/d1-auth-service

# 5. Если нужно изменить shared
git checkout dev/shared
git pull origin dev/shared
# ... изменения в shared/ ...
git add shared/
git commit -m "feat(shared): add user.proto"
git push origin dev/shared

# 6. Вернуться к своему сервису
git checkout dev/d1-auth-service
git merge origin/dev/shared
git push origin dev/d1-auth-service
```

---

## 👤 Developer 2 (AI/3D) — Команды

### Создание веток

```powershell
# Compliance Service
git checkout dev/shared && git pull
git checkout -b dev/d2-compliance-service

# AI Service
git checkout dev/shared && git pull
git checkout -b dev/d2-ai-service

# Floor Plan Service
git checkout dev/shared && git pull
git checkout -b dev/d2-floor-plan-service

# Scene Service
git checkout dev/shared && git pull
git checkout -b dev/d2-scene-service

# Branch Service
git checkout dev/shared && git pull
git checkout -b dev/d2-branch-service
```

### Типичный день

```powershell
# 1. Начало работы
git checkout dev/d2-ai-service
git fetch origin
git merge origin/dev/shared

# 2. Работа
# ... пишете код ...
git add ai-service/
git commit -m "feat(ai): implement OpenRouter client"
git push origin dev/d2-ai-service

# 3. Нужен новый тип в proto
git checkout dev/shared
git pull origin dev/shared
# ... редактируете shared/proto/ai/v1/ai.proto ...
make proto  # перегенерировать
git add shared/
git commit -m "feat(shared): add streaming to ai.proto"
git push origin dev/shared

# 4. Вернуться
git checkout dev/d2-ai-service
git merge origin/dev/shared
```

---

## 🔀 Слияние веток

### Когда сервис готов

```powershell
# 1. Убедиться что всё запушено
git status  # должно быть чисто
git push origin dev/d1-auth-service

# 2. Merge в develop
git checkout develop
git pull origin develop
git merge dev/d1-auth-service
git push origin develop

# 3. Удалить ветку (опционально)
git branch -d dev/d1-auth-service
git push origin --delete dev/d1-auth-service
```

### В конце хакатона (все сервисы → main)

```powershell
# 1. Все merge в develop
git checkout develop
git pull origin develop

# 2. Merge develop в main
git checkout main
git pull origin main
git merge develop
git push origin main

# 3. Создать тег релиза
git tag -a v1.0.0 -m "Hackathon release"
git push origin v1.0.0
```

---

## ⚠️ Решение проблем

### Конфликт при merge

```powershell
git merge origin/dev/shared
# CONFLICT in shared/proto/common/v1/common.proto

# 1. Открыть файл, найти маркеры конфликта
# <<<<<<< HEAD
# ваш код
# =======
# их код
# >>>>>>> origin/dev/shared

# 2. Отредактировать, оставив нужное

# 3. Сохранить и продолжить
git add shared/proto/common/v1/common.proto
git commit -m "merge: resolve proto conflict"
```

### Отменить последний коммит (ещё не запушен)

```powershell
git reset --soft HEAD~1  # сохранить изменения
# или
git reset --hard HEAD~1  # удалить изменения
```

### Отменить изменения в файле

```powershell
git checkout -- path/to/file.go
```

### Посмотреть историю

```powershell
git log --oneline -20
git log --oneline --graph --all
```

### Посмотреть изменения

```powershell
git diff                    # незакоммиченные
git diff --staged           # staged (после git add)
git diff origin/dev/shared  # по сравнению с remote
```

---

## 📊 Визуализация веток

```
main
│
└── develop
    │
    ├── dev/shared ◄─────────────────────────────────┐
    │   │                                            │
    │   ├── D1: common.proto, auth.proto            │
    │   ├── D2: compliance.proto, ai.proto          │
    │   └── D1+D2: shared/pkg/*                      │
    │                                                │
    ├── dev/d1-auth-service ◄────────────────┐      │
    │   └── uses shared via replace          │      │
    │                                        │      │
    ├── dev/d1-user-service ◄───────────────┐│      │
    │                                       ││      │
    ├── dev/d1-workspace-service           ││      │
    │                                       ││      │
    ├── dev/d1-request-service             ││      │
    │                                       ││      │
    ├── dev/d1-notification-service        ││      │
    │                                       ││      │
    ├── dev/d1-api-gateway ◄────────────────┼┼──────┘
    │                                       ││
    ├── dev/d2-compliance-service ◄─────────┼┼──────┐
    │                                       ││      │
    ├── dev/d2-ai-service ◄─────────────────┼┼─────┐│
    │                                       ││     ││
    ├── dev/d2-floor-plan-service ◄────────┐││     ││
    │                                      │││     ││
    ├── dev/d2-scene-service ◄─────────────┼┼┼─────┼┘
    │                                      │││     │
    └── dev/d2-branch-service ◄────────────┴┴┴─────┘
```

---

## 🔔 Уведомления в Cursor

Cursor автоматически показывает:
- Синяя точка на Source Control = есть изменения
- Стрелки вверх/вниз = есть что пушить/пулить
- GitLens показывает автора каждой строки

**Горячие клавиши:**
- `Ctrl+Shift+G` — открыть Git панель
- `Ctrl+Enter` (в Git панели) — коммит
- `Ctrl+Shift+P` → "Git: Pull" — pull
- `Ctrl+Shift+P` → "Git: Push" — push


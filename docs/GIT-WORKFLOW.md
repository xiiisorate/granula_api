# Git Workflow для команды Granula

> **Команда:** 2 backend-разработчика  
> **IDE:** Cursor / VS Code  
> **Стратегия:** Feature Branches

---

## Содержание

1. [Структура веток](#структура-веток)
2. [Настройка окружения](#настройка-окружения)
3. [Начало работы](#начало-работы)
4. [Ежедневный workflow](#ежедневный-workflow)
5. [Синхронизация между разработчиками](#синхронизация-между-разработчиками)
6. [Разрешение конфликтов](#разрешение-конфликтов)
7. [Полезные команды](#полезные-команды)
8. [Чеклист](#чеклист)

---

## Структура веток

```
main (protected)
  │
  ├── develop                     ← Основная ветка разработки
  │     │
  │     ├── dev/shared            ← Общие библиотеки (proto, pkg)
  │     │
  │     ├── dev/d1-auth           ← Developer 1: Auth Service
  │     ├── dev/d1-user           ← Developer 1: User Service
  │     ├── dev/d1-workspace      ← Developer 1: Workspace Service
  │     ├── dev/d1-request        ← Developer 1: Request Service
  │     ├── dev/d1-notification   ← Developer 1: Notification Service
  │     ├── dev/d1-gateway        ← Developer 1: API Gateway
  │     │
  │     ├── dev/d2-compliance     ← Developer 2: Compliance Service
  │     ├── dev/d2-ai             ← Developer 2: AI Service
  │     ├── dev/d2-floor-plan     ← Developer 2: Floor Plan Service
  │     ├── dev/d2-scene          ← Developer 2: Scene Service
  │     └── dev/d2-branch         ← Developer 2: Branch Service
  │
  └── release/v1.0.0              ← Релизные ветки
```

### Правила именования веток

| Тип | Формат | Пример |
|-----|--------|--------|
| Shared | `dev/shared` | `dev/shared` |
| Feature D1 | `dev/d1-{service}` | `dev/d1-auth` |
| Feature D2 | `dev/d2-{service}` | `dev/d2-compliance` |
| Hotfix | `hotfix/{issue}` | `hotfix/jwt-validation` |
| Release | `release/v{version}` | `release/v1.0.0` |

---

## Настройка окружения

### 1. Установка расширений в Cursor/VS Code

Обязательные:
- **GitLens** — расширенная работа с Git
- **Git Graph** — визуализация веток

```bash
# Или через командную строку
code --install-extension eamodio.gitlens
code --install-extension mhutchie.git-graph
```

### 2. Настройка VS Code

Создайте файл `.vscode/settings.json` в корне проекта:

```json
{
  "git.autofetch": true,
  "git.autofetchPeriod": 60,
  "git.fetchOnPull": true,
  "git.pruneOnFetch": true,
  "git.confirmSync": false,
  "git.enableSmartCommit": true,
  "git.postCommitCommand": "none",
  "gitlens.hovers.currentLine.over": "line",
  "gitlens.codeLens.enabled": true,
  "gitlens.currentLine.enabled": true
}
```

### 3. Настройка Git

```bash
# Глобальные настройки (один раз)
git config --global user.name "Your Name"
git config --global user.email "your.email@example.com"
git config --global pull.rebase false
git config --global push.autoSetupRemote true
git config --global init.defaultBranch main

# Алиасы для удобства
git config --global alias.st "status -sb"
git config --global alias.co "checkout"
git config --global alias.br "branch"
git config --global alias.cm "commit -m"
git config --global alias.lg "log --oneline --graph --all"
git config --global alias.sync "!git fetch origin && git status"
```

### 4. SSH ключи (рекомендуется)

```bash
# Генерация SSH ключа
ssh-keygen -t ed25519 -C "your.email@example.com"

# Копирование публичного ключа
# Windows:
type %USERPROFILE%\.ssh\id_ed25519.pub | clip
# Linux/Mac:
cat ~/.ssh/id_ed25519.pub | pbcopy

# Добавьте ключ в GitHub: Settings → SSH and GPG keys → New SSH key
```

---

## Начало работы

### Шаг 1: Создание репозитория (Developer 1)

#### 🪟 Windows (PowerShell)

```powershell
# Инициализация
mkdir granula
cd granula
git init

# Запуск скрипта создания структуры
# (скопируйте scripts/init-project.ps1 и запустите)
.\scripts\init-project.ps1

# Или создайте структуру вручную:
# Первый коммит
git add .
git commit -m "chore: initial project structure"

# Создание remote (замените на ваш URL)
git remote add origin git@github.com:your-org/granula.git
git branch -M main
git push -u origin main

# Создание develop ветки
git checkout -b develop
git push -u origin develop
```

#### 🐧 Linux/macOS (Bash)

```bash
# Инициализация
mkdir granula && cd granula
git init

# Создание базовой структуры
mkdir -p shared/{proto,pkg,gen}
mkdir -p api-gateway/{cmd/server,internal}
mkdir -p auth-service/{cmd/server,internal,migrations}
mkdir -p user-service/{cmd/server,internal,migrations}
mkdir -p workspace-service/{cmd/server,internal,migrations}
mkdir -p floor-plan-service/{cmd/server,internal,migrations}
mkdir -p scene-service/{cmd/server,internal}
mkdir -p branch-service/{cmd/server,internal}
mkdir -p ai-service/{cmd/server,internal}
mkdir -p compliance-service/{cmd/server,internal,migrations}
mkdir -p request-service/{cmd/server,internal,migrations}
mkdir -p notification-service/{cmd/server,internal,migrations}
mkdir -p scripts docs .vscode

# Создание .gitignore
cat > .gitignore << 'EOF'
# Binaries
bin/
*.exe
*.dll
*.so
*.dylib

# Go
go.work
vendor/

# IDE
.idea/
*.swp
*.swo
.vscode/*
!.vscode/settings.json
!.vscode/extensions.json

# Environment
.env
.env.local
*.env

# OS
.DS_Store
Thumbs.db

# Logs
*.log
logs/

# Generated
shared/gen/

# Test
coverage.out
coverage.html

# Temp
tmp/
temp/

# Keys (NEVER commit!)
*.pem
*.key
ed25519
ed25519.pub
EOF

# Первый коммит
git add .
git commit -m "chore: initial project structure"

# Создание remote
git remote add origin git@github.com:your-org/granula.git
git push -u origin main

# Создание develop ветки
git checkout -b develop
git push -u origin develop
```

### Шаг 2: Клонирование (Developer 2)

```bash
git clone git@github.com:your-org/granula.git
cd granula
git checkout develop
```

### Шаг 3: Совместная работа над shared (Час 0-1)

**Developer 1 создаёт ветку:**

```bash
git checkout develop
git checkout -b dev/shared
```

**Developer 2 переключается на ту же ветку:**

```bash
git fetch origin
git checkout dev/shared
```

**Распределение задач в shared:**

| Developer 1 | Developer 2 |
|-------------|-------------|
| `shared/proto/common/v1/common.proto` | `shared/proto/scene/v1/scene.proto` |
| `shared/proto/auth/v1/auth.proto` | `shared/proto/branch/v1/branch.proto` |
| `shared/proto/user/v1/user.proto` | `shared/proto/ai/v1/ai.proto` |
| `shared/proto/workspace/v1/workspace.proto` | `shared/proto/compliance/v1/compliance.proto` |
| `shared/proto/floor_plan/v1/floor_plan.proto` | `shared/pkg/grpc/server.go` |
| `shared/proto/request/v1/request.proto` | `shared/pkg/grpc/client.go` |
| `shared/proto/notification/v1/notification.proto` | `shared/pkg/grpc/interceptors.go` |
| `shared/pkg/logger/logger.go` | |
| `shared/pkg/errors/errors.go` | |
| `shared/pkg/config/config.go` | |
| `shared/pkg/validator/validator.go` | |
| `docker-compose.yml` | |
| `scripts/init-databases.sql` | |
| `Makefile` | |

**Синхронизация во время работы над shared:**

```bash
# Перед началом работы
git pull origin dev/shared

# После каждого логического блока
git add .
git commit -m "feat(shared): add auth.proto"
git push origin dev/shared

# Получение изменений партнёра
git pull origin dev/shared
```

**Завершение shared:**

```bash
# После завершения всех задач в shared
git checkout develop
git merge dev/shared
git push origin develop
```

---

## Ежедневный workflow

### Утро: Начало работы

```bash
# 1. Обновить локальный репозиторий
git checkout develop
git pull origin develop

# 2. Получить все удалённые ветки
git fetch origin --prune

# 3. Переключиться на свою ветку или создать новую
git checkout dev/d1-auth
# или
git checkout -b dev/d1-auth
```

### В процессе работы

```bash
# Частые коммиты (каждые 30-60 минут)
git add .
git commit -m "feat(auth): implement user registration"

# Push в конце логического блока
git push origin dev/d1-auth
```

### Правила коммитов

Используйте [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

| Type | Когда использовать |
|------|-------------------|
| `feat` | Новая функциональность |
| `fix` | Исправление бага |
| `refactor` | Рефакторинг кода |
| `docs` | Изменения документации |
| `test` | Добавление/изменение тестов |
| `chore` | Обслуживание (deps, configs) |

**Примеры:**

```bash
git commit -m "feat(auth): implement JWT token generation"
git commit -m "fix(auth): handle expired refresh tokens"
git commit -m "refactor(user): extract avatar upload logic"
git commit -m "docs(api): add authentication endpoints"
git commit -m "test(auth): add unit tests for login"
git commit -m "chore(deps): update fiber to v2.52"
```

### Вечер: Завершение работы

```bash
# 1. Закоммитить все изменения
git add .
git commit -m "wip: auth service progress"

# 2. Push в свою ветку
git push origin dev/d1-auth

# 3. Обновить develop (если готов merge)
git checkout develop
git pull origin develop
git merge dev/d1-auth
git push origin develop
```

---

## Синхронизация между разработчиками

### Сценарий 1: Нужны изменения из shared

```bash
# Developer 2 нужны proto файлы от Developer 1

# Переключиться на свою ветку
git checkout dev/d2-compliance

# Получить обновления
git fetch origin

# Merge shared в свою ветку
git merge origin/dev/shared

# Или через develop
git merge origin/develop
```

### Сценарий 2: Нужен сервис партнёра

Например, AI Service (D2) нужен Auth Service (D1) для валидации токенов:

```bash
# Developer 2
git fetch origin
git checkout dev/d2-ai

# Проверить что auth-service готов
git log origin/dev/d1-auth --oneline -5

# Merge develop (где уже есть auth-service)
git merge origin/develop
```

### Сценарий 3: Срочный hotfix

```bash
# Кто-то нашёл баг в shared

# Создать hotfix ветку от develop
git checkout develop
git pull origin develop
git checkout -b hotfix/proto-validation

# Исправить
git add .
git commit -m "fix(shared): correct proto field types"
git push origin hotfix/proto-validation

# Merge в develop
git checkout develop
git merge hotfix/proto-validation
git push origin develop

# Уведомить партнёра!
# Партнёр делает:
git fetch origin
git merge origin/develop
```

### Автоматические уведомления

Настройте GitHub Webhooks → Telegram/Slack для уведомлений о push.

---

## Разрешение конфликтов

### В Cursor/VS Code

1. При возникновении конфликта VS Code покажет файлы с конфликтами
2. Откройте файл — увидите маркеры конфликта
3. Используйте кнопки:
   - **Accept Current Change** — оставить ваши изменения
   - **Accept Incoming Change** — взять изменения партнёра
   - **Accept Both Changes** — объединить оба
4. После разрешения:

```bash
git add .
git commit -m "merge: resolve conflicts in proto files"
```

### Типичные конфликты и решения

| Ситуация | Решение |
|----------|---------|
| Оба изменили `go.mod` | Принять оба, запустить `go mod tidy` |
| Оба изменили proto | Обсудить, объединить вручную |
| Конфликт в `docker-compose.yml` | Обычно Accept Both, проверить порты |
| Конфликт в `Makefile` | Accept Both, проверить дубли команд |

### Предотвращение конфликтов

1. **Чёткое разделение файлов** — каждый работает только со своими сервисами
2. **Частые pull** — `git pull origin develop` минимум 2 раза в день
3. **Маленькие коммиты** — легче разрешать конфликты
4. **Коммуникация** — предупреждайте если меняете shared файлы

---

## Полезные команды

### Просмотр состояния

```bash
# Короткий статус
git status -sb

# История коммитов (граф)
git log --oneline --graph --all

# Что изменилось в файле
git diff path/to/file

# Кто изменял файл
git blame path/to/file

# Список веток
git branch -a

# Какие файлы изменены между ветками
git diff develop..dev/d1-auth --name-only
```

### Работа с ветками

```bash
# Создать и переключиться
git checkout -b dev/d1-auth

# Переключиться на существующую
git checkout dev/d1-auth

# Удалить локальную ветку
git branch -d dev/d1-auth

# Удалить remote ветку
git push origin --delete dev/d1-auth

# Переименовать ветку
git branch -m old-name new-name
```

### Отмена изменений

```bash
# Отменить изменения в файле (до commit)
git checkout -- path/to/file

# Отменить все изменения (до commit)
git checkout -- .

# Отменить последний коммит (сохранить изменения)
git reset --soft HEAD~1

# Отменить последний коммит (удалить изменения)
git reset --hard HEAD~1

# Отменить push (создаёт новый коммит)
git revert HEAD
git push
```

### Stash (временное сохранение)

```bash
# Сохранить изменения
git stash push -m "work in progress on auth"

# Список stash
git stash list

# Применить последний stash
git stash pop

# Применить конкретный stash
git stash apply stash@{0}

# Удалить stash
git stash drop stash@{0}
```

### Синхронизация

```bash
# Получить все обновления (без merge)
git fetch origin

# Получить и merge текущую ветку
git pull origin

# Принудительный push (осторожно!)
git push --force-with-lease origin dev/d1-auth
```

---

## Чеклист

### Перед началом работы

- [ ] `git fetch origin` — получить обновления
- [ ] `git checkout develop && git pull` — обновить develop
- [ ] `git checkout dev/d1-*` — переключиться на свою ветку
- [ ] `git merge develop` — синхронизировать с develop

### Во время работы

- [ ] Коммиты каждые 30-60 минут
- [ ] Осмысленные сообщения коммитов
- [ ] Push минимум 2 раза в день
- [ ] `git fetch origin` перед важными merge

### Перед merge в develop

- [ ] Все тесты проходят (`make test`)
- [ ] Линтер не ругается (`make lint`)
- [ ] Код компилируется (`make build-service SERVICE=...`)
- [ ] Обновлён из develop (`git merge origin/develop`)
- [ ] Нет конфликтов

### В конце дня

- [ ] Все изменения закоммичены
- [ ] Всё запушено в remote
- [ ] Если сервис готов — merge в develop
- [ ] Сообщить партнёру о важных изменениях

---

## Горячие клавиши в Cursor/VS Code

| Действие | Windows | Mac |
|----------|---------|-----|
| Открыть Git панель | `Ctrl+Shift+G` | `Cmd+Shift+G` |
| Открыть терминал | `` Ctrl+` `` | `` Cmd+` `` |
| Command Palette | `Ctrl+Shift+P` | `Cmd+Shift+P` |
| Поиск по файлам | `Ctrl+P` | `Cmd+P` |
| GitLens: File History | `Alt+H` | `Option+H` |
| GitLens: Line History | `Alt+Shift+H` | `Option+Shift+H` |

### Git команды через Command Palette

1. Нажмите `Ctrl+Shift+P`
2. Введите "Git:"
3. Выберите нужную команду:
   - `Git: Pull`
   - `Git: Push`
   - `Git: Fetch`
   - `Git: Checkout to...`
   - `Git: Create Branch...`
   - `Git: Merge Branch...`

---

## Визуализация workflow

```
Час 0-1: Shared (совместно)
─────────────────────────────────────────────────────────────────────
                    dev/shared
                        │
        D1 ────────────►├──────────────► D2
        (proto, pkg)    │                (grpc, proto)
                        │
                        ▼
                    develop
                        │
─────────────────────────────────────────────────────────────────────

Час 1+: Раздельная работа
─────────────────────────────────────────────────────────────────────
                    develop
                        │
        ┌───────────────┼───────────────┐
        │               │               │
        ▼               │               ▼
    dev/d1-auth         │        dev/d2-compliance
        │               │               │
        ▼               │               ▼
    dev/d1-user         │        dev/d2-ai
        │               │               │
        ▼               │               ▼
       ...              │              ...
        │               │               │
        └───────────────┼───────────────┘
                        │
                        ▼
                    develop (merge)
                        │
                        ▼
                      main (release)
─────────────────────────────────────────────────────────────────────
```

---

## Контакты и помощь

- **Git документация:** https://git-scm.com/doc
- **GitHub Docs:** https://docs.github.com
- **GitLens:** https://gitlens.amod.io

При возникновении проблем:
1. `git status` — посмотреть состояние
2. `git log --oneline -10` — последние коммиты
3. `git reflog` — история всех действий (для восстановления)


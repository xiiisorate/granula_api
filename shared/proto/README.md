# Proto файлы - Асинхронная работа

## Структура

Все proto файлы находятся в `shared/proto/` и разделены по сервисам:

```
shared/proto/
├── common/v1/common.proto          # Общие типы (базовый файл)
├── auth/v1/auth.proto              # D1: Auth Service
├── user/v1/user.proto              # D1: User Service
├── workspace/v1/workspace.proto    # D1: Workspace Service
├── floor_plan/v1/floor_plan.proto  # D1: Floor Plan Service
├── request/v1/request.proto        # D1: Request Service
├── notification/v1/notification.proto  # D1: Notification Service
├── scene/v1/scene.proto            # D2: Scene Service
├── branch/v1/branch.proto          # D2: Branch Service
├── ai/v1/ai.proto                  # D2: AI Service
└── compliance/v1/compliance.proto  # D2: Compliance Service
```

## Распределение задач

### Developer 1 (D1)
- ✅ `common/v1/common.proto` - базовые типы
- ✅ `auth/v1/auth.proto` - сервис аутентификации
- ✅ `user/v1/user.proto` - сервис пользователей
- ✅ `workspace/v1/workspace.proto` - сервис воркспейсов
- ✅ `floor_plan/v1/floor_plan.proto` - сервис планировок
- ✅ `request/v1/request.proto` - сервис заявок
- ✅ `notification/v1/notification.proto` - сервис уведомлений

### Developer 2 (D2)
- ✅ `scene/v1/scene.proto` - сервис сцен
- ✅ `branch/v1/branch.proto` - сервис веток
- ✅ `ai/v1/ai.proto` - сервис AI
- ✅ `compliance/v1/compliance.proto` - сервис compliance

## Правила работы

### 1. Все работают в ветке `dev/shared`

```bash
git checkout dev/shared
git pull origin dev/shared
```

### 2. Каждый работает со своими файлами

- **D1** работает только с файлами: `common`, `auth`, `user`, `workspace`, `floor_plan`, `request`, `notification`
- **D2** работает только с файлами: `scene`, `branch`, `ai`, `compliance`

### 3. Частые коммиты и синхронизация

```bash
# После каждого логического блока (каждые 30-60 минут)
git add shared/proto/your-service/v1/your-service.proto
git commit -m "feat(shared): update auth.proto - add OAuth methods"
git push origin dev/shared

# Перед началом работы - получить изменения партнёра
git pull origin dev/shared
```

### 4. Использование common.proto

Все proto файлы импортируют `common/v1/common.proto`:

```protobuf
import "common/v1/common.proto";
```

Если нужно добавить общий тип:
1. Добавьте в `common/v1/common.proto`
2. Закоммитьте и запушьте
3. Уведомите партнёра

### 5. Избегайте конфликтов

- **Не редактируйте** файлы партнёра
- **Не изменяйте** `common.proto` без согласования
- **Используйте** разные поля/методы в своих сервисах

## Генерация Go кода

После изменения proto файлов:

```bash
make proto
```

Или вручную:

```bash
find shared/proto -name "*.proto" -exec protoc \
  --go_out=shared/gen --go_opt=paths=source_relative \
  --go-grpc_out=shared/gen --go-grpc_opt=paths=source_relative \
  -I shared/proto {} \;
```

## Проверка перед коммитом

1. ✅ Proto файлы компилируются без ошибок
2. ✅ Go код генерируется успешно (`make proto`)
3. ✅ Нет конфликтов с изменениями партнёра (`git pull`)
4. ✅ Сообщение коммита следует [Conventional Commits](https://www.conventionalcommits.org/)

## Примеры коммитов

```bash
git commit -m "feat(shared): add Register and Login methods to auth.proto"
git commit -m "feat(shared): add streaming support to ai.proto"
git commit -m "fix(shared): correct field types in workspace.proto"
git commit -m "refactor(shared): reorganize messages in scene.proto"
```

## Завершение работы над shared

Когда все proto файлы готовы:

```bash
# 1. Убедитесь, что все изменения закоммичены
git status

# 2. Синхронизируйтесь с партнёром
git pull origin dev/shared

# 3. Сгенерируйте Go код
make proto

# 4. Merge в develop
git checkout develop
git merge dev/shared
git push origin develop
```

## Полезные команды

```bash
# Посмотреть изменения в proto файлах
git diff shared/proto/

# Посмотреть историю конкретного файла
git log --oneline shared/proto/auth/v1/auth.proto

# Проверить статус
git status -sb

# Посмотреть кто что редактирует
git log --all --oneline --graph --decorate
```


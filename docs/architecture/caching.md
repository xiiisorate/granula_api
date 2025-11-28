# Стратегии кэширования

## Обзор

Granula API использует многоуровневое кэширование для оптимизации производительности и снижения нагрузки на базы данных.

## Архитектура кэширования

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Client    │────►│   API       │────►│   Redis     │
└─────────────┘     └──────┬──────┘     └──────┬──────┘
                           │                    │
                           │                    │ Cache Miss
                           │                    ▼
                           │            ┌─────────────┐
                           │            │ PostgreSQL  │
                           │            │   MongoDB   │
                           │            └─────────────┘
                           │                    │
                           │◄───────────────────┘
                           │
                    ┌──────┴──────┐
                    │  Response   │
                    └─────────────┘
```

## Типы кэшей

### 1. Кэш сущностей (Entity Cache)

```go
// internal/repository/redis/cache.go

// EntityCache кэш доменных сущностей.
type EntityCache struct {
    client *redis.Client
    logger *zap.Logger
}

// CacheConfig конфигурация кэша.
type CacheConfig struct {
    // TTL по типам сущностей
    UserTTL      time.Duration // 15 min
    WorkspaceTTL time.Duration // 10 min
    SceneTTL     time.Duration // 5 min
    
    // Размеры для LRU eviction
    MaxKeys int // Максимум ключей (по умолчанию без лимита в Redis)
}

// Ключи кэша
const (
    // Пользователи
    // user:cache:{user_id} -> JSON
    keyUserCache = "user:cache:%s"
    
    // Воркспейсы
    // workspace:cache:{workspace_id} -> JSON
    keyWorkspaceCache = "workspace:cache:%s"
    
    // Сцены (полные данные в MongoDB, кэшируем метаданные)
    // scene:meta:{scene_id} -> JSON
    keySceneMeta = "scene:meta:%s"
    
    // Списки
    // user:workspaces:{user_id} -> JSON array of IDs
    keyUserWorkspaces = "user:workspaces:%s"
)

// GetUser получает пользователя из кэша.
func (c *EntityCache) GetUser(ctx context.Context, id string) (*entity.User, error) {
    key := fmt.Sprintf(keyUserCache, id)
    
    data, err := c.client.Get(ctx, key).Bytes()
    if err != nil {
        if errors.Is(err, redis.Nil) {
            return nil, ErrCacheMiss
        }
        return nil, fmt.Errorf("redis get: %w", err)
    }
    
    var user entity.User
    if err := json.Unmarshal(data, &user); err != nil {
        // Инвалидируем битые данные
        c.client.Del(ctx, key)
        return nil, ErrCacheMiss
    }
    
    return &user, nil
}

// SetUser сохраняет пользователя в кэш.
func (c *EntityCache) SetUser(ctx context.Context, user *entity.User) error {
    key := fmt.Sprintf(keyUserCache, user.ID)
    
    data, err := json.Marshal(user)
    if err != nil {
        return fmt.Errorf("marshal user: %w", err)
    }
    
    return c.client.Set(ctx, key, data, c.config.UserTTL).Err()
}

// InvalidateUser удаляет пользователя из кэша.
func (c *EntityCache) InvalidateUser(ctx context.Context, id string) error {
    keys := []string{
        fmt.Sprintf(keyUserCache, id),
        fmt.Sprintf(keyUserWorkspaces, id),
    }
    return c.client.Del(ctx, keys...).Err()
}
```

### 2. Кэш запросов (Query Cache)

```go
// internal/repository/redis/query_cache.go

// QueryCache кэш результатов запросов.
type QueryCache struct {
    client *redis.Client
    hasher hash.Hash
}

// CacheKey генерирует ключ кэша из параметров запроса.
func (c *QueryCache) CacheKey(prefix string, params interface{}) string {
    data, _ := json.Marshal(params)
    hash := sha256.Sum256(data)
    return fmt.Sprintf("query:%s:%x", prefix, hash[:8])
}

// GetList получает закэшированный список.
func (c *QueryCache) GetList(
    ctx context.Context,
    key string,
    dest interface{},
) error {
    data, err := c.client.Get(ctx, key).Bytes()
    if err != nil {
        if errors.Is(err, redis.Nil) {
            return ErrCacheMiss
        }
        return err
    }
    
    return json.Unmarshal(data, dest)
}

// SetList кэширует список с TTL.
func (c *QueryCache) SetList(
    ctx context.Context,
    key string,
    data interface{},
    ttl time.Duration,
) error {
    bytes, err := json.Marshal(data)
    if err != nil {
        return err
    }
    
    return c.client.Set(ctx, key, bytes, ttl).Err()
}

// Пример использования в репозитории
func (r *WorkspaceRepository) List(
    ctx context.Context,
    userID string,
    params ListParams,
) ([]entity.Workspace, error) {
    // Формируем ключ кэша
    cacheKey := r.queryCache.CacheKey("workspaces", struct {
        UserID string
        Params ListParams
    }{userID, params})
    
    // Пробуем получить из кэша
    var workspaces []entity.Workspace
    err := r.queryCache.GetList(ctx, cacheKey, &workspaces)
    if err == nil {
        return workspaces, nil
    }
    
    // Cache miss - идём в БД
    workspaces, err = r.listFromDB(ctx, userID, params)
    if err != nil {
        return nil, err
    }
    
    // Сохраняем в кэш
    _ = r.queryCache.SetList(ctx, cacheKey, workspaces, 5*time.Minute)
    
    return workspaces, nil
}
```

### 3. Кэш сессий

```go
// internal/repository/redis/session.go

// SessionCache кэш сессий пользователей.
type SessionCache struct {
    client *redis.Client
}

// Session данные сессии.
type Session struct {
    UserID    string    `json:"user_id"`
    DeviceID  string    `json:"device_id"`
    IP        string    `json:"ip"`
    UserAgent string    `json:"user_agent"`
    CreatedAt time.Time `json:"created_at"`
}

// Create создаёт новую сессию.
func (c *SessionCache) Create(
    ctx context.Context,
    sessionID string,
    session *Session,
    ttl time.Duration,
) error {
    key := fmt.Sprintf("session:%s", sessionID)
    
    data, err := json.Marshal(session)
    if err != nil {
        return err
    }
    
    pipe := c.client.Pipeline()
    
    // Сохраняем сессию
    pipe.Set(ctx, key, data, ttl)
    
    // Добавляем в список сессий пользователя
    userSessionsKey := fmt.Sprintf("user:sessions:%s", session.UserID)
    pipe.SAdd(ctx, userSessionsKey, sessionID)
    pipe.Expire(ctx, userSessionsKey, ttl)
    
    _, err = pipe.Exec(ctx)
    return err
}

// Get получает сессию.
func (c *SessionCache) Get(ctx context.Context, sessionID string) (*Session, error) {
    key := fmt.Sprintf("session:%s", sessionID)
    
    data, err := c.client.Get(ctx, key).Bytes()
    if err != nil {
        if errors.Is(err, redis.Nil) {
            return nil, ErrSessionNotFound
        }
        return nil, err
    }
    
    var session Session
    if err := json.Unmarshal(data, &session); err != nil {
        return nil, err
    }
    
    return &session, nil
}

// Delete удаляет сессию.
func (c *SessionCache) Delete(ctx context.Context, sessionID string) error {
    key := fmt.Sprintf("session:%s", sessionID)
    
    // Получаем сессию для user_id
    session, err := c.Get(ctx, sessionID)
    if err != nil && !errors.Is(err, ErrSessionNotFound) {
        return err
    }
    
    pipe := c.client.Pipeline()
    pipe.Del(ctx, key)
    
    if session != nil {
        userSessionsKey := fmt.Sprintf("user:sessions:%s", session.UserID)
        pipe.SRem(ctx, userSessionsKey, sessionID)
    }
    
    _, err = pipe.Exec(ctx)
    return err
}

// DeleteAllForUser удаляет все сессии пользователя.
func (c *SessionCache) DeleteAllForUser(ctx context.Context, userID string) error {
    userSessionsKey := fmt.Sprintf("user:sessions:%s", userID)
    
    // Получаем все session IDs
    sessionIDs, err := c.client.SMembers(ctx, userSessionsKey).Result()
    if err != nil {
        return err
    }
    
    if len(sessionIDs) == 0 {
        return nil
    }
    
    // Формируем ключи для удаления
    keys := make([]string, len(sessionIDs)+1)
    for i, id := range sessionIDs {
        keys[i] = fmt.Sprintf("session:%s", id)
    }
    keys[len(sessionIDs)] = userSessionsKey
    
    return c.client.Del(ctx, keys...).Err()
}
```

## Стратегии инвалидации

### 1. Time-Based (TTL)

```go
// Автоматическая инвалидация по времени
const (
    // Короткий TTL для часто меняющихся данных
    ShortTTL = 5 * time.Minute
    
    // Средний TTL для относительно стабильных данных
    MediumTTL = 15 * time.Minute
    
    // Длинный TTL для редко меняющихся данных
    LongTTL = 1 * time.Hour
)

// TTL по типам данных
var CacheTTL = map[string]time.Duration{
    "user":           MediumTTL,  // 15 min
    "workspace":      MediumTTL,  // 15 min
    "workspace_list": ShortTTL,   // 5 min
    "scene_meta":     ShortTTL,   // 5 min
    "compliance":     LongTTL,    // 1 hour (редко меняются правила)
}
```

### 2. Event-Based (Pub/Sub)

```go
// internal/service/cache/invalidator.go

// CacheInvalidator инвалидатор кэша по событиям.
type CacheInvalidator struct {
    cache  *EntityCache
    pubsub *redis.PubSub
    logger *zap.Logger
}

// Events каналы событий.
const (
    ChannelUserUpdated      = "cache:user:updated"
    ChannelWorkspaceUpdated = "cache:workspace:updated"
    ChannelSceneUpdated     = "cache:scene:updated"
)

// InvalidationEvent событие инвалидации.
type InvalidationEvent struct {
    EntityType string   `json:"type"`
    EntityIDs  []string `json:"ids"`
    Timestamp  int64    `json:"ts"`
}

// Start запускает слушатель событий.
func (i *CacheInvalidator) Start(ctx context.Context) error {
    i.pubsub = i.cache.client.Subscribe(ctx,
        ChannelUserUpdated,
        ChannelWorkspaceUpdated,
        ChannelSceneUpdated,
    )
    
    ch := i.pubsub.Channel()
    
    go func() {
        for msg := range ch {
            var event InvalidationEvent
            if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
                i.logger.Error("Invalid invalidation event", zap.Error(err))
                continue
            }
            
            if err := i.handleEvent(ctx, &event); err != nil {
                i.logger.Error("Failed to handle invalidation",
                    zap.String("type", event.EntityType),
                    zap.Error(err),
                )
            }
        }
    }()
    
    return nil
}

// PublishInvalidation публикует событие инвалидации.
func (i *CacheInvalidator) PublishInvalidation(
    ctx context.Context,
    entityType string,
    ids ...string,
) error {
    event := InvalidationEvent{
        EntityType: entityType,
        EntityIDs:  ids,
        Timestamp:  time.Now().UnixMilli(),
    }
    
    data, _ := json.Marshal(event)
    
    channel := fmt.Sprintf("cache:%s:updated", entityType)
    return i.cache.client.Publish(ctx, channel, data).Err()
}

// Использование в сервисе
func (s *UserService) Update(ctx context.Context, user *entity.User) error {
    // Обновляем в БД
    if err := s.repo.Update(ctx, user); err != nil {
        return err
    }
    
    // Публикуем инвалидацию
    return s.invalidator.PublishInvalidation(ctx, "user", user.ID.String())
}
```

### 3. Write-Through Cache

```go
// internal/repository/cached/workspace.go

// CachedWorkspaceRepository репозиторий с write-through кэшированием.
type CachedWorkspaceRepository struct {
    repo  repository.WorkspaceRepository
    cache *EntityCache
}

// Create создаёт воркспейс и сразу кэширует.
func (r *CachedWorkspaceRepository) Create(
    ctx context.Context,
    workspace *entity.Workspace,
) error {
    // Сначала сохраняем в БД
    if err := r.repo.Create(ctx, workspace); err != nil {
        return err
    }
    
    // Сразу кэшируем (write-through)
    if err := r.cache.SetWorkspace(ctx, workspace); err != nil {
        // Логируем, но не фейлим операцию
        r.logger.Warn("Failed to cache workspace",
            zap.String("id", workspace.ID.String()),
            zap.Error(err),
        )
    }
    
    // Инвалидируем списки
    r.cache.InvalidateUserWorkspaces(ctx, workspace.OwnerID.String())
    
    return nil
}

// GetByID получает с кэша или БД.
func (r *CachedWorkspaceRepository) GetByID(
    ctx context.Context,
    id uuid.UUID,
) (*entity.Workspace, error) {
    // Пробуем кэш
    workspace, err := r.cache.GetWorkspace(ctx, id.String())
    if err == nil {
        return workspace, nil
    }
    
    // Cache miss - идём в БД
    workspace, err = r.repo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }
    
    // Кэшируем результат
    _ = r.cache.SetWorkspace(ctx, workspace)
    
    return workspace, nil
}
```

## Cache Aside Pattern

```go
// internal/pkg/cache/cache_aside.go

// CacheAside реализует паттерн cache-aside.
type CacheAside[T any] struct {
    cache   Cache
    loader  func(ctx context.Context, key string) (T, error)
    ttl     time.Duration
    metrics *CacheMetrics
}

// Get получает значение из кэша или загружает.
func (c *CacheAside[T]) Get(ctx context.Context, key string) (T, error) {
    var result T
    
    // Попытка получить из кэша
    data, err := c.cache.Get(ctx, key)
    if err == nil {
        if err := json.Unmarshal(data, &result); err == nil {
            c.metrics.Hits.Inc()
            return result, nil
        }
    }
    
    c.metrics.Misses.Inc()
    
    // Загрузка из источника
    result, err = c.loader(ctx, key)
    if err != nil {
        return result, err
    }
    
    // Сохранение в кэш (асинхронно)
    go func() {
        data, _ := json.Marshal(result)
        _ = c.cache.Set(context.Background(), key, data, c.ttl)
    }()
    
    return result, nil
}

// Пример использования
func NewUserCacheAside(cache Cache, repo UserRepository) *CacheAside[*entity.User] {
    return &CacheAside[*entity.User]{
        cache: cache,
        loader: func(ctx context.Context, key string) (*entity.User, error) {
            id, _ := uuid.Parse(key)
            return repo.GetByID(ctx, id)
        },
        ttl: 15 * time.Minute,
    }
}
```

## Мониторинг кэша

```go
// internal/pkg/cache/metrics.go

// CacheMetrics метрики кэша.
type CacheMetrics struct {
    Hits        prometheus.Counter
    Misses      prometheus.Counter
    Evictions   prometheus.Counter
    Size        prometheus.Gauge
    Latency     prometheus.Histogram
}

func NewCacheMetrics(name string) *CacheMetrics {
    return &CacheMetrics{
        Hits: prometheus.NewCounter(prometheus.CounterOpts{
            Name: fmt.Sprintf("granula_cache_%s_hits_total", name),
            Help: "Total cache hits",
        }),
        Misses: prometheus.NewCounter(prometheus.CounterOpts{
            Name: fmt.Sprintf("granula_cache_%s_misses_total", name),
            Help: "Total cache misses",
        }),
        Evictions: prometheus.NewCounter(prometheus.CounterOpts{
            Name: fmt.Sprintf("granula_cache_%s_evictions_total", name),
            Help: "Total cache evictions",
        }),
        Size: prometheus.NewGauge(prometheus.GaugeOpts{
            Name: fmt.Sprintf("granula_cache_%s_size", name),
            Help: "Current cache size",
        }),
        Latency: prometheus.NewHistogram(prometheus.HistogramOpts{
            Name:    fmt.Sprintf("granula_cache_%s_latency_seconds", name),
            Help:    "Cache operation latency",
            Buckets: []float64{.0001, .0005, .001, .005, .01, .05, .1},
        }),
    }
}

// Hit Rate = Hits / (Hits + Misses)
// Цель: > 90% для entity cache, > 80% для query cache
```

## Best Practices

### 1. Консистентность

```go
// Используем транзакции Redis для атомарности
func (c *EntityCache) UpdateUserWithWorkspaces(
    ctx context.Context,
    user *entity.User,
    workspaceIDs []string,
) error {
    pipe := c.client.TxPipeline()
    
    // Обновляем пользователя
    userData, _ := json.Marshal(user)
    pipe.Set(ctx, fmt.Sprintf(keyUserCache, user.ID), userData, c.config.UserTTL)
    
    // Обновляем список воркспейсов
    wsData, _ := json.Marshal(workspaceIDs)
    pipe.Set(ctx, fmt.Sprintf(keyUserWorkspaces, user.ID), wsData, c.config.WorkspaceTTL)
    
    _, err := pipe.Exec(ctx)
    return err
}
```

### 2. Защита от cache stampede

```go
// internal/pkg/cache/singleflight.go

// SingleFlightCache предотвращает cache stampede.
type SingleFlightCache struct {
    cache  Cache
    group  singleflight.Group
    loader func(ctx context.Context, key string) (interface{}, error)
}

// Get загружает данные с защитой от stampede.
func (c *SingleFlightCache) Get(ctx context.Context, key string) (interface{}, error) {
    // Пробуем кэш
    data, err := c.cache.Get(ctx, key)
    if err == nil {
        return data, nil
    }
    
    // Используем singleflight для предотвращения множественных запросов
    result, err, _ := c.group.Do(key, func() (interface{}, error) {
        // Ещё раз проверяем кэш (мог быть загружен другой горутиной)
        if data, err := c.cache.Get(ctx, key); err == nil {
            return data, nil
        }
        
        // Загружаем из источника
        return c.loader(ctx, key)
    })
    
    if err != nil {
        return nil, err
    }
    
    // Сохраняем в кэш
    _ = c.cache.Set(ctx, key, result, 0)
    
    return result, nil
}
```

### 3. Graceful degradation

```go
// При ошибках Redis продолжаем работу без кэша
func (r *CachedRepository) GetByID(ctx context.Context, id string) (*Entity, error) {
    // Пробуем кэш с таймаутом
    cacheCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
    defer cancel()
    
    cached, err := r.cache.Get(cacheCtx, id)
    if err == nil {
        return cached, nil
    }
    
    // При любой ошибке (timeout, unavailable) идём в БД
    // Не блокируем пользователя из-за проблем с кэшем
    return r.repo.GetByID(ctx, id)
}
```


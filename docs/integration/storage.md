# Интеграция с файловым хранилищем

## Обзор

Granula использует S3-совместимое хранилище (MinIO для self-hosted, AWS S3 для production) для:
- Планировок (PDF, изображения)
- Превью и рендеры
- Экспорты документов
- Аватары пользователей
- 3D модели мебели

## Конфигурация

```env
# Storage
S3_ENDPOINT=localhost:9000
S3_ACCESS_KEY=minioadmin
S3_SECRET_KEY=minioadmin
S3_BUCKET=granula
S3_REGION=us-east-1
S3_USE_SSL=false

# Presigned URLs
S3_PRESIGN_EXPIRY=1h
S3_UPLOAD_MAX_SIZE=52428800  # 50MB
```

## Архитектура

```
┌─────────────────────────────────────────────────────────────┐
│                      Granula API                             │
│  ┌─────────────────────────────────────────────────────────┐│
│  │                   Storage Service                        ││
│  │  ┌────────────┐  ┌────────────┐  ┌────────────┐        ││
│  │  │ FloorPlan  │  │   Render   │  │   Avatar   │        ││
│  │  │  Storage   │  │  Storage   │  │  Storage   │        ││
│  │  └─────┬──────┘  └─────┬──────┘  └─────┬──────┘        ││
│  │        │               │               │                ││
│  │        └───────────────┼───────────────┘                ││
│  │                        │                                ││
│  │                 ┌──────┴──────┐                         ││
│  │                 │  S3 Client  │                         ││
│  │                 └──────┬──────┘                         ││
│  └────────────────────────┼────────────────────────────────┘│
└───────────────────────────┼─────────────────────────────────┘
                            │
                            ▼
              ┌─────────────────────────┐
              │    MinIO / AWS S3       │
              │                         │
              │  granula/               │
              │  ├── floor-plans/       │
              │  ├── renders/           │
              │  ├── avatars/           │
              │  ├── models/            │
              │  └── exports/           │
              └─────────────────────────┘
```

## S3 клиент

```go
// internal/integration/storage/client.go

// Client клиент S3 хранилища.
type Client struct {
    s3          *s3.Client
    presigner   *s3.PresignClient
    bucket      string
    endpoint    string
    useSSL      bool
    presignTTL  time.Duration
    logger      *zap.Logger
}

// Config конфигурация клиента.
type Config struct {
    Endpoint     string        `env:"S3_ENDPOINT,required"`
    AccessKey    string        `env:"S3_ACCESS_KEY,required"`
    SecretKey    string        `env:"S3_SECRET_KEY,required"`
    Bucket       string        `env:"S3_BUCKET,required"`
    Region       string        `env:"S3_REGION" envDefault:"us-east-1"`
    UseSSL       bool          `env:"S3_USE_SSL" envDefault:"false"`
    PresignTTL   time.Duration `env:"S3_PRESIGN_EXPIRY" envDefault:"1h"`
    MaxUploadSize int64        `env:"S3_UPLOAD_MAX_SIZE" envDefault:"52428800"`
}

// NewClient создаёт новый клиент.
func NewClient(cfg *Config, logger *zap.Logger) (*Client, error) {
    resolver := aws.EndpointResolverWithOptionsFunc(
        func(service, region string, options ...interface{}) (aws.Endpoint, error) {
            return aws.Endpoint{
                URL:               buildEndpointURL(cfg.Endpoint, cfg.UseSSL),
                HostnameImmutable: true,
            }, nil
        },
    )
    
    awsCfg, err := config.LoadDefaultConfig(context.Background(),
        config.WithRegion(cfg.Region),
        config.WithEndpointResolverWithOptions(resolver),
        config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
            cfg.AccessKey,
            cfg.SecretKey,
            "",
        )),
    )
    if err != nil {
        return nil, fmt.Errorf("load aws config: %w", err)
    }
    
    s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
        o.UsePathStyle = true // Для MinIO
    })
    
    return &Client{
        s3:         s3Client,
        presigner:  s3.NewPresignClient(s3Client),
        bucket:     cfg.Bucket,
        endpoint:   cfg.Endpoint,
        useSSL:     cfg.UseSSL,
        presignTTL: cfg.PresignTTL,
        logger:     logger,
    }, nil
}

// Upload загружает файл в хранилище.
func (c *Client) Upload(ctx context.Context, key string, data io.Reader, contentType string) error {
    _, err := c.s3.PutObject(ctx, &s3.PutObjectInput{
        Bucket:      aws.String(c.bucket),
        Key:         aws.String(key),
        Body:        data,
        ContentType: aws.String(contentType),
    })
    if err != nil {
        return fmt.Errorf("put object: %w", err)
    }
    
    c.logger.Debug("File uploaded",
        zap.String("bucket", c.bucket),
        zap.String("key", key),
    )
    
    return nil
}

// Download скачивает файл.
func (c *Client) Download(ctx context.Context, key string) (io.ReadCloser, error) {
    result, err := c.s3.GetObject(ctx, &s3.GetObjectInput{
        Bucket: aws.String(c.bucket),
        Key:    aws.String(key),
    })
    if err != nil {
        var nsk *types.NoSuchKey
        if errors.As(err, &nsk) {
            return nil, ErrNotFound
        }
        return nil, fmt.Errorf("get object: %w", err)
    }
    
    return result.Body, nil
}

// Delete удаляет файл.
func (c *Client) Delete(ctx context.Context, key string) error {
    _, err := c.s3.DeleteObject(ctx, &s3.DeleteObjectInput{
        Bucket: aws.String(c.bucket),
        Key:    aws.String(key),
    })
    if err != nil {
        return fmt.Errorf("delete object: %w", err)
    }
    
    return nil
}

// GetPresignedURL генерирует presigned URL для скачивания.
func (c *Client) GetPresignedURL(ctx context.Context, key string) (string, error) {
    result, err := c.presigner.PresignGetObject(ctx, &s3.GetObjectInput{
        Bucket: aws.String(c.bucket),
        Key:    aws.String(key),
    }, s3.WithPresignExpires(c.presignTTL))
    if err != nil {
        return "", fmt.Errorf("presign get: %w", err)
    }
    
    return result.URL, nil
}

// GetUploadPresignedURL генерирует presigned URL для загрузки.
func (c *Client) GetUploadPresignedURL(ctx context.Context, key string, contentType string) (string, error) {
    result, err := c.presigner.PresignPutObject(ctx, &s3.PutObjectInput{
        Bucket:      aws.String(c.bucket),
        Key:         aws.String(key),
        ContentType: aws.String(contentType),
    }, s3.WithPresignExpires(c.presignTTL))
    if err != nil {
        return "", fmt.Errorf("presign put: %w", err)
    }
    
    return result.URL, nil
}

// GetPublicURL возвращает публичный URL (для публичных файлов).
func (c *Client) GetPublicURL(key string) string {
    scheme := "http"
    if c.useSSL {
        scheme = "https"
    }
    return fmt.Sprintf("%s://%s/%s/%s", scheme, c.endpoint, c.bucket, key)
}

// Exists проверяет существование файла.
func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
    _, err := c.s3.HeadObject(ctx, &s3.HeadObjectInput{
        Bucket: aws.String(c.bucket),
        Key:    aws.String(key),
    })
    if err != nil {
        var nsk *types.NoSuchKey
        var notFound *types.NotFound
        if errors.As(err, &nsk) || errors.As(err, &notFound) {
            return false, nil
        }
        return false, err
    }
    return true, nil
}
```

## Сервис хранения планировок

```go
// internal/service/storage/floor_plan.go

// FloorPlanStorage сервис хранения планировок.
type FloorPlanStorage struct {
    client *storage.Client
    logger *zap.Logger
}

// UploadFloorPlan загружает файл планировки.
func (s *FloorPlanStorage) UploadFloorPlan(
    ctx context.Context,
    workspaceID string,
    floorPlanID string,
    file io.Reader,
    filename string,
    contentType string,
) (*UploadResult, error) {
    // Определяем расширение
    ext := filepath.Ext(filename)
    if ext == "" {
        ext = getExtensionFromContentType(contentType)
    }
    
    // Формируем путь
    key := fmt.Sprintf("floor-plans/%s/%s/original%s", workspaceID, floorPlanID, ext)
    
    // Загружаем
    if err := s.client.Upload(ctx, key, file, contentType); err != nil {
        return nil, fmt.Errorf("upload file: %w", err)
    }
    
    return &UploadResult{
        Key:      key,
        URL:      s.client.GetPublicURL(key),
        Filename: filename,
    }, nil
}

// UploadProcessedImage загружает обработанное изображение.
func (s *FloorPlanStorage) UploadProcessedImage(
    ctx context.Context,
    workspaceID string,
    floorPlanID string,
    image []byte,
) (string, error) {
    key := fmt.Sprintf("floor-plans/%s/%s/processed.png", workspaceID, floorPlanID)
    
    if err := s.client.Upload(ctx, key, bytes.NewReader(image), "image/png"); err != nil {
        return "", err
    }
    
    return s.client.GetPublicURL(key), nil
}

// UploadThumbnail загружает превью.
func (s *FloorPlanStorage) UploadThumbnail(
    ctx context.Context,
    workspaceID string,
    floorPlanID string,
    thumbnail []byte,
) (string, error) {
    key := fmt.Sprintf("floor-plans/%s/%s/thumbnail.png", workspaceID, floorPlanID)
    
    if err := s.client.Upload(ctx, key, bytes.NewReader(thumbnail), "image/png"); err != nil {
        return "", err
    }
    
    return s.client.GetPublicURL(key), nil
}

// GetDownloadURL возвращает URL для скачивания оригинала.
func (s *FloorPlanStorage) GetDownloadURL(ctx context.Context, key string) (string, error) {
    return s.client.GetPresignedURL(ctx, key)
}

// DeleteFloorPlan удаляет все файлы планировки.
func (s *FloorPlanStorage) DeleteFloorPlan(
    ctx context.Context,
    workspaceID string,
    floorPlanID string,
) error {
    prefix := fmt.Sprintf("floor-plans/%s/%s/", workspaceID, floorPlanID)
    
    // Получаем список файлов
    result, err := s.client.ListObjects(ctx, prefix)
    if err != nil {
        return err
    }
    
    // Удаляем все файлы
    for _, obj := range result {
        if err := s.client.Delete(ctx, obj.Key); err != nil {
            s.logger.Warn("Failed to delete file",
                zap.String("key", obj.Key),
                zap.Error(err),
            )
        }
    }
    
    return nil
}
```

## Сервис хранения рендеров

```go
// internal/service/storage/render.go

// RenderStorage сервис хранения рендеров.
type RenderStorage struct {
    client *storage.Client
    logger *zap.Logger
}

// SaveRender сохраняет рендер сцены.
func (s *RenderStorage) SaveRender(
    ctx context.Context,
    sceneID string,
    branchID string,
    renderID string,
    image []byte,
    viewType string,
) (*RenderResult, error) {
    // Основное изображение
    key := fmt.Sprintf("renders/%s/%s/%s.png", sceneID, branchID, renderID)
    if err := s.client.Upload(ctx, key, bytes.NewReader(image), "image/png"); err != nil {
        return nil, err
    }
    
    // Генерируем thumbnail
    thumbnail, err := s.generateThumbnail(image, 400, 300)
    if err != nil {
        s.logger.Warn("Failed to generate thumbnail", zap.Error(err))
    } else {
        thumbKey := fmt.Sprintf("renders/%s/%s/%s_thumb.png", sceneID, branchID, renderID)
        s.client.Upload(ctx, thumbKey, bytes.NewReader(thumbnail), "image/png")
    }
    
    return &RenderResult{
        URL:          s.client.GetPublicURL(key),
        ThumbnailURL: s.client.GetPublicURL(fmt.Sprintf("renders/%s/%s/%s_thumb.png", sceneID, branchID, renderID)),
        ExpiresAt:    time.Now().Add(7 * 24 * time.Hour), // 7 дней
    }, nil
}

// SavePreview сохраняет превью ветки.
func (s *RenderStorage) SavePreview(
    ctx context.Context,
    branchID string,
    image []byte,
) (string, error) {
    key := fmt.Sprintf("previews/%s.png", branchID)
    
    if err := s.client.Upload(ctx, key, bytes.NewReader(image), "image/png"); err != nil {
        return "", err
    }
    
    return s.client.GetPublicURL(key), nil
}

// generateThumbnail создаёт миниатюру.
func (s *RenderStorage) generateThumbnail(data []byte, width, height int) ([]byte, error) {
    img, _, err := image.Decode(bytes.NewReader(data))
    if err != nil {
        return nil, err
    }
    
    // Используем resize
    thumbnail := resize.Thumbnail(uint(width), uint(height), img, resize.Lanczos3)
    
    var buf bytes.Buffer
    if err := png.Encode(&buf, thumbnail); err != nil {
        return nil, err
    }
    
    return buf.Bytes(), nil
}
```

## Сервис аватаров

```go
// internal/service/storage/avatar.go

// AvatarStorage сервис хранения аватаров.
type AvatarStorage struct {
    client    *storage.Client
    maxSize   int64
    dimension int
    logger    *zap.Logger
}

// UploadAvatar загружает и обрабатывает аватар.
func (s *AvatarStorage) UploadAvatar(
    ctx context.Context,
    userID string,
    file io.Reader,
    contentType string,
) (string, error) {
    // Читаем файл
    data, err := io.ReadAll(io.LimitReader(file, s.maxSize))
    if err != nil {
        return "", fmt.Errorf("read file: %w", err)
    }
    
    // Валидируем тип
    if !isAllowedImageType(contentType) {
        return "", ErrInvalidImageType
    }
    
    // Декодируем изображение
    img, _, err := image.Decode(bytes.NewReader(data))
    if err != nil {
        return "", fmt.Errorf("decode image: %w", err)
    }
    
    // Обрезаем до квадрата и ресайзим
    processed := s.processAvatar(img)
    
    // Энкодим в JPEG
    var buf bytes.Buffer
    if err := jpeg.Encode(&buf, processed, &jpeg.Options{Quality: 85}); err != nil {
        return "", fmt.Errorf("encode jpeg: %w", err)
    }
    
    // Загружаем с версионированием
    version := time.Now().Unix()
    key := fmt.Sprintf("avatars/%s.jpg", userID)
    
    if err := s.client.Upload(ctx, key, &buf, "image/jpeg"); err != nil {
        return "", err
    }
    
    // URL с версией для обхода кэша
    return fmt.Sprintf("%s?v=%d", s.client.GetPublicURL(key), version), nil
}

// processAvatar обрабатывает изображение аватара.
func (s *AvatarStorage) processAvatar(img image.Image) image.Image {
    bounds := img.Bounds()
    width := bounds.Dx()
    height := bounds.Dy()
    
    // Определяем квадратную область
    var cropped image.Image
    if width > height {
        offset := (width - height) / 2
        cropped = imaging.Crop(img, image.Rect(offset, 0, offset+height, height))
    } else if height > width {
        offset := (height - width) / 2
        cropped = imaging.Crop(img, image.Rect(0, offset, width, offset+width))
    } else {
        cropped = img
    }
    
    // Ресайзим до целевого размера
    return imaging.Resize(cropped, s.dimension, s.dimension, imaging.Lanczos)
}

// DeleteAvatar удаляет аватар.
func (s *AvatarStorage) DeleteAvatar(ctx context.Context, userID string) error {
    key := fmt.Sprintf("avatars/%s.jpg", userID)
    return s.client.Delete(ctx, key)
}
```

## Очистка устаревших файлов

```go
// internal/service/storage/cleanup.go

// CleanupService сервис очистки устаревших файлов.
type CleanupService struct {
    client *storage.Client
    logger *zap.Logger
}

// CleanupExpiredRenders удаляет устаревшие рендеры.
func (s *CleanupService) CleanupExpiredRenders(ctx context.Context) error {
    cutoff := time.Now().Add(-7 * 24 * time.Hour) // 7 дней
    
    objects, err := s.client.ListObjects(ctx, "renders/")
    if err != nil {
        return err
    }
    
    var deleted int
    for _, obj := range objects {
        if obj.LastModified.Before(cutoff) {
            if err := s.client.Delete(ctx, obj.Key); err != nil {
                s.logger.Warn("Failed to delete render",
                    zap.String("key", obj.Key),
                    zap.Error(err),
                )
                continue
            }
            deleted++
        }
    }
    
    s.logger.Info("Cleanup completed",
        zap.Int("deleted", deleted),
        zap.String("prefix", "renders/"),
    )
    
    return nil
}
```

## Bucket Policy

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "PublicReadAvatars",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "s3:GetObject",
      "Resource": "arn:aws:s3:::granula/avatars/*"
    },
    {
      "Sid": "PublicReadModels",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "s3:GetObject",
      "Resource": "arn:aws:s3:::granula/models/*"
    },
    {
      "Sid": "PublicReadPreviews",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "s3:GetObject",
      "Resource": "arn:aws:s3:::granula/previews/*"
    }
  ]
}
```


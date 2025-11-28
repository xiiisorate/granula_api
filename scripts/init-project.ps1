# =============================================================================
# Granula Project Initialization Script (Windows PowerShell)
# =============================================================================
# Usage: .\scripts\init-project.ps1
# Run from project root directory
# =============================================================================

Write-Host "🚀 Initializing Granula project structure..." -ForegroundColor Cyan

# Создание директорий
$dirs = @(
    # Shared
    "shared/proto/auth/v1",
    "shared/proto/user/v1",
    "shared/proto/workspace/v1",
    "shared/proto/floor_plan/v1",
    "shared/proto/scene/v1",
    "shared/proto/branch/v1",
    "shared/proto/ai/v1",
    "shared/proto/compliance/v1",
    "shared/proto/request/v1",
    "shared/proto/notification/v1",
    "shared/proto/common/v1",
    "shared/pkg/logger",
    "shared/pkg/errors",
    "shared/pkg/config",
    "shared/pkg/grpc",
    "shared/pkg/validator",
    "shared/gen",
    
    # API Gateway
    "api-gateway/cmd/server",
    "api-gateway/internal/config",
    "api-gateway/internal/handler/http/v1",
    "api-gateway/internal/handler/http/middleware",
    "api-gateway/internal/grpc",
    "api-gateway/internal/websocket",
    
    # Auth Service
    "auth-service/cmd/server",
    "auth-service/internal/config",
    "auth-service/internal/domain/entity",
    "auth-service/internal/repository/postgres",
    "auth-service/internal/service",
    "auth-service/internal/grpc",
    "auth-service/internal/jwt",
    "auth-service/internal/oauth",
    "auth-service/migrations",
    
    # User Service
    "user-service/cmd/server",
    "user-service/internal/config",
    "user-service/internal/domain/entity",
    "user-service/internal/repository/postgres",
    "user-service/internal/service",
    "user-service/internal/grpc",
    "user-service/internal/storage",
    "user-service/migrations",
    
    # Workspace Service
    "workspace-service/cmd/server",
    "workspace-service/internal/config",
    "workspace-service/internal/domain/entity",
    "workspace-service/internal/repository/postgres",
    "workspace-service/internal/service",
    "workspace-service/internal/grpc",
    "workspace-service/migrations",
    
    # Floor Plan Service
    "floor-plan-service/cmd/server",
    "floor-plan-service/internal/config",
    "floor-plan-service/internal/domain/entity",
    "floor-plan-service/internal/repository/postgres",
    "floor-plan-service/internal/service",
    "floor-plan-service/internal/grpc",
    "floor-plan-service/internal/storage",
    "floor-plan-service/migrations",
    
    # Scene Service
    "scene-service/cmd/server",
    "scene-service/internal/config",
    "scene-service/internal/domain/entity",
    "scene-service/internal/repository/mongodb",
    "scene-service/internal/service",
    "scene-service/internal/grpc",
    
    # Branch Service
    "branch-service/cmd/server",
    "branch-service/internal/config",
    "branch-service/internal/domain/entity",
    "branch-service/internal/repository/mongodb",
    "branch-service/internal/service",
    "branch-service/internal/grpc",
    "branch-service/internal/engine",
    
    # AI Service
    "ai-service/cmd/server",
    "ai-service/internal/config",
    "ai-service/internal/domain/entity",
    "ai-service/internal/repository/mongodb",
    "ai-service/internal/service",
    "ai-service/internal/grpc",
    "ai-service/internal/openrouter",
    "ai-service/internal/worker",
    
    # Compliance Service
    "compliance-service/cmd/server",
    "compliance-service/internal/config",
    "compliance-service/internal/domain/entity",
    "compliance-service/internal/repository/postgres",
    "compliance-service/internal/service",
    "compliance-service/internal/grpc",
    "compliance-service/internal/engine",
    "compliance-service/migrations",
    
    # Request Service
    "request-service/cmd/server",
    "request-service/internal/config",
    "request-service/internal/domain/entity",
    "request-service/internal/repository/postgres",
    "request-service/internal/service",
    "request-service/internal/grpc",
    "request-service/migrations",
    
    # Notification Service
    "notification-service/cmd/server",
    "notification-service/internal/config",
    "notification-service/internal/domain/entity",
    "notification-service/internal/repository/postgres",
    "notification-service/internal/service",
    "notification-service/internal/grpc",
    "notification-service/internal/email",
    "notification-service/internal/pubsub",
    "notification-service/migrations",
    
    # Other
    "scripts",
    # "docs",
    ".vscode"
)

foreach ($dir in $dirs) {
    if (!(Test-Path $dir)) {
        New-Item -ItemType Directory -Force -Path $dir | Out-Null
        Write-Host "  Created: $dir" -ForegroundColor Green
    }
}

# Создание .gitignore
$gitignore = @"
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
"@

if (!(Test-Path .gitignore)) {
    $gitignore | Out-File -FilePath .gitignore -Encoding utf8NoBOM
    Write-Host "  Created: .gitignore" -ForegroundColor Green
}

# Создание .vscode/settings.json
$vscodeSettings = @"
{
  "git.autofetch": true,
  "git.autofetchPeriod": 60,
  "git.fetchOnPull": true,
  "git.pruneOnFetch": true,
  "git.confirmSync": false,
  "editor.formatOnSave": true,
  "go.formatTool": "goimports",
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "package",
  "go.useLanguageServer": true,
  "[go]": {
    "editor.defaultFormatter": "golang.go",
    "editor.codeActionsOnSave": {
      "source.organizeImports": "explicit"
    }
  },
  "protoc": {
    "options": [
      "--proto_path=shared/proto"
    ]
  },
  "files.exclude": {
    "**/bin": true,
    "**/vendor": true,
    "**/.git": true
  }
}
"@

$vscodeSettings | Out-File -FilePath .vscode/settings.json -Encoding utf8NoBOM
Write-Host "  Created: .vscode/settings.json" -ForegroundColor Green

# Создание .vscode/extensions.json
$vscodeExtensions = @"
{
  "recommendations": [
    "golang.go",
    "eamodio.gitlens",
    "zxh404.vscode-proto3",
    "ms-azuretools.vscode-docker",
    "redhat.vscode-yaml",
    "streetsidesoftware.code-spell-checker"
  ]
}
"@

$vscodeExtensions | Out-File -FilePath .vscode/extensions.json -Encoding utf8NoBOM
Write-Host "  Created: .vscode/extensions.json" -ForegroundColor Green

Write-Host ""
Write-Host "✅ Project structure created!" -ForegroundColor Green
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Yellow
Write-Host "  1. git add ."
Write-Host "  2. git commit -m 'chore: initial project structure'"
Write-Host "  3. git remote add origin <your-repo-url>"
Write-Host "  4. git push -u origin main"
Write-Host "  5. git checkout -b develop"
Write-Host "  6. git push -u origin develop"
Write-Host ""


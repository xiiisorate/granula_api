#!/bin/bash
# =============================================================================
# Granula API - Server Setup Script
# =============================================================================
# Первоначальная настройка сервера для деплоя.
# Запускать под root на свежем Ubuntu/Debian сервере.
#
# Usage:
#   chmod +x setup-server.sh
#   ./setup-server.sh
# =============================================================================

set -e

echo "╔══════════════════════════════════════════════════════════════════════════════╗"
echo "║                    Granula API - Server Setup                                ║"
echo "╚══════════════════════════════════════════════════════════════════════════════╝"

# =============================================================================
# Configuration
# =============================================================================
DOMAIN="api.granula.raitokyokai.tech"
EMAIL="admin@granula.ru"
APP_DIR="/opt/granula/api"

# =============================================================================
# System Updates
# =============================================================================
echo ""
echo "📦 Updating system packages..."
apt-get update
apt-get upgrade -y

# =============================================================================
# Install Dependencies
# =============================================================================
echo ""
echo "📦 Installing dependencies..."
apt-get install -y \
    apt-transport-https \
    ca-certificates \
    curl \
    gnupg \
    lsb-release \
    git \
    htop \
    vim \
    ufw \
    fail2ban \
    certbot

# =============================================================================
# Install Docker
# =============================================================================
echo ""
echo "🐳 Installing Docker..."
if ! command -v docker &> /dev/null; then
    curl -fsSL https://get.docker.com | bash
    systemctl enable docker
    systemctl start docker
fi

# Install Docker Compose
if ! command -v docker-compose &> /dev/null; then
    curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    chmod +x /usr/local/bin/docker-compose
fi

echo "Docker version: $(docker --version)"
echo "Docker Compose version: $(docker-compose --version)"

# =============================================================================
# Configure Firewall
# =============================================================================
echo ""
echo "🔥 Configuring firewall..."
ufw default deny incoming
ufw default allow outgoing
ufw allow ssh
ufw allow 80/tcp
ufw allow 443/tcp
ufw --force enable
ufw status

# =============================================================================
# Configure Fail2Ban
# =============================================================================
echo ""
echo "🛡️ Configuring Fail2Ban..."
cat > /etc/fail2ban/jail.local << 'EOF'
[DEFAULT]
bantime = 3600
findtime = 600
maxretry = 5

[sshd]
enabled = true
port = ssh
filter = sshd
logpath = /var/log/auth.log
maxretry = 3
EOF

systemctl enable fail2ban
systemctl restart fail2ban

# =============================================================================
# Create Application Directory
# =============================================================================
echo ""
echo "📁 Creating application directory..."
mkdir -p ${APP_DIR}
mkdir -p ${APP_DIR}/images
mkdir -p ${APP_DIR}/nginx/ssl
mkdir -p ${APP_DIR}/certbot/www
mkdir -p ${APP_DIR}/certbot/conf

# =============================================================================
# SSL Certificate (Let's Encrypt)
# =============================================================================
echo ""
echo "🔐 Setting up SSL certificate..."

# Create temporary nginx for certbot verification
cat > /tmp/nginx-certbot.conf << 'NGINX'
events { worker_connections 1024; }
http {
    server {
        listen 80;
        server_name api.granula.raitokyokai.tech;
        location /.well-known/acme-challenge/ {
            root /var/www/certbot;
        }
        location / {
            return 200 'OK';
        }
    }
}
NGINX

# Run temporary nginx
docker run -d --name nginx-certbot \
    -p 80:80 \
    -v /tmp/nginx-certbot.conf:/etc/nginx/nginx.conf:ro \
    -v ${APP_DIR}/certbot/www:/var/www/certbot:ro \
    nginx:alpine

sleep 5

# Get certificate
certbot certonly --webroot \
    -w ${APP_DIR}/certbot/www \
    -d ${DOMAIN} \
    --email ${EMAIL} \
    --agree-tos \
    --non-interactive \
    || echo "Certificate already exists or domain not pointing to this server yet"

# Stop temporary nginx
docker stop nginx-certbot
docker rm nginx-certbot

# Copy certificates
if [ -d "/etc/letsencrypt/live/${DOMAIN}" ]; then
    cp -rL /etc/letsencrypt/live/${DOMAIN}/* ${APP_DIR}/nginx/ssl/
    echo "✅ SSL certificates copied"
else
    echo "⚠️ SSL certificates not found. Run certbot manually after DNS is configured."
fi

# =============================================================================
# Setup Auto-renewal
# =============================================================================
echo ""
echo "🔄 Setting up SSL auto-renewal..."
cat > /etc/cron.d/certbot-renew << 'EOF'
0 0,12 * * * root certbot renew --quiet && docker exec granula-nginx nginx -s reload
EOF

# =============================================================================
# Create Deploy Script
# =============================================================================
echo ""
echo "📝 Creating deploy script..."
cat > ${APP_DIR}/deploy.sh << 'DEPLOY'
#!/bin/bash
# Quick deploy script
cd /opt/granula/api

# Pull latest changes (if using git)
# git pull origin main

# Load new images
for img in images/*.tar.gz; do
    if [ -f "$img" ]; then
        echo "Loading $img..."
        gunzip -c "$img" | docker load
    fi
done

# Restart services
docker-compose --env-file .env -f docker-compose.yml down
docker-compose --env-file .env -f docker-compose.yml up -d

# Wait and check health
sleep 30
curl -sf http://localhost:8080/health && echo "✅ API is healthy" || echo "❌ API health check failed"
DEPLOY

chmod +x ${APP_DIR}/deploy.sh

# =============================================================================
# Create Backup Script
# =============================================================================
echo ""
echo "💾 Creating backup script..."
cat > ${APP_DIR}/backup.sh << 'BACKUP'
#!/bin/bash
# Database backup script
BACKUP_DIR="/opt/granula/backups"
DATE=$(date +%Y%m%d_%H%M%S)

mkdir -p ${BACKUP_DIR}

# Backup PostgreSQL databases
for db in auth users workspaces notifications compliance floorplan requests; do
    docker exec granula-postgres-${db} pg_dump -U granula ${db}_db | gzip > ${BACKUP_DIR}/${db}_${DATE}.sql.gz
done

# Backup MongoDB
docker exec granula-mongodb mongodump --archive --gzip > ${BACKUP_DIR}/mongodb_${DATE}.archive.gz

# Cleanup old backups (keep 7 days)
find ${BACKUP_DIR} -type f -mtime +7 -delete

echo "Backup completed: ${DATE}"
BACKUP

chmod +x ${APP_DIR}/backup.sh

# Setup daily backup cron
echo "0 3 * * * root ${APP_DIR}/backup.sh >> /var/log/granula-backup.log 2>&1" > /etc/cron.d/granula-backup

# =============================================================================
# System Optimizations
# =============================================================================
echo ""
echo "⚡ Applying system optimizations..."

# Increase file limits
cat >> /etc/security/limits.conf << 'EOF'
* soft nofile 65535
* hard nofile 65535
EOF

# Optimize sysctl
cat >> /etc/sysctl.conf << 'EOF'
# Network optimizations
net.core.somaxconn = 65535
net.ipv4.tcp_max_syn_backlog = 65535
net.ipv4.ip_local_port_range = 1024 65535
net.ipv4.tcp_tw_reuse = 1
net.ipv4.tcp_fin_timeout = 15
net.core.netdev_max_backlog = 65535
vm.swappiness = 10
EOF

sysctl -p

# =============================================================================
# Summary
# =============================================================================
echo ""
echo "╔══════════════════════════════════════════════════════════════════════════════╗"
echo "║                         Setup Complete! ✅                                    ║"
echo "╠══════════════════════════════════════════════════════════════════════════════╣"
echo "║  Application directory: ${APP_DIR}                                           ║"
echo "║                                                                              ║"
echo "║  Next steps:                                                                 ║"
echo "║  1. Copy .env file with secrets to ${APP_DIR}/.env                           ║"
echo "║  2. Copy docker-compose.yml to ${APP_DIR}/                                   ║"
echo "║  3. Copy Docker images to ${APP_DIR}/images/                                 ║"
echo "║  4. Run: cd ${APP_DIR} && ./deploy.sh                                        ║"
echo "║                                                                              ║"
echo "║  Useful commands:                                                            ║"
echo "║  - View logs: docker-compose logs -f                                         ║"
echo "║  - Restart: docker-compose restart                                           ║"
echo "║  - Backup: ${APP_DIR}/backup.sh                                              ║"
echo "╚══════════════════════════════════════════════════════════════════════════════╝"


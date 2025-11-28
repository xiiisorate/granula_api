-- =============================================================================
-- Granula: PostgreSQL Database Initialization
-- Creates separate databases for each microservice
-- =============================================================================

-- Auth Service Database
CREATE DATABASE auth_db;
GRANT ALL PRIVILEGES ON DATABASE auth_db TO granula;

-- User Service Database
CREATE DATABASE users_db;
GRANT ALL PRIVILEGES ON DATABASE users_db TO granula;

-- Workspace Service Database
CREATE DATABASE workspaces_db;
GRANT ALL PRIVILEGES ON DATABASE workspaces_db TO granula;

-- Floor Plan Service Database
CREATE DATABASE floor_plans_db;
GRANT ALL PRIVILEGES ON DATABASE floor_plans_db TO granula;

-- Compliance Service Database
CREATE DATABASE compliance_db;
GRANT ALL PRIVILEGES ON DATABASE compliance_db TO granula;

-- Request Service Database
CREATE DATABASE requests_db;
GRANT ALL PRIVILEGES ON DATABASE requests_db TO granula;

-- Notification Service Database
CREATE DATABASE notifications_db;
GRANT ALL PRIVILEGES ON DATABASE notifications_db TO granula;

-- =============================================================================
-- Extensions for each database
-- =============================================================================

\c auth_db
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

\c users_db
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

\c workspaces_db
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

\c floor_plans_db
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

\c compliance_db
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

\c requests_db
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

\c notifications_db
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";


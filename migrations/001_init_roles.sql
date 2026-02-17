-- 001_init_roles.sql
-- 创建 PostgreSQL 角色（需要超级用户执行）

-- 创建角色
CREATE ROLE linkflow_admin WITH LOGIN PASSWORD 'change_me_admin';

CREATE ROLE linkflow_app WITH LOGIN PASSWORD 'change_me_app';

CREATE ROLE linkflow_read WITH LOGIN PASSWORD 'change_me_read';

-- 创建数据库
CREATE DATABASE linkflow OWNER linkflow_admin;

-- 连接到 linkflow 数据库后执行以下内容
\c linkflow

-- 启用扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE EXTENSION IF NOT EXISTS timescaledb;

-- 授权 schema 使用权限
GRANT USAGE ON SCHEMA public TO linkflow_app, linkflow_read;

-- linkflow_admin: 完全控制
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO linkflow_admin;

GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO linkflow_admin;

ALTER DEFAULT PRIVILEGES IN SCHEMA public
GRANT ALL ON TABLES TO linkflow_admin;

ALTER DEFAULT PRIVILEGES IN SCHEMA public
GRANT ALL ON SEQUENCES TO linkflow_admin;

-- linkflow_app: CRUD 权限（后续建表后授权）
ALTER DEFAULT PRIVILEGES FOR ROLE linkflow_admin IN SCHEMA public
GRANT
SELECT, INSERT,
UPDATE, DELETE ON TABLES TO linkflow_app;

ALTER DEFAULT PRIVILEGES FOR ROLE linkflow_admin IN SCHEMA public
GRANT USAGE,
SELECT ON SEQUENCES TO linkflow_app;

-- linkflow_read: 只读权限
ALTER DEFAULT PRIVILEGES FOR ROLE linkflow_admin IN SCHEMA public
GRANT
SELECT ON TABLES TO linkflow_read;
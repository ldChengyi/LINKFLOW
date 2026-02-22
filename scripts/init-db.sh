#!/bin/bash
set -e

# 使用 sed 替换环境变量（使用 | 作为分隔符避免密码中的 / 字符冲突）
sed -e "s|\${DB_ADMIN_PASSWORD}|${DB_ADMIN_PASSWORD}|g" \
    -e "s|\${DB_APP_PASSWORD}|${DB_APP_PASSWORD}|g" \
    -e "s|\${DB_READ_PASSWORD}|${DB_READ_PASSWORD}|g" \
    /docker-entrypoint-initdb.d/init-db.sql.template > /tmp/init-db.sql

# 执行 SQL
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" -f /tmp/init-db.sql

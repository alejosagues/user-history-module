FROM postgres:16.1

COPY ../docker-entrypoint-initdb.d/init.sql docker-entrypoint-initdb.d/
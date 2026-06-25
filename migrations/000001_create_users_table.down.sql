-- Down migration: откат начальной схемы
DROP TABLE IF EXISTS task_tags CASCADE;
DROP TABLE IF EXISTS task_overrides CASCADE;
DROP TABLE IF EXISTS tags CASCADE;
DROP TABLE IF EXISTS tasks CASCADE;
DROP TYPE IF EXISTS task_status;
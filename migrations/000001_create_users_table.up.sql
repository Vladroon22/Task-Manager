-- Active: 1765748558180@@127.0.0.1@5432

CREATE TYPE task_status AS ENUM ('new', 'in_progress', 'done', 'cancelled');

-- Основная таблица задач с полями для повторений
CREATE TABLE IF NOT EXISTS tasks (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    due_date DATE NOT NULL,
    status task_status NOT NULL DEFAULT 'new',
    
    -- Поля для повторяющихся задач
    is_recurring BOOLEAN NOT NULL DEFAULT FALSE,
    recurrence_type VARCHAR(20) CHECK (recurrence_type IN ('daily', 'monthly', 'specific', 'even_odd')),
    recurrence_interval INTEGER DEFAULT 1,
    recurrence_days INTEGER[] DEFAULT NULL,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS task_overrides (
    id SERIAL PRIMARY KEY,
    task_id INTEGER NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    override_date DATE NOT NULL,
    status task_status NOT NULL,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(task_id, override_date)
);

CREATE TABLE IF NOT EXISTS tags (
    id SERIAL PRIMARY KEY,
    names TEXT[] NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS task_tags (
    task_id INTEGER NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    tag_id INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (task_id, tag_id)
);

-- Индексы
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
CREATE INDEX IF NOT EXISTS idx_tasks_due_date ON tasks(due_date);
CREATE INDEX IF NOT EXISTS idx_tasks_is_recurring ON tasks(is_recurring);
CREATE INDEX IF NOT EXISTS idx_tasks_recurrence_type ON tasks(recurrence_type);
CREATE INDEX IF NOT EXISTS idx_task_overrides_task_id ON task_overrides(task_id);
CREATE INDEX IF NOT EXISTS idx_task_overrides_date ON task_overrides(override_date);
CREATE INDEX IF NOT EXISTS idx_task_overrides_task_date ON task_overrides(task_id, override_date);
CREATE INDEX IF NOT EXISTS idx_task_tags_task_id ON task_tags(task_id);
CREATE INDEX IF NOT EXISTS idx_task_tags_tag_id ON task_tags(tag_id);

-- Заполнение обычных задач
INSERT INTO tasks (title, description, due_date, status, created_at, updated_at) VALUES
('Запись к терапевту', 'Плановый осмотр у терапевта. Взять с собой паспорт и полис ОМС.', '2026-07-15', 'new', NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days'),
('Визит к стоматологу', 'Лечение кариеса, 2 зуба. Клиника "Дентал-Профи"', '2026-07-10', 'in_progress', NOW() - INTERVAL '5 days', NOW() - INTERVAL '1 day'),
('Анализы крови', 'Сдать общий анализ крови и биохимию. Натощак! Лаборатория на 2 этаже.', '2026-06-25', 'done', NOW() - INTERVAL '10 days', NOW() - INTERVAL '8 days'),
('Прием у кардиолога', 'Консультация по результатам ЭКГ. Направил терапевт.', '2026-07-20', 'new', NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day'),
('МРТ коленного сустава', 'По направлению от хирурга. Центр МРТ на Ленина 15.', '2026-07-05', 'cancelled', NOW() - INTERVAL '7 days', NOW() - INTERVAL '3 days'),
('Вакцинация от гриппа', 'Прививка в поликлинике №5, кабинет 302.', '2026-09-01', 'new', NOW(), NOW()),
('УЗИ брюшной полости', 'Подготовка: за 3 дня исключить газообразующие продукты.', '2026-07-12', 'in_progress', NOW() - INTERVAL '4 days', NOW() - INTERVAL '1 day'),
('Консультация эндокринолога', 'Проверить щитовидную железу, принести результаты УЗИ.', '2026-08-15', 'new', NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day'),
('Флюорография', 'Ежегодный профилактический осмотр.', '2026-06-20', 'done', NOW() - INTERVAL '15 days', NOW() - INTERVAL '13 days'),
('Прием у окулиста', 'Проверка зрения, возможна замена очков.', '2026-07-18', 'in_progress', NOW() - INTERVAL '3 days', NOW() - INTERVAL '2 days');

-- Заполнение повторяющихся задач

-- 11. Ежедневная задача (каждый день)
INSERT INTO tasks (title, description, due_date, status, is_recurring, recurrence_type, recurrence_interval) VALUES
('Ежедневный обход', 'Утренний обход пациентов в палатах', '2026-06-20', 'new', TRUE, 'daily', 1);

-- 12. Каждый 3-й день
INSERT INTO tasks (title, description, due_date, status, is_recurring, recurrence_type, recurrence_interval) VALUES
('Дежурство', 'Дежурство в отделении терапии', '2026-06-20', 'new', TRUE, 'daily', 3);

-- 13. Ежемесячно 5 и 20 числа
INSERT INTO tasks (title, description, due_date, status, is_recurring, recurrence_type, recurrence_days) VALUES
('Отчет о работе', 'Подготовка ежемесячного отчета для главврача', '2026-06-20', 'new', TRUE, 'monthly', ARRAY[5, 20]);

-- 14. На конкретные даты (3, 15, 28 число каждого месяца)
INSERT INTO tasks (title, description, due_date, status, is_recurring, recurrence_type, recurrence_days) VALUES
('Совещание отделения', 'Общее совещание врачей отделения', '2026-06-20', 'new', TRUE, 'specific', ARRAY[3, 15, 28]);

-- 15. Только по четным дням
INSERT INTO tasks (title, description, due_date, status, is_recurring, recurrence_type, recurrence_interval) VALUES
('Плановые процедуры', 'Проведение плановых процедур пациентам', '2026-06-20', 'new', TRUE, 'even_odd', 0);

-- 16. Только по нечетным дням
INSERT INTO tasks (title, description, due_date, status, is_recurring, recurrence_type, recurrence_interval) VALUES
('Забор анализов', 'Забор анализов у пациентов (натощак)', '2026-06-20', 'new', TRUE, 'even_odd', 1);

-- Заполнение тегов
INSERT INTO tags (names) VALUES
(ARRAY['звонок', 'терапевт']),
(ARRAY['стоматолог']),
(ARRAY['отчетность']),
(ARRAY['кардиолог']),
(ARRAY['обследование']),
(ARRAY['хирург']),
(ARRAY['вакцинация']),
(ARRAY['невролог']),
(ARRAY['окулист']),
(ARRAY['эндокринолог']),
(ARRAY['гастроэнтеролог']),
(ARRAY['дерматолог']),
(ARRAY['диетолог']),
(ARRAY['массаж']),
(ARRAY['справка']),
(ARRAY['лфк']),
(ARRAY['операции']),
(ARRAY['узи']),
(ARRAY['мрт']),
(ARRAY['аллерголог']),
(ARRAY['обход']),
(ARRAY['дежурство']),
(ARRAY['отчет']),
(ARRAY['совещание']),
(ARRAY['процедуры']),
(ARRAY['анализы']);

-- Привязка тегов к обычным задачам
INSERT INTO task_tags (task_id, tag_id) VALUES
(1, 1), (1, 5),   -- Запись к терапевту: звонок+терапевт + обследование
(2, 2),             -- Визит к стоматологу: стоматолог
(3, 3), (3, 1),    -- Анализы крови: отчетность + звонок+терапевт
(4, 4), (4, 5),    -- Прием у кардиолога: кардиолог + обследование
(5, 19), (5, 5), (5, 6),  -- МРТ: мрт + обследование + хирург
(6, 7),             -- Вакцинация: вакцинация
(7, 18), (7, 5),   -- УЗИ: узи + обследование
(8, 9), (8, 5),    -- Консультация эндокринолога: эндокринолог + обследование
(9, 17), (9, 5),   -- Флюорография: операции + обследование
(10, 8), (10, 5);  -- Прием у окулиста: окулист + обследование

-- Привязка тегов к повторяющимся задачам
INSERT INTO task_tags (task_id, tag_id) VALUES
(11, 21),                    -- Ежедневный обход: обход
(12, 22),                    -- Дежурство: дежурство
(13, 23),                    -- Отчет: отчет
(14, 24),                    -- Совещание: совещание
(15, 25),                    -- Процедуры: процедуры
(16, 26);                    -- Забор анализов: анализы

-- Примеры переопределений для повторяющихся задач
INSERT INTO task_overrides (task_id, override_date, status, notes) VALUES
-- Ежедневный обход: отмечаем некоторые дни
(11, '2026-06-22', 'done', 'Все пациенты осмотрены'),
(11, '2026-06-23', 'done', 'Обход выполнен'),
(11, '2026-06-25', 'cancelled', 'Врач на конференции'),
(11, '2026-06-28', 'done', 'Выходной день, дежурный врач'),
(11, '2026-07-01', 'in_progress', 'Обход начат'),

-- Дежурство
(12, '2026-06-23', 'done', 'Дежурство прошло без происшествий'),
(12, '2026-06-26', 'done', 'Ночное дежурство'),

-- Плановые процедуры (четные дни)
(15, '2026-06-20', 'done', 'Процедуры выполнены'),
(15, '2026-06-22', 'done', 'Процедуры выполнены'),
(15, '2026-06-24', 'in_progress', 'Выполняются процедуры'),
(15, '2026-06-26', 'cancelled', 'Нет пациентов'),

-- Забор анализов (нечетные дни)
(16, '2026-06-21', 'done', 'Анализы взяты у 5 пациентов'),
(16, '2026-06-23', 'done', 'Анализы взяты у 3 пациентов'),
(16, '2026-06-25', 'done', 'Анализы взяты у 7 пациентов');
package repository

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/Vladroon22/TaskTracker/internal/generator"
	"github.com/Vladroon22/TaskTracker/internal/models"
	"github.com/lib/pq"
)

type TaskRepo struct {
	db      *sql.DB
	tagRepo *TagRepo
}

func NewTaskRepo(db *sql.DB, tagRepo *TagRepo) *TaskRepo {
	return &TaskRepo{
		db:      db,
		tagRepo: tagRepo,
	}
}

// Create создает новую задачу с тегами
func (r *TaskRepo) Create(ctx context.Context, task *models.CreateTaskRequest) (*models.Task, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Определяем статус
	status := models.StatusNew
	if task.Status != "" && task.Status.IsValid() {
		status = task.Status
	}

	// Создаем задачу
	query := `
		INSERT INTO 
			tasks (title, description, due_date, status)
		VALUES
			($1, $2, $3, $4)
		RETURNING 
			id, title, description, due_date, status, created_at, updated_at
	`

	var created models.Task
	if err := tx.QueryRowContext(ctx, query,
		task.Title,
		task.Description,
		task.DueDate.Format(time.DateTime),
		status,
	).Scan(
		&created.ID,
		&created.Title,
		&created.Description,
		&created.DueDate,
		&created.Status,
		&created.CreatedAt,
		&created.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	// Добавляем теги к задаче
	if len(task.Tags) > 0 {
		for _, tagName := range task.Tags {
			// Ищем существующий тег по имени в массиве names
			var tagID int
			err := tx.QueryRowContext(ctx,
				"SELECT id FROM tags WHERE $1 = ANY(names) LIMIT 1",
				tagName,
			).Scan(&tagID)

			if err == sql.ErrNoRows {
				// Создаем новый тег с массивом из одного имени
				err = tx.QueryRowContext(ctx,
					"INSERT INTO tags (names) VALUES ($1) RETURNING id",
					pq.Array([]string{tagName}),
				).Scan(&tagID)
				if err != nil {
					return nil, fmt.Errorf("failed to create tag '%s': %w", tagName, err)
				}
			} else if err != nil {
				return nil, fmt.Errorf("failed to find tag '%s': %w", tagName, err)
			}

			// Связываем тег с задачей
			if _, err := tx.ExecContext(ctx,
				"INSERT INTO task_tags (task_id, tag_id) VALUES ($1, $2) ON CONFLICT DO NOTHING",
				created.ID, tagID,
			); err != nil {
				return nil, fmt.Errorf("failed to link tag '%s' to task: %w", tagName, err)
			}
		}
	}

	// Фиксируем транзакцию
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Загружаем теги созданной задачи
	tags, err := r.tagRepo.GetTaskTags(ctx, created.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load task tags: %w", err)
	}
	created.Tags = tags

	return &created, nil
}

// List возвращает список задач с фильтрацией и пагинацией
func (r *TaskRepo) List(ctx context.Context, filter *models.TaskFilter) ([]models.Task, error) {
	whereClauses := []string{}
	args := []interface{}{}
	argIdx := 1

	// Фильтр по статусу
	if filter.Status != nil && filter.Status.IsValid() {
		whereClauses = append(whereClauses, fmt.Sprintf("t.status = $%d", argIdx))
		args = append(args, *filter.Status)
		argIdx++
	}

	// Фильтр по дате "с"
	if filter.DateFrom != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("t.due_date >= $%d", argIdx))
		args = append(args, filter.DateFrom.Format(time.DateOnly))
		argIdx++
	}

	// Фильтр по дате "по"
	if filter.DateTo != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("t.due_date <= $%d", argIdx))
		args = append(args, filter.DateTo.Format(time.DateOnly))
		argIdx++
	}

	// Фильтр по тегам
	if len(filter.Tags) > 0 {
		tagPlaceholders := make([]string, len(filter.Tags))
		for i, tagID := range filter.Tags {
			tagPlaceholders[i] = fmt.Sprintf("$%d", argIdx)
			args = append(args, tagID)
			argIdx++
		}
		whereClauses = append(whereClauses, fmt.Sprintf(
			"t.id IN (SELECT task_id FROM task_tags WHERE tag_id IN (%s))",
			strings.Join(tagPlaceholders, ","),
		))
	}

	// Формируем WHERE clause
	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Основной запрос с агрегацией тегов
	query := fmt.Sprintf(`
		SELECT 
			t.id, 
			t.title, 
			t.description, 
			t.due_date, 
			t.status, 
			t.created_at, 
			t.updated_at,
			COALESCE(
				(
				SELECT 
					array_agg(tg.id ORDER BY tg.id) 
				FROM
					task_tags tt2 
				JOIN
					tags tg 
				ON 
					tt2.tag_id = tg.id 
				WHERE
					tt2.task_id = t.id
				), '{}'::int[]
			) as tag_ids
		FROM 
			tasks t
		%s
		GROUP BY 
			t.id
		ORDER BY
			t.due_date DESC, t.id DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var task models.Task
		var tagIDs []int
		if err := rows.Scan(
			&task.ID,
			&task.Title,
			&task.Description,
			&task.DueDate,
			&task.Status,
			&task.CreatedAt,
			&task.UpdatedAt,
			pq.Array(&tagIDs),
		); err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}

		// Загружаем теги по ID
		if len(tagIDs) > 0 {
			task.Tags = make([]models.Tag, 0, len(tagIDs))
			for _, tagID := range tagIDs {
				tag, err := r.tagRepo.GetTagByID(ctx, tagID)
				if err == nil {
					task.Tags = append(task.Tags, *tag)
				}
			}
		} else {
			task.Tags = []models.Tag{}
		}

		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tasks: %w", err)
	}

	// Возвращаем пустой массив вместо null
	if tasks == nil {
		tasks = []models.Task{}
	}

	return tasks, nil
}

// GetByID возвращает задачу по ID
func (r *TaskRepo) GetByID(ctx context.Context, id int) (*models.Task, error) {
	query := `
		SELECT 
            id, title, description, due_date, status,
            is_recurring, recurrence_type, recurrence_interval, recurrence_days,
            created_at, updated_at
        FROM
			tasks 
        WHERE 
			id = $1
	`

	var task models.Task
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&task.DueDate,
		&task.Status,
		&task.IsRecurring,
		&task.RecurrenceType,
		&task.RecurrenceInterval,
		pq.Array(&task.RecurrenceDays),
		&task.CreatedAt,
		&task.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, models.ErrTaskNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	// Загружаем теги задачи
	tags, err := r.tagRepo.GetTaskTags(ctx, task.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task tags: %w", err)
	}
	task.Tags = tags

	return &task, nil
}

// Update обновляет задачу
func (r *TaskRepo) Update(ctx context.Context, id int, update *models.UpdateTaskRequest) (*models.Task, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	setClauses := []string{}
	args := []interface{}{}
	argIdx := 1

	// Обновление title
	if update.Title != nil {
		setClauses = append(setClauses, fmt.Sprintf("title = $%d", argIdx))
		args = append(args, *update.Title)
		argIdx++
	}

	// Обновление description
	if update.Description != nil {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", argIdx))
		args = append(args, *update.Description)
		argIdx++
	}

	// Обновление due_date
	if update.DueDate != nil {
		setClauses = append(setClauses, fmt.Sprintf("due_date = $%d", argIdx))
		args = append(args, update.DueDate.Format(time.DateTime))
		argIdx++
	}

	// Обновление status
	if update.Status != nil {
		if !update.Status.IsValid() {
			return nil, fmt.Errorf("invalid status: %s", *update.Status)
		}
		setClauses = append(setClauses, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *update.Status)
		argIdx++
	}

	if len(setClauses) > 0 {
		// Добавляем обновление updated_at
		setClauses = append(setClauses, "updated_at = NOW()")

		// Добавляем ID в конец аргументов
		args = append(args, id)

		query := fmt.Sprintf(`
			UPDATE
				tasks
			SET 
				%s
			WHERE
				id = $%d
			RETURNING 
				id, title, description, due_date, status, created_at, updated_at
		`, strings.Join(setClauses, ", "), argIdx)

		var row models.Task

		err = tx.QueryRowContext(ctx, query, args...).Scan(
			&row.ID,
			&row.Title,
			&row.Description,
			&row.DueDate,
			&row.Status,
			&row.IsRecurring,
			&row.RecurrenceType,
			&row.RecurrenceInterval,
			&row.RecurrenceDays,
			&row.CreatedAt,
			&row.UpdatedAt,
		)

		if err == sql.ErrNoRows {
			return nil, models.ErrTaskNotFound
		}
		if err != nil {
			return nil, fmt.Errorf("failed to update task: %w", err)
		}

		// Фиксируем транзакцию
		if err = tx.Commit(); err != nil {
			return nil, fmt.Errorf("failed to commit transaction: %w", err)
		}

		// Загружаем обновленную задачу с тегами
		return r.GetByID(ctx, id)
	}

	return r.GetByID(ctx, id)
}

func (r *TaskRepo) Delete(ctx context.Context, id int) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM tasks WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return models.ErrTaskNotFound
	}

	return nil
}

////////// periodicity

func (r *TaskRepo) CreateTaskWithPeriod(ctx context.Context, req *models.CreateTaskPeriodRequest) (*models.Task, error) {

	status := models.StatusNew
	if req.Status != "" {
		status = models.TaskStatus(req.Status)
	}

	var task models.Task
	var recurrenceDays pq.Int32Array // Используем pq.Int32Array

	if err := r.db.QueryRowContext(ctx, `
        INSERT INTO tasks (
            title, description, due_date, status, is_recurring, recurrence_type, recurrence_interval, recurrence_days
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING
			id, title, description, due_date, status, is_recurring, recurrence_type, recurrence_interval, recurrence_days, created_at, updated_at
    `,
		req.Title,
		req.Description,
		req.DueDate.Time(),
		status,
		req.IsRecurring,
		req.RecurrenceType,
		req.RecurrenceInterval,
		pq.Array(req.RecurrenceDays),
	).Scan(
		&task.ID, &task.Title, &task.Description, &task.DueDate, &task.Status,
		&task.IsRecurring, &task.RecurrenceType, &task.RecurrenceInterval,
		&recurrenceDays,
		&task.CreatedAt, &task.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	// Конвертируем pq.Int32Array в []int
	if req.RecurrenceDays != nil {
		task.RecurrenceDays = make([]int, len(recurrenceDays))
		for i, d := range recurrenceDays {
			task.RecurrenceDays[i] = int(d)
		}
	}

	// Добавляем теги
	if len(req.Tags) > 0 {
		for _, tagName := range req.Tags {
			r.addTagToTask(ctx, task.ID, tagName)
		}
	}

	task.Tags, _ = r.tagRepo.GetTaskTags(ctx, task.ID)

	return &task, nil
}

// GetTasksForPeriod получает все задачи за период (включая повторяющиеся)
func (r *TaskRepo) GetTasksForPeriod(ctx context.Context, from, to time.Time) ([]models.Task, error) {
	var tasks []models.Task

	regularTasks, err := r.getRegularTasks(ctx, from, to)
	if err != nil {
		return nil, err
	}
	tasks = append(tasks, regularTasks...)

	// 2. Получаем повторяющиеся задачи
	recurringTasks, err := r.getRecurringTasks(ctx, from, to)
	if err != nil {
		return nil, err
	}
	tasks = append(tasks, recurringTasks...)

	// 3. Сортируем по дате
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].DueDate.Time().Before(tasks[j].DueDate.Time())
	})

	// 4. Добавляем теги для каждой задачи
	for i := range tasks {
		tags, _ := r.tagRepo.GetTaskTags(ctx, tasks[i].ID)
		tasks[i].Tags = tags
	}

	return tasks, nil
}

// getRegularTasks получает обычные задачи
func (r *TaskRepo) getRegularTasks(ctx context.Context, from, to time.Time) ([]models.Task, error) {
	rows, err := r.db.QueryContext(ctx, `
        SELECT 
            id, title, COALESCE(description, ''), due_date, status, 
            is_recurring, recurrence_type, recurrence_interval, recurrence_days, 
            created_at, updated_at
        FROM 
            tasks 
        WHERE 
            due_date BETWEEN $1::date AND $2::date
        AND 
            is_recurring = FALSE
        ORDER BY
            due_date
    `, from, to)

	if err != nil {
		return nil, fmt.Errorf("failed to get regular tasks: %w", err)
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var task models.Task
		var recurrenceDays []int64 // Изменено на []int64 для pq.Array

		err := rows.Scan(
			&task.ID, &task.Title, &task.Description, &task.DueDate, &task.Status,
			&task.IsRecurring, &task.RecurrenceType, &task.RecurrenceInterval,
			pq.Array(&recurrenceDays), // Используем pq.Array
			&task.CreatedAt, &task.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}

		// Конвертируем []int64 в []int
		task.RecurrenceDays = make([]int, len(recurrenceDays))
		for i, d := range recurrenceDays {
			task.RecurrenceDays[i] = int(d)
		}

		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating regular tasks: %w", err)
	}

	return tasks, nil
}

// getRecurringTasks получает повторяющиеся задачи
func (r *TaskRepo) getRecurringTasks(ctx context.Context, from, to time.Time) ([]models.Task, error) {
	rows, err := r.db.QueryContext(ctx, `
        SELECT 
            id, title, COALESCE(description, ''), due_date, status,
            is_recurring, recurrence_type, recurrence_interval, recurrence_days,
            created_at, updated_at
        FROM 
            tasks 
        WHERE 
            is_recurring = TRUE 
        AND 
            due_date <= $1::date
    `, to) // Исправлено: используем только $1 (to)

	if err != nil {
		return nil, fmt.Errorf("failed to get recurring tasks: %w", err)
	}
	defer rows.Close()

	var tasks []models.Task

	for rows.Next() {
		var template models.Task
		var recurrenceDays []int64 // Изменено на []int64 для pq.Array

		err := rows.Scan(
			&template.ID, &template.Title, &template.Description,
			&template.DueDate, &template.Status,
			&template.IsRecurring, &template.RecurrenceType,
			&template.RecurrenceInterval, pq.Array(&recurrenceDays), // Используем pq.Array
			&template.CreatedAt, &template.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan recurring task: %w", err)
		}

		// Конвертируем []int64 в []int
		template.RecurrenceDays = make([]int, len(recurrenceDays))
		for i, d := range recurrenceDays {
			template.RecurrenceDays[i] = int(d)
		}

		// Генерируем даты повторений
		dates := generator.GenerateRecurringDates(template, from, to)

		// Получаем переопределения для этого шаблона
		overrides, err := r.getOverridesForPeriod(ctx, template.ID, from, to)
		if err != nil {
			return nil, fmt.Errorf("failed to get overrides for task %d: %w", template.ID, err)
		}

		// Создаем задачи для каждой даты
		for _, date := range dates {
			task := template

			// Устанавливаем новую дату
			task.DueDate = models.CustomDate(date)

			// Устанавливаем статус на основе переопределений
			dateKey := date.Format(time.DateOnly)
			if override, exists := overrides[dateKey]; exists {
				task.Status = override.Status
			} else {
				task.Status = models.StatusNew
			}

			tasks = append(tasks, task)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating recurring tasks: %w", err)
	}

	return tasks, nil
}

// getOverridesForPeriod получает переопределения статусов
func (r *TaskRepo) getOverridesForPeriod(ctx context.Context, taskID int, from, to time.Time) (map[string]models.TaskOverride, error) {
	rows, err := r.db.QueryContext(ctx, `
        SELECT 
            task_id, override_date, status
        FROM 
            task_overrides 
        WHERE
            task_id = $1 
        AND 
            override_date BETWEEN $2::date AND $3::date
    `, taskID, from, to)

	if err != nil {
		return nil, fmt.Errorf("failed to get task overrides: %w", err)
	}
	defer rows.Close()

	overrides := make(map[string]models.TaskOverride)
	for rows.Next() {
		var o models.TaskOverride
		var overrideDate time.Time

		if err := rows.Scan(&o.TaskID, &overrideDate, &o.Status); err != nil {
			return nil, fmt.Errorf("failed to scan override: %w", err)
		}

		o.OverrideDate = overrideDate
		overrides[overrideDate.Format(time.DateOnly)] = o
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating overrides: %w", err)
	}

	return overrides, nil
}

// SetTaskStatus устанавливает статус для конкретной даты повторяющейся задачи
func (r *TaskRepo) SetTaskStatus(ctx context.Context, taskID int, date time.Time, status models.TaskStatus) error {
	// Проверяем, что задача существует и она повторяющаяся
	var isRecurring bool
	if err := r.db.QueryRowContext(ctx, `SELECT is_recurring FROM tasks WHERE id = $1`, taskID).Scan(&isRecurring); err == sql.ErrNoRows {
		return fmt.Errorf("task not found")
	} else if err != nil {
		return err
	}

	if !isRecurring {
		// Для обычных задач просто обновляем статус
		_, err := r.db.ExecContext(ctx, `UPDATE tasks SET status = $1, updated_at = NOW() WHERE id = $2`, status, taskID)
		return err
	}

	query := `
        INSERT INTO task_overrides (task_id, override_date, status)
        VALUES ($1, $2, $3)
        ON CONFLICT (task_id, override_date) 
        DO UPDATE SET status = $3
    `

	// Для повторяющихся - создаем/обновляем переопределение
	_, err := r.db.ExecContext(ctx, query, taskID, date, status)

	return err
}

// addTagToTask добавляет тег к задаче
func (r *TaskRepo) addTagToTask(ctx context.Context, taskID int, tagName string) error {
	var tagID int

	// Вставляем или получаем тег
	err := r.db.QueryRowContext(ctx, `
        INSERT INTO tags (name) VALUES ($1)
        ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
        RETURNING id
    `, tagName).Scan(&tagID)

	if err != nil {
		return err
	}

	// Связываем тег с задачей
	_, err = r.db.ExecContext(ctx, `
        INSERT INTO task_tags (task_id, tag_id) VALUES ($1, $2)
        ON CONFLICT DO NOTHING
    `, taskID, tagID)

	return err
}

// GetTaskTags получает теги задачи
func (r *TaskRepo) GetTaskTags(ctx context.Context, taskID int) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
        SELECT t.name 
        FROM tags t
        JOIN task_tags tt ON t.id = tt.tag_id
        WHERE tt.task_id = $1
    `, taskID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		rows.Scan(&tag)
		tags = append(tags, tag)
	}

	if tags == nil {
		tags = []string{}
	}

	return tags, rows.Err()
}

func (r *TaskRepo) DeleteTaskOverride(ctx context.Context, taskID int, date time.Time) error {
	result, err := r.db.ExecContext(ctx, `
        DELETE FROM task_overrides 
        WHERE task_id = $1 AND override_date = $2
    `, taskID, date)

	if err != nil {
		return fmt.Errorf("failed to delete task override: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("override not found for task %d on date %s", taskID, date.Format(time.DateOnly))
	}

	return nil
}

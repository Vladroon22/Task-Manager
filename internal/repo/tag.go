package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Vladroon22/TaskTracker/internal/models"
	"github.com/lib/pq"
)

type TagRepo struct {
	db *sql.DB
}

func NewTagRepo(db *sql.DB) *TagRepo {
	return &TagRepo{db: db}
}

func (r *TagRepo) CreateTag(ctx context.Context, names []string) (*models.Tag, error) {
	query := `
        INSERT INTO 
            tags (names)
        VALUES
            ($1)
        RETURNING
            id, names, created_at, updated_at
    `

	var created models.Tag
	var dbNames []string
	err := r.db.QueryRowContext(ctx, query, pq.Array(names)).Scan(
		&created.ID,
		pq.Array(&dbNames),
		&created.CreatedAt,
		&created.UpdateAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create tag: %w", err)
	}

	created.Names = dbNames
	return &created, nil
}

func (r *TagRepo) GetTagByID(ctx context.Context, id int) (*models.Tag, error) {
	query := `
        SELECT
            id, names, created_at, updated_at
        FROM 
            tags
        WHERE
            id = $1
    `

	var tag models.Tag
	var names []string
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&tag.ID,
		pq.Array(&names),
		&tag.CreatedAt,
		&tag.UpdateAt,
	)

	if err == sql.ErrNoRows {
		return nil, models.ErrTagNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tag: %w", err)
	}

	tag.Names = names
	return &tag, nil
}

func (r *TagRepo) ListTags(ctx context.Context) ([]models.Tag, error) {
	query := `
		SELECT 
			t.id, 
			t.names,  
			t.created_at,
			t.updated_at,
			COUNT(tt.task_id) as task_count
		FROM 
			tags t
		LEFT JOIN 
			task_tags tt ON t.id = tt.tag_id
		GROUP BY 
			t.id, t.names, t.created_at, t.updated_at
		ORDER BY 
			t.id
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list tags: %w", err)
	}
	defer rows.Close()

	var tags []models.Tag
	for rows.Next() {
		var tag models.Tag
		var names []string
		var taskCount int
		err := rows.Scan(
			&tag.ID,
			pq.Array(&names),
			&tag.CreatedAt,
			&tag.UpdateAt,
			&taskCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tag.Names = names
		tags = append(tags, tag)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tags: %w", err)
	}

	if tags == nil {
		tags = []models.Tag{}
	}

	return tags, nil
}

func (r *TagRepo) DeleteTag(ctx context.Context, id int) error {
	if _, err := r.db.ExecContext(ctx, "DELETE FROM task_tags WHERE tag_id = $1", id); err != nil {
		return fmt.Errorf("failed to remove tag associations: %w", err)
	}

	result, err := r.db.ExecContext(ctx, "DELETE FROM tags WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete tag: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return models.ErrTagNotFound
	}

	return nil
}

func (r *TagRepo) AddTagToTask(ctx context.Context, taskID, tagID int, names []string) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var tagExists bool
	err = tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM tags WHERE id = $1)", tagID).Scan(&tagExists)
	if err != nil {
		return fmt.Errorf("failed to check tag existence: %w", err)
	}
	if !tagExists {
		return models.ErrTaskNotFound
	}

	var taskExists bool
	err = tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM tasks WHERE id = $1)", taskID).Scan(&taskExists)
	if err != nil {
		return fmt.Errorf("failed to check task existence: %w", err)
	}
	if !taskExists {
		return models.ErrTaskNotFound
	}

	query := `
		INSERT INTO
			task_tags (task_id, tag_id) 
		VALUES ($1, $2) 
			ON CONFLICT DO NOTHING
	`

	result, err := tx.ExecContext(ctx, query, taskID, tagID)
	if err != nil {
		return fmt.Errorf("failed to add tag to task: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("task %d already has tag %d", taskID, tagID)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *TagRepo) RemoveTagFromTask(ctx context.Context, taskID int, tagID int) error {
	query := `
		DELETE FROM 
			task_tags 
		WHERE 
			task_id = $1 AND tag_id = $2
	`

	result, err := r.db.ExecContext(ctx, query, taskID, tagID)
	if err != nil {
		return fmt.Errorf("failed to remove tag from task: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return models.ErrTagNotOnTask
	}

	return nil
}

func (r *TagRepo) GetTaskTags(ctx context.Context, taskID int) ([]models.Tag, error) {
	query := `
		SELECT
			t.id,
			t.names,
			t.created_at,
			t.updated_at
		FROM 
			tags t
		JOIN 
			task_tags tt
		ON
			t.id = tt.tag_id
		WHERE
			tt.task_id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task tags: %w", err)
	}
	defer rows.Close()

	var tags []models.Tag
	for rows.Next() {
		var tag models.Tag
		var names []string
		if err := rows.Scan(&tag.ID, pq.Array(&names), &tag.CreatedAt, &tag.UpdateAt); err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tag.Names = names
		tags = append(tags, tag)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tags: %w", err)
	}

	if tags == nil {
		tags = []models.Tag{}
	}

	return tags, nil
}

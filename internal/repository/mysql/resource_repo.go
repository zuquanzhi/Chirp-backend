package mysql

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/zuquanzhi/Chirp/backend/internal/domain"
)

type resourceRepository struct {
	db *sql.DB
}

func NewResourceRepository(db *sql.DB) domain.ResourceRepository {
	return &resourceRepository{db: db}
}

func (r *resourceRepository) Create(ctx context.Context, res *domain.Resource) error {
	stmt := `INSERT INTO resources(owner_id,title,description,filename,original_name,size,file_hash,status,created_at,subject,type) VALUES(?,?,?,?,?,?,?,?,?,?,?)`
	result, err := r.db.ExecContext(ctx, stmt, res.OwnerID, res.Title, res.Description, res.Filename, res.OriginalName, res.Size, res.FileHash, res.Status, time.Now(), res.Subject, res.Type)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	res.ID = id
	return nil
}

func (r *resourceRepository) List(ctx context.Context, status domain.ResourceStatus, search string) ([]domain.Resource, error) {
	query := `SELECT id,owner_id,title,description,filename,original_name,size,file_hash,status,created_at,COALESCE(subject,''),COALESCE(type,'') FROM resources WHERE 1=1`
	args := []interface{}{}

	if status != "" {
		query += ` AND status = ?`
		args = append(args, status)
	}
	if search != "" {
		query += ` AND (title LIKE ? OR description LIKE ?)`
		args = append(args, "%"+search+"%", "%"+search+"%")
	}
	query += ` ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.Resource
	for rows.Next() {
		var res domain.Resource
		if err := rows.Scan(&res.ID, &res.OwnerID, &res.Title, &res.Description, &res.Filename, &res.OriginalName, &res.Size, &res.FileHash, &res.Status, &res.CreatedAt, &res.Subject, &res.Type); err != nil {
			return nil, err
		}
		list = append(list, res)
	}
	return list, nil
}

func (r *resourceRepository) GetByID(ctx context.Context, id int64) (*domain.Resource, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id,owner_id,title,description,filename,original_name,size,file_hash,status,created_at,COALESCE(subject,''),COALESCE(type,'') FROM resources WHERE id = ?`, id)
	var res domain.Resource
	if err := row.Scan(&res.ID, &res.OwnerID, &res.Title, &res.Description, &res.Filename, &res.OriginalName, &res.Size, &res.FileHash, &res.Status, &res.CreatedAt, &res.Subject, &res.Type); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &res, nil
}

func (r *resourceRepository) UpdateStatus(ctx context.Context, id int64, status domain.ResourceStatus) error {
	_, err := r.db.ExecContext(ctx, `UPDATE resources SET status = ? WHERE id = ?`, status, id)
	return err
}

func (r *resourceRepository) GetByHash(ctx context.Context, hash string) ([]domain.Resource, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id,owner_id,title,description,filename,original_name,size,file_hash,status,created_at,COALESCE(subject,''),COALESCE(type,'') FROM resources WHERE file_hash = ?`, hash)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.Resource
	for rows.Next() {
		var res domain.Resource
		if err := rows.Scan(&res.ID, &res.OwnerID, &res.Title, &res.Description, &res.Filename, &res.OriginalName, &res.Size, &res.FileHash, &res.Status, &res.CreatedAt, &res.Subject, &res.Type); err != nil {
			return nil, err
		}
		list = append(list, res)
	}
	return list, nil
}

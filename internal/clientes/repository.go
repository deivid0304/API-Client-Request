package clientes

import (
	"context"
	"database/sql"
	"time"
)

type Cliente struct {
	ID        int64     `json:"id"`
	Nome      string    `json:"nome"`
	Email     string    `json:"email"`
	Telefone  string    `json:"telefone"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, nome, email, telefone string) (Cliente, error) {
	var cliente Cliente
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO clientes (nome, email, telefone)
		VALUES ($1, $2, $3)
		RETURNING id, nome, email, telefone, created_at, updated_at
	`, nome, email, telefone).Scan(&cliente.ID, &cliente.Nome, &cliente.Email, &cliente.Telefone, &cliente.CreatedAt, &cliente.UpdatedAt)
	return cliente, err
}

func (r *Repository) List(ctx context.Context) ([]Cliente, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, nome, email, telefone, created_at, updated_at
		FROM clientes
		ORDER BY id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	clientes := make([]Cliente, 0)
	for rows.Next() {
		var cliente Cliente
		if err := rows.Scan(&cliente.ID, &cliente.Nome, &cliente.Email, &cliente.Telefone, &cliente.CreatedAt, &cliente.UpdatedAt); err != nil {
			return nil, err
		}
		clientes = append(clientes, cliente)
	}
	return clientes, rows.Err()
}

func (r *Repository) FindByID(ctx context.Context, id int64) (Cliente, error) {
	var cliente Cliente
	err := r.db.QueryRowContext(ctx, `
		SELECT id, nome, email, telefone, created_at, updated_at
		FROM clientes
		WHERE id = $1
	`, id).Scan(&cliente.ID, &cliente.Nome, &cliente.Email, &cliente.Telefone, &cliente.CreatedAt, &cliente.UpdatedAt)
	return cliente, err
}

func (r *Repository) Update(ctx context.Context, id int64, nome, email, telefone string) (Cliente, error) {
	var cliente Cliente
	err := r.db.QueryRowContext(ctx, `
		UPDATE clientes
		SET nome = $1, email = $2, telefone = $3, updated_at = now()
		WHERE id = $4
		RETURNING id, nome, email, telefone, created_at, updated_at
	`, nome, email, telefone, id).Scan(&cliente.ID, &cliente.Nome, &cliente.Email, &cliente.Telefone, &cliente.CreatedAt, &cliente.UpdatedAt)
	return cliente, err
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM clientes WHERE id = $1`, id)
	return err
}

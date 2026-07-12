package clientes

import (
	"context"
	"errors"
	"strings"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, nome, email, telefone string) (Cliente, error) {
	if err := validate(nome, email, telefone); err != nil {
		return Cliente{}, err
	}
	return s.repo.Create(ctx, strings.TrimSpace(nome), strings.TrimSpace(email), strings.TrimSpace(telefone))
}

func (s *Service) List(ctx context.Context) ([]Cliente, error) {
	return s.repo.List(ctx)
}

func (s *Service) FindByID(ctx context.Context, id int64) (Cliente, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id int64, nome, email, telefone string) (Cliente, error) {
	if err := validate(nome, email, telefone); err != nil {
		return Cliente{}, err
	}
	return s.repo.Update(ctx, id, strings.TrimSpace(nome), strings.TrimSpace(email), strings.TrimSpace(telefone))
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func validate(nome, email, telefone string) error {
	if strings.TrimSpace(nome) == "" || strings.TrimSpace(email) == "" || strings.TrimSpace(telefone) == "" {
		return errors.New("nome, email e telefone sao obrigatorios")
	}
	if !strings.Contains(email, "@") {
		return errors.New("email invalido")
	}
	return nil
}

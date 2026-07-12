package clientes

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) (any, int, error) {
	var req struct {
		Nome     string `json:"nome"`
		Email    string `json:"email"`
		Telefone string `json:"telefone"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, http.StatusBadRequest, errors.New("json invalido")
	}
	cliente, err := h.service.Create(r.Context(), req.Nome, req.Email, req.Telefone)
	return cliente, http.StatusCreated, err
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) (any, int, error) {
	clientes, err := h.service.List(r.Context())
	return clientes, http.StatusOK, err
}

func (h *Handler) FindByID(w http.ResponseWriter, r *http.Request) (any, int, error) {
	id, err := idFromRequest(r)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	cliente, err := h.service.FindByID(r.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, http.StatusNotFound, errors.New("cliente nao encontrado")
	}
	return cliente, http.StatusOK, err
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) (any, int, error) {
	id, err := idFromRequest(r)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	var req struct {
		Nome     string `json:"nome"`
		Email    string `json:"email"`
		Telefone string `json:"telefone"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, http.StatusBadRequest, errors.New("json invalido")
	}
	cliente, err := h.service.Update(r.Context(), id, req.Nome, req.Email, req.Telefone)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, http.StatusNotFound, errors.New("cliente nao encontrado")
	}
	return cliente, http.StatusOK, err
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) (any, int, error) {
	id, err := idFromRequest(r)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	if err := h.service.Delete(r.Context(), id); err != nil {
		return nil, http.StatusInternalServerError, err
	}
	return map[string]string{"message": "cliente removido"}, http.StatusOK, nil
}

func idFromRequest(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("id invalido")
	}
	return id, nil
}

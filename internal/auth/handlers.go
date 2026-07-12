package auth

import (
	"encoding/json"
	"errors"
	"net/http"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) (any, int, error) {
	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, http.StatusBadRequest, errors.New("json invalido")
	}
	if req.Name == "" || req.Email == "" || len(req.Password) < 8 {
		return nil, http.StatusBadRequest, errors.New("nome, email e senha com 8+ caracteres sao obrigatorios")
	}
	res, err := h.service.Register(r.Context(), req.Name, req.Email, req.Password)
	return res, http.StatusCreated, err
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) (any, int, error) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, http.StatusBadRequest, errors.New("json invalido")
	}
	res, err := h.service.Login(r.Context(), req.Email, req.Password)
	if errors.Is(err, ErrInvalidCredentials) {
		return nil, http.StatusUnauthorized, err
	}
	return res, http.StatusOK, err
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) (any, int, error) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, http.StatusBadRequest, errors.New("json invalido")
	}
	res, err := h.service.Refresh(r.Context(), req.RefreshToken)
	return res, http.StatusOK, err
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) (any, int, error) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, http.StatusBadRequest, errors.New("json invalido")
	}
	if err := h.service.Logout(r.Context(), req.RefreshToken); err != nil {
		return nil, http.StatusInternalServerError, err
	}
	return map[string]string{"message": "sessao encerrada"}, http.StatusOK, nil
}

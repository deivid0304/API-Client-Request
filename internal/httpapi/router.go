package httpapi

import (
	"log/slog"
	"net/http"

	"api-cliente/internal/auth"
	"api-cliente/internal/clientes"
)

type endpoint func(http.ResponseWriter, *http.Request) (any, int, error)

func NewRouter(authHandler *auth.Handler, clienteHandler *clientes.Handler, tokens *auth.TokenManager) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("POST /auth/register", adapt(authHandler.Register))
	mux.HandleFunc("POST /auth/login", adapt(authHandler.Login))
	mux.HandleFunc("POST /auth/refresh", adapt(authHandler.Refresh))
	mux.HandleFunc("POST /auth/logout", adapt(authHandler.Logout))

	protected := AuthMiddleware(tokens)
	mux.Handle("POST /cliente", protected(http.HandlerFunc(adapt(clienteHandler.Create))))
	mux.Handle("GET /cliente", protected(http.HandlerFunc(adapt(clienteHandler.List))))
	mux.Handle("GET /cliente/{id}", protected(http.HandlerFunc(adapt(clienteHandler.FindByID))))
	mux.Handle("PUT /cliente/{id}", protected(http.HandlerFunc(adapt(clienteHandler.Update))))
	mux.Handle("DELETE /cliente/{id}", protected(http.HandlerFunc(adapt(clienteHandler.Delete))))

	return securityHeaders(mux)
}

func adapt(fn endpoint) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		payload, status, err := fn(w, r)
		if err != nil {
			if status == 0 {
				status = http.StatusInternalServerError
			}
			slog.Warn("request failed", "method", r.Method, "path", r.URL.Path, "status", status, "error", err)
			writeError(w, status, err.Error())
			return
		}
		if status == 0 {
			status = http.StatusOK
		}
		writeJSON(w, status, payload)
	}
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		next.ServeHTTP(w, r)
	})
}

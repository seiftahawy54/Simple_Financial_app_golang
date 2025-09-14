package routes

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"

	"finance_app/src/handlers"
)

func Routes(router chi.Router, h *handlers.AppHandler) {
	// Add middleware
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(60 * time.Second))

	// Add CORS middleware
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	router.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			resp := map[string]interface{}{
				"message":   "OK",
				"timestamp": time.Now(),
				"status":    "healthy",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(resp); err != nil {
				logrus.Error("Failed to encode health response: ", err)
			}
		})

		r.Route("/transactions", func(sub chi.Router) {
			sub.Get("/", h.TransactionService.GetAllTransactions)
			sub.Post("/", h.TransactionService.CreateTransaction)
			sub.Get("/{id}", h.TransactionService.GetTransactionByID)
			sub.Get("/account/{accountId}", h.TransactionService.GetTransactionsByAccountID)
		})
	})
}

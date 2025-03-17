package hermes

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/themedef/go-hermes/internal/contracts"
	"net/http"
)

type APIHandler struct {
	db  contracts.Store
	ctx context.Context
}

func NewAPIHandler(ctx context.Context, db contracts.Store) *APIHandler {
	return &APIHandler{db: db, ctx: ctx}
}

func (h *APIHandler) RunServer(port, prefix string, middlewares ...func(http.Handler) http.Handler) {
	mux := http.NewServeMux()
	if prefix != "" {
		prefix = "/" + prefix
	}

	handlers := map[string]http.HandlerFunc{
		prefix + "/set":     h.SetHandler,
		prefix + "/setnx":   h.SetNXHandler,
		prefix + "/setxx":   h.SetXXHandler,
		prefix + "/setcas":  h.SetCASHandler,
		prefix + "/get":     h.GetHandler,
		prefix + "/delete":  h.DeleteHandler,
		prefix + "/incr":    h.IncrHandler,
		prefix + "/decr":    h.DecrHandler,
		prefix + "/lpush":   h.LPushHandler,
		prefix + "/rpush":   h.RPushHandler,
		prefix + "/lpop":    h.LPopHandler,
		prefix + "/rpop":    h.RPopHandler,
		prefix + "/find":    h.FindByValueHandler,
		prefix + "/ttl":     h.UpdateTTLHandler,
		prefix + "/hset":    h.HSetHandler,
		prefix + "/hget":    h.HGetHandler,
		prefix + "/hdel":    h.HDelHandler,
		prefix + "/hgetall": h.HGetAllHandler,
	}

	for pattern, handler := range handlers {
		wrappedHandler := applyMiddleware(handler, middlewares...)
		mux.Handle(pattern, wrappedHandler)
	}

	fmt.Println("Server running on :"+port+" with prefix:", prefix)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		fmt.Println("Server failed to start:", err)
	}
}

func applyMiddleware(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for _, middleware := range middlewares {
		h = middleware(h)
	}
	return h
}

func helperEncodeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode JSON response", http.StatusInternalServerError)
		fmt.Println("Encode error:", err)
	}
}

func (h *APIHandler) SetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Key   string      `json:"key"`
		Value interface{} `json:"value"`
		TTL   int         `json:"ttl"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	if err := h.db.Set(h.ctx, req.Key, req.Value, req.TTL); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	helperEncodeJSON(w, map[string]interface{}{
		"message": "Set OK",
		"key":     req.Key,
		"ttl":     req.TTL,
	})
}

func (h *APIHandler) SetNXHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Key   string      `json:"key"`
		Value interface{} `json:"value"`
		TTL   int         `json:"ttl"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	ok, err := h.db.SetNX(h.ctx, req.Key, req.Value, req.TTL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	helperEncodeJSON(w, map[string]interface{}{
		"success": ok,
		"key":     req.Key,
		"ttl":     req.TTL,
	})
}

func (h *APIHandler) SetXXHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Key   string      `json:"key"`
		Value interface{} `json:"value"`
		TTL   int         `json:"ttl"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	ok, err := h.db.SetXX(h.ctx, req.Key, req.Value, req.TTL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	helperEncodeJSON(w, map[string]interface{}{
		"success": ok,
		"key":     req.Key,
		"ttl":     req.TTL,
	})
}

func (h *APIHandler) SetCASHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Key      string      `json:"key"`
		OldValue interface{} `json:"old_value"`
		NewValue interface{} `json:"new_value"`
		TTL      int         `json:"ttl"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	success, err := h.db.SetCAS(h.ctx, req.Key, req.OldValue, req.NewValue, req.TTL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	helperEncodeJSON(w, map[string]interface{}{
		"success": success,
		"key":     req.Key,
	})
}

func (h *APIHandler) GetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	key := r.URL.Query().Get("key")

	value, exists, _ := h.db.Get(h.ctx, key)
	if !exists {
		http.Error(w, "Key not found or expired", http.StatusNotFound)
		return
	}

	helperEncodeJSON(w, map[string]interface{}{
		"key":   key,
		"value": value,
	})
}

func (h *APIHandler) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Key string `json:"key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	success, _ := h.db.Delete(h.ctx, req.Key)
	if !success {
		http.Error(w, "Key not found", http.StatusNotFound)
		return
	}

	helperEncodeJSON(w, map[string]interface{}{
		"message": "Deleted",
		"key":     req.Key,
	})
}

func (h *APIHandler) IncrHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Key string `json:"key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	newVal, ok, err := h.db.Incr(h.ctx, req.Key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if !ok {
		http.Error(w, "INCR failed (not a number?)", http.StatusBadRequest)
		return
	}

	helperEncodeJSON(w, map[string]interface{}{
		"key":   req.Key,
		"value": newVal,
	})
}

func (h *APIHandler) DecrHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Key string `json:"key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	newVal, ok, err := h.db.Decr(h.ctx, req.Key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if !ok {
		http.Error(w, "DECR failed (not a number?)", http.StatusBadRequest)
		return
	}

	helperEncodeJSON(w, map[string]interface{}{
		"key":   req.Key,
		"value": newVal,
	})
}

func (h *APIHandler) LPushHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Key   string      `json:"key"`
		Value interface{} `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	if err := h.db.LPush(h.ctx, req.Key, req.Value); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	helperEncodeJSON(w, map[string]interface{}{
		"message": "LPUSH success",
		"key":     req.Key,
		"value":   req.Value,
	})
}

func (h *APIHandler) RPushHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Key   string      `json:"key"`
		Value interface{} `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	if err := h.db.RPush(h.ctx, req.Key, req.Value); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	helperEncodeJSON(w, map[string]interface{}{
		"message": "RPUSH success",
		"key":     req.Key,
		"value":   req.Value,
	})
}

func (h *APIHandler) LPopHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Key string `json:"key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	value, exists, err := h.db.LPop(h.ctx, req.Key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !exists {
		http.Error(w, "LPOP: list empty or not found", http.StatusNotFound)
		return
	}

	helperEncodeJSON(w, map[string]interface{}{
		"message": "LPOP success",
		"key":     req.Key,
		"value":   value,
	})
}

func (h *APIHandler) RPopHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Key string `json:"key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	value, exists, err := h.db.RPop(h.ctx, req.Key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !exists {
		http.Error(w, "RPOP: list empty or not found", http.StatusNotFound)
		return
	}

	helperEncodeJSON(w, map[string]interface{}{
		"message": "RPOP success",
		"key":     req.Key,
		"value":   value,
	})
}

func (h *APIHandler) FindByValueHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Value interface{} `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	keys, _ := h.db.FindByValue(h.ctx, req.Value)
	helperEncodeJSON(w, map[string]interface{}{
		"value": req.Value,
		"keys":  keys,
	})
}

func (h *APIHandler) UpdateTTLHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Key string `json:"key"`
		TTL int    `json:"ttl"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	if err := h.db.UpdateTTL(h.ctx, req.Key, req.TTL); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	helperEncodeJSON(w, map[string]interface{}{
		"message": "TTL updated",
		"key":     req.Key,
		"ttl":     req.TTL,
	})
}

func (h *APIHandler) HSetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Key   string      `json:"key"`
		Field string      `json:"field"`
		Value interface{} `json:"value"`
		TTL   int         `json:"ttl"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	if err := h.db.HSet(h.ctx, req.Key, req.Field, req.Value, req.TTL); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	helperEncodeJSON(w, map[string]interface{}{
		"message": "HSET success",
		"key":     req.Key,
		"field":   req.Field,
	})
}

func (h *APIHandler) HGetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	key := r.URL.Query().Get("key")
	field := r.URL.Query().Get("field")

	value, exists, err := h.db.HGet(h.ctx, key, field)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !exists {
		http.Error(w, "Field not found", http.StatusNotFound)
		return
	}

	helperEncodeJSON(w, map[string]interface{}{
		"key":   key,
		"field": field,
		"value": value,
	})
}

func (h *APIHandler) HDelHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Key   string `json:"key"`
		Field string `json:"field"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	if err := h.db.HDel(h.ctx, req.Key, req.Field); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	helperEncodeJSON(w, map[string]interface{}{
		"message": "HDEL success",
		"key":     req.Key,
		"field":   req.Field,
	})
}

func (h *APIHandler) HGetAllHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	key := r.URL.Query().Get("key")

	result, err := h.db.HGetAll(h.ctx, key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	helperEncodeJSON(w, map[string]interface{}{
		"key":    key,
		"fields": result,
	})
}

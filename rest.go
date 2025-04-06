package hermes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/themedef/go-hermes/internal/contracts"
	"net/http"
)

type APIHandler struct {
	db  contracts.StoreHandler
	ctx context.Context
}

func NewAPIHandler(ctx context.Context, db contracts.StoreHandler) *APIHandler {
	return &APIHandler{db: db, ctx: ctx}
}

func applyMiddleware(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for _, middleware := range middlewares {
		h = middleware(h)
	}
	return h
}

func decodeRequest(r *http.Request, v interface{}) error {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return errors.New("Invalid JSON body")
	}
	return nil
}

func requireMethod(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method != method {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return false
	}
	return true
}

func helperEncodeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode JSON response", http.StatusInternalServerError)
		fmt.Println("Encode error:", err)
	}
}

func (h *APIHandler) RunServer(port, prefix string, middlewares ...func(http.Handler) http.Handler) {
	mux := http.NewServeMux()
	if prefix != "" {
		prefix = "/" + prefix
	}
	handlers := map[string]http.HandlerFunc{
		prefix + "/set":           h.SetHandler,
		prefix + "/setnx":         h.SetNXHandler,
		prefix + "/setxx":         h.SetXXHandler,
		prefix + "/get":           h.GetHandler,
		prefix + "/setcas":        h.SetCASHandler,
		prefix + "/getset":        h.GetSetHandler,
		prefix + "/incr":          h.IncrHandler,
		prefix + "/decr":          h.DecrHandler,
		prefix + "/incrby":        h.IncrByHandler,
		prefix + "/decrby":        h.DecrByHandler,
		prefix + "/lpush":         h.LPushHandler,
		prefix + "/rpush":         h.RPushHandler,
		prefix + "/lpop":          h.LPopHandler,
		prefix + "/rpop":          h.RPopHandler,
		prefix + "/llen":          h.LLenHandler,
		prefix + "/lrange":        h.LRangeHandler,
		prefix + "/hset":          h.HSetHandler,
		prefix + "/hget":          h.HGetHandler,
		prefix + "/hdel":          h.HDelHandler,
		prefix + "/hgetall":       h.HGetAllHandler,
		prefix + "/hexists":       h.HExistsHandler,
		prefix + "/hlen":          h.HLenHandler,
		prefix + "/exists":        h.ExistsHandler,
		prefix + "/expire":        h.ExpireHandler,
		prefix + "/persist":       h.PersistHandler,
		prefix + "/type":          h.TypeHandler,
		prefix + "/details":       h.GetWithDetailsHandler,
		prefix + "/rename":        h.RenameHandler,
		prefix + "/find":          h.FindByValueHandler,
		prefix + "/delete":        h.DeleteHandler,
		prefix + "/dropall":       h.DropAllHandler,
		prefix + "/subscribe":     h.SubscribeHandler,
		prefix + "/subscriptions": h.ListSubscriptionsHandler,
		prefix + "/closeallsub":   h.CloseAllSubscriptionsHandler,
	}
	for pattern, handler := range handlers {
		mux.Handle(pattern, applyMiddleware(handler, middlewares...))
	}
	fmt.Println("Server running on :"+port+" with prefix:", prefix)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		fmt.Println("Server failed to start:", err)
	}
}

func (h *APIHandler) SetHandler(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	var req struct {
		Key   string      `json:"key"`
		Value interface{} `json:"value"`
		TTL   int         `json:"ttl"`
	}
	if err := decodeRequest(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	var req struct {
		Key   string      `json:"key"`
		Value interface{} `json:"value"`
		TTL   int         `json:"ttl"`
	}
	if err := decodeRequest(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	var req struct {
		Key   string      `json:"key"`
		Value interface{} `json:"value"`
		TTL   int         `json:"ttl"`
	}
	if err := decodeRequest(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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

func (h *APIHandler) GetHandler(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	key := r.URL.Query().Get("key")
	value, err := h.db.Get(h.ctx, key)
	if err != nil {
		if IsKeyNotFound(err) || IsKeyExpired(err) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	helperEncodeJSON(w, map[string]interface{}{
		"key":   key,
		"value": value,
	})
}

func (h *APIHandler) SetCASHandler(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	var req struct {
		Key      string      `json:"key"`
		OldValue interface{} `json:"old_value"`
		NewValue interface{} `json:"new_value"`
		TTL      int         `json:"ttl"`
	}
	if err := decodeRequest(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.db.SetCAS(h.ctx, req.Key, req.OldValue, req.NewValue, req.TTL); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	helperEncodeJSON(w, map[string]interface{}{
		"success": true,
		"key":     req.Key,
	})
}

func (h *APIHandler) GetSetHandler(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	var req struct {
		Key      string      `json:"key"`
		NewValue interface{} `json:"new_value"`
		TTL      int         `json:"ttl"`
	}
	if err := decodeRequest(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	oldVal, err := h.db.GetSet(h.ctx, req.Key, req.NewValue, req.TTL)
	if err != nil {
		if IsKeyExpired(err) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	helperEncodeJSON(w, map[string]interface{}{
		"key":      req.Key,
		"oldValue": oldVal,
		"newValue": req.NewValue,
	})
}

func (h *APIHandler) IncrHandler(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	var req struct {
		Key string `json:"key"`
	}
	if err := decodeRequest(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	newVal, err := h.db.Incr(h.ctx, req.Key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	helperEncodeJSON(w, map[string]interface{}{
		"key":   req.Key,
		"value": newVal,
	})
}

func (h *APIHandler) DecrHandler(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	var req struct {
		Key string `json:"key"`
	}
	if err := decodeRequest(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	newVal, err := h.db.Decr(h.ctx, req.Key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	helperEncodeJSON(w, map[string]interface{}{
		"key":   req.Key,
		"value": newVal,
	})
}

func (h *APIHandler) IncrByHandler(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	var req struct {
		Key       string `json:"key"`
		Increment int64  `json:"increment"`
	}
	if err := decodeRequest(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	newVal, err := h.db.IncrBy(h.ctx, req.Key, req.Increment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	helperEncodeJSON(w, map[string]interface{}{
		"key":   req.Key,
		"value": newVal,
	})
}

func (h *APIHandler) DecrByHandler(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	var req struct {
		Key       string `json:"key"`
		Decrement int64  `json:"decrement"`
	}
	if err := decodeRequest(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	newVal, err := h.db.DecrBy(h.ctx, req.Key, req.Decrement)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	helperEncodeJSON(w, map[string]interface{}{
		"key":   req.Key,
		"value": newVal,
	})
}

func (h *APIHandler) LPushHandler(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	var req struct {
		Key    string        `json:"key"`
		Values []interface{} `json:"values"`
	}
	if err := decodeRequest(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if len(req.Values) == 0 {
		http.Error(w, "At least one value required", http.StatusBadRequest)
		return
	}
	if err := h.db.LPush(h.ctx, req.Key, req.Values...); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	helperEncodeJSON(w, map[string]interface{}{
		"message": "LPUSH success",
		"key":     req.Key,
		"count":   len(req.Values),
	})
}

func (h *APIHandler) RPushHandler(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	var req struct {
		Key    string        `json:"key"`
		Values []interface{} `json:"values"`
	}
	if err := decodeRequest(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if len(req.Values) == 0 {
		http.Error(w, "At least one value required", http.StatusBadRequest)
		return
	}
	if err := h.db.RPush(h.ctx, req.Key, req.Values...); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	helperEncodeJSON(w, map[string]interface{}{
		"message": "RPUSH success",
		"key":     req.Key,
		"count":   len(req.Values),
	})
}

func (h *APIHandler) LPopHandler(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	var req struct {
		Key string `json:"key"`
	}
	if err := decodeRequest(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	value, err := h.db.LPop(h.ctx, req.Key)
	if err != nil {
		if IsKeyNotFound(err) || IsEmptyList(err) {
			http.Error(w, "LPOP: list empty or not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	helperEncodeJSON(w, map[string]interface{}{
		"message": "LPOP success",
		"key":     req.Key,
		"value":   value,
	})
}

func (h *APIHandler) RPopHandler(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	var req struct {
		Key string `json:"key"`
	}
	if err := decodeRequest(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	value, err := h.db.RPop(h.ctx, req.Key)
	if err != nil {
		if IsKeyNotFound(err) || IsEmptyList(err) {
			http.Error(w, "RPOP: list empty or not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	helperEncodeJSON(w, map[string]interface{}{
		"message": "RPOP success",
		"key":     req.Key,
		"value":   value,
	})
}

func (h *APIHandler) LLenHandler(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	key := r.URL.Query().Get("key")
	length, err := h.db.LLen(h.ctx, key)
	if err != nil {
		if IsKeyNotFound(err) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	helperEncodeJSON(w, map[string]interface{}{
		"key":    key,
		"length": length,
	})
}

func (h *APIHandler) LRangeHandler(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	var req struct {
		Key   string `json:"key"`
		Start int    `json:"start"`
		End   int    `json:"end"`
	}
	if err := decodeRequest(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	result, err := h.db.LRange(h.ctx, req.Key, req.Start, req.End)
	if err != nil {
		if IsKeyNotFound(err) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	helperEncodeJSON(w, map[string]interface{}{
		"key":    req.Key,
		"start":  req.Start,
		"end":    req.End,
		"result": result,
	})
}

func (h *APIHandler) HSetHandler(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	var req struct {
		Key   string      `json:"key"`
		Field string      `json:"field"`
		Value interface{} `json:"value"`
		TTL   int         `json:"ttl"`
	}
	if err := decodeRequest(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	key := r.URL.Query().Get("key")
	field := r.URL.Query().Get("field")
	value, err := h.db.HGet(h.ctx, key, field)
	if err != nil {
		if IsKeyNotFound(err) {
			http.Error(w, "Field not found or key not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	helperEncodeJSON(w, map[string]interface{}{
		"key":   key,
		"field": field,
		"value": value,
	})
}

func (h *APIHandler) HDelHandler(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodDelete) {
		return
	}
	var req struct {
		Key   string `json:"key"`
		Field string `json:"field"`
	}
	if err := decodeRequest(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.db.HDel(h.ctx, req.Key, req.Field); err != nil {
		if IsKeyNotFound(err) {
			http.Error(w, "Field or key not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	helperEncodeJSON(w, map[string]interface{}{
		"message": "HDEL success",
		"key":     req.Key,
		"field":   req.Field,
	})
}

func (h *APIHandler) HGetAllHandler(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	key := r.URL.Query().Get("key")
	result, err := h.db.HGetAll(h.ctx, key)
	if err != nil {
		if IsKeyNotFound(err) {
			http.Error(w, "Key not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	helperEncodeJSON(w, map[string]interface{}{
		"key":    key,
		"fields": result,
	})
}

func (h *APIHandler) HExistsHandler(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	key := r.URL.Query().Get("key")
	field := r.URL.Query().Get("field")
	exists, err := h.db.HExists(h.ctx, key, field)
	if err != nil {
		if IsKeyNotFound(err) {
			http.Error(w, "Key not found (or expired)", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	helperEncodeJSON(w, map[string]interface{}{
		"key":    key,
		"field":  field,
		"exists": exists,
	})
}

func (h *APIHandler) HLenHandler(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	key := r.URL.Query().Get("key")
	length, err := h.db.HLen(h.ctx, key)
	if err != nil {
		if IsKeyNotFound(err) {
			http.Error(w, "Key not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	helperEncodeJSON(w, map[string]interface{}{
		"key":    key,
		"length": length,
	})
}

func (h *APIHandler) ExistsHandler(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	key := r.URL.Query().Get("key")
	exists, err := h.db.Exists(h.ctx, key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	helperEncodeJSON(w, map[string]interface{}{
		"key":    key,
		"exists": exists,
	})
}

func (h *APIHandler) ExpireHandler(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	var req struct {
		Key string `json:"key"`
		TTL int    `json:"ttl"`
	}
	if err := decodeRequest(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	success, err := h.db.Expire(h.ctx, req.Key, req.TTL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !success {
		http.Error(w, "Key not found", http.StatusNotFound)
		return
	}
	helperEncodeJSON(w, map[string]interface{}{
		"key":     req.Key,
		"ttl":     req.TTL,
		"success": success,
	})
}

func (h *APIHandler) PersistHandler(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	var req struct {
		Key string `json:"key"`
	}
	if err := decodeRequest(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	success, err := h.db.Persist(h.ctx, req.Key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !success {
		http.Error(w, "Key not found or already persistent", http.StatusNotFound)
		return
	}
	helperEncodeJSON(w, map[string]interface{}{
		"key":     req.Key,
		"success": success,
	})
}

func (h *APIHandler) TypeHandler(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	key := r.URL.Query().Get("key")
	typ, err := h.db.Type(h.ctx, key)
	if err != nil {
		if IsKeyNotFound(err) {
			http.Error(w, "Key not found or expired", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	helperEncodeJSON(w, map[string]interface{}{
		"key":  key,
		"type": typ,
	})
}

func (h *APIHandler) GetWithDetailsHandler(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	key := r.URL.Query().Get("key")
	value, ttl, err := h.db.GetWithDetails(h.ctx, key)
	if err != nil {
		if IsKeyNotFound(err) || IsKeyExpired(err) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	helperEncodeJSON(w, map[string]interface{}{
		"key":   key,
		"value": value,
		"ttl":   ttl,
	})
}

func (h *APIHandler) RenameHandler(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	var req struct {
		OldKey string `json:"old_key"`
		NewKey string `json:"new_key"`
	}
	if err := decodeRequest(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.db.Rename(h.ctx, req.OldKey, req.NewKey); err != nil {
		if IsKeyNotFound(err) {
			http.Error(w, "Old key not found", http.StatusNotFound)
		} else if IsKeyExists(err) {
			http.Error(w, "New key already exists", http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	helperEncodeJSON(w, map[string]interface{}{
		"message": "Rename success",
		"oldKey":  req.OldKey,
		"newKey":  req.NewKey,
	})
}

func (h *APIHandler) FindByValueHandler(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	var req struct {
		Value interface{} `json:"value"`
	}
	if err := decodeRequest(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	keys, err := h.db.FindByValue(h.ctx, req.Value)
	if err != nil {
		if IsKeyNotFound(err) {
			http.Error(w, "No keys found for this value", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	helperEncodeJSON(w, map[string]interface{}{
		"value": req.Value,
		"keys":  keys,
	})
}

func (h *APIHandler) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	var req struct {
		Key string `json:"key"`
	}
	if err := decodeRequest(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.db.Delete(h.ctx, req.Key); err != nil {
		if IsKeyNotFound(err) {
			http.Error(w, "Key not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	helperEncodeJSON(w, map[string]interface{}{
		"message": "Deleted",
		"key":     req.Key,
	})
}

func (h *APIHandler) DropAllHandler(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	if err := h.db.DropAll(h.ctx); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	helperEncodeJSON(w, map[string]interface{}{
		"message": "All keys dropped",
	})
}

func (h *APIHandler) SubscribeHandler(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Missing key parameter", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}
	msgChan := h.db.Subscribe(key)
	defer h.db.Unsubscribe(key, msgChan)
	fmt.Fprintf(w, "data: Subscribed to %s\n\n", key)
	flusher.Flush()
	ctx := r.Context()
	for {
		select {
		case msg, ok := <-msgChan:
			if !ok {
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		case <-ctx.Done():
			return
		}
	}
}

func (h *APIHandler) ListSubscriptionsHandler(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	subs := h.db.ListSubscriptions()
	helperEncodeJSON(w, map[string]interface{}{
		"subscriptions": subs,
	})
}

func (h *APIHandler) CloseAllSubscriptionsHandler(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	var req struct {
		Key string `json:"key"`
	}
	if err := decodeRequest(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.Key == "" {
		http.Error(w, "Missing key", http.StatusBadRequest)
		return
	}
	h.db.CloseAllSubscriptionsForKey(req.Key)
	helperEncodeJSON(w, map[string]interface{}{
		"message": "All subscriptions closed for key",
		"key":     req.Key,
	})
}

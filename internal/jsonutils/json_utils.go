package jsonutils

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/GFernandesS/go-bid/internal/validator"
)

func EncodeJson[T any](w http.ResponseWriter, r *http.Request, statusCode int, data T) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		return fmt.Errorf("failed to encode json %w", err)
	}

	return nil
}

func EncodeInternalError(w http.ResponseWriter, r *http.Request, customMessage ...string) {
	message := "unexpected error"

	if len(customMessage) > 0 {
		message = customMessage[0]
	}

	_ = EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
		"message": message,
	})
}

func DecodeValidJson[T validator.Validator](r *http.Request) (T, map[string]string, error) {
	var data T

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		return data, nil, fmt.Errorf("failed to decode json %w", err)
	}

	if problems := data.Valid(r.Context()); len(problems) > 0 {
		return data, problems, fmt.Errorf("invalid %T %d problems", data, len(problems))
	}

	return data, nil, nil
}

func DecodeJson[T any](r *http.Request) (T, error) {
	var data T

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		return data, fmt.Errorf("failed to decode json %w", err)
	}

	return data, nil
}

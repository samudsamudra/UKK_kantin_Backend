package app

import (
	"strings"

	"github.com/go-playground/validator/v10"
)

// ValidationErrorResponse mengubah error validator jadi pesan user-friendly
func ValidationErrorResponse(err error) map[string]string {
	out := map[string]string{}

	verrs, ok := err.(validator.ValidationErrors)
	if !ok {
		out["error"] = "input tidak valid"
		return out
	}

	for _, e := range verrs {
		field := strings.ToLower(e.Field())

		switch e.Tag() {
		case "required":
			out[field] = "wajib diisi"
		case "email":
			out[field] = "format email tidak valid"
		case "min":
			out[field] = "panjang atau nilai terlalu kecil"
		case "gt":
			out[field] = "nilai harus lebih besar"
		case "oneof":
			out[field] = "nilai tidak valid"
		default:
			out[field] = "input tidak valid"
		}
	}

	return out
}

// SimpleErrorResponse untuk error umum
func SimpleErrorResponse(msg string) map[string]string {
	return map[string]string{
		"error": msg,
	}
}

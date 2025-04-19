package auth

import (
	"errors"
	"net/http"
	"strings"
)

func GetBearerToken(headers http.Header) (string, error) {
	bearerToken := headers.Get("Authorization")

	token, found := strings.CutPrefix(bearerToken, "Bearer ")
	if !found {
		return "", errors.New("no token found")
	}

	return token, nil
}

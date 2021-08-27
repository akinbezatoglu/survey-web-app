package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/sirupsen/logrus"
)

// ContextKey implements type for context key
type ContextKey string

// ContextJWTKey is the key for the jwt context value
const ContextJWTKey ContextKey = "jwt"

// logMiddleware handles logging
func (a *API) logMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logrus.WithFields(logrus.Fields{
			"host":       r.Host,
			"address":    r.RemoteAddr,
			"method":     r.Method,
			"requestURI": r.RequestURI,
			"proto":      r.Proto,
			"useragent":  r.UserAgent(),
		}).Info("HTTP request information")

		next.ServeHTTP(w, r)
	})
}

// corsMiddleware handles preflight
func (a *API) corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Authorization")
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")

		next.ServeHTTP(w, r)
	})
}

// jwtMiddleware handles authentication via jwt's
func (a *API) jwtMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get authentication header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Check if authentication token is present
		authHeaderParts := strings.Split(authHeader, " ")
		if len(authHeaderParts) != 2 || strings.ToLower(authHeaderParts[0]) != "bearer" {
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Validate authentication token
		claims, err := a.parseJWT(authHeaderParts[1])
		if err != nil {
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), ContextJWTKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// parseJWT parses and validates a token using the HMAC signing method
func (a *API) parseJWT(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(a.config.SigningSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("jwt validation failed")
}

// createJWT creates, signs, and encodes a JWT token using the HMAC signing method
func (a *API) createJWT(claims jwt.MapClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(a.config.SigningSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

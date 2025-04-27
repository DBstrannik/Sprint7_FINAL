package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCafeNegative(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		request string
		status  int
		message string
	}{
		{"/cafe", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=omsk", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=tula&count=na", http.StatusBadRequest, "incorrect count"},
	}

	for _, v := range requests {
		t.Run(v.request, func(t *testing.T) {
			response := httptest.NewRecorder()
			req := httptest.NewRequest("GET", v.request, nil)
			handler.ServeHTTP(response, req)

			assert.Equal(t, v.status, response.Code)
			assert.Equal(t, v.message, strings.TrimSpace(response.Body.String()))
		})
	}
}

func TestCafeWhenOk(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		name   string
		url    string
		status int
	}{
		{"Valid count and city", "/cafe?count=2&city=moscow", http.StatusOK},
		{"Only city", "/cafe?city=tula", http.StatusOK},
		{"City and search", "/cafe?city=moscow&search=ложка", http.StatusOK},
	}

	for _, v := range requests {
		t.Run(v.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			req := httptest.NewRequest("GET", v.url, nil)
			handler.ServeHTTP(response, req)

			require.Equal(t, v.status, response.Code)
			if v.status == http.StatusOK {
				assert.NotEmpty(t, strings.TrimSpace(response.Body.String()))
			}
		})
	}
}

func TestCafeCount(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	tests := []struct {
		name     string
		url      string
		wantLen  int
		wantCode int
	}{
		{"Count 0", "/cafe?city=moscow&count=0", 0, http.StatusOK},
		{"Count 1", "/cafe?city=moscow&count=1", 1, http.StatusOK},
		{"Count 2", "/cafe?city=moscow&count=2", 2, http.StatusOK},
		{"Count more than available", "/cafe?city=moscow&count=100", len(cafeList["moscow"]), http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			req := httptest.NewRequest("GET", tt.url, nil)
			handler.ServeHTTP(response, req)

			require.Equal(t, tt.wantCode, response.Code)
			body := strings.TrimSpace(response.Body.String())

			if tt.wantLen == 0 {
				assert.Empty(t, body)
			} else {
				cafes := strings.Split(body, ",")
				assert.Equal(t, tt.wantLen, len(cafes))
			}
		})
	}
}

func TestCafeSearch(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	tests := []struct {
		name     string
		url      string
		wantLen  int
		contains string
	}{
		{"Search 'фасоль' (no results)", "/cafe?city=moscow&search=фасоль", 0, ""},
		{"Search 'кофе' (2 results)", "/cafe?city=moscow&search=кофе", 2, "кофе"},
		{"Search 'вилка' (1 result)", "/cafe?city=moscow&search=вилка", 1, "вилка"},
		{"Search case insensitive", "/cafe?city=moscow&search=ЛОЖКА", 1, "ложка"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			req := httptest.NewRequest("GET", tt.url, nil)
			handler.ServeHTTP(response, req)

			require.Equal(t, http.StatusOK, response.Code)
			body := strings.TrimSpace(response.Body.String())

			if tt.wantLen == 0 {
				assert.Empty(t, body)
			} else {
				cafes := strings.Split(body, ",")
				assert.Equal(t, tt.wantLen, len(cafes))
				for _, cafe := range cafes {
					assert.Contains(t, strings.ToLower(cafe), strings.ToLower(tt.contains))
				}
			}
		})
	}
}

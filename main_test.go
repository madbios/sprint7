package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v.request, nil)
		handler.ServeHTTP(response, req)
		assert.Equal(t, v.status, response.Code)
		assert.Equal(t, v.message, strings.TrimSpace(response.Body.String()))
	}
}

func TestCafeWhenOk(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)
	requests := []string{
		"/cafe?count=2&city=moscow",
		"/cafe?city=tula",
		"/cafe?city=moscow&search=ложка",
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v, nil)
		handler.ServeHTTP(response, req)
		assert.Equal(t, http.StatusOK, response.Code)
	}
}
func TestCafeCount(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)
	requests := []struct {
		count int
		want  int
	}{
		{count: 0, want: 0},
		{count: 1, want: 1},
		{count: 2, want: 2},
		{count: 100, want: min(100, len(cafeList["tula"]))},
	}
	for _, v := range requests {
		var got int
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", fmt.Sprintf("/?city=%s&count=%d", "tula", v.count), nil)
		handler.ServeHTTP(response, req)
		assert.Equal(t, http.StatusOK, response.Code, "count=%d", v.count)
		body := response.Body.String()
		if body != "" {
			got = len(strings.Split(body, ","))
		}
		assert.Equal(t, v.want, got, "count=%d", v.count)
	}
}
func TestCafeSearch(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)
	requests := []struct {
		search    string
		wantCount int
	}{
		{search: "фасоль", wantCount: 0},
		{search: "кофе", wantCount: 2},
		{search: "вилка", wantCount: 1},
	}
	for _, v := range requests {
		var got int
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", fmt.Sprintf("/?city=moscow&search=%s", url.QueryEscape(v.search)), nil)
		handler.ServeHTTP(response, req)
		assert.Equal(t, http.StatusOK, response.Code, "search=%s", v.search)
		body := response.Body.String()
		if body != "" {
			parts := strings.Split(body, ",")
			got = len(parts)

			searchUpper := strings.ToUpper(v.search)
			for _, name := range parts {
				assert.True(t,
					strings.Contains(strings.ToUpper(name), searchUpper),
					"search=%s: %q не содержит подстроку", v.search, name,
				)
			}
		}
		assert.Equal(t, v.wantCount, got, "search=%s", v.search)
	}
}

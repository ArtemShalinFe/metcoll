package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdate(t *testing.T) {

	type want struct {
		code int
	}
	tests := []struct {
		name    string
		request *http.Request
		want    want
	}{
		{
			name:    "positive #1",
			request: httptest.NewRequest(http.MethodPost, "/update/gauge/metric/1.0", nil),
			want:    want{code: 200},
		},
		{
			name:    "negative #2",
			request: httptest.NewRequest(http.MethodPost, "/update/gauge/", nil),
			want:    want{code: 404},
		},
		{
			name:    "negative #3",
			request: httptest.NewRequest(http.MethodPost, "/update/gauge/metric/1e10", nil),
			want:    want{code: 200},
		},
		{
			name:    "negative #4",
			request: httptest.NewRequest(http.MethodPost, "/update/gauge/metric/novalue", nil),
			want:    want{code: 400},
		},
		{
			name:    "positive #5",
			request: httptest.NewRequest(http.MethodPost, "/update/counter/metric/1", nil),
			want:    want{code: 200},
		},
		{
			name:    "negative #6",
			request: httptest.NewRequest(http.MethodPost, "/update/counter/", nil),
			want:    want{code: 404},
		},
		{
			name:    "negative #7",
			request: httptest.NewRequest(http.MethodPost, "/update/counter/metric/novalue", nil),
			want:    want{code: 400},
		},
		{
			name:    "negative #8",
			request: httptest.NewRequest(http.MethodPost, "/update/summary/metric/novalue", nil),
			want:    want{code: 501},
		},
		{
			name:    "negative #9",
			request: httptest.NewRequest(http.MethodGet, "/update/summary/metric/novalue", nil),
			want:    want{code: 405},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			Update(w, tt.request)
			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, res.StatusCode, tt.want.code)
		})
	}
}

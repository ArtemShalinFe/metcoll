package main

import "testing"

func TestMetric_UriPathForPush(t *testing.T) {

	s := &Server{
		host: "localhost",
		port: "8080",
	}

	tests := []struct {
		name   string
		fields *Metric
		want   string
	}{
		{
			name:   "positive #1",
			fields: NewMetric("metric1", "1", counter),
			want:   "http://localhost:8080/update/counter/metric1/1",
		},
		{
			name:   "positive #2",
			fields: NewMetric("metric2", "2", counter),
			want:   "http://localhost:8080/update/counter/metric2/2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Metric{
				name:  tt.fields.name,
				value: tt.fields.value,
				mType: tt.fields.mType,
			}
			if got := m.URIPathForPush(s); got != tt.want {
				t.Errorf("Metric.UriPathForPush() = %v, want %v", got, tt.want)
			}
		})
	}
}

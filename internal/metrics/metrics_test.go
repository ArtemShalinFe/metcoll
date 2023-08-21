package metrics

import (
	"testing"

	"github.com/go-playground/assert"
)

func TestMetrics_IsPollCount(t *testing.T) {
	type fields struct {
		ID    string
		MType string
		Delta int64
		Value float64
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "negative case #1",
			fields: fields{
				ID:    "Alloc",
				MType: GaugeMetric,
			},
			want: false,
		},
		{
			name: "positive case #1",
			fields: fields{
				ID:    PollCount,
				MType: CounterMetric,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Metrics{
				ID:    tt.fields.ID,
				MType: tt.fields.MType,
				Delta: &tt.fields.Delta,
				Value: &tt.fields.Value,
			}
			if got := m.IsPollCount(); got != tt.want {
				t.Errorf("Metrics.IsPollCount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetMetric(t *testing.T) {
	type args struct {
		id    string
		mType string
	}
	tests := []struct {
		name    string
		args    args
		want    *Metrics
		wantErr bool
	}{
		{
			name: "counter",
			args: args{
				id:    "counter",
				mType: CounterMetric,
			},
			want:    NewCounterMetric("counter", 0),
			wantErr: false,
		},
		{
			name: "gauge",
			args: args{
				id:    "gauge",
				mType: GaugeMetric,
			},
			want:    NewGaugeMetric("gauge", 0),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetMetric(tt.args.id, tt.args.mType)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetMetric() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want.ID, got.ID)
			assert.Equal(t, tt.want.MType, got.MType)
		})
	}
}

func TestMetrics_String(t *testing.T) {

	tests := []struct {
		name    string
		metrics *Metrics
		want    string
	}{
		{
			name:    "",
			metrics: NewCounterMetric("counter", 1),
			want:    "1",
		},
		{
			name:    "",
			metrics: NewGaugeMetric("gauge", 1.1),
			want:    "1.1",
		},
		{
			name:    "",
			metrics: &Metrics{MType: "nope"},
			want:    "unknow metric type",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.metrics.String(); got != tt.want {
				t.Errorf("Metrics.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

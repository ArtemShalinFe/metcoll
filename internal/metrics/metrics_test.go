package metrics

import (
	"context"
	"testing"

	"github.com/go-playground/assert"
	gomock "go.uber.org/mock/gomock"
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

func TestMetrics_Get(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := NewMockStorage(ctrl)
	db.EXPECT().GetInt64Value(gomock.Any(), "counter").AnyTimes().Return(int64(1), nil)
	db.EXPECT().GetFloat64Value(gomock.Any(), "gauge").AnyTimes().Return(float64(1.1), nil)

	type fields struct {
		ID    string
		MType string
		Value string
	}
	type args struct {
		ctx     context.Context
		storage Storage
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		want    *Metrics
	}{
		{
			name: "#1 case",
			fields: fields{
				ID:    "counter",
				MType: CounterMetric,
				Value: "1",
			},
			args: args{
				ctx:     context.Background(),
				storage: db,
			},
			want: NewCounterMetric("counter", 1),
		},
		{
			name: "#2 case",
			fields: fields{
				ID:    "gauge",
				MType: GaugeMetric,
				Value: "1.1",
			},
			args: args{
				ctx:     context.Background(),
				storage: db,
			},
			want: NewGaugeMetric("gauge", 1.1),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			m, err := NewMetric(tt.fields.ID, tt.fields.MType, tt.fields.Value)
			if err != nil {
				t.Error(err)
			}
			if err := m.Get(tt.args.ctx, tt.args.storage); (err != nil) != tt.wantErr {
				t.Errorf("Metrics.Get() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

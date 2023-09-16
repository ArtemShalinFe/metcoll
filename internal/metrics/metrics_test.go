package metrics

import (
	"context"
	"testing"

	"github.com/go-playground/assert"
	gomock "go.uber.org/mock/gomock"
)

const (
	counter = "counter"
	gauge   = "gauge"
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
		want    *Metrics
		args    args
		name    string
		wantErr bool
	}{
		{
			name: counter,
			args: args{
				id:    counter,
				mType: CounterMetric,
			},
			want:    NewCounterMetric(counter, 0),
			wantErr: false,
		},
		{
			name: gauge,
			args: args{
				id:    gauge,
				mType: GaugeMetric,
			},
			want:    NewGaugeMetric(gauge, 0),
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
			metrics: NewCounterMetric(counter, 1),
			want:    "1",
		},
		{
			name:    "",
			metrics: NewGaugeMetric(gauge, 1.1),
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
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := NewMockStorage(ctrl)
	db.EXPECT().GetInt64Value(gomock.Any(), counter).AnyTimes().Return(int64(1), nil)
	db.EXPECT().GetFloat64Value(gomock.Any(), gauge).AnyTimes().Return(float64(1.1), nil)

	type fields struct {
		ID    string
		MType string
		Value string
	}
	type args struct {
		storage Storage
	}

	tests := []struct {
		args    args
		want    *Metrics
		fields  fields
		name    string
		wantErr bool
	}{
		{
			name: "#1 case",
			fields: fields{
				ID:    counter,
				MType: CounterMetric,
				Value: "1",
			},
			args: args{

				storage: db,
			},
			want: NewCounterMetric(counter, 1),
		},
		{
			name: "#2 case",
			fields: fields{
				ID:    gauge,
				MType: GaugeMetric,
				Value: "1.1",
			},
			args: args{

				storage: db,
			},
			want: NewGaugeMetric(gauge, 1.1),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			m, err := NewMetric(tt.fields.ID, tt.fields.MType, tt.fields.Value)
			if err != nil {
				t.Error(err)
			}
			if err := m.Get(ctx, tt.args.storage); (err != nil) != tt.wantErr {
				t.Errorf("Metrics.Get() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMetrics_Update(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := NewMockStorage(ctrl)
	db.EXPECT().AddInt64Value(gomock.Any(), counter, int64(1)).AnyTimes().Return(int64(2), nil)
	db.EXPECT().SetFloat64Value(gomock.Any(), gauge, float64(1.2)).AnyTimes().Return(float64(1.2), nil)

	type fields struct {
		ID    string
		MType string
		Value string
	}
	type args struct {
		storage Storage
	}

	tests := []struct {
		args    args
		want    *Metrics
		fields  fields
		name    string
		wantErr bool
	}{
		{
			name: "#1 case",
			fields: fields{
				ID:    counter,
				MType: CounterMetric,
				Value: "1",
			},
			args: args{

				storage: db,
			},
			want: NewCounterMetric(counter, 2),
		},
		{
			name: "#2 case",
			fields: fields{
				ID:    gauge,
				MType: GaugeMetric,
				Value: "1.2",
			},
			args: args{
				storage: db,
			},
			want: NewGaugeMetric(gauge, 1.2),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			m, err := NewMetric(tt.fields.ID, tt.fields.MType, tt.fields.Value)
			if err != nil {
				t.Error(err)
			}
			if err := m.Update(ctx, tt.args.storage); (err != nil) != tt.wantErr {
				t.Errorf("Metrics.Update() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBatchUpdate(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := NewMockStorage(ctrl)

	var ms []*Metrics
	ms = append(ms, NewCounterMetric(counter, 1),
		NewGaugeMetric(gauge, 1.2))

	var wantms []*Metrics
	wantms = append(wantms, NewCounterMetric(counter, 2),
		NewGaugeMetric(gauge, 1.2))

	counters := make(map[string]int64)
	counters[counter] = 1
	db.EXPECT().BatchAddInt64Value(gomock.Any(), counters).AnyTimes().Return(counters, nil, nil)

	gauges := make(map[string]float64)
	gauges[gauge] = 1.2
	db.EXPECT().BatchSetFloat64Value(gomock.Any(), gauges).AnyTimes().Return(gauges, nil, nil)

	tests := []struct {
		storage Storage
		name    string
		want    []*Metrics
		metrics []*Metrics
		wantErr bool
	}{
		{
			name:    "#1 case",
			storage: db,
			want:    wantms,
			metrics: ms,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ms, errs, err := BatchUpdate(ctx, tt.metrics, tt.storage)
			if err != nil {
				t.Errorf("test of batch update was failed, err: %v", err)
			}
			for _, er := range errs {
				t.Errorf("test of update was failed, err: %v", er)
			}
			for _, wantMetric := range tt.want {
				found := false
				for _, metric := range ms {
					if wantMetric.ID == metric.ID {
						found = true
					}
				}
				if !found {
					t.Errorf("metric %s not found", wantMetric.ID)
				}
			}
		})
	}
}

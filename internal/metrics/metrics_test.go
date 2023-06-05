package metrics

import "testing"

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

package storage

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v2"
	"go.uber.org/zap"
)

func TestDB_GetInt64Value(t *testing.T) {
	ctx := context.Background()

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	mock.ExpectBegin().WillReturnError(errors.New("some error"))

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT value FROM counters").WithArgs("keyOne").WillReturnRows(mock.NewRows([]string{"value"}).AddRow(int64(1)))
	mock.ExpectCommit()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT value FROM counters").WithArgs("keyTwo").WillReturnError(pgx.ErrNoRows)
	mock.ExpectCommit()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT value FROM counters").WithArgs("keyTwo").WillReturnError(errors.New("bad querry"))
	mock.ExpectRollback()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT value FROM counters").WithArgs("keyTwo").WillReturnError(errors.New("bad querry"))
	mock.ExpectRollback().WillReturnError(errors.New("fail rollback"))

	type fields struct {
		pool   PgxIface
		logger *zap.SugaredLogger
	}
	type args struct {
		ctx context.Context
		key string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "begin fail case",
			fields: fields{
				pool:   mock,
				logger: zap.L().Sugar(),
			},
			args: args{
				ctx: ctx,
				key: "keyOne",
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "positive case",
			fields: fields{
				pool:   mock,
				logger: zap.L().Sugar(),
			},
			args: args{
				ctx: ctx,
				key: "keyOne",
			},
			want:    1,
			wantErr: false,
		},
		{
			name: "negative case",
			fields: fields{
				pool:   mock,
				logger: zap.L().Sugar(),
			},
			args: args{
				ctx: ctx,
				key: "keyTwo",
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "negative case bad querry with rollback",
			fields: fields{
				pool:   mock,
				logger: zap.L().Sugar(),
			},
			args: args{
				ctx: ctx,
				key: "keyTwo",
			},
			want:    0,
			wantErr: false,
		},
		{
			name: "negative case bad querry with fail rollback",
			fields: fields{
				pool:   mock,
				logger: zap.L().Sugar(),
			},
			args: args{
				ctx: ctx,
				key: "keyTwo",
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			db := &DB{
				pool:   tt.fields.pool,
				logger: tt.fields.logger,
			}
			got, err := db.GetInt64Value(tt.args.ctx, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("DB.GetInt64Value() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DB.GetInt64Value() = %v, want %v", got, tt.want)
			}
		})
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDB_GetFloat64Value(t *testing.T) {
	ctx := context.Background()

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	mock.ExpectBegin().WillReturnError(errors.New("some error"))

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT delta FROM gauges").WithArgs("gaugeOne").WillReturnRows(mock.NewRows([]string{"delta"}).AddRow(float64(1.1)))
	mock.ExpectCommit()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT delta FROM gauges").WithArgs("gaugeTwo").WillReturnError(pgx.ErrNoRows)
	mock.ExpectCommit()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT delta FROM gauges").WithArgs("gaugeTwo").WillReturnError(errors.New("bad querry"))
	mock.ExpectRollback()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT delta FROM gauges").WithArgs("gaugeTwo").WillReturnError(errors.New("bad querry"))
	mock.ExpectRollback().WillReturnError(errors.New("fail rollback"))

	type fields struct {
		pool   PgxIface
		logger *zap.SugaredLogger
	}
	type args struct {
		ctx context.Context
		key string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    float64
		wantErr bool
	}{
		{
			name: "begin fail case",
			fields: fields{
				pool:   mock,
				logger: zap.L().Sugar(),
			},
			args: args{
				ctx: ctx,
				key: "gaugeOne",
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "positive case",
			fields: fields{
				pool:   mock,
				logger: zap.L().Sugar(),
			},
			args: args{
				ctx: ctx,
				key: "gaugeOne",
			},
			want:    1.1,
			wantErr: false,
		},
		{
			name: "negative case",
			fields: fields{
				pool:   mock,
				logger: zap.L().Sugar(),
			},
			args: args{
				ctx: ctx,
				key: "gaugeTwo",
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "negative case bad querry with rollback",
			fields: fields{
				pool:   mock,
				logger: zap.L().Sugar(),
			},
			args: args{
				ctx: ctx,
				key: "gaugeTwo",
			},
			want:    0,
			wantErr: false,
		},
		{
			name: "negative case bad querry with fail rollback",
			fields: fields{
				pool:   mock,
				logger: zap.L().Sugar(),
			},
			args: args{
				ctx: ctx,
				key: "gaugeTwo",
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			db := &DB{
				pool:   tt.fields.pool,
				logger: tt.fields.logger,
			}
			got, err := db.GetFloat64Value(tt.args.ctx, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("DB.GetFloat64Value() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DB.GetFloat64Value() = %v, want %v", got, tt.want)
			}
		})
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDB_GetDataList(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id, delta FROM gauges").WillReturnRows(mock.NewRows([]string{"id", "delta"}).AddRow("gaugeOne", float64(1.1)).AddRow("gaugeTwo", float64(1.2)))
	mock.ExpectCommit()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id, value FROM counters").WillReturnRows(mock.NewRows([]string{"id", "value"}).AddRow("counterOne", int64(1)).AddRow("counterTwo", int64(2)))
	mock.ExpectCommit()

	mock.ExpectBegin().WillReturnError(errors.New("some error"))

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id, delta FROM gauges").WillReturnRows(mock.NewRows([]string{"id", "delta"}).AddRow("gaugeOne", float64(1.1)).AddRow("gaugeTwo", float64(1.2)))
	mock.ExpectCommit()

	mock.ExpectBegin().WillReturnError(errors.New("some error"))

	type fields struct {
		pool   PgxIface
		logger *zap.SugaredLogger
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "positive case",
			fields: fields{
				pool:   mock,
				logger: zap.L().Sugar(),
			},
			args: args{
				ctx: context.Background(),
			},
			want:    []string{"gaugeOne 1.1", "gaugeTwo 1.2", "counterOne 1", "counterTwo 2"},
			wantErr: false,
		},
		{
			name: "negative case #1",
			fields: fields{
				pool:   mock,
				logger: zap.L().Sugar(),
			},
			args: args{
				ctx: context.Background(),
			},
			want:    []string{},
			wantErr: true,
		},
		{
			name: "negative case #2",
			fields: fields{
				pool:   mock,
				logger: zap.L().Sugar(),
			},
			args: args{
				ctx: context.Background(),
			},
			want:    []string{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			db := &DB{
				pool:   tt.fields.pool,
				logger: tt.fields.logger,
			}
			got, err := db.GetDataList(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("DB.GetFloat64Value() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DB.GetFloat64Value() = %v, want %v", got, tt.want)
			}
		})
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDB_AddInt64Value(t *testing.T) {
	ctx := context.Background()

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	mock.ExpectBegin().WillReturnError(errors.New("some error"))

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT delta FROM gauges").WithArgs("gaugeOne").WillReturnRows(mock.NewRows([]string{"delta"}).AddRow(float64(1.1)))
	mock.ExpectCommit()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT delta FROM gauges").WithArgs("gaugeTwo").WillReturnError(pgx.ErrNoRows)
	mock.ExpectCommit()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT delta FROM gauges").WithArgs("gaugeTwo").WillReturnError(errors.New("bad querry"))
	mock.ExpectRollback()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT delta FROM gauges").WithArgs("gaugeTwo").WillReturnError(errors.New("bad querry"))
	mock.ExpectRollback().WillReturnError(errors.New("fail rollback"))

	type fields struct {
		pool   PgxIface
		logger *zap.SugaredLogger
	}
	type args struct {
		ctx context.Context
		key string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    float64
		wantErr bool
	}{
		{
			name: "begin fail case",
			fields: fields{
				pool:   mock,
				logger: zap.L().Sugar(),
			},
			args: args{
				ctx: ctx,
				key: "gaugeOne",
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "positive case",
			fields: fields{
				pool:   mock,
				logger: zap.L().Sugar(),
			},
			args: args{
				ctx: ctx,
				key: "gaugeOne",
			},
			want:    1.1,
			wantErr: false,
		},
		{
			name: "negative case",
			fields: fields{
				pool:   mock,
				logger: zap.L().Sugar(),
			},
			args: args{
				ctx: ctx,
				key: "gaugeTwo",
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "negative case bad querry with rollback",
			fields: fields{
				pool:   mock,
				logger: zap.L().Sugar(),
			},
			args: args{
				ctx: ctx,
				key: "gaugeTwo",
			},
			want:    0,
			wantErr: false,
		},
		{
			name: "negative case bad querry with fail rollback",
			fields: fields{
				pool:   mock,
				logger: zap.L().Sugar(),
			},
			args: args{
				ctx: ctx,
				key: "gaugeTwo",
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			db := &DB{
				pool:   tt.fields.pool,
				logger: tt.fields.logger,
			}
			got, err := db.GetFloat64Value(tt.args.ctx, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("DB.GetFloat64Value() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DB.GetFloat64Value() = %v, want %v", got, tt.want)
			}
		})
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

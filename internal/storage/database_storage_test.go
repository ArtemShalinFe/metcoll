package storage

import (
	"context"
	"errors"
	"reflect"
	"syscall"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v2"
	"go.uber.org/zap"
)

func TestDB_createTables(t *testing.T) {
	ctx := context.Background()

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	mock.ExpectBegin().WillReturnError(errors.New("some error"))

	mock.ExpectBegin()
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS counters (.+)").WillReturnResult(pgxmock.NewResult("CREATE", 1))
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS gauges (.+)").WillReturnResult(pgxmock.NewResult("CREATE", 1))
	mock.ExpectCommit()

	mock.ExpectBegin()
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS counters").WillReturnError(syscall.ECONNREFUSED)
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS counters").WillReturnResult(pgxmock.NewResult("CREATE", 1))
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS gauges").WillReturnResult(pgxmock.NewResult("CREATE", 1))
	mock.ExpectCommit()

	mock.ExpectBegin()
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS counters").WillReturnError(errors.New("some bad error"))
	mock.ExpectRollback()

	mock.ExpectBegin()
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS counters").WillReturnResult(pgxmock.NewResult("CREATE", 1))
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS gauges").WillReturnError(syscall.ECONNREFUSED)
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS gauges").WillReturnResult(pgxmock.NewResult("CREATE", 1))
	mock.ExpectCommit()

	mock.ExpectBegin()
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS counters").WillReturnResult(pgxmock.NewResult("CREATE", 1))
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS gauges").WillReturnError(errors.New("some bad error"))
	mock.ExpectRollback()

	mock.ExpectBegin()
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS counters").WillReturnResult(pgxmock.NewResult("CREATE", 1))
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS gauges").WillReturnError(errors.New("some bad error"))
	mock.ExpectRollback().WillReturnError(errors.New("fail rollback"))

	type fields struct {
		pool   PgxIface
		logger *zap.SugaredLogger
	}
	tests := []struct {
		fields  fields
		ctx     context.Context
		name    string
		wantErr bool
	}{
		{
			name: "begin fail case",
			fields: fields{
				pool:   mock,
				logger: zap.L().Sugar(),
			},
			ctx:     ctx,
			wantErr: true,
		},
		{
			name: "positive case",
			fields: fields{
				pool:   mock,
				logger: zap.L().Sugar(),
			},
			ctx:     ctx,
			wantErr: false,
		},
		{
			name: "positive case with retry",
			fields: fields{
				pool:   mock,
				logger: zap.L().Sugar(),
			},
			ctx:     ctx,
			wantErr: false,
		},
		{
			name: "negative case creating counters",
			fields: fields{
				pool:   mock,
				logger: zap.L().Sugar(),
			},
			ctx:     ctx,
			wantErr: false,
		},
		{
			name: "positive case with retry #2",
			fields: fields{
				pool:   mock,
				logger: zap.L().Sugar(),
			},
			ctx:     ctx,
			wantErr: false,
		},
		{
			name: "negative case creating gauges",
			fields: fields{
				pool:   mock,
				logger: zap.L().Sugar(),
			},
			ctx:     ctx,
			wantErr: false,
		},
		{
			name: "negative case with failed rollback",
			fields: fields{
				pool:   mock,
				logger: zap.L().Sugar(),
			},
			ctx:     ctx,
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
			err := db.createTables(tt.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("DB.createTables() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

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
	mock.ExpectQuery("INSERT (.+)").WithArgs("counterOne", int64(1)).WillReturnRows(mock.NewRows([]string{"value"}).AddRow(int64(2)))
	mock.ExpectCommit()

	mock.ExpectBegin()
	mock.ExpectQuery("INSERT (.+)").WithArgs("counterTwo", int64(1)).WillReturnError(errors.New("some insert errors"))
	mock.ExpectCommit()

	mock.ExpectBegin()
	mock.ExpectQuery("INSERT (.+)").WithArgs("counterThree", int64(3)).WillReturnError(errors.New("bad querry"))
	mock.ExpectRollback().WillReturnError(errors.New("fail rollback"))

	type fields struct {
		pool   PgxIface
		logger *zap.SugaredLogger
	}
	type args struct {
		ctx   context.Context
		key   string
		value int64
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
			args:    args{},
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
				ctx:   ctx,
				key:   "counterOne",
				value: 1,
			},
			want:    2,
			wantErr: false,
		},
		{
			name: "negative case",
			fields: fields{
				pool:   mock,
				logger: zap.L().Sugar(),
			},
			args: args{
				ctx:   ctx,
				key:   "counterTwo",
				value: 1,
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "negative case with fail rollback",
			fields: fields{
				pool:   mock,
				logger: zap.L().Sugar(),
			},
			args: args{
				ctx:   ctx,
				key:   "counterThree",
				value: 3,
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
			got, err := db.AddInt64Value(tt.args.ctx, tt.args.key, tt.args.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("DB.AddInt64Value() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DB.AddInt64Value() = %v, want %v", got, tt.want)
			}
		})
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDB_SetFloat64Value(t *testing.T) {
	ctx := context.Background()

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	mock.ExpectBegin().WillReturnError(errors.New("some error"))

	mock.ExpectBegin()
	mock.ExpectQuery("INSERT (.+)").WithArgs("gaugeOne", float64(1.1)).WillReturnRows(mock.NewRows([]string{"value"}).AddRow(float64(1.1)))
	mock.ExpectCommit()

	mock.ExpectBegin()
	mock.ExpectQuery("INSERT (.+)").WithArgs("gaugeTwo", float64(1.2)).WillReturnError(errors.New("some insert errors"))
	mock.ExpectCommit()

	mock.ExpectBegin()
	mock.ExpectQuery("INSERT (.+)").WithArgs("gaugeThree", float64(1.3)).WillReturnError(errors.New("bad querry"))
	mock.ExpectRollback().WillReturnError(errors.New("fail rollback"))

	type fields struct {
		pool   PgxIface
		logger *zap.SugaredLogger
	}
	type args struct {
		ctx   context.Context
		key   string
		value float64
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
			args:    args{},
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
				ctx:   ctx,
				key:   "gaugeOne",
				value: 1.1,
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
				ctx:   ctx,
				key:   "gaugeTwo",
				value: 1.2,
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "negative case with fail rollback",
			fields: fields{
				pool:   mock,
				logger: zap.L().Sugar(),
			},
			args: args{
				ctx:   ctx,
				key:   "gaugeThree",
				value: 1.3,
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
			got, err := db.SetFloat64Value(tt.args.ctx, tt.args.key, tt.args.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("DB.SetFloat64Value() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DB.SetFloat64Value() = %v, want %v", got, tt.want)
			}
		})
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

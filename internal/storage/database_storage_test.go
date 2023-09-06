package storage

import (
	"context"
	"errors"
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

	mock.ExpectBegin().WillReturnError(errors.New("create tables tx begin error"))

	mock.ExpectBegin()
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS counters (.+)").
		WillReturnResult(pgxmock.NewResult("CREATE", 1))
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS gauges (.+)").
		WillReturnResult(pgxmock.NewResult("CREATE", 1))
	mock.ExpectCommit()

	const cgq = "CREATE TABLE IF NOT EXISTS gauges"
	const ccq = "CREATE TABLE IF NOT EXISTS counters"

	mock.ExpectBegin()
	mock.ExpectExec(ccq).
		WillReturnError(syscall.ECONNREFUSED)
	mock.ExpectExec(ccq).
		WillReturnResult(pgxmock.NewResult("CREATE", 1))
	mock.ExpectExec(cgq).
		WillReturnResult(pgxmock.NewResult("CREATE", 1))
	mock.ExpectCommit()

	mock.ExpectBegin()
	mock.ExpectExec(ccq).
		WillReturnError(errors.New("exec querry error"))
	mock.ExpectRollback()

	mock.ExpectBegin()
	mock.ExpectExec(ccq).
		WillReturnResult(pgxmock.NewResult("CREATE", 1))
	mock.ExpectExec(cgq).
		WillReturnError(syscall.ECONNREFUSED)
	mock.ExpectExec(cgq).
		WillReturnResult(pgxmock.NewResult("CREATE", 1))
	mock.ExpectCommit()

	mock.ExpectBegin()
	mock.ExpectExec(ccq).
		WillReturnResult(pgxmock.NewResult("CREATE", 1))
	mock.ExpectExec(cgq).
		WillReturnError(errors.New("error syntax querry"))
	mock.ExpectRollback()

	mock.ExpectBegin()
	mock.ExpectExec(ccq).
		WillReturnResult(pgxmock.NewResult("CREATE", 1))
	mock.ExpectExec(cgq).
		WillReturnError(errors.New("some bad error"))
	mock.ExpectRollback().WillReturnError(errors.New("fail rollback"))

	type fields struct {
		pool   PgxIface
		logger *zap.SugaredLogger
	}
	tests := []struct {
		fields  fields
		name    string
		wantErr bool
	}{
		{
			name: "begin fail case",
			fields: fields{
				pool:   mock,
				logger: zap.L().Sugar(),
			},

			wantErr: true,
		},
		{
			name: "positive case",
			fields: fields{
				pool:   mock,
				logger: zap.L().Sugar(),
			},

			wantErr: false,
		},
		{
			name: "positive case with retry",
			fields: fields{
				pool:   mock,
				logger: zap.L().Sugar(),
			},

			wantErr: false,
		},
		{
			name: "negative case creating counters",
			fields: fields{
				pool:   mock,
				logger: zap.L().Sugar(),
			},

			wantErr: false,
		},
		{
			name: "positive case with retry #2",
			fields: fields{
				pool:   mock,
				logger: zap.L().Sugar(),
			},

			wantErr: false,
		},
		{
			name: "negative case creating gauges",
			fields: fields{
				pool:   mock,
				logger: zap.L().Sugar(),
			},

			wantErr: false,
		},
		{
			name: "negative case with failed rollback",
			fields: fields{
				pool:   mock,
				logger: zap.L().Sugar(),
			},

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
			err := db.createTables(ctx)
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

	var sqc = "SELECT value FROM counters"

	mock.ExpectBegin()
	mock.ExpectQuery(sqc).WithArgs("keyOne").WillReturnRows(mock.NewRows([]string{"value"}).AddRow(int64(1)))
	mock.ExpectCommit()

	mock.ExpectBegin()
	mock.ExpectQuery(sqc).WithArgs("keyTwo").WillReturnError(pgx.ErrNoRows)
	mock.ExpectCommit()

	mock.ExpectBegin()
	mock.ExpectQuery(sqc).WithArgs("keyTwo").WillReturnError(errors.New("bad querry"))
	mock.ExpectRollback()

	mock.ExpectBegin()
	mock.ExpectQuery(sqc).WithArgs("keyTwo").WillReturnError(errors.New("bad syntax querry"))
	mock.ExpectRollback().WillReturnError(errors.New("fail rollback"))

	type fields struct {
		pool   PgxIface
		logger *zap.SugaredLogger
	}
	type args struct {
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
			got, err := db.GetInt64Value(ctx, tt.args.key)
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

const gaugeOne = "gaugeOne"
const gaugeTwo = "gaugeTwo"

func TestDB_GetFloat64Value(t *testing.T) {
	ctx := context.Background()

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	mock.ExpectBegin().WillReturnError(errors.New("just error"))

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT delta FROM gauges").WithArgs(gaugeOne).
		WillReturnRows(mock.NewRows([]string{"delta"}).AddRow(float64(1.1)))
	mock.ExpectCommit()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT delta FROM gauges").WithArgs(gaugeTwo).
		WillReturnError(pgx.ErrNoRows)
	mock.ExpectCommit()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT delta FROM gauges").WithArgs(gaugeTwo).
		WillReturnError(errors.New("bad querry"))
	mock.ExpectRollback()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT delta FROM gauges").WithArgs(gaugeTwo).
		WillReturnError(errors.New("bad querry"))
	mock.ExpectRollback().WillReturnError(errors.New("fail rollback"))

	type fields struct {
		pool   PgxIface
		logger *zap.SugaredLogger
	}
	type args struct {
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

				key: gaugeOne,
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

				key: gaugeOne,
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

				key: gaugeTwo,
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

				key: gaugeTwo,
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
			got, err := db.GetFloat64Value(ctx, tt.args.key)
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

const counterOne = "counterOne"

func TestDB_GetDataList(t *testing.T) {
	ctx := context.Background()

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	const gq = "SELECT id, delta FROM gauges"

	mock.ExpectBegin()
	mock.ExpectQuery(gq).
		WillReturnRows(mock.NewRows([]string{"id", "delta"}).AddRow(gaugeOne, float64(1.1)).AddRow("gaugeTwo", float64(1.2)))
	mock.ExpectCommit()

	const cq = "SELECT id, value FROM counters"

	mock.ExpectBegin()
	mock.ExpectQuery(cq).
		WillReturnRows(mock.NewRows([]string{"id", "value"}).AddRow(counterOne, int64(1)).AddRow("counterTwo", int64(2)))
	mock.ExpectCommit()

	mock.ExpectBegin().WillReturnError(errors.New("transaction begin error"))

	mock.ExpectBegin()
	mock.ExpectQuery(gq).
		WillReturnRows(mock.NewRows([]string{"id", "delta"}).AddRow(gaugeOne, float64(1.1)).AddRow("gaugeTwo", float64(1.2)))
	mock.ExpectCommit()

	mock.ExpectBegin().WillReturnError(errors.New("begin error"))

	type fields struct {
		pool   PgxIface
		logger *zap.SugaredLogger
	}

	tests := []struct {
		name   string
		fields fields

		want    []string
		wantErr bool
	}{
		{
			name: "positive case",
			fields: fields{
				pool:   mock,
				logger: zap.L().Sugar(),
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

			want:    []string{},
			wantErr: true,
		},
		{
			name: "negative case #2",
			fields: fields{
				pool:   mock,
				logger: zap.L().Sugar(),
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
			got, err := db.GetDataList(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("DB.GetDataList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for wantIdx := range tt.want {
				finded := false
				for gotIdx := range got {
					if !finded {
						finded = tt.want[wantIdx] == got[gotIdx]
					}
				}
				if !finded {
					t.Errorf("DB.GetDataList() value %v not found in %v", tt.want[wantIdx], got)
				}
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

	mock.ExpectBegin().WillReturnError(errors.New("expect begin error"))

	const iq = "INSERT (.+)"

	mock.ExpectBegin()
	mock.ExpectQuery(iq).WithArgs(counterOne, int64(1)).
		WillReturnRows(mock.NewRows([]string{"value"}).AddRow(int64(2)))
	mock.ExpectCommit()

	mock.ExpectBegin()
	mock.ExpectQuery(iq).WithArgs("counterTwo", int64(1)).
		WillReturnError(errors.New("some insert errors"))
	mock.ExpectCommit()

	mock.ExpectBegin()
	mock.ExpectQuery(iq).WithArgs("counterThree", int64(3)).
		WillReturnError(errors.New("bad querry"))
	mock.ExpectRollback().WillReturnError(errors.New("fail rollback"))

	type fields struct {
		pool   PgxIface
		logger *zap.SugaredLogger
	}
	type args struct {
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

				key:   counterOne,
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
			got, err := db.AddInt64Value(ctx, tt.args.key, tt.args.value)
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

	mock.ExpectBegin().WillReturnError(errors.New("set float 64 begin tx error"))

	const iq = "INSERT (.+)"

	mock.ExpectBegin()
	mock.ExpectQuery(iq).WithArgs("gaugeOne", float64(1.1)).
		WillReturnRows(mock.NewRows([]string{"value"}).AddRow(float64(1.1)))
	mock.ExpectCommit()

	mock.ExpectBegin()
	mock.ExpectQuery(iq).WithArgs("gaugeTwo", float64(1.2)).
		WillReturnError(errors.New("some insert errors"))
	mock.ExpectCommit()

	mock.ExpectBegin()
	mock.ExpectQuery(iq).WithArgs("gaugeThree", float64(1.3)).
		WillReturnError(errors.New("bad querry"))
	mock.ExpectRollback().WillReturnError(errors.New("fail rollback"))

	type fields struct {
		pool   PgxIface
		logger *zap.SugaredLogger
	}
	type args struct {
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
			got, err := db.SetFloat64Value(ctx, tt.args.key, tt.args.value)
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

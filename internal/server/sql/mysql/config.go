package mysql

import (
	"fmt"
	"strconv"
	"strings"

	"vitess.io/vitess/go/mysql"
	"vitess.io/vitess/go/sqltypes"
)

type ServerConfig struct {
	Name      string      `yaml:"name"`
	Record    []string    `yaml:"record"`
	Port      int         `yaml:"port"`
	Databases []*Database `yaml:"databases"`
}

func NewServerConfig(m map[string]interface{}) (*ServerConfig, error) {
	var sc ServerConfig

	if x, ok := m["name"]; ok {
		sc.Name = x.(string)
	}

	if x, ok := m["port"]; ok {
		sc.Port = x.(int)
	}

	if x, ok := m["databases"]; ok {
		var (
			ifaces = x.([]interface{})
			dbs    = make([]*Database, len(ifaces))
		)

		for idx := range ifaces {
			var iface = ifaces[idx]
			d, err := newDatabase(iface.(map[interface{}]interface{}))
			if err != nil {
				return nil, err
			}

			dbs[idx] = d
		}

		sc.Databases = dbs
	}

	return &sc, nil
}

func (sc *ServerConfig) Type() string {
	return "mysql"
}

type Database struct {
	Name           string                `yaml:"name"`
	ReturnSequence []*RepeatableResponse `yaml:"returnSequence"`
}

func newDatabase(m map[interface{}]interface{}) (*Database, error) {
	var db Database

	if x, ok := m["name"]; ok {
		db.Name = x.(string)
	}

	if x, ok := m["returnSequence"]; ok {
		var (
			ifaces = x.([]interface{})
			rs     = make([]*RepeatableResponse, len(ifaces))
			err    error
		)

		for idx := range ifaces {
			rs[idx], err = newRepeatableResponse(db.Name, x.(map[interface{}]interface{}))
			if err != nil {
				return nil, err
			}
		}
	}

	return &db, nil
}

type RepeatableResponse struct {
	Repeat   *Repeat  `yaml:"repeat"`
	Response Response `yaml:"-"`
}

func newRepeatableResponse(db string, m map[interface{}]interface{}) (*RepeatableResponse, error) {
	var (
		rr  RepeatableResponse
		err error
	)

	if x, ok := m["repeat"]; ok {
		rr.Repeat, err = NewRepeat(x.(map[interface{}]interface{}))
		if err != nil {
			return nil, err
		}
	}

	if x, ok := m["error"]; ok {
		rr.Response = newError(x.(map[interface{}]interface{}))
		return &rr, nil
	}

	if x, ok := m["result"]; ok {
		rr.Response = newResult(x.(map[interface{}]interface{}))
		return &rr, nil
	}

	if x, ok := m["rows"]; ok {
		rr.Response, err = newRows(db, x.(map[interface{}]interface{}))
		if err != nil {
			return nil, err
		}
	}

	return &rr, nil
}

type Rows struct {
	_response

	Fields []*Field           `yaml:"fields"`
	Rows   [][]sqltypes.Value `yaml:"rows"`
}

func newRows(db string, m map[interface{}]interface{}) (*Rows, error) {
	var r Rows

	if x, ok := m["fields"]; ok {
		var (
			ifaces = x.([]interface{})
			fields = make([]*Field, len(ifaces))
		)

		for idx := range ifaces {
			var mm = ifaces[idx].(map[interface{}]interface{})
			fields[idx] = newField(db, mm)
		}
	}

	if x, ok := m["rows"]; ok {
		var (
			rowIfaces = x.([]interface{})
			rows      = make([][]sqltypes.Value, len(rowIfaces))
		)

		for idx := range rowIfaces {
			var (
				colIfaces = rowIfaces[idx].([]interface{})
				row       = make([]sqltypes.Value, len(colIfaces))
			)

			for jdx := range colIfaces {
				var (
					col   = colIfaces[jdx]
					field = r.Fields[jdx]
					t     = field.Type
				)

				t = strings.ReplaceAll(t, " ", "")
				t = strings.ToLower(t)

				switch t {
				case "null":
					row[jdx] = sqltypes.NULL
				case "tinyint":
					row[jdx] = sqltypes.MakeTrusted(
						sqltypes.Int8,
						strconv.AppendInt(nil, int64(x.(int)), 10),
					)
				case "tinyintunsinged":
					row[jdx] = sqltypes.MakeTrusted(
						sqltypes.Uint8,
						strconv.AppendInt(nil, int64(x.(int)), 10),
					)
				case "smallint":
					row[jdx] = sqltypes.MakeTrusted(
						sqltypes.Int16,
						strconv.AppendInt(nil, int64(x.(int)), 10),
					)
				case "smallintunsigned":
					row[jdx] = sqltypes.MakeTrusted(
						sqltypes.Uint16,
						strconv.AppendInt(nil, int64(x.(int)), 10),
					)
				case "mediumint":
					row[jdx] = sqltypes.MakeTrusted(
						sqltypes.Int24,
						strconv.AppendInt(nil, int64(x.(int)), 10),
					)
				case "mediumintunsigned":
					row[jdx] = sqltypes.MakeTrusted(
						sqltypes.Uint24,
						strconv.AppendInt(nil, int64(x.(int)), 10),
					)
				case "integer":
					row[jdx] = sqltypes.MakeTrusted(
						sqltypes.Int32,
						strconv.AppendInt(nil, int64(x.(int)), 10),
					)
				case "integerunsigned":
					row[jdx] = sqltypes.MakeTrusted(
						sqltypes.Uint32,
						strconv.AppendInt(nil, int64(x.(int)), 10),
					)
				case "bigint":
					row[jdx] = sqltypes.MakeTrusted(
						sqltypes.Int64,
						strconv.AppendInt(nil, int64(x.(int)), 10),
					)
				case "bigintunsigned":
					row[jdx] = sqltypes.MakeTrusted(
						sqltypes.Uint64,
						strconv.AppendInt(nil, int64(x.(int)), 10),
					)
				case "float":
					row[jdx] = sqltypes.MakeTrusted(
						sqltypes.Float32,
						strconv.AppendFloat(nil, float64(x.(int)), 'E', 10, 32),
					)
				case "real":
					fallthrough
				case "double":
					row[jdx] = sqltypes.MakeTrusted(
						sqltypes.Float64,
						strconv.AppendFloat(nil, float64(x.(int)), 'E', 10, 64),
					)
				case "timestamp":
					row[jdx] = sqltypes.MakeTrusted(
						sqltypes.Timestamp,
						[]byte(x.(string)),
					)
				case "date":
					row[jdx] = sqltypes.MakeTrusted(
						sqltypes.Date,
						[]byte(x.(string)),
					)
				case "time":
					row[jdx] = sqltypes.MakeTrusted(
						sqltypes.Time,
						[]byte(x.(string)),
					)
				case "datetime":
					row[jdx] = sqltypes.MakeTrusted(
						sqltypes.Datetime,
						[]byte(x.(string)),
					)
				case "year":
					row[jdx] = sqltypes.MakeTrusted(
						sqltypes.Year,
						strconv.AppendInt(nil, int64(x.(int)), 10),
					)
				case "numeric":
					fallthrough
				case "decimal":
					row[jdx] = sqltypes.MakeTrusted(
						sqltypes.Float64,
						strconv.AppendFloat(nil, float64(x.(int)), 'E', 10, 64),
					)
				case "text":
					row[jdx] = sqltypes.NewVarChar(col.(string))
				case "blob":
					sqltypes.NewVarBinary(col.(string))
				case "varchar":
					row[jdx] = sqltypes.NewVarChar(col.(string))
				case "varbinary":
					sqltypes.NewVarBinary(col.(string))
				case "char":
					row[jdx] = sqltypes.MakeTrusted(
						sqltypes.Char,
						[]byte(x.(string))[0:1],
					)
				case "binary":
					row[jdx] = sqltypes.MakeTrusted(
						sqltypes.Binary,
						[]byte(x.(string)),
					)
				case "bit":
					row[jdx] = sqltypes.MakeTrusted(
						sqltypes.Bit,
						strconv.AppendInt(nil, int64(x.(int)), 10),
					)
				case "enum":
					row[jdx] = sqltypes.MakeTrusted(
						sqltypes.Binary,
						[]byte(x.(string)),
					)
				default:
					return nil, fmt.Errorf("unsupported type: %+v", field.Type)
				}
			}

			rows[idx] = row
		}

		r.Rows = rows
	}

	return &r, nil
}

type Result struct {
	_response

	RowsAffected int `yaml:"rowsAffected"`
	InsertID     int `yaml:"insertID"`
}

func newResult(m map[interface{}]interface{}) *Result {
	var r Result

	if x, ok := m["rowsAffected"]; ok {
		r.RowsAffected = x.(int)
	}

	if x, ok := m["type"]; ok {
		r.InsertID = x.(int)
	}

	return &r
}

type Field struct {
	_response

	Name     string `yaml:"name"`
	Type     string `yaml:"type"`
	Table    string `yaml:"table"`
	Database string `yaml:"-"`
}

func newField(db string, m map[interface{}]interface{}) *Field {
	var f = Field{
		Database: db,
	}

	if x, ok := m["name"]; ok {
		f.Name = x.(string)
	}

	if x, ok := m["type"]; ok {
		f.Type = x.(string)
	}

	if x, ok := m["table"]; ok {
		f.Table = x.(string)
	}

	return &f
}

type Error struct {
	_response

	SQLError *mysql.SQLError `yaml:"sqlError"`
}

func newError(m map[interface{}]interface{}) *Error {
	var e Error

	if x, ok := m["sqlError"]; ok {
		var (
			mm = x.(map[interface{}]interface{})
			se mysql.SQLError
		)

		if x, ok := mm["num"]; ok {
			se.Num = x.(int)
		}

		if x, ok := mm["state"]; ok {
			se.State = x.(string)
		}

		if x, ok := mm["message"]; ok {
			se.Message = x.(string)
		}

		e.SQLError = &se
	}

	return &e
}

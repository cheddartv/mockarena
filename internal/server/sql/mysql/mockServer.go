package mysql

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"vitess.io/vitess/go/mysql"
	"vitess.io/vitess/go/sqltypes"
	querypb "vitess.io/vitess/go/vt/proto/query"
)

type MockServer struct {
	stats mysqlServerStats

	conf      ServerConfig
	startedAt time.Time

	listener          *mysql.Listener
	responseSequences map[string]*responseSequence
	responseSequence  *responseSequence

	sync.Mutex
	sync.WaitGroup
}

func NewMockServer(c ServerConfig) (*MockServer, error) {
	var (
		s MockServer
	)

	s.conf = c
	s.stats.Name = c.Name
	s.responseSequences = make(map[string]*responseSequence)

	for _, db := range c.Databases {
		var rs = newResponseSequence(db.ReturnSequence)

		s.responseSequences[db.Name] = rs

		s.Add(1)
		go func(rs *responseSequence) {
			<-rs.doneChan
			s.Lock()
			defer s.Unlock()
			s.Done()
		}(rs)
	}

	l, err := mysql.NewListener("tcp", "0:8032", mysql.NewAuthServerNone(), &s, time.Second, time.Second, false)
	if err != nil {
		return nil, err
	}

	s.listener = l

	return &s, nil
}

func (ms *MockServer) Start() {
	ms.startedAt = time.Now()
	ms.listener.Accept()
}

func (ms *MockServer) ServerStats() interface{} {
	return nil
}

/// mysql.Handler iface

func (ms *MockServer) NewConnection(c *mysql.Conn) {}

func (ms *MockServer) ConnectionClosed(c *mysql.Conn) {}

func (ms *MockServer) ComQuery(c *mysql.Conn, query string, callback func(*sqltypes.Result) error) error {
	ms.Lock()
	defer ms.Unlock()

	{
		var db string
		n, err := fmt.Sscanf(strings.ToLower(query), "use %s", &db)
		if err != nil {
			return mysql.NewSQLErrorFromError(err)
		}
		if n == 1 {
			db = strings.ReplaceAll(db, "`", "")
			rs, ok := ms.responseSequences[db]
			if !ok {
				return &mysql.SQLError{
					Num:     1049,
					State:   "42000",
					Message: fmt.Sprintf("Unknown database '%s'", db),
				}
			}

			ms.responseSequence = rs
		}
	}

	if ms.responseSequence == nil {
		return &mysql.SQLError{
			Num:     1046,
			State:   "3D000",
			Message: "No database selected",
		}
	}

	var response = ms.responseSequence.next()
	if response == nil {
		c.Close()
		return nil
	}

	switch r := response.(type) {
	case *Error:
	case *Rows:
		var fields = make([]*querypb.Field, len(r.Fields))
		for idx := range r.Fields {
			var (
				rf = r.Fields[idx]
				t  = rf.Type
				f  = querypb.Field{
					Name:     rf.Name,
					Table:    rf.Table,
					Database: rf.Database,
				}
			)

			t = strings.ReplaceAll(t, " ", "")
			t = strings.ToLower(t)

			switch t {
			case "null":
				f.Type = sqltypes.Null
			case "tinyint":
				f.Type = sqltypes.Int8
			case "tinyintunsinged":
				f.Type = sqltypes.Uint8
			case "smallint":
				f.Type = sqltypes.Int16
			case "smallintunsigned":
				f.Type = sqltypes.Uint16
			case "mediumint":
				f.Type = sqltypes.Int24
			case "mediumintunsigned":
				f.Type = sqltypes.Uint24
			case "integer":
				f.Type = sqltypes.Int32
			case "integerunsigned":
				f.Type = sqltypes.Uint32
			case "bigint":
				f.Type = sqltypes.Int64
			case "bigintunsigned":
				f.Type = sqltypes.Uint64
			case "float":
				f.Type = sqltypes.Float32
			case "real":
				f.Type = sqltypes.Float64
			case "double":
				f.Type = sqltypes.Float64
			case "timestamp":
				f.Type = sqltypes.Timestamp
			case "date":
				f.Type = sqltypes.Date
			case "time":
				f.Type = sqltypes.Time
			case "datetime":
				f.Type = sqltypes.Datetime
			case "year":
				f.Type = sqltypes.Year
			case "numeric":
				f.Type = sqltypes.Decimal
			case "decimal":
				f.Type = sqltypes.Decimal
			case "text":
				f.Type = sqltypes.Text
			case "blob":
				f.Type = sqltypes.Blob
			case "varchar":
				f.Type = sqltypes.VarChar
			case "varbinary":
				f.Type = sqltypes.VarBinary
			case "char":
				f.Type = sqltypes.Char
			case "binary":
				f.Type = sqltypes.Binary
			case "bit":
				f.Type = sqltypes.Bit
			case "enum":
				f.Type = sqltypes.Enum
			default:
			}
		}

		callback(&sqltypes.Result{
			Fields: fields,
			Rows: [][]sqltypes.Value{
				{
					sqltypes.NewVarChar("hello"),
				},
			},
		})
	case *Result:
	}

	return nil
}

func (ms *MockServer) ComPrepare(c *mysql.Conn, query string, bindVars map[string]*querypb.BindVariable) ([]*querypb.Field, error) {
	fmt.Printf("(*MockServer)ComPrepare was called: %+v\n", *c.PrepareData[1])
	return nil, nil
}

func (ms *MockServer) ComStmtExecute(c *mysql.Conn, prepare *mysql.PrepareData, callback func(*sqltypes.Result) error) error {
	fmt.Println("(*MockServer)ComStmtExecute was called")
	callback(&sqltypes.Result{
		Fields: []*querypb.Field{
			{
				Name:         "my_column",
				Type:         sqltypes.VarChar,
				Table:        "my_table",
				Database:     "my_database",
				ColumnLength: 10,
			},
		},
		Rows: [][]sqltypes.Value{
			{
				sqltypes.NewVarChar("hello"),
			},
		},
	})
	return nil
}

func (ms *MockServer) WarningCount(c *mysql.Conn) uint16 {
	fmt.Println("(*MockServer)WarningCount was called")
	return 0
}

func (ms *MockServer) ComResetConnection(c *mysql.Conn) {
	fmt.Println("(*MockServer)ComResetConnection was called")
}

type mysqlServerStats struct {
	Name string
}

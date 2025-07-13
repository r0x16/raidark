package connection

import "testing"

func TestDSNBuilders(t *testing.T) {
	pg := GormPostgresConnection{Host: "h", Port: "p", Username: "u", Password: "pw", Database: "d"}
	if pg.GetDsn() != "host='h' user='u' password='pw' dbname='d' port='p' sslmode='disable'" {
		t.Fatal("bad pg dsn")
	}
	my := GormMysqlConnection{Host: "h", Port: "p", Username: "u", Password: "pw", Database: "d"}
	if my.GetDsn() != "u:pw@tcp(h:p)/d?charset=utf8mb4&parseTime=True&loc=Local" {
		t.Fatal("bad mysql dsn")
	}
	sq := GormSqliteConnection{DatabasePath: "file.db"}
	if sq.GetDsn() != "file.db" {
		t.Fatal("bad sqlite dsn")
	}
}

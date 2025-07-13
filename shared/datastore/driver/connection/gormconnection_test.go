package connection

import "testing"

func TestMysqlDsn(t *testing.T) {
	c := GormMysqlConnection{Host: "h", Port: "1", Username: "u", Password: "p", Database: "d"}
	dsn := c.GetDsn()
	if dsn == "" || dsn != "u:p@tcp(h:1)/d?charset=utf8mb4&parseTime=True&loc=Local" {
		t.Fatalf("unexpected dsn: %s", dsn)
	}
}

func TestPostgresDsn(t *testing.T) {
	c := GormPostgresConnection{Host: "h", Port: "2", Username: "u", Password: "p", Database: "d"}
	dsn := c.GetDsn()
	if dsn == "" {
		t.Fatal("empty dsn")
	}
}

func TestSqliteDsn(t *testing.T) {
	c := GormSqliteConnection{DatabasePath: "db.sqlite"}
	if c.GetDsn() != "db.sqlite" {
		t.Fatalf("unexpected dsn")
	}
}

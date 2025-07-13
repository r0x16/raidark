package connection

/*
 * All the data needed to connect to a sqlite database
 */
type GormSqliteConnection struct {
	DatabasePath string
}

/*
 * Creates a new dsn string for the sqlite driver
 * using the connection struct data
 */
func (g *GormSqliteConnection) GetDsn() string {
	return g.DatabasePath
}

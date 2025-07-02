package connection

/*
 * All the data needed to connect to a mysql database
 */
type GormPostgresConnection struct {
	Host     string
	Port     string
	Username string
	Password string
	Database string
}

/*
 * Creates a new dsn string for the mysql driver
 * using the connection struct data
 */
func (g *GormPostgresConnection) GetDsn() string {
	return "host='" + g.Host + "' user='" + g.Username + "' password='" + g.Password + "' dbname='" + g.Database + "' port='" + g.Port + "' sslmode='disable'"
}

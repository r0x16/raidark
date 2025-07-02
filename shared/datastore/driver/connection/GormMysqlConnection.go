package connection

/*
 * All the data needed to connect to a mysql database
 */
type GormMysqlConnection struct {
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
func (g *GormMysqlConnection) GetDsn() string {
	return g.Username + ":" + g.Password + "@tcp(" + g.Host + ":" + g.Port + ")/" + g.Database + "?charset=utf8mb4&parseTime=True&loc=Local"
}

package database

type Driver string

const (
	DriverMongoDB   Driver = "mongodb"
	DriverPostgres Driver = "postgres"
	DriverMySQL    Driver = "mysql"
	DriverSQLite   Driver = "sqlite"
)

func IsSupportedDriver(driver string) bool {
	switch Driver(driver) {
	case DriverMongoDB, DriverPostgres, DriverMySQL, DriverSQLite:
		return true
	default:
		return false
	}
}

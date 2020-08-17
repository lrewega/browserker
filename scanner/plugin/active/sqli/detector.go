package sqli

import (
	"regexp"

	"gitlab.com/browserker/browserk"
)

// Detector for sql injection via error strings
type Detector struct {
	errors map[browserk.TechType]*regexp.Regexp
}

// NewDetector for detecting sql error strings
// TODO: add more
func NewDetector() *Detector {
	d := &Detector{errors: make(map[browserk.TechType]*regexp.Regexp, 0)}
	oracleRe := regexp.MustCompile("oracle\\.jdbc|ORA-01756|quoted string not properly terminated|ORA-00923|FROM keyword not found where expected|ORA-01722|invalid number|ORA-01741|illegal zero-length identifier")
	d.errors[browserk.DBOracle] = oracleRe

	mysqlRe := regexp.MustCompile("You have an error in your SQL syntax")
	d.errors[browserk.DBMySQL] = mysqlRe

	pgsqlRe := regexp.MustCompile("org.postgresql.util.PSQLException|Unterminated string literal started at position|Unterminated identifier started at position")
	d.errors[browserk.DBPostgres] = pgsqlRe

	mssqlRe := regexp.MustCompile("com.microsoft.sqlserver.jdbc|com.microsoft.jdbc|com.microsoft.sqlserver.jdbc|weblogic.jdbc.mssqlserver")
	d.errors[browserk.DBMSSQL] = mssqlRe

	sqliteRe := regexp.MustCompile("near \".+\": syntax error")
	d.errors[browserk.DBSQLite] = sqliteRe
	return d
}

// Detect error strings against the input, returns Unknown if nothing detected
func (d *Detector) Detect(input []byte) (browserk.TechType, string) {
	for tech, regex := range d.errors {
		if regex.Match(input) {
			return tech, regex.String()
		}
	}
	return browserk.Unknown, ""
}

package browserk

type TechType int

const (
	Unknown TechType = iota
	LangJava
	LangDotNet
	LangPHP
	LangGo
	LangPython
	LangRuby
	LangJS
	LangC
	LangCPP
	LangRust

	DBMySQL
	DBMSSQL
	DBPostgres
	DBOracle
	DBSQLite
	DBMongo
	DBRedis
	DBElastic

	ServerApache
	ServerNginx
	ServerIIS
	ServerGo
	ServerGunicorn

	AppTomcat
	AppJetty
	AppJBoss
	AppGlassfish
	AppWebLogic
	AppDjango
	AppFlask
)

type Tech struct {
	TechTypes map[string][]TechType
}

func NewTech() *Tech {
	return &Tech{TechTypes: make(map[string][]TechType, 0)}
}

func (t TechType) String() string {
	switch t {
	case LangJava:
		return "Java"
	case LangDotNet:
		return "DotNet"
	case LangPHP:
		return "PHP"
	case LangGo:
		return "Go"
	case LangPython:
		return "Python"
	case LangRuby:
		return "Ruby"
	case LangJS:
		return "JavaScript"
	case LangC:
		return "C"
	case LangCPP:
		return "C++"
	case LangRust:
		return "Rust"
	case DBMySQL:
		return "MySQL"
	case DBMSSQL:
		return "MS-SQL"
	case DBPostgres:
		return "PostgreSQL"
	case DBOracle:
		return "Oracle"
	case DBSQLite:
		return "SQLite"
	case DBMongo:
		return "Mongo"
	case DBRedis:
		return "Redis"
	case DBElastic:
		return "ElasticSearch"
	case ServerApache:
		return "Apache"
	case ServerNginx:
		return "Nginx"
	case ServerIIS:
		return "IIS"
	case ServerGo:
		return "Go"
	case ServerGunicorn:
		return "Gunicorn"
	case AppTomcat:
		return "Tomcat"
	case AppJetty:
		return "Jetty"
	case AppJBoss:
		return "JBoss"
	case AppGlassfish:
		return "GlassFish"
	case AppWebLogic:
		return "WebLogic"
	case AppDjango:
		return "Django"
	case AppFlask:
		return "Flask"
	}
	return "Unknown"
}

// PathHas determines if a particular path has a tech type defined, or if
// the all techs has it
func (t *Tech) PathHas(path string, tech TechType) bool {
	techs, hasPath := t.TechTypes[path]
	if !hasPath {
		var hasInAll bool
		techs, hasInAll = t.TechTypes["all"]
		if !hasInAll {
			return false
		}
	}

	for _, t := range techs {
		if t == tech {
			return true
		}
	}
	return false
}

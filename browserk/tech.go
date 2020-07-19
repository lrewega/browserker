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

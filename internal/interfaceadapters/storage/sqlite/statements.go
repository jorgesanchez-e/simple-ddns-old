package sqlite

const (
	createTable string = `CREATE TABLE IF NOT EXISTS domains (
		fqdn TEXT NOT NULL, 
		update_time TEXT NOT NULL,
		register_type TEXT NOT NULL,
		ip TEXT NOT NULL,
		active BOOL NOT NULL
	)`

	lastRegister string = `SELECT ip FROM domains WHERE 
		fqdn = ? AND
		register_type = ? AND 
		active = true
	`

	insertRegister string = `INSERT INTO domains 
		(fqdn, update_time, register_type, ip, active)
		VALUES(?,?,?,?,?)	
	`

	updateRegister string = `UPDATE domains SET active = false
		WHERE fqdn = ? AND register_type = ? AND active = true
	`
)

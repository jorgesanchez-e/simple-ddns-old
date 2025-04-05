package sqlite

const (
	createTable string = `CREATE TABLE IF NOT EXISTS ddns_domains (
		fqdn TEXT NOT NULL, 
		update_time TEXT NOT NULL,
		register_type TEXT NOT NULL,
		ip TEXT NOT NULL,
		active BOOL NOT NULL
	)`

	lastRegister string = `SELECT ip, register_type FROM ddns_domains WHERE 
		fqdn = ? AND
		active = true
	`

	insertRegister string = `INSERT INTO ddns_domains 
		(fqdn, update_time, register_type, ip, active)
		VALUES(?,?,?,?,?)	
	`

	updateRegister string = `UPDATE ddns_domains SET active = false
		WHERE fqdn = ? AND register_type = ? AND active = true
	`
)

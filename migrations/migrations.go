// Package migrations embebe los archivos SQL de las migraciones del esquema
// de RodriTV (SPECS-004.6). La API los aplica al arrancar via golang-migrate
// con el driver iofs (internal/platform.MigrateUp) y sqlc los usa como
// fuente de esquema.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS

// Package migrations embute as migrations goose no binário (`ominicore migrate`).
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS

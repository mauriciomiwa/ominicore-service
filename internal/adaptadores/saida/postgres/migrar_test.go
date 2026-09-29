package postgres

import (
	"context"
	"testing"
)

func TestSchemaInvalidoERecusadoAntesDoSQL(t *testing.T) {
	for _, s := range []string{"", "Ominicore", "ominicore; DROP TABLE x", "om-hml", "1abc"} {
		if _, err := Conectar(context.Background(), "postgres://x@127.0.0.1:1/x", s); err == nil {
			t.Errorf("schema %q aceito", s)
		}
	}
	for _, s := range []string{"ominicore", "ominicore_hml"} {
		if err := validarSchema(s); err != nil {
			t.Errorf("schema %q recusado: %v", s, err)
		}
	}
}

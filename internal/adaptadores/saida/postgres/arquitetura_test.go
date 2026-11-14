package postgres

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"testing"
)

// consultasDaAPI são as queries de `notificacoes` feitas em nome de um Sistema
// de origem: só valem dentro de naOrganizacao, com a Organização fixada para a
// política por_organizacao (REQ-52, design D3). O worker, a retenção e os
// Avisos de status usam outras queries e veem todas as Organizações.
var consultasDaAPI = map[string]bool{
	"InserirNotificacao":   true,
	"BuscarPorChave":       true,
	"ConsultarNotificacao": true,
	"ReivindicarChave":     true,
	"ListarTentativas":     true,
}

func TestArquitetura_LeiturasDaAPISoComOrganizacaoFixada(t *testing.T) {
	arquivos, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	vistas := map[string]bool{}
	for _, nome := range arquivos {
		if strings.HasSuffix(nome, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, nome, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		permitidas := dentroDeNaOrganizacao(f)
		for _, c := range chamadasDaAPI(f) {
			nomeQuery := c.Fun.(*ast.SelectorExpr).Sel.Name
			vistas[nomeQuery] = true
			if !cobre(permitidas, c) {
				t.Errorf("%s: %s fora de naOrganizacao", fset.Position(c.Pos()), nomeQuery)
			}
		}
	}
	for q := range consultasDaAPI {
		if !vistas[q] {
			t.Errorf("%s não é chamada no adaptador: atualize consultasDaAPI", q)
		}
	}
}

// dentroDeNaOrganizacao devolve os trechos das funções passadas a naOrganizacao.
func dentroDeNaOrganizacao(f *ast.File) [][2]token.Pos {
	var trechos [][2]token.Pos
	ast.Inspect(f, func(n ast.Node) bool {
		c, ok := n.(*ast.CallExpr)
		if id, ehIdent := funcDe(c).(*ast.Ident); !ok || !ehIdent || id.Name != "naOrganizacao" {
			return true
		}
		for _, a := range c.Args {
			if fl, ok := a.(*ast.FuncLit); ok {
				trechos = append(trechos, [2]token.Pos{fl.Pos(), fl.End()})
			}
		}
		return true
	})
	return trechos
}

// chamadasDaAPI devolve as chamadas `q.<query>` de consultasDaAPI.
func chamadasDaAPI(f *ast.File) []*ast.CallExpr {
	var chamadas []*ast.CallExpr
	ast.Inspect(f, func(n ast.Node) bool {
		c, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if sel, ok := c.Fun.(*ast.SelectorExpr); ok && consultasDaAPI[sel.Sel.Name] {
			if _, ehQueries := sel.X.(*ast.Ident); ehQueries {
				chamadas = append(chamadas, c)
			}
		}
		return true
	})
	return chamadas
}

func funcDe(c *ast.CallExpr) ast.Expr {
	if c == nil {
		return nil
	}
	return c.Fun
}

func cobre(trechos [][2]token.Pos, c *ast.CallExpr) bool {
	for _, p := range trechos {
		if c.Pos() >= p[0] && c.End() <= p[1] {
			return true
		}
	}
	return false
}

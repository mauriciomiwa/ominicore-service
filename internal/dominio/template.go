package dominio

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
)

// ConteudoCanal é a versão do Template para um Canal. No WhatsApp ela tem uma
// ou as duas formas (REQ-11): para a Meta, o Modelo aprovado (nome + idioma) e
// Parametros, que lista, na ordem de {{1}}, {{2}}…, as variáveis do Template
// (depois de Preencher, os valores); para o UltraMsg, o texto em Corpo.
type ConteudoCanal struct {
	Assunto        string
	Corpo          string
	ModeloAprovado string
	Idioma         string
	Parametros     []string
}

// Template é o conteúdo cadastrado, com variáveis no formato {{nome}}, uma
// versão por Canal e a Ordem de Canais padrão.
type Template struct {
	ID          string
	OrdemCanais []Canal
	Conteudos   map[Canal]ConteudoCanal
}

var reVariavel = regexp.MustCompile(`\{\{\s*([a-zA-Z0-9_]+)\s*\}\}`)

// Variaveis lista, ordenadas e sem repetição, as variáveis usadas em qualquer
// versão do Template.
func (t Template) Variaveis() []string {
	var nomes []string
	for _, c := range t.Conteudos {
		for _, m := range reVariavel.FindAllStringSubmatch(c.Assunto+" "+c.Corpo, -1) {
			nomes = append(nomes, m[1])
		}
		nomes = append(nomes, c.Parametros...)
	}
	slices.Sort(nomes)
	return slices.Compact(nomes)
}

// ValidarDados confere se todas as variáveis do Template vieram nos dados.
func (t Template) ValidarDados(dados map[string]string) error {
	var faltando []string
	for _, v := range t.Variaveis() {
		if _, ok := dados[v]; !ok {
			faltando = append(faltando, v)
		}
	}
	if len(faltando) > 0 {
		return fmt.Errorf("%w: %s", ErrVariavelFaltando, strings.Join(faltando, ", "))
	}
	return nil
}

// Preencher devolve a versão do Canal com as variáveis substituídas.
func (t Template) Preencher(c Canal, dados map[string]string) (ConteudoCanal, error) {
	conteudo, ok := t.Conteudos[c]
	if !ok {
		return ConteudoCanal{}, fmt.Errorf("template %s sem versão para o canal %s", t.ID, c)
	}
	substituir := func(s string) string {
		return reVariavel.ReplaceAllStringFunc(s, func(m string) string {
			return dados[reVariavel.FindStringSubmatch(m)[1]]
		})
	}
	valores := make([]string, len(conteudo.Parametros))
	for i, v := range conteudo.Parametros {
		valores[i] = dados[v]
	}
	return ConteudoCanal{
		Assunto: substituir(conteudo.Assunto), Corpo: substituir(conteudo.Corpo),
		ModeloAprovado: conteudo.ModeloAprovado, Idioma: conteudo.Idioma, Parametros: valores,
	}, nil
}

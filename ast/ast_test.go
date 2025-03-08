package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ChaosNyaruko/monkey/token"
)

func TestString(t *testing.T) {
	// let foo = bar;
	p := &Program{
		Statements: []Statement{
			&LetStatement{
				Token: token.Token{
					Type:    token.LET,
					Literal: "let",
				},
				Name: &Identifier{
					Token: token.Token{
						Type:    token.IDENT,
						Literal: "foo",
					},
					Value: "foo",
				},
				Value: &Identifier{
					Token: token.Token{
						Type:    token.IDENT,
						Literal: "bar",
					},
					Value: "bar",
				},
			},
		},
	}

	t.Logf("program: %v", p.String())
	assert.Equal(t, "let foo = bar;", p.String())
}

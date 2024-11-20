package main

import (
	"go/ast"
	"go/token"
)

type UpdateKind int

const (
	UpdateUnknown UpdateKind = iota
	UpdateGrow
	UpdateShrink
)

// Target points to a growable resource, such as a struct field or a global variable.
type Target struct {
	// Identity is the identifier for the target.
	Identity *ast.Ident

	// Field is the field under the target which is growable. Is nil if target does not have fields (ie is a variable).
	Field *ast.Field
}

type Update struct {
	Target

	Kind UpdateKind
	Pos  token.Pos
}

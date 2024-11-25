package main

import (
	"go/ast"
	"go/token"
)

// Target points to a growable resource, such as a struct field or a global variable.
type Target struct {
	// Identity is the identifier for the target.
	Identity *ast.Ident

	// Field is the field under the target which is growable. Is nil if target does not have fields (ie is a variable).
	Field *ast.Field
}

// Equals checks if the two targets point to the same resource or resoruce field. Since Targets do not specify a package, this method is not safe to use when comparing targets from different packages.
func (t Target) Equals(ot Target) bool {
	if t.Identity.Name != ot.Identity.Name {
		return false
	}

	if t.Field == nil && ot.Field == nil {
		return true
	}

	if t.Field == nil || ot.Field == nil {
		return false
	}

	return t.Field.Names[0].Name == ot.Field.Names[0].Name
}

func (t Target) String() string {
	if t.Field == nil {
		return t.Identity.Name
	}

	return t.Identity.Name + "." + t.Field.Names[0].Name
}

func (t Target) Position(fset *token.FileSet) token.Position {
	if t.Field == nil {
		return fset.Position(t.Identity.Pos())
	}

	return fset.Position(t.Field.Pos())
}

type UpdateKind int

const (
	UpdateUnknown UpdateKind = iota
	UpdateGrow
	UpdateShrink
)

type Update struct {
	Target

	Kind UpdateKind

	Pos token.Pos
}

type Result struct {
	Target Target

	Pos token.Position
}

type Context struct {
	Parent *Context

	Targets map[string]Target

	Recv *ast.Field
}

func (c *Context) AddTarget(t Target) *Context {
	c.Targets[t.String()] = t

	return c
}

package ast

import "fmt"

type AuthorizationSchema struct {
	TypeDefinitions []*TypeDefinition `json:"typeDefinitions"`
}

type TypeDefinition struct {

	// The name of the type, typically a unique domain/entity name within the system.
	Name string `json:"name"`

	// Direct relationships (e.g. RelationshipTuples which are writable)
	Relations map[string]*Relation

	Permissions map[string]*Permission
}

type Relation struct {

	// The name of the relation
	Name string `json:"name"`

	// The set of subject types which constrain to whom this relation can be associated with
	SubjectTypeRestrictions []*SubjectTypeRestriction `json:"typeRestrictions"`
}

type SubjectTypeRestriction struct {

	// The subject type
	Type string `json:"type"`

	// An optional relation (if the subject restriction refers to a SubjectSet)
	Relation string `json:"relation"`
}

type Permission struct {

	// The name of the permission
	Name string `json:"name"`

	Expression *RelationshipRewrite `json:"expression"`
}

type RelationshipRewrite struct {

	// unary rewrite (a --> b) e.g. computed_userset
	// binary rewrite (a#parent --> b#relation) e.g. tuple_to_userset
	// set rewrites:
	//  1) union ([N] children expressions)
	//  2) intersection ([N] children expressions)
	//  3) exclusion ([2] children expressions, but each child can be any compound)
	Rewrite RewriteOperation
}

type RewriteOperation interface {
}

/*
typedef user

typedef group
  relation member: user | group#member
*/

type UnaryRule struct {
	PrerequisiteRelation string
	DerivedRelation      string
}

// prereq1(x, y) and prereq2(y, z) --> derived(x, z)
// folder-viewer(user:jon, folder:x) and document-parent(folder:x, document:1) --> document-viewer(user:jon, document:1)
//
// or
//
// prereq1(x, y) and prereq2(x, y) --> derived(x, y)
// document-viewer(user:jon, document:1) and document-allowed(user:jon, document:1) --> document-can-view(user:jon, document:1)
type BinaryRule struct {
	PrerequisiteRelation1 string
	PrerequisiteRelation2 string
	DerivedRelation       string
}

type GeneratedRules struct {
	UnaryRules         []UnaryRule
	BinaryRules        []BinaryRule
	NegatedBinaryRules []BinaryRule
}

func SchemaToRules(schema *AuthorizationSchema) (GeneratedRules, error) {
	return GeneratedRules{}, fmt.Errorf("not implemented")
}

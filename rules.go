package main

import (
	"fmt"
)

type SchemaQueryRules struct {

	// The type restrictions which apply to relationship tuples.
	RelationTypeRestrictions []RelationTypeRestriction `json:"relation_type_restrictions"`
	UnaryRules               []UnaryRule               `json:"unary_rules"`
	BinaryRules              []BinaryRule              `json:"binary_rules"`
}

func (s SchemaQueryRules) ToSQL() string {

	var sql string

	if len(s.RelationTypeRestrictions) > 0 {
		sql += "INSERT INTO type_restrictions VALUES\n"
		for i, typeRestriction := range s.RelationTypeRestrictions {
			if i == len(s.RelationTypeRestrictions)-1 {
				sql += fmt.Sprintf("('%s', '%s', '%s', '%s');", typeRestriction.ResourceType, typeRestriction.Relation, typeRestriction.SubjectType, typeRestriction.SubjectRelation)
			} else {
				sql += fmt.Sprintf("('%s', '%s', '%s', '%s'),\n", typeRestriction.ResourceType, typeRestriction.Relation, typeRestriction.SubjectType, typeRestriction.SubjectRelation)
			}
		}
	}

	if len(s.UnaryRules) > 0 {
		sql += "\n\n"

		sql += "INSERT INTO unary_rules VALUES\n"
		for i, unaryRule := range s.UnaryRules {
			if i == len(s.UnaryRules)-1 {
				sql += fmt.Sprintf("('%s', '%s', '%s');", unaryRule.ResourceType, unaryRule.SourceRelation, unaryRule.DerivedRelation)
			} else {
				sql += fmt.Sprintf("('%s', '%s', '%s'),\n", unaryRule.ResourceType, unaryRule.SourceRelation, unaryRule.DerivedRelation)
			}
		}
	}

	if len(s.BinaryRules) > 0 {
		sql += "\n\n"

		sql += "INSERT INTO binary_rules VALUES\n"
		for i, binaryRule := range s.BinaryRules {
			if i == len(s.BinaryRules)-1 {
				sql += fmt.Sprintf("('%s', '%s', '%s', '%s', '%s');", binaryRule.FirstResourceType, binaryRule.FirstRelation, binaryRule.SecondResourceType, binaryRule.SecondRelation, binaryRule.DerivedRelation)
			} else {
				sql += fmt.Sprintf("('%s', '%s', '%s', '%s', '%s'),\n", binaryRule.FirstResourceType, binaryRule.FirstRelation, binaryRule.SecondResourceType, binaryRule.SecondRelation, binaryRule.DerivedRelation)
			}
		}
	}

	return sql
}

type RelationTypeRestriction struct {
	ResourceType    string `json:"resource_type"`
	Relation        string `json:"relation"`
	SubjectType     string `json:"subject_type"`
	SubjectRelation string `json:"subject_relation"`
}

func (r RelationTypeRestriction) String() string {
	if r.SubjectRelation != "" {
		return fmt.Sprintf("%s(%s#%s, %s)", r.Relation, r.SubjectType, r.SubjectRelation, r.ResourceType)
	}

	return fmt.Sprintf("%s(%s, %s)", r.Relation, r.SubjectType, r.ResourceType)
}

type UnaryRule struct {
	ResourceType    string `json:"resource_type"`
	SourceRelation  string `json:"source_relation"`
	DerivedRelation string `json:"derived_relation"`
}

/*
(1, hierarchy) - can_view(subject, folder), parent(folder, document) :- can_view(subject, document)
(2, intersection) - viewer(subject, document), allowed(subject, document) :- can_view(subject, object)
*/
type BinaryRule struct {
	FirstResourceType  string `json:"first_resource_type"`
	FirstRelation      string `json:"first_relation"`
	SecondResourceType string `json:"second_resource_type"`
	SecondRelation     string `json:"second_relation"`
	DerivedRelation    string `json:"derived_relation"`
}

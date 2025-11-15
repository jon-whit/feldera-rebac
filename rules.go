package main

import (
	"fmt"

	authorizerpb "github.com/jon-whit/feldera-rebac/protos/gen/go/authorizer/v1alpha1"
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

func expandPermissionExpressionRefV2(
	schema *authorizerpb.Schema,
	typedef *authorizerpb.TypeDefinition,
	permissionName string,
	exp *authorizerpb.PermissionExpressionRef,
) (string, []UnaryRule, []BinaryRule) {
	var unaryRules []UnaryRule
	var binaryRules []BinaryRule
	var compositeKey string

	resourceType := typedef.GetName()

	switch permissionExp := exp.GetExpression().(type) {
	case *authorizerpb.PermissionExpressionRef_UnaryExpression:
		rule := UnaryRule{
			ResourceType:    resourceType,
			SourceRelation:  permissionExp.UnaryExpression.GetSourceRelation(),
			DerivedRelation: permissionName,
		}
		unaryRules = append(unaryRules, rule)

		compositeKey = permissionName
	case *authorizerpb.PermissionExpressionRef_HierarchicalExpression:
		// get the type restrictions for base relation (e.g. parent)
		baseRelationName := permissionExp.HierarchicalExpression.GetBase()
		baseRelation, ok := typedef.GetRelations()[baseRelationName]
		if !ok {
			panic(fmt.Sprintf("undefined relation '%s'", baseRelationName))
		}

		for _, typeRestriction := range baseRelation.GetTypeRestrictions() {
			targetRelation := permissionExp.HierarchicalExpression.GetTarget()

			// todo: get permissionexpref from target relation
			targetPermissionExpRef := schema.GetTypeDefinitions()[typeRestriction.GetResourceType()].GetPermissions()[targetRelation].GetExpression()

			targetCompositeKey, _, _ := expandPermissionExpressionRefV2(schema, typedef, targetRelation, targetPermissionExpRef)

			binaryRules = append(binaryRules, BinaryRule{
				FirstResourceType:  typeRestriction.GetResourceType(),
				FirstRelation:      targetCompositeKey,
				SecondResourceType: resourceType,
				SecondRelation:     baseRelationName,
				DerivedRelation:    permissionName,
			})
		}

		compositeKey = permissionName
	case *authorizerpb.PermissionExpressionRef_SetExpression:
		unary, binary := expandSetExpression(typedef, permissionName, permissionExp.SetExpression)
		unaryRules = append(unaryRules, unary...)
		binaryRules = append(binaryRules, binary...)
	default:
		panic("unexpected PermissionExpressionRef type")
	}

	return compositeKey, unaryRules, binaryRules
}

func expandSetExpressionV2(
	schema *authorizerpb.Schema,
	typedef *authorizerpb.TypeDefinition,
	permissionName string,
	setExp *authorizerpb.PermissionSetExpressionRef,
) ([]UnaryRule, []BinaryRule) {
	var unaryRules []UnaryRule
	var binaryRules []BinaryRule

	switch exp := setExp.SetExpression.(type) {
	case *authorizerpb.PermissionSetExpressionRef_Union_:
		for _, operand := range exp.Union.Operands {
			compositeKey, operandUnaryRules, operandBinaryRules := expandPermissionExpressionRefV2(schema, typedef, permissionName, operand)
			unaryRules = append(unaryRules, operandUnaryRules...)
			binaryRules = append(binaryRules, operandBinaryRules...)
		}
	case *authorizerpb.PermissionSetExpressionRef_Intersection_:
		if len(exp.Intersection.Operands) != 2 {
			panic("intersection must have exactly two operands")
		}

		for _, operand := range exp.Intersection.Operands {
			compositeKey, operandUnaryRules, operandBinaryRules := expandPermissionExpressionRefV2(schema, typedef, permissionName, operand)
			unaryRules = append(unaryRules, operandUnaryRules...)
			binaryRules = append(binaryRules, operandBinaryRules...)
		}
	default:
		panic("unexpected PermissionSetExpression type")
	}

	return unaryRules, binaryRules
}

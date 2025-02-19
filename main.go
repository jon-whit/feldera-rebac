package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	authorizerpb "github.com/jon-whit/feldera-rebac/protos/gen/go/authorizer/v1alpha1"
	"google.golang.org/protobuf/encoding/protojson"
)

var schemaPathFlag = flag.String("schema-path", "schema.json", "Path to the (.json) schema file")

func main() {
	flag.Parse()

	schemaPath := *schemaPathFlag

	jsonBytes, err := os.ReadFile(schemaPath)
	if err != nil {
		log.Fatalf("failed to open schema file: %v", err)
	}

	var schema authorizerpb.Schema
	if err := protojson.Unmarshal(jsonBytes, &schema); err != nil {
		log.Fatalf("failed to unmarshal schema: %v", err)
	}

	rules := mapSchemaToQueryRules(&schema)

	fmt.Println(rules.ToSQL())
}

// map authorizerpb.Schema to SchemaQueryRules
func mapSchemaToQueryRules(schema *authorizerpb.Schema) SchemaQueryRules {

	var typeRestrictions []RelationTypeRestriction
	for typeName, typeDefinition := range schema.GetTypeDefinitions() {
		for relationName, relationDefinition := range typeDefinition.GetRelations() {
			for _, subjectTypeRestriction := range relationDefinition.GetTypeRestrictions() {
				typeRestrictions = append(typeRestrictions, RelationTypeRestriction{
					ResourceType:    typeName,
					Relation:        relationName,
					SubjectType:     subjectTypeRestriction.GetResourceType(),
					SubjectRelation: subjectTypeRestriction.GetRelation(),
				})
			}
		}
	}

	unaryRules, binaryRules := rules(schema)

	return SchemaQueryRules{
		RelationTypeRestrictions: typeRestrictions,
		UnaryRules:               unaryRules,
		BinaryRules:              binaryRules,
	}
}

func expandPermissionExpressionRef(
	resourceType string,
	permissionName string,
	exp *authorizerpb.PermissionExpressionRef,
) ([]UnaryRule, []BinaryRule) {
	var unaryRules []UnaryRule
	var binaryRules []BinaryRule

	switch permissionExp := exp.GetExpression().(type) {
	case *authorizerpb.PermissionExpressionRef_UnaryExpression:
		rule := UnaryRule{
			ResourceType:    resourceType,
			SourceRelation:  permissionExp.UnaryExpression.GetSourceRelation(),
			DerivedRelation: permissionName,
		}
		unaryRules = append(unaryRules, rule)
	case *authorizerpb.PermissionExpressionRef_HierarchicalExpression:
	case *authorizerpb.PermissionExpressionRef_SetExpression:
		unary, binary := expandSetExpression(resourceType, permissionName, permissionExp.SetExpression)
		unaryRules = append(unaryRules, unary...)
		binaryRules = append(binaryRules, binary...)
	default:
		panic("unexpected PermissionExpressionRef type")
	}

	return unaryRules, binaryRules
}

func expandSetExpression(
	resourceType string,
	permissionName string,
	setExp *authorizerpb.PermissionSetExpressionRef,
) ([]UnaryRule, []BinaryRule) {
	var unaryRules []UnaryRule
	var binaryRules []BinaryRule

	switch exp := setExp.SetExpression.(type) {
	case *authorizerpb.PermissionSetExpressionRef_Union_:
		for _, operand := range exp.Union.Operands {
			operandUnaryRules, operandBinaryRules := expandPermissionExpressionRef(resourceType, permissionName, operand)
			unaryRules = append(unaryRules, operandUnaryRules...)
			binaryRules = append(binaryRules, operandBinaryRules...)
		}
	case *authorizerpb.PermissionSetExpressionRef_Intersection_:
		panic("intersection is not yet a supported PermissionSetExpression")
	default:
		panic("unexpected PermissionSetExpression type")
	}

	return unaryRules, binaryRules
}

func rules(schema *authorizerpb.Schema) ([]UnaryRule, []BinaryRule) {
	var unaryRules []UnaryRule
	var binaryRules []BinaryRule

	for typeName, typeDefinition := range schema.GetTypeDefinitions() {
		for permissionName, permission := range typeDefinition.GetPermissions() {

			permissionExp := permission.GetExpression()
			unary, binary := expandPermissionExpressionRef(typeName, permissionName, permissionExp)
			unaryRules = append(unaryRules, unary...)
			binaryRules = append(binaryRules, binary...)
		}
	}

	return unaryRules, binaryRules
}

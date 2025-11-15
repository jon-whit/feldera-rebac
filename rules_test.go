package main

import (
	"log"
	"reflect"
	"testing"

	authorizerpb "github.com/jon-whit/feldera-rebac/protos/gen/go/authorizer/v1alpha1"
)

func TestSchemaQueryRules_ToSQL(t *testing.T) {
	rules := SchemaQueryRules{
		RelationTypeRestrictions: []RelationTypeRestriction{
			{
				ResourceType:    "subreddit",
				Relation:        "moderator",
				SubjectType:     "account",
				SubjectRelation: "",
			},
			{
				ResourceType:    "subreddit",
				Relation:        "community_appearance_editor",
				SubjectType:     "account",
				SubjectRelation: "",
			},
			{
				ResourceType:    "group",
				Relation:        "member",
				SubjectType:     "group",
				SubjectRelation: "member",
			},
		},
		UnaryRules: []UnaryRule{
			{
				ResourceType:    "subreddit",
				SourceRelation:  "moderator",
				DerivedRelation: "can_edit_community_appearance",
			},
			{
				ResourceType:    "subreddit",
				SourceRelation:  "community_appearance_editor",
				DerivedRelation: "can_edit_community_appearance",
			},
		},
	}

	log.Println(rules.ToSQL())
}

func TestExpandPermissionExpressionRefV2(t *testing.T) {
	schema := &authorizerpb.Schema{
		TypeDefinitions: map[string]*authorizerpb.TypeDefinition{
			"subreddit": &authorizerpb.TypeDefinition{
				Name: "subreddit",
				Relations: map[string]*authorizerpb.Relation{
					"moderator": {
						TypeRestrictions: []*authorizerpb.RelationTypeRestriction{
							{
								ResourceType: "account",
								Relation:     "",
							},
						},
					},
					"community_appearance_editor": {
						TypeRestrictions: []*authorizerpb.RelationTypeRestriction{
							{
								ResourceType: "account",
								Relation:     "",
							},
						},
					},
				},
				Permissions: map[string]*authorizerpb.Permission{
					"can_edit_community_appearance": {
						Expression: &authorizerpb.PermissionExpressionRef{
							Expression: &authorizerpb.PermissionExpressionRef_UnaryExpression{
								UnaryExpression: &authorizerpb.UnaryPermissionExpression{
									SourceRelation: "moderator",
								},
							},
						},
					},
				},
			},
		},
	}

	permissionName := "can_edit_community_appearance"

	typedef := schema.GetTypeDefinitions()["subreddit"]
	compositeKey, unaryRules, _ := expandPermissionExpressionRefV2(schema, typedef, permissionName, schema.GetTypeDefinitions()["subreddit"].Permissions[permissionName].Expression)

	if compositeKey != permissionName {
		t.Errorf("expected %s, got %s", permissionName, compositeKey)
	}

	expectedUnaryRules := []UnaryRule{
		{
			ResourceType:    "subreddit",
			SourceRelation:  "moderator",
			DerivedRelation: "can_edit_community_appearance",
		},
	}

	if !reflect.DeepEqual(unaryRules, expectedUnaryRules) {
		t.Errorf("expected %v, got %v", expectedUnaryRules, unaryRules)
	}

	schema = &authorizerpb.Schema{
		TypeDefinitions: map[string]*authorizerpb.TypeDefinition{
			"user": {
				Name:        "user",
				Relations:   map[string]*authorizerpb.Relation{},
				Permissions: map[string]*authorizerpb.Permission{},
			},
			"folder": {
				Name: "folder",
				Relations: map[string]*authorizerpb.Relation{
					"viewer": {
						TypeRestrictions: []*authorizerpb.RelationTypeRestriction{
							{
								ResourceType: "user",
								Relation:     "",
							},
						},
					},
				},
				Permissions: map[string]*authorizerpb.Permission{
					"can_view": {
						Expression: &authorizerpb.PermissionExpressionRef{
							Expression: &authorizerpb.PermissionExpressionRef_UnaryExpression{
								UnaryExpression: &authorizerpb.UnaryPermissionExpression{
									SourceRelation: "viewer",
								},
							},
						},
					},
				},
			},
			"document": {
				Name: "document",
				Relations: map[string]*authorizerpb.Relation{
					"parent": {
						TypeRestrictions: []*authorizerpb.RelationTypeRestriction{
							{
								ResourceType: "folder",
								Relation:     "",
							},
						},
					},
				},
				Permissions: map[string]*authorizerpb.Permission{
					"can_view": {
						Expression: &authorizerpb.PermissionExpressionRef{
							Expression: &authorizerpb.PermissionExpressionRef_HierarchicalExpression{
								HierarchicalExpression: &authorizerpb.HierarchicalPermissionExpression{
									Base:   "parent",
									Target: "can_view",
								},
							},
						},
					},
				},
			},
		},
	}

	typedef = schema.GetTypeDefinitions()["document"]
	compositeKey, unaryRules, binaryRules := expandPermissionExpressionRefV2(schema, typedef, "can_view", schema.GetTypeDefinitions()["document"].Permissions["can_view"].Expression)

	_ = compositeKey
	_ = unaryRules
	_ = binaryRules

	unaryRules, binaryRules = rules(schema)
	expectedUnaryRules = []UnaryRule{
		{
			ResourceType:    "folder",
			SourceRelation:  "viewer",
			DerivedRelation: "can_view",
		},
	}

	expectedBinaryRules := []BinaryRule{
		{
			FirstResourceType:  "folder",
			FirstRelation:      "can_view",
			SecondResourceType: "document",
			SecondRelation:     "parent",
			DerivedRelation:    "can_view",
		},
	}

	if !reflect.DeepEqual(unaryRules, expectedUnaryRules) {
		t.Errorf("expected %v, got %v", expectedUnaryRules, unaryRules)
	}

	if !reflect.DeepEqual(binaryRules, expectedBinaryRules) {
		t.Errorf("expected %v, got %v", expectedBinaryRules, binaryRules)
	}
}

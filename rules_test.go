package main

import (
	"log"
	"testing"
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

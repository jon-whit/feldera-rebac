package ast

import "testing"

func TestAstToRules(t *testing.T) {
	tests := []struct {
		name          string
		schema        string
		expectedRules GeneratedRules
	}{
		{
			name: "computed_userset_rule",
			schema: `
			  typeDefinitions:
			    - name: user
			    - name: document
				  relations:
				    editor:
					  - name: editor
					    typeRestrictions:
						  - type: user
				  permissions:
				    viewer:
					  - name: viewer
					    expression: ...
			`,
			expectedRules: GeneratedRules{
				UnaryRules: []UnaryRule{
					{
						PrerequisiteRelation: "document-editor",
						DerivedRelation:      "document-viewer",
					},
				},
			},
		},
		{
			name: "nested_userset_rule",
			schema: `
			  typeDefinitions:
			    - name: user
			    - name: group
				  relations:
				    member:
					  - name: member
					    typeRestrictions:
						  - type: user
						  - type: group
						    relation: member
			`,
			expectedRules: GeneratedRules{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_ = test.schema // todo: unmarshal yaml

			_ = test.expectedRules //
		})
	}
}

/*
typedef user

typedef document
  relation editor: user
  permission viewer: editor

typedef group
  relation member: document#viewer

INSERT INTO unary_rules VALUES
  ('document-editor', 'document-viewer');

INSERT INTO relationships VALUES
  ('user:jon', '', 'document:1', 'document-editor'),
  ('user:bob', '', 'document:2', 'document-editor'),
  ('document:1', 'viewer', 'group:foo', 'group-member'),
  ('document:2', '', 'group:bar', 'group-member');
  ;
*/

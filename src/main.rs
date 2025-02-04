// todo: generate the Feldera rules from the parsed ReBAC TypeDefinitions

use std::collections::HashMap;

#[derive(Debug)]
struct Relationship {
    resource_type: String,
    resource_id: String,
    relation: String,
    subject_type: String,
    subject_id: String,
    subject_relation: String,
}

#[derive(Debug)]
struct Schema {
    type_definitions: HashMap<String, TypeDefinition>
}

#[derive(Debug)]
struct TypeDefinition {
    name: String,
    relations: HashMap<String, Relation>,
    permissions: HashMap<String, Permission>,
}

#[derive(Debug)]
struct Relation {
    name: String,
    type_restrictions: Vec<RelationTypeRestriction>
}

// RelationTypeRestriction constrains a relationship on some type to another related
// type.
#[derive(Debug)]
struct RelationTypeRestriction {
    type_name: String,
    relation: String,
}

// UnaryExpression, relA(subject, resource) :- permission(subject, resource)
// SetExpression
#[derive(Debug)]
enum PermissionExpression {
    UnaryExpression(String),

    // HierarchicalExpression defines the (target, source) pair for hierarchical permissions.
    //
    // Given a binary expression of the form:
    //   viewer(user:jon, folder:x), parent(folder:x, document:readme) :- can_view(user:jon, document:readme)
    // the HierarchicalExpression pair would be (viewer, parent), because 'viewer' is the target relation/permission
    // and 'parent' is the source relation/permission.
    HierarchicalExpression(HierarchicalExpression),
    SetExpression(SetExpression),
}

#[derive(Debug)]
struct HierarchicalExpression {
    target_relation: String,
    source_relation: String,
}

#[derive(Debug)]
enum SetExpression {
    Union(Vec<PermissionExpression>),
    Intersection(Vec<PermissionExpression>),
}

#[derive(Debug)]
struct Permission {
    name: String,
    expression: PermissionExpression,
}

fn main() {
    // input to the program is a grammar file
    // todo: parse a grammar file

    // start: parsed AST
    let type_definitions: HashMap<String, TypeDefinition> = [
        ("org", TypeDefinition {
            name: "org".to_string(),
            relations: [
                ("viewer", Relation {
                    name: "viewer".to_string(),
                    type_restrictions: vec![
                        RelationTypeRestriction {
                            type_name: "account".to_string(),
                            relation: "".to_string(),
                        },
                    ],
                }),
            ].into_iter().map(|(k, v)| (k.to_string(), v)).collect(),
            permissions: HashMap::new(),
        }),
        ("folder", TypeDefinition {
            name: "folder".to_string(),
            relations: [
                ("viewer", Relation {
                    name: "viewer".to_string(),
                    type_restrictions: vec![
                        RelationTypeRestriction {
                            type_name: "account".to_string(),
                            relation: "".to_string(),
                        },
                    ],
                }),
            ].into_iter().map(|(k, v)| (k.to_string(), v)).collect(),
            permissions: HashMap::new(),
        }),
        ("document", TypeDefinition {
            name: "document".to_string(),
            relations: [
                ("parent", Relation {
                    name: "parent".to_string(),
                    type_restrictions: vec![
                        RelationTypeRestriction {
                            type_name: "org".to_string(),
                            relation: "".to_string(),
                        },
                        RelationTypeRestriction {
                            type_name: "folder".to_string(),
                            relation: "".to_string(),
                        },
                    ],
                }),
            ].into_iter().map(|(k, v)| (k.to_string(), v)).collect(),
            permissions: [
                ("can_view", Permission {
                    name: "can_view".to_string(),
                    expression: PermissionExpression::HierarchicalExpression(HierarchicalExpression {
                        target_relation: "viewer".to_string(),
                        source_relation: "parent".to_string(),
                    }),
                }),
            ].into_iter().map(|(k, v)| (k.to_string(), v)).collect(),
        }),
        ("account", TypeDefinition {
            name: "account".to_string(),
            relations: HashMap::new(),
            permissions: HashMap::new(),
        }),
        ("subreddit", TypeDefinition{
            name: "subreddit".to_string(),
            relations: [
                ("moderator", Relation {
                    name: "moderator".to_string(),
                    type_restrictions: vec![
                        RelationTypeRestriction {
                            type_name: "account".to_string(),
                            relation: "".to_string(),
                        },
                    ],
                }),
                ("community_appearance_editor", Relation {
                    name: "community_appearance_editor".to_string(),
                    type_restrictions: vec![
                        RelationTypeRestriction {
                            type_name: "account".to_string(),
                            relation: "".to_string(),
                        },
                    ],
                }),
                ("org", Relation {
                    name: "org".to_string(),
                    type_restrictions: vec![
                        RelationTypeRestriction {
                            type_name: "org".to_string(),
                            relation: "".to_string(),
                        },
                    ],
                }),
            ].into_iter().map(|(k, v)| (k.to_string(), v)).collect(),
            permissions: [
                ("can_view", Permission {
                    name: "can_view".to_string(),
                    expression: PermissionExpression::SetExpression(SetExpression::Union(vec![
                        PermissionExpression::UnaryExpression("moderator".to_string()),
                        PermissionExpression::UnaryExpression("community_appearance_editor".to_string()),
                    ])),
                }),
                ("can_moderate", Permission {
                    name: "can_moderate".to_string(),
                    expression: PermissionExpression::SetExpression(SetExpression::Union(vec![
                        PermissionExpression::HierarchicalExpression(HierarchicalExpression{
                            target_relation: "admin".to_string(),
                            source_relation: "org".to_string(),
                        }),
                        PermissionExpression::SetExpression(SetExpression::Union(vec![
                            PermissionExpression::UnaryExpression("moderator".to_string()),
                            PermissionExpression::UnaryExpression("community_appearance_editor".to_string()),
                        ])),
                    ])),
                }),
            ].into_iter().map(|(k, v)| (k.to_string(), v)).collect(),
        })
    ].into_iter().map(|(k, v)| (k.to_string(), v)).collect();
    
    let schema = Schema {
        type_definitions,
    };
    // end: parsed AST

    for (type_name, type_definition) in schema.type_definitions.into_iter() {
        println!("Type: {}", type_name);

        for (permission_name, permission) in &type_definition.permissions {
            println!("  Permission: {}", permission_name);

            let rules = rules_from_permission_expression(&type_definition, permission, &permission.expression);
            for rule in rules {
                println!("    {}", rule);
            }
        }
    }

}

fn rules_from_permission_expression(
    type_definition: &TypeDefinition, 
    permission: &Permission, 
    expression: &PermissionExpression,
) -> Vec<String> {
    match expression {
        PermissionExpression::UnaryExpression(relation) => {
            // UnaryExpression produces a single unary rule
            vec![format!("{resource_type}#{relation}(subject, resource) :- {resource_type}#{permission}(subject, resource)", resource_type=type_definition.name, relation=relation, permission=permission.name)]
        },
        PermissionExpression::HierarchicalExpression(h) => {
            // HierarchicalExpression produces a binary rule(s)

            // foreach of the type restrictions on the target relation, generate a binary rule
            type_definition.relations.get(&h.source_relation).unwrap().type_restrictions.iter().map(|type_restriction| {
                format!("{target_type}#{target_relation}(subject, {target_type}), {source_type}#{source_relation}({target_type}, {source_type}) :- {source_type}#{permission}(subject, {source_type})", target_type=type_restriction.type_name, target_relation=h.target_relation, source_type=type_definition.name, source_relation=h.source_relation, permission=permission.name)
            }).collect()
        },
        PermissionExpression::SetExpression(set_expression) => {
            match set_expression {
                SetExpression::Union(expressions) => {
                    expressions.into_iter().flat_map(|expression| rules_from_permission_expression(type_definition, permission, &expression)).collect()
                },
                SetExpression::Intersection(expressions) => {
                    expressions.into_iter().flat_map(|expression| rules_from_permission_expression(type_definition, permission, &expression)).collect()
                },
            }
        },
    }
}
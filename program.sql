CREATE TYPE id_t AS string;

CREATE TABLE relationships (
    subject_type id_t not null,
    subject_id id_t not null,
    subject_relation id_t not null DEFAULT '',
    resource_type id_t not null,
    resource_id id_t not null,
    relationship id_t not null
) WITH (
    'materialized' = 'true',
    'connectors' = '[{
    "transport": {
      "name": "postgres_input",
      "config": {
        "uri": "postgresql://postgres:password@postgres:5432/postgres",
        "query": "select subject_type, subject_id, subject_relation, resource_type, resource_id, relationship from relationships;"
      }
    }
  }]'
);

CREATE TABLE type_restrictions (
	resource_type id_t not null,
	relation id_t not null,
	subject_type id_t not null,
	subject_relation id_t not null
) WITH (
    'materialized' = 'true',
    'connectors' = '[{
    "transport": {
      "name": "postgres_input",
      "config": {
        "uri": "postgresql://postgres:password@postgres:5432/postgres",
        "query": "select resource_type, relation, subject_type, subject_relation from type_restrictions;"
      }
    }
  }]'
);

CREATE TABLE unary_rules (
    resource_type id_t not null,
    prerequisite_relationship id_t not null,
    derived_relationship id_t not null
) WITH (
    'materialized' = 'true',
    'connectors' = '[{
    "transport": {
      "name": "postgres_input",
      "config": {
        "uri": "postgresql://postgres:password@postgres:5432/postgres",
        "query": "select resource_type, prerequisite_relationship, derived_relationship from unary_rules;"
      }
    }
  }]'
);

CREATE TABLE binary_rules (
    prerequisite1_resource_type id_t not null,
    prerequisite1_relationship id_t not null,
    prerequisite2_resource_type id_t not null,
    prerequisite2_relationship id_t not null,
    derived_relationship id_t not null
) WITH (
    'materialized' = 'true',
    'connectors' = '[{
    "transport": {
      "name": "postgres_input",
      "config": {
        "uri": "postgresql://postgres:password@postgres:5432/postgres",
        "query": "select prerequisite1_resource_type, prerequisite1_relationship, prerequisite2_resource_type, prerequisite2_relationship, derived_relationship from binary_rules;"
      }
    }
  }]'
);

DECLARE RECURSIVE VIEW derived_unary_relationships (
    subject_type id_t not null,
    subject_id id_t not null,
    subject_relation id_t not null,
    resource_type id_t not null,
    resource_id id_t not null,
    relationship id_t not null
);
DECLARE RECURSIVE view derived_binary_relationships (
    subject_type id_t not null,
    subject_id id_t not null,
    subject_relation id_t not null,
    resource_type id_t not null,
    resource_id id_t not null,
    relationship id_t not null
);
DECLARE RECURSIVE VIEW derived_relationships (
    subject_type id_t not null,
    subject_id id_t not null,
    subject_relation id_t not null,
    resource_type id_t not null,
    resource_id id_t not null,
    relationship id_t not null
);

CREATE MATERIALIZED VIEW derived_unary_relationships AS
SELECT
    derived_relationships.subject_type,
    derived_relationships.subject_id,
    derived_relationships.subject_relation,
    derived_relationships.resource_type,
    derived_relationships.resource_id,
    unary_rules.derived_relationship as relationship
FROM derived_relationships, unary_rules
WHERE
    derived_relationships.resource_type = unary_rules.resource_type AND
    derived_relationships.relationship = unary_rules.prerequisite_relationship;

CREATE materialized view derived_binary_relationships AS WITH
    lhs AS (
        SELECT 
            derived_relationships.subject_type,
            derived_relationships.subject_id,
            derived_relationships.subject_relation,
            derived_relationships.resource_type,
            derived_relationships.resource_id,
            binary_rules.derived_relationship AS relationship
        FROM derived_relationships, binary_rules 
        WHERE
            derived_relationships.resource_type = binary_rules.prerequisite1_resource_type AND
            derived_relationships.relationship = binary_rules.prerequisite1_relationship
    ), 
    rhs AS (
        SELECT 
            derived_relationships.subject_type,
            derived_relationships.subject_id,
            derived_relationships.subject_relation,
            derived_relationships.resource_type,
            derived_relationships.resource_id,
            binary_rules.derived_relationship AS relationship
        FROM derived_relationships, binary_rules 
        WHERE
            derived_relationships.resource_type = binary_rules.prerequisite2_resource_type AND
            derived_relationships.relationship = binary_rules.prerequisite2_relationship
    )
SELECT 
    lhs.subject_type,
    lhs.subject_id,
    lhs.subject_relation,
    rhs.resource_type,
    rhs.resource_id,
    lhs.relationship
FROM lhs, rhs
WHERE
    (lhs.resource_type = rhs.resource_type AND lhs.resource_id = rhs.resource_id) OR
    (lhs.resource_type = rhs.subject_type AND lhs.resource_id = rhs.subject_id);

CREATE MATERIALIZED VIEW derived_relationships WITH (
'connectors' = '[
  {
    "transport": {
      "name": "redis_output",
      "config": {
        "connection_string": "redis://redis:6379/0",
        "key_separator": ":"
      }
    },
    "format": {
        "name": "json",
        "config": {
          "key_fields": ["subject_type","subject_id","subject_relation", "relationship", "resource_type", "resource_id"]
        }
    }
  },
  {
    "transport": {
      "name": "redis_output",
      "config": {
        "connection_string": "redis://redis:6379/0",
        "key_separator": ":"
      }
    },
    "format": {
        "name": "json",
        "config": {
          "key_fields": ["resource_type","resource_id","relationship", "subject_type", "subject_relation", "subject_id"]
        }
    }
  }
]'
) AS
SELECT 
    relationships.subject_type,
    relationships.subject_id,
    relationships.subject_relation,
    relationships.resource_type,
    relationships.resource_id,
    relationships.relationship
FROM relationships
INNER JOIN type_restrictions
ON
    relationships.resource_type = type_restrictions.resource_type AND
    relationships.relationship = type_restrictions.relation AND
    relationships.subject_type = type_restrictions.subject_type AND
    relationships.subject_relation = type_restrictions.subject_relation
UNION ALL
SELECT
    derived_relationships.subject_type,
    derived_relationships.subject_id,
    derived_relationships.subject_relation,
    relationships.resource_type,
    relationships.resource_id,
    relationships.relationship
FROM derived_relationships, relationships 
WHERE
    derived_relationships.resource_type = relationships.subject_type AND 
    derived_relationships.resource_id = relationships.subject_id AND
    derived_relationships.relationship = relationships.subject_relation
UNION ALL
SELECT * FROM derived_unary_relationships
UNION ALL
SELECT * FROM derived_binary_relationships;
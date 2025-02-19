CREATE TYPE id_t AS string;
CREATE TABLE relationships (
    subject_type id_t not null,
    subject_id id_t not null,
    subject_relation id_t not null DEFAULT '',
    resource_type id_t not null,
    resource_id id_t not null,
    relationship id_t not null
);
CREATE TABLE type_restrictions (
	resource_type id_t not null,
	relation id_t not null,
	subject_type id_t not null,
	subject_relation id_t not null
);
CREATE TABLE unary_rules (
    resource_type id_t not null,
    prerequisite_relationship id_t not null,
    derived_relationship id_t not null
);
DECLARE RECURSIVE VIEW derived_unary_relationships (
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

CREATE MATERIALIZED VIEW derived_relationships AS
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
SELECT * FROM derived_unary_relationships;
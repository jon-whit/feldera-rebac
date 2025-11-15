CREATE TABLE relationships (
    subject_type TEXT not null,
    subject_id TEXT not null,
    subject_relation TEXT not null DEFAULT '',
    resource_type TEXT not null,
    resource_id TEXT not null,
    relationship TEXT not null
);

CREATE TABLE type_restrictions (
	resource_type TEXT not null,
	relation TEXT not null,
	subject_type TEXT not null,
	subject_relation TEXT not null
);

CREATE TABLE unary_rules (
    resource_type TEXT not null,
    prerequisite_relationship TEXT not null,
    derived_relationship TEXT not null
);

CREATE TABLE binary_rules (
    prerequisite1_resource_type TEXT not null,
    prerequisite1_relationship TEXT not null,
    prerequisite2_resource_type TEXT not null,
    prerequisite2_relationship TEXT not null,
    derived_relationship TEXT not null
);
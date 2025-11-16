create type id_t as string;

create table relationships (
    subject_type id_t not null,
    subject_id id_t not null,
    subject_relation id_t not null default '',
    resource_type id_t not null,
    resource_id id_t not null,
    relationship id_t not null
);

-- unary_rules handle the 'computed_userset' model of Zanzibar
--
-- relA(x, y) :- relB(x, y)
create table unary_rules (
    resource_type id_t not null,
    prerequisite_relationship id_t not null,
    derived_relationship id_t not null
);

-- Rules with two pre-requisites.
--
-- prerequisite1_relationship(x, y), prerequisite2_relationship(y, z) :- derived_relationship(x, z)
--
-- viewer(user:jon, folder:x) and parent(folder:x, document:1) --> can_view(user:jon, document:1)
--
-- or, alternatively,
--
-- prerequisite1_relationship(x, y), prerequisite2_relationship(x, y) :- derived_relationship(x, y)
--
-- viewer(user:jon, document:x) and allowed(user:jon, document:x) --> can_view(user:jon, document:x)
create table binary_rules (
    prerequisite1_resource_type id_t not null,
    prerequisite1_relationship id_t not null,
    prerequisite2_resource_type id_t not null,
    prerequisite2_relationship id_t not null,
    derived_relationship id_t not null
);

create table negated_binary_rules (
   prerequisite1_resource_type id_t not null,
   prerequisite1_relationship id_t not null,
   prerequisite2_resource_type id_t not null,
   prerequisite2_relationship id_t not null,
   derived_relationship id_t not null
);

-- relA(x, y) :- relB(y, x)
create table bidirectional_unary_rules (
    resource_type id_t not null,
    relation id_t not null,
    inverse_relation id_t not null
);

declare recursive view derived_unary_relationships (
    subject_type id_t not null,
    subject_id id_t not null,
    subject_relation id_t not null,
    resource_type id_t not null,
    resource_id id_t not null,
    relationship id_t not null
);

-- Relationships derived using binary rules.
declare recursive view derived_binary_relationships (
    subject_type id_t not null,
    subject_id id_t not null,
    subject_relation id_t not null,
    resource_type id_t not null,
    resource_id id_t not null,
    relationship id_t not null
);


declare recursive view derived_negated_binary_relationships (
    subject_type id_t not null,
    subject_id id_t not null,
    subject_relation id_t not null,
    resource_type id_t not null,
    resource_id id_t not null,
    relationship id_t not null
);

declare recursive view derived_bidirectional_relationships (
    subject_type id_t not null,
    subject_id id_t not null,
    subject_relation id_t not null,
    resource_type id_t not null,
    resource_id id_t not null,
    relationship id_t not null
);

-- All derived relationships.
declare recursive view derived_relationships (
    subject_type id_t not null,
    subject_id id_t not null,
    subject_relation id_t not null,
    resource_type id_t not null,
    resource_id id_t not null,
    relationship id_t not null
);

create materialized view derived_unary_relationships as
select
    derived_relationships.subject_type,
    derived_relationships.subject_id,
    derived_relationships.subject_relation,
    derived_relationships.resource_type,
    derived_relationships.resource_id,
    unary_rules.derived_relationship as relationship
from derived_relationships, unary_rules
where
    derived_relationships.resource_type = unary_rules.resource_type and
    derived_relationships.relationship = unary_rules.prerequisite_relationship;

create materialized view derived_binary_relationships as with
    lhs AS (
        select 
            derived_relationships.subject_type,
            derived_relationships.subject_id,
            derived_relationships.subject_relation,
            derived_relationships.resource_type,
            derived_relationships.resource_id,
            binary_rules.derived_relationship as relationship
        from derived_relationships, binary_rules 
        where
            derived_relationships.resource_type = binary_rules.prerequisite1_resource_type and
            derived_relationships.relationship = binary_rules.prerequisite1_relationship
    ), 
    rhs AS (
        select 
            derived_relationships.subject_type,
            derived_relationships.subject_id,
            derived_relationships.subject_relation,
            derived_relationships.resource_type,
            derived_relationships.resource_id,
            binary_rules.derived_relationship as relationship
        from derived_relationships, binary_rules 
        where
            derived_relationships.resource_type = binary_rules.prerequisite2_resource_type and
            derived_relationships.relationship = binary_rules.prerequisite2_relationship
    )
select 
    lhs.subject_type,
    lhs.subject_id,
    lhs.subject_relation,
    rhs.resource_type,
    rhs.resource_id,
    lhs.relationship
from lhs, rhs
where
    (lhs.resource_type = rhs.resource_type and lhs.resource_id = rhs.resource_id) or
    (lhs.resource_type = rhs.subject_type and lhs.resource_id = rhs.subject_id);

create materialized view derived_negated_binary_relationships as with
    base as (
        select
            derived_relationships.subject_type,
            derived_relationships.subject_id,
            derived_relationships.subject_relation,
            derived_relationships.resource_type,
            derived_relationships.resource_id,
            negated_binary_rules.derived_relationship as relationship
        from derived_relationships, negated_binary_rules
        where derived_relationships.relationship = negated_binary_rules.prerequisite1_relationship
    ),
    sub as (
        select
            derived_relationships.subject_type,
            derived_relationships.subject_id,
            derived_relationships.subject_relation,
            derived_relationships.resource_type,
            derived_relationships.resource_id,
            negated_binary_rules.derived_relationship as relationship
        from derived_relationships, negated_binary_rules
        where derived_relationships.relationship = negated_binary_rules.prerequisite2_relationship
    )
select * from base
where not exists (
    select * from sub where base.resource_type = sub.resource_type and base.resource_id = sub.resource_id
);

-- note that this doesn't work for definitions such as
--
-- @bidirectional("blockedby")
-- define blocks: [group#member]
--
-- it assumes the relationship only involved with objects blocks(account:x, account:y) :- blockedby(account:y, account:x)
create materialized view derived_bidirectional_relationships as
select 
    derived_relationships.resource_type as subject_type,
    derived_relationships.resource_id as subject_id,
    derived_relationships.subject_relation,
    derived_relationships.subject_type as resource_type,
    derived_relationships.subject_id as resource_id,
    bidirectional_unary_rules.inverse_relation as relationship
from derived_relationships, bidirectional_unary_rules
where
    derived_relationships.resource_type = bidirectional_unary_rules.resource_type and
    derived_relationships.relationship = bidirectional_unary_rules.relation;

create materialized view derived_relationships as
select * from relationships where subject_relation = ''
union all
select
    derived_relationships.subject_type,
    derived_relationships.subject_id,
    derived_relationships.subject_relation,
    relationships.resource_type,
    relationships.resource_id,
    relationships.relationship
from derived_relationships, relationships 
where
    derived_relationships.resource_type = relationships.subject_type and 
    derived_relationships.resource_id = relationships.subject_id and
    derived_relationships.relationship = relationships.subject_relation
union all
select * from derived_unary_relationships
union all
select * from derived_binary_relationships
union all
select * from derived_negated_binary_relationships
union all
select * from derived_bidirectional_relationships;
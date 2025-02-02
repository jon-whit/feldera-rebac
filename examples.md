## Example 1
typedef user

typedef document
  relation viewer: user
  relation restricted: user

  permission can_view = viewer but not restricted


INSERT INTO negated_binary_rules VALUES
  ('viewer', 'restricted', 'can_view');

INSERT INTO relationships VALUES
  ('user:jon', 'document:1', 'viewer');

SELECT * FROM derived_relationships;

subject_id	resource_id	relationship
user:jon	document:1	can_view
user:jon	document:1	viewer

INSERT INTO relationships VALUES
  ('user:jon', 'document:1', 'restricted');

SELECT * FROM derived_relationships;

subject_id	resource_id	relationship
user:jon	document:1	restricted
user:jon	document:1	viewer


## Example 2

typedef user

typedef document
  relation editor: user
  relation restricted: user
  relation veryrestricted: user

  relation viewer: editor
  relation blocked: restricted and veryrestricted

  relation can_view: viewer but not blocked


INSERT INTO unary_rules VALUES
  ('editor', 'viewer');

INSERT INTO binary_rules VALUES
  ('restricted', 'veryrestricted', 'blocked');

INSERT INTO negated_binary_rules VALUES
  ('viewer', 'blocked', 'can_view');

INSERT INTO relationships VALUES
  ('user:jon', 'document:1', 'editor');

SELECT * FROM derived_relationships;

subject_id	resource_id	relationship
user:jon	document:1	viewer
user:jon	document:1	can_view
user:jon	document:1	editor

INSERT INTO relationships VALUES
  ('user:jon', 'document:1', 'restricted');

SELECT * FROM derived_relationships;

subject_id	resource_id	relationship
user:jon	document:1	viewer
user:jon	document:1	restricted
user:jon	document:1	can_view
user:jon	document:1	editor

INSERT INTO relationships VALUES
  ('user:jon', 'document:1', 'veryrestricted');

SELECT * FROM derived_relationships;

## Example 3

typedef user

typedef folder
  relation editor: user
  relation blocked: user

  relation viewer: editor

  permission can_view = viewer

typedef document
  relation parent: folder
  relation owner: user
  
  relation can_view: owner or can_view from parent

INSERT INTO unary_rules VALUES
  ('folder-editor', 'folder-viewer'),
  ('folder-viewer', 'folder-can_view'),
  ('document-owner', 'document-can_view');

-- prereq1_relationship(object1, object2) and prereq2_relationship(object2, object3) --> derived_relationship(object1, object3)
--
-- (prereq1, prereq2, derived)
--
-- (folder-can_view, document-parent, document-can_view)
--
-- folder-can_view(user:bob, folder:1) and document-parent(folder:1, document:y) --> document-can_view(user:bob, document:y)
INSERT INTO binary_rules VALUES
  ('folder-can_view', 'document-parent', 'document-can_view');

INSERT INTO relationships VALUES
  ('user:jill', 'document:x', 'document-owner');

SELECT * FROM derived_relationships;

subject_id | resource_id | relationship
user:jill	 | document:x	 | document-owner
user:jill	 | document:x	 | document-can_view

INSERT INTO relationships VALUES
  ('folder:1', 'document:y', 'document-parent'),
  ('user:bob', 'folder:1', 'folder-editor');

## Example 4

NOTE TO SELF: 
  How do we handle arbitrarily nested expressions, for example 'permission can_view = a or (b but not ((c and d) or e))'. Such expressions with the unary and binary rule definitions semantics would require many intermediate rules derived.

typedef org
  relation banned: user

typedef group
  relation org: org
  relation member: user

  permission is_member = member but not banned from org -- have to represent the intermediate binary_rule 'banned from org' using a placeholder derived relation
  
  -- this alternative is more ergonomic because of the split of the explicit intermediate definition, thus no intermediate derived relation is necessary
  --
  -- permission banned = banned from org
  -- permission is_member = member but not banned


INSERT INTO binary_rules VALUES
  ('org-banned', 'group-org', 'group-is_member[1]'); -- group-is_member[1] is notation for the operand at index 1 in the rule definition, 'member' is index 0 and 'banned from org' is index 1
  
  -- alternatively, it would reduce the number of instances of intermediate relationships if we used a static placeholder for '
  -- ('org-banned', 'group-org', 'group-[banned-from-org]')

INSERT INTO negated_binary_rules VALUES
  ('group-member', 'group-is_member[1]', 'group-is_member');

  -- alternatively,
  --('group-member', 'group-[banned-from-org]', 'group-is_member');


INSERT INTO relationships VALUES
  ('user:jon', 'group:x', 'group-member'),
  ('org:acme', 'group:x', 'group-org');

SELECT * FROM derived_relationships;

subject_id	resource_id	relationship
org:acme	group:x	group-org
user:jon	group:x	group-member
user:jon	group:x	group-is_member

INSERT INTO relationships VALUES
  ('user:jon', 'org:acme', 'org-banned');

SELECT * FROM derived_relationships;

subject_id	resource_id	relationship
org:acme	group:x	group-org
user:jon	org:acme	org-banned
user:jon	group:x	group-member
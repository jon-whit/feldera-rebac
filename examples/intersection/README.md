# Relationships with Intersection
This example demonstrates how we can model relationships that must intersect. If you are familiar with [Google Zanzibar](https://storage.googleapis.com/gweb-research2023-media/pubtools/5068.pdf), these are relationships of the form `a and b`.

The example [schema.json](./schema.json) represents the following declarative relationship model:
```
typedef user {}

typedef document {

    relation viewer: [user]
    relation allowed: [user]

    permission can_view = viewer and allowed
}
```

## Try It Out
> ℹ️ The commands assume you are running from the root path of this repository.

1. Start Feldera
```
docker run -p 8080:8080 --tty --rm -it ghcr.io/feldera/pipeline-manager:0.33.0
```

2. Create a Feldera Pipeline called `rebac`
```
curl -L 'http://localhost:8080/v0/pipelines' \
-H 'Content-Type: application/json' \
-H 'Accept: application/json' \
-d '{
  "description": "A Feldera sample that demonstrates ReBAC models.",
  "name": "rebac",
  "program_code": ""
}'
```

3. Create the necessary tables and views for the relationship graph SQL program.

Copy the SQL program.
```
cat program.sql | pbcopy
```

Navigate to http://localhost:8080/pipelines/rebac/ and paste the SQL program into the `program.sql` file.
![](../../docs/program-sql-screenshot.png)

4. Start the Feldera Pipeline by hitting the "Start" button

5. Run the rules generator.
The `schema.json` file contains a sample schema definition for a ReBAC model. We convert this schema definition into data that defines the rules for how the relationship subgraphs should be derived.

```
go run main.go rules.go --schema-path ./examples/hierarchical-relationships/schema.json
```

This will output:

```
INSERT INTO type_restrictions VALUES
('document', 'viewer', 'user', ''),
('document', 'allowed', 'user', '');

INSERT INTO binary_rules VALUES
('document', 'viewer', 'document', 'allowed', 'can_view');
```

6. Copy the `INSERT` statements from step 5 into the "Ad-Hoc Queries" window in the Feldera Pipeline dashboard, and run them.

7. Insert some relationships
```
INSERT INTO relationships VALUES
    ('user', 'jon', '', 'document', 'readme', 'viewer'),
    ('user', 'jon', '', 'document', 'readme', 'allowed'),
    ('user', 'bob', '', 'document', 'readme', 'viewer');
```

8. Check the status of the relationship graph by querying the `dervied_relationships` table.
```
SELECT * FROM derived_relationships;
```


> ℹ️ Notice that `user:jon` can_view the `document:readme` because he has `viewer` and `allowed`, but `user:bob` cannot because he only has `viewer`.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	schemav2 "github.com/authzed/spicedb/pkg/schema/v2"
	"github.com/authzed/spicedb/pkg/schemadsl/compiler"
	"github.com/authzed/spicedb/pkg/schemadsl/input"
)

var schemaPathFlag = flag.String("schema-path", "schema.zed", "Path to the (.zed) schema file")

func main() {
	flag.Parse()

	schemaPath := *schemaPathFlag

	schemaBytes, err := os.ReadFile(schemaPath)
	if err != nil {
		log.Fatalf("failed to open schema file: %v", err)
	}

	compiledSchema, err := compiler.Compile(
		compiler.InputSchema{
			Source:       input.Source(schemaPath),
			SchemaString: string(schemaBytes),
		},
		compiler.AllowUnprefixedObjectType(),
	)
	if err != nil {
		log.Fatalf("failed to compile schema: %v", err)
	}

	s, err := schemav2.BuildSchemaFromCompiledSchema(*compiledSchema)
	if err != nil {
		log.Fatalf("failed to build schema from compiled source: %v", err)
	}

	schema, err := schemav2.ResolveSchema(s)
	if err != nil {
		log.Fatalf("failed to resolve schema: %v", err)
	}

	flattenedSchema, err := schemav2.FlattenSchema(schema, schemav2.FlattenSeparatorDollar)
	if err != nil {
		log.Fatalf("failed to flatten schema: %v", err)
	}

	visitor := &visitor{}

	r, err := schemav2.WalkFlattenedSchema(flattenedSchema, visitor, visitor.rules)
	if err != nil {
		log.Fatalf("failed to walk flattened schema: %v", err)
	}

	for _, rule := range r {
		log.Printf("%+v", rule)
	}
}

type UnaryRule struct {
	ResourceType    string `json:"resource_type"`
	SourceRelation  string `json:"source_relation"`
	DerivedRelation string `json:"derived_relation"`
}

func (u *UnaryRule) isExprRule() {}

type BinaryRule struct {
	FirstResourceType  string `json:"first_resource_type"`
	FirstRelation      string `json:"first_relation"`
	SecondResourceType string `json:"second_resource_type"`
	SecondRelation     string `json:"second_relation"`
	DerivedRelation    string `json:"derived_relation"`
}

func (b *BinaryRule) isExprRule() {}

type ExprRule interface {
	isExprRule()
}

type visitor struct {
	permissions []*schemav2.Permission

	rules []ExprRule
}

func (v *visitor) VisitPermission(p *schemav2.Permission, value []ExprRule) ([]ExprRule, bool, error) {
	v.permissions = append(v.permissions, p)

	// visit permission should return either a unary rule, binary rule, or a negated binary rule

	switch op := p.Operation().(type) {
	case *schemav2.RelationReference:
		sourceType := p.Parent().Name()
		derivedRelation := p.Name()
		sourceRelation := op.RelationName()

		return []ExprRule{
			&UnaryRule{
				ResourceType:    sourceType,
				SourceRelation:  sourceRelation,
				DerivedRelation: derivedRelation,
			},
		}, true, nil
	case *schemav2.ResolvedRelationReference:
		sourceType := p.Parent().Name()
		derivedRelation := p.Name()
		sourceRelation := op.RelationName()

		v.rules = append(v.rules, &UnaryRule{
			ResourceType:    sourceType,
			SourceRelation:  sourceRelation,
			DerivedRelation: derivedRelation,
		})

		return v.rules, true, nil

	case *schemav2.ResolvedArrowReference:
		var binaryRules []ExprRule
		for _, baseRelation := range op.ResolvedLeft().BaseRelations() {
			binaryRules = append(binaryRules, &BinaryRule{
				FirstResourceType:  p.Parent().Name(),
				FirstRelation:      op.Left(),
				SecondResourceType: baseRelation.Type(),
				SecondRelation:     op.Right(),
				DerivedRelation:    p.Name(),
			})
		}

		v.rules = append(v.rules, binaryRules...)

		return v.rules, true, nil
	case *schemav2.UnionOperation:
		// produce a unary rule for each child

		children := op.Children()
		if len(children) != 2 {
			return nil, false, fmt.Errorf("union must have exactly two operands")
		}

		child1Op, child2Op := children[0], children[1]

		child1, ok := child1Op.(*schemav2.ResolvedRelationReference)
		if !ok {
			return nil, false, fmt.Errorf("expected a resolved relation reference")
		}

		child2, ok := child2Op.(*schemav2.ResolvedRelationReference)
		if !ok {
			return nil, false, fmt.Errorf("expected a resolved relation reference")
		}

		sourceType := p.Parent().Name()
		derivedRelation := p.Name()

		sourceRelation1 := child1.RelationName()
		sourceRelation2 := child2.RelationName()

		v.rules = append(v.rules, []ExprRule{
			&UnaryRule{
				ResourceType:    sourceType,
				SourceRelation:  sourceRelation1,
				DerivedRelation: derivedRelation,
			},
			&UnaryRule{
				ResourceType:    sourceType,
				SourceRelation:  sourceRelation2,
				DerivedRelation: derivedRelation,
			},
		}...)

		return v.rules, true, nil

	case *schemav2.IntersectionOperation:
		// produce a binary rule for the children

		children := op.Children()
		if len(children) != 2 {
			return nil, false, fmt.Errorf("intersection must have exactly two operands")
		}

		child1Op, child2Op := children[0], children[1]

		child1, ok := child1Op.(*schemav2.ResolvedRelationReference)
		if !ok {
			return nil, false, fmt.Errorf("expected a resolved relation reference")
		}

		child2, ok := child2Op.(*schemav2.ResolvedRelationReference)
		if !ok {
			return nil, false, fmt.Errorf("expected a resolved relation reference")
		}

		sourceType := p.Parent().Name()
		derivedRelation := p.Name()

		v.rules = append(v.rules, &BinaryRule{
			FirstResourceType:  sourceType,
			FirstRelation:      child1.RelationName(),
			SecondResourceType: sourceType,
			SecondRelation:     child2.RelationName(),
			DerivedRelation:    derivedRelation,
		})

	//case *schemav2.ExclusionOperation:
	// produce a negated binary rule for the children

	default:
		return nil, false, fmt.Errorf("unsupported operation type: %T", op)
	}

	return value, true, nil
}

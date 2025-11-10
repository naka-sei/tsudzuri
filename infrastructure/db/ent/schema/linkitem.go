package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	guuid "github.com/google/uuid"
	tsuuid "github.com/naka-sei/tsudzuri/pkg/uuid"
)

// LinkItem represents an item (URL) inside a Page.
type LinkItem struct{ ent.Schema }

func (LinkItem) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "link_items", Schema: "tsudzuri"},
	}
}

func (LinkItem) Mixin() []ent.Mixin { return []ent.Mixin{TimeMixin{}} }

func (LinkItem) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", guuid.UUID{}).Default(tsuuid.NewV7),
		field.UUID("page_id", guuid.UUID{}),
		field.Text("url").NotEmpty(),
		field.Text("memo").Optional().Nillable(),
		field.Int("priority").Default(0),
	}
}

func (LinkItem) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("page", Page.Type).Ref("link_items").Field("page_id").Unique().Required(),
	}
}

// Indexes could be added here if needed (e.g., composite index page_id+priority).

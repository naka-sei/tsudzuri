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

// Page holds the schema definition for the Page entity.
type Page struct{ ent.Schema }

func (Page) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "pages", Schema: "tsudzuri"},
	}
}

func (Page) Mixin() []ent.Mixin { return []ent.Mixin{TimeMixin{}} }

func (Page) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", guuid.UUID{}).Default(tsuuid.NewV7),
		field.String("title").NotEmpty().MaxLen(50),
		field.UUID("creator_id", guuid.UUID{}), // FK for creator edge
		field.String("invite_code").NotEmpty().Unique().MaxLen(8),
	}
}

func (Page) Edges() []ent.Edge {
	return []ent.Edge{
		// Creator of the page.
		edge.From("creator", User.Type).Ref("created_pages").Field("creator_id").Unique().Required(),
		// Link items contained in this page.
		edge.To("link_items", LinkItem.Type),
		// M2M invited users via existing join table tsudzuri.page_users.
		// Define the storage key here; the inverse side is User.invited_pages.
		edge.To("invited_users", User.Type).
			Annotations(entsql.Annotation{Table: "page_users", Schema: "tsudzuri"}).
			StorageKey(
				edge.Table("page_users"),
				edge.Columns("page_id", "user_id"),
			),
	}
}

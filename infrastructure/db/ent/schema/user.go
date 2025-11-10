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

// User holds the schema definition for the User entity.
type User struct{ ent.Schema }

// Annotations sets the SQL schema and table name.
func (User) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "users", Schema: "tsudzuri"},
	}
}

// Mixin defines common timestamp fields.
func (User) Mixin() []ent.Mixin { return []ent.Mixin{TimeMixin{}} }

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", guuid.UUID{}).Default(tsuuid.NewV7),
		field.String("uid").NotEmpty().Unique(),
		field.Enum("provider").Values("anonymous", "google", "facebook").Default("anonymous"),
		field.String("email").Optional().Nillable().Unique(),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("created_pages", Page.Type),
		// M2M invited pages via existing join table tsudzuri.page_users
		// The join table definition is configured on Page.invited_users.
		edge.From("invited_pages", Page.Type).Ref("invited_users"),
	}
}

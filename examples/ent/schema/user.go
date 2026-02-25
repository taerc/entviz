package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Comment("用户姓名"),
		field.String("email").
			Comment("用户邮箱地址"),
		field.Int8("role").
			Comment("用户角色"),
		field.Time("created").
			Default(time.Now).
			Comment("创建时间"),
		field.Int("age").
			Range(0, 1000).
			Optional().
			Nillable().
			Comment("用户年龄"),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("pets", Pet.Type),
		edge.To("posts", Post.Type),
		edge.To("parent", User.Type).
			Unique(),
		edge.To("cars", Car.Type),
	}
}

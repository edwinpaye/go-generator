package dbml

import (
	"context"
	"testing"
)

func TestParseDBML_FullFeatures(t *testing.T) {
	sampleDBML := `
Project ECommerce {
	Note: 'Enterprise REST API Backend Schema'
}

Enum auth.UserStatus {
	ACTIVE [note: 'Active User']
	INACTIVE
	SUSPENDED
}

TablePartial AuditFields {
	created_at timestamp [not null, default: 'now()']
	updated_at timestamp [not null]
}

Table auth.users {
	id uuid [pk, increment]
	email varchar(255) [not null, unique]
	status auth.UserStatus [default: 'ACTIVE']
	mixin AuditFields

	indexes {
		email [unique, name: 'idx_user_email']
		(status, created_at) [name: 'idx_status_created']
	}
}

Table orders {
	id uuid [pk]
	user_id uuid [not null, ref: > auth.users.id]
	total_amount decimal(10,2) [not null]
	mixin AuditFields
}

TableGroup AuthDomain {
	auth.users
}

Ref fk_orders_user: orders.user_id > auth.users.id [delete: cascade]
`

	parser := NewParser()
	doc, err := parser.Parse(context.Background(), sampleDBML)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if doc.Project == nil || doc.Project.Name != "ECommerce" {
		t.Errorf("Expected project name ECommerce, got %v", doc.Project)
	}

	if len(doc.Enums) != 1 || doc.Enums[0].Name != "UserStatus" || doc.Enums[0].Schema != "auth" {
		t.Errorf("Enum parse mismatch: %+v", doc.Enums)
	}

	if len(doc.TablePartials) != 1 || doc.TablePartials[0].Name != "AuditFields" {
		t.Errorf("TablePartial parse mismatch: %+v", doc.TablePartials)
	}

	if len(doc.Tables) != 2 {
		t.Fatalf("Expected 2 tables, got %d", len(doc.Tables))
	}

	userTable := doc.FindTable("auth", "users")
	if userTable == nil {
		t.Fatalf("Could not find table auth.users")
	}

	if len(userTable.PartialsUsed) != 1 || userTable.PartialsUsed[0] != "AuditFields" {
		t.Errorf("Table mixin not recorded correctly: %v", userTable.PartialsUsed)
	}

	if len(userTable.Indexes) != 2 {
		t.Errorf("Expected 2 indexes on auth.users, got %d", len(userTable.Indexes))
	}

	if len(doc.TableGroups) != 1 || doc.TableGroups[0].Name != "AuthDomain" {
		t.Errorf("TableGroup parse mismatch: %+v", doc.TableGroups)
	}
}

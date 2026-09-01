package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new kogen workspace with sample DBML schema and kogen.yaml configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		sampleDBML := `Project ECommerceBackend {
  Note: 'Enterprise-Grade Microservice DBML Schema'
}

Enum auth.UserStatus {
  ACTIVE [note: 'Account is active']
  SUSPENDED [note: 'Suspended by admin']
  PENDING
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
  status varchar(50) [not null, default: 'CREATED']
  mixin AuditFields
}

TableGroup AuthDomain {
  auth.users
}

TableGroup OrderDomain {
  orders
}

Ref fk_orders_user: orders.user_id > auth.users.id [delete: cascade]
`

		sampleConfig := `project_name: "ecommerce-api"
package_name: "com.enterprise.ecommerce"
target_dir: "./output/ecommerce-api"
database: "postgres"
architecture: "hexagonal"
native_target: "graalvm"
include_in_memory: true
include_openapi: true
include_docker: true
include_migrations: true
server_port: 8080
`

		if err := os.WriteFile("schema.dbml", []byte(sampleDBML), 0644); err != nil {
			return err
		}
		fmt.Println("✔ Created 'schema.dbml'")

		if err := os.WriteFile("kogen.yaml", []byte(sampleConfig), 0644); err != nil {
			return err
		}
		fmt.Println("✔ Created 'kogen.yaml'")

		fmt.Println("\n🚀 Workspace initialized! Run `kogen generate` to create your Kotlin REST API backend.")
		return nil
	},
}

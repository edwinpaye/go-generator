package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var domainTemplate string

var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "Generate enterprise-grade DBML templates for specific industry domain verticals (e-commerce, fintech, saas, healthcare)",
	RunE: func(cmd *cobra.Command, args []string) error {
		var schemaContent string
		domainName := strings.ToLower(domainTemplate)

		switch domainName {
		case "fintech":
			schemaContent = `Project FintechDomain {
  Note: 'Fintech Core Banking & Digital Wallet DBML Schema'
}

Enum bank.AccountType {
  SAVINGS
  CHECKING
  INVESTMENT
  CREDIT
}

Enum bank.TransactionStatus {
  PENDING
  COMPLETED
  FAILED
  REVERSED
}

TablePartial Audit {
  created_at timestamp [not null, default: 'now()']
  updated_at timestamp [not null]
}

Table bank.accounts {
  id uuid [pk]
  account_number varchar(50) [not null, unique]
  type bank.AccountType [not null, default: 'CHECKING']
  balance decimal(18,4) [not null, default: '0.0000']
  currency varchar(3) [not null, default: 'USD']
  mixin Audit
}

Table bank.transactions {
  id uuid [pk]
  account_id uuid [not null, ref: > bank.accounts.id]
  amount decimal(18,4) [not null]
  status bank.TransactionStatus [not null, default: 'PENDING']
  reference_code varchar(100) [unique]
  mixin Audit
}

TableGroup Banking {
  bank.accounts
  bank.transactions
}
`

		case "saas":
			schemaContent = `Project SaasDomain {
  Note: 'Multi-Tenant SaaS Subscription Platform Schema'
}

Enum saas.SubscriptionPlan {
  FREE
  STARTER
  PRO
  ENTERPRISE
}

TablePartial Audit {
  created_at timestamp [not null, default: 'now()']
  updated_at timestamp [not null]
}

Table saas.tenants {
  id uuid [pk]
  name varchar(255) [not null]
  domain varchar(100) [unique]
  plan saas.SubscriptionPlan [not null, default: 'FREE']
  mixin Audit
}

Table saas.users {
  id uuid [pk]
  tenant_id uuid [not null, ref: > saas.tenants.id]
  email varchar(255) [not null, unique]
  role varchar(50) [not null, default: 'MEMBER']
  mixin Audit
}

TableGroup SaaSPlatform {
  saas.tenants
  saas.users
}
`

		case "healthcare":
			schemaContent = `Project HealthcareDomain {
  Note: 'EHR & Patient Care Management DBML Schema'
}

Enum health.AppointmentStatus {
  SCHEDULED
  COMPLETED
  CANCELLED
}

TablePartial Audit {
  created_at timestamp [not null, default: 'now()']
  updated_at timestamp [not null]
}

Table health.patients {
  id uuid [pk]
  medical_record_number varchar(50) [not null, unique]
  full_name varchar(255) [not null]
  birth_date date [not null]
  mixin Audit
}

Table health.appointments {
  id uuid [pk]
  patient_id uuid [not null, ref: > health.patients.id]
  doctor_name varchar(255) [not null]
  appointment_time timestamp [not null]
  status health.AppointmentStatus [not null, default: 'SCHEDULED']
  mixin Audit
}

TableGroup HealthCare {
  health.patients
  health.appointments
}
`

		default: // E-Commerce
			schemaContent = `Project ECommerce {
  Note: 'Enterprise Microservice E-Commerce Schema'
}

Enum store.OrderStatus {
  CREATED
  PROCESSING
  SHIPPED
  DELIVERED
  CANCELLED
}

TablePartial Audit {
  created_at timestamp [not null, default: 'now()']
  updated_at timestamp [not null]
}

Table store.products {
  id uuid [pk]
  name varchar(255) [not null]
  price decimal(10,2) [not null]
  sku varchar(50) [unique]
  stock_quantity int [not null, default: 0]
  mixin Audit
}

Table store.orders {
  id uuid [pk]
  total_amount decimal(10,2) [not null]
  status store.OrderStatus [not null, default: 'CREATED']
  mixin Audit
}

TableGroup Storefront {
  store.products
  store.orders
}
`
		}

		targetFile := fmt.Sprintf("schema-%s.dbml", domainName)
		if err := os.WriteFile(targetFile, []byte(schemaContent), 0644); err != nil {
			return err
		}

		fmt.Printf("✨ Successfully generated '%s' domain DBML template!\n", targetFile)
		return nil
	},
}

func init() {
	templateCmd.Flags().StringVarP(&domainTemplate, "domain", "d", "ecommerce", "Domain vertical template (ecommerce, fintech, saas, healthcare)")
}

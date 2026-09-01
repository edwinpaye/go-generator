package domain

import (
	"fmt"
	"strings"
)

// DBMLDocument represents a parsed DBML schema file
type DBMLDocument struct {
	Project       *Project
	Tables        []Table
	TablePartials []TablePartial
	Enums         []Enum
	TableGroups   []TableGroup
	Refs          []Ref
}

// Project contains DBML project metadata
type Project struct {
	Name     string
	Note     string
	Settings map[string]string
}

// Table represents a database table definition (with optional schema scope e.g., auth.users)
type Table struct {
	Schema       string
	Name         string
	Note         string
	Columns      []Column
	Indexes      []Index
	PartialsUsed []string
}

// FullName returns the schema-qualified table name e.g. "auth.users" or "users"
func (t Table) FullName() string {
	if t.Schema != "" {
		return t.Schema + "." + t.Name
	}
	return t.Name
}

// TablePartial represents a partial table definition (mixin reusable schema block)
type TablePartial struct {
	Name    string
	Note    string
	Columns []Column
}

// Column represents a single table column definition
type Column struct {
	Name       string
	Type       string
	PK         bool
	Nullable   bool
	Unique     bool
	Increment  bool
	DefaultVal string
	Note       string
	Ref        *RefInline
	Settings   map[string]string
}

// RefInline represents a relationship specified inline in column options [ref: > users.id]
type RefInline struct {
	Type         string // ">", "<", "-", "<>"
	TargetSchema string
	TargetTable  string
	TargetColumn string
}

// RefType defines relationship multiplicity
type RefType string

const (
	RefManyToOne  RefType = ">"  // N:1
	RefOneToMany  RefType = "<"  // 1:N
	RefOneToOne   RefType = "-"  // 1:1
	RefManyToMany RefType = "<>" // N:M
)

// Ref represents a foreign key relationship between tables
type Ref struct {
	Name          string
	Type          RefType
	SourceSchema  string
	SourceTable   string
	SourceColumns []string
	TargetSchema  string
	TargetTable   string
	TargetColumns []string
	OnDelete      string
	OnUpdate      string
}

func (r Ref) SourceFullName() string {
	if r.SourceSchema != "" {
		return r.SourceSchema + "." + r.SourceTable
	}
	return r.SourceTable
}

func (r Ref) TargetFullName() string {
	if r.TargetSchema != "" {
		return r.TargetSchema + "." + r.TargetTable
	}
	return r.TargetTable
}

// Enum represents a custom database enum type definition
type Enum struct {
	Schema string
	Name   string
	Values []EnumValue
}

func (e Enum) FullName() string {
	if e.Schema != "" {
		return e.Schema + "." + e.Name
	}
	return e.Name
}

// EnumValue represents a single value within an Enum
type EnumValue struct {
	Name string
	Note string
}

// TableGroup represents a logical group of tables (mapped to Bounded Contexts / Sub-packages)
type TableGroup struct {
	Name  string
	Note  string
	Items []TableGroupItem
}

type TableGroupItem struct {
	Schema string
	Table  string
}

func (item TableGroupItem) FullName() string {
	if item.Schema != "" {
		return item.Schema + "." + item.Table
	}
	return item.Table
}

// Index represents an index on a table
type Index struct {
	Columns []string
	Name    string
	Unique  bool
	Type    string // e.g. btree, hash, gin
	Note    string
}

// FindTable returns a pointer to a table matching schema and name
func (doc *DBMLDocument) FindTable(schema, name string) *Table {
	for i := range doc.Tables {
		if (doc.Tables[i].Schema == schema || (doc.Tables[i].Schema == "" && schema == "")) &&
			strings.EqualFold(doc.Tables[i].Name, name) {
			return &doc.Tables[i]
		}
	}
	// Fallback check by simple name if schema isn't explicit
	for i := range doc.Tables {
		if strings.EqualFold(doc.Tables[i].Name, name) {
			return &doc.Tables[i]
		}
	}
	return nil
}

// FindEnum returns an enum matching name
func (doc *DBMLDocument) FindEnum(name string) *Enum {
	for i := range doc.Enums {
		if strings.EqualFold(doc.Enums[i].Name, name) || strings.EqualFold(doc.Enums[i].FullName(), name) {
			return &doc.Enums[i]
		}
	}
	return nil
}

func (doc *DBMLDocument) Summary() string {
	return fmt.Sprintf("Project: %s, Tables: %d, Partials: %d, Enums: %d, TableGroups: %d, Refs: %d",
		func() string {
			if doc.Project != nil {
				return doc.Project.Name
			}
			return "Unnamed"
		}(),
		len(doc.Tables), len(doc.TablePartials), len(doc.Enums), len(doc.TableGroups), len(doc.Refs))
}

package dbml

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/antigravity/kogen/internal/core/domain"
)

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

type tokenStream struct {
	tokens []Token
	pos    int
}

func (s *tokenStream) peek() Token {
	if s.pos >= len(s.tokens) {
		return Token{Type: TokenEOF}
	}
	return s.tokens[s.pos]
}

func (s *tokenStream) next() Token {
	tok := s.peek()
	if tok.Type != TokenEOF {
		s.pos++
	}
	return tok
}

func (s *tokenStream) match(val string) bool {
	if strings.EqualFold(s.peek().Value, val) {
		s.next()
		return true
	}
	return false
}

func (s *tokenStream) expect(val string) error {
	tok := s.next()
	if !strings.EqualFold(tok.Value, val) {
		return fmt.Errorf("expected '%s' at line %d col %d, got '%s'", val, tok.Line, tok.Col, tok.Value)
	}
	return nil
}

func (s *Parser) ParseFile(ctx context.Context, filePath string) (*domain.DBMLDocument, error) {
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read DBML file %s: %w", filePath, err)
	}
	return s.Parse(ctx, string(bytes))
}

func (s *Parser) Parse(ctx context.Context, dbmlContent string) (*domain.DBMLDocument, error) {
	lexer := NewLexer(dbmlContent)
	tokens := lexer.Tokenize()
	stream := &tokenStream{tokens: tokens, pos: 0}

	doc := &domain.DBMLDocument{
		Tables:        make([]domain.Table, 0),
		TablePartials: make([]domain.TablePartial, 0),
		Enums:         make([]domain.Enum, 0),
		TableGroups:   make([]domain.TableGroup, 0),
		Refs:          make([]domain.Ref, 0),
	}

	for stream.peek().Type != TokenEOF {
		tok := stream.peek()
		val := strings.ToLower(tok.Value)

		switch val {
		case "project":
			p, err := parseProject(stream)
			if err != nil {
				return nil, err
			}
			doc.Project = p

		case "tablepartial":
			tp, err := parseTablePartial(stream)
			if err != nil {
				return nil, err
			}
			doc.TablePartials = append(doc.TablePartials, tp)

		case "table":
			t, err := parseTable(stream)
			if err != nil {
				return nil, err
			}
			doc.Tables = append(doc.Tables, t)

		case "enum":
			e, err := parseEnum(stream)
			if err != nil {
				return nil, err
			}
			doc.Enums = append(doc.Enums, e)

		case "tablegroup":
			tg, err := parseTableGroup(stream)
			if err != nil {
				return nil, err
			}
			doc.TableGroups = append(doc.TableGroups, tg)

		case "ref":
			refs, err := parseRefBlock(stream)
			if err != nil {
				return nil, err
			}
			doc.Refs = append(doc.Refs, refs...)

		default:
			// Skip unknown token
			stream.next()
		}
	}

	return doc, nil
}

func parseProject(stream *tokenStream) (*domain.Project, error) {
	stream.next() // consume 'Project'
	p := &domain.Project{Settings: make(map[string]string)}

	// Optional project name before '{'
	if stream.peek().Value != "{" {
		p.Name = stream.next().Value
	}

	if stream.peek().Value == "{" {
		stream.next() // consume '{'
		for stream.peek().Value != "}" && stream.peek().Type != TokenEOF {
			key := stream.next().Value
			if stream.peek().Value == ":" {
				stream.next()
			}
			val := stream.next().Value
			if strings.EqualFold(key, "note") {
				p.Note = val
			} else {
				p.Settings[key] = val
			}
		}
		if stream.peek().Value == "}" {
			stream.next()
		}
	}
	return p, nil
}

func parseTablePartial(stream *tokenStream) (domain.TablePartial, error) {
	stream.next() // consume 'TablePartial'
	tp := domain.TablePartial{
		Name:    stream.next().Value,
		Columns: make([]domain.Column, 0),
	}

	if err := stream.expect("{"); err != nil {
		return tp, err
	}

	for stream.peek().Value != "}" && stream.peek().Type != TokenEOF {
		col, err := parseColumn(stream)
		if err != nil {
			return tp, err
		}
		tp.Columns = append(tp.Columns, col)
	}

	if stream.peek().Value == "}" {
		stream.next()
	}
	return tp, nil
}

func parseTable(stream *tokenStream) (domain.Table, error) {
	stream.next() // consume 'Table'
	tbl := domain.Table{
		Columns:      make([]domain.Column, 0),
		Indexes:      make([]domain.Index, 0),
		PartialsUsed: make([]string, 0),
	}

	// Schema & Name parse: e.g. "public.users" or "users"
	firstName := stream.next().Value
	if stream.peek().Value == "." {
		stream.next() // consume '.'
		tbl.Schema = firstName
		tbl.Name = stream.next().Value
	} else {
		tbl.Name = firstName
	}

	// Optional table alias "as U"
	if strings.EqualFold(stream.peek().Value, "as") {
		stream.next()
		stream.next() // consume alias
	}

	// Optional header settings or mixins before '{'
	if err := stream.expect("{"); err != nil {
		return tbl, err
	}

	for stream.peek().Value != "}" && stream.peek().Type != TokenEOF {
		tokVal := strings.ToLower(stream.peek().Value)

		if tokVal == "mixin" || tokVal == "use" || tokVal == "partial" {
			stream.next() // consume mixin
			partialName := stream.next().Value
			tbl.PartialsUsed = append(tbl.PartialsUsed, partialName)
			continue
		}

		if tokVal == "indexes" {
			indexes, err := parseIndexesBlock(stream)
			if err != nil {
				return tbl, err
			}
			tbl.Indexes = append(tbl.Indexes, indexes...)
			continue
		}

		if tokVal == "note:" {
			stream.next()
			tbl.Note = stream.next().Value
			continue
		}

		col, err := parseColumn(stream)
		if err != nil {
			return tbl, err
		}
		tbl.Columns = append(tbl.Columns, col)
	}

	if stream.peek().Value == "}" {
		stream.next()
	}

	// Optional Note at table end
	if stream.peek().Value == "[" {
		stream.next()
		for stream.peek().Value != "]" && stream.peek().Type != TokenEOF {
			k := stream.next().Value
			if stream.peek().Value == ":" {
				stream.next()
				v := stream.next().Value
				if strings.EqualFold(k, "note") {
					tbl.Note = v
				}
			}
		}
		if stream.peek().Value == "]" {
			stream.next()
		}
	}

	return tbl, nil
}

func parseColumn(stream *tokenStream) (domain.Column, error) {
	col := domain.Column{
		Name:     stream.next().Value,
		Settings: make(map[string]string),
		Nullable: true, // Default nullable unless explicitly not null or pk
	}

	// Column Type
	if stream.peek().Value != "[" && stream.peek().Value != "{" && stream.peek().Value != "}" {
		colType := stream.next().Value
		// Handle schema-qualified type like auth.UserStatus
		if stream.peek().Value == "." {
			stream.next() // consume '.'
			colType += "." + stream.next().Value
		}
		// Handle parametric type like varchar(255)
		if stream.peek().Value == "(" {
			colType += stream.next().Value // '('
			for stream.peek().Value != ")" && stream.peek().Type != TokenEOF {
				colType += stream.next().Value
			}
			if stream.peek().Value == ")" {
				colType += stream.next().Value // ')'
			}
		}
		col.Type = colType
	}

	// Column settings inside [...]
	if stream.peek().Value == "[" {
		stream.next() // consume '['
		for stream.peek().Value != "]" && stream.peek().Type != TokenEOF {
			tok := stream.peek()
			tokVal := strings.ToLower(tok.Value)

			switch tokVal {
			case "pk", "primary key":
				col.PK = true
				col.Nullable = false
				stream.next()

			case "not null":
				col.Nullable = false
				stream.next()

			case "null":
				col.Nullable = true
				stream.next()

			case "unique":
				col.Unique = true
				stream.next()

			case "increment", "autoincrement":
				col.Increment = true
				stream.next()

			case "default":
				stream.next()
				if stream.peek().Value == ":" {
					stream.next()
				}
				col.DefaultVal = stream.next().Value

			case "note":
				stream.next()
				if stream.peek().Value == ":" {
					stream.next()
				}
				col.Note = stream.next().Value

			case "ref":
				stream.next()
				if stream.peek().Value == ":" {
					stream.next()
				}
				refInline := parseRefInline(stream)
				col.Ref = &refInline

			default:
				// Generic setting key or key: value
				key := stream.next().Value
				val := ""
				if stream.peek().Value == ":" {
					stream.next()
					val = stream.next().Value
				}
				col.Settings[key] = val
			}

			if stream.peek().Value == "," {
				stream.next()
			}
		}

		if stream.peek().Value == "]" {
			stream.next()
		}
	}

	return col, nil
}

func parseRefInline(stream *tokenStream) domain.RefInline {
	ref := domain.RefInline{}
	if stream.peek().Value == ">" || stream.peek().Value == "<" || stream.peek().Value == "-" || stream.peek().Value == "<>" {
		ref.Type = stream.next().Value
	}

	// Target e.g. "users.id" or "public.users.id"
	targetParts := make([]string, 0)
	targetParts = append(targetParts, stream.next().Value)
	for stream.peek().Value == "." {
		stream.next() // consume '.'
		targetParts = append(targetParts, stream.next().Value)
	}

	if len(targetParts) == 3 {
		ref.TargetSchema = targetParts[0]
		ref.TargetTable = targetParts[1]
		ref.TargetColumn = targetParts[2]
	} else if len(targetParts) == 2 {
		ref.TargetTable = targetParts[0]
		ref.TargetColumn = targetParts[1]
	}

	return ref
}

func parseIndexesBlock(stream *tokenStream) ([]domain.Index, error) {
	stream.next() // consume 'indexes'
	var indexes []domain.Index

	if err := stream.expect("{"); err != nil {
		return nil, err
	}

	for stream.peek().Value != "}" && stream.peek().Type != TokenEOF {
		idx := domain.Index{Columns: make([]string, 0)}

		if stream.peek().Value == "(" {
			stream.next() // consume '('
			for stream.peek().Value != ")" && stream.peek().Type != TokenEOF {
				idx.Columns = append(idx.Columns, stream.next().Value)
				if stream.peek().Value == "," {
					stream.next()
				}
			}
			if stream.peek().Value == ")" {
				stream.next()
			}
		} else {
			idx.Columns = append(idx.Columns, stream.next().Value)
		}

		// Index settings [name: '...', unique, type: hash]
		if stream.peek().Value == "[" {
			stream.next()
			for stream.peek().Value != "]" && stream.peek().Type != TokenEOF {
				k := strings.ToLower(stream.next().Value)
				if k == "unique" {
					idx.Unique = true
				} else if stream.peek().Value == ":" {
					stream.next()
					v := stream.next().Value
					if k == "name" {
						idx.Name = v
					} else if k == "type" {
						idx.Type = v
					}
				}
				if stream.peek().Value == "," {
					stream.next()
				}
			}
			if stream.peek().Value == "]" {
				stream.next()
			}
		}
		indexes = append(indexes, idx)
	}

	if stream.peek().Value == "}" {
		stream.next()
	}
	return indexes, nil
}

func parseEnum(stream *tokenStream) (domain.Enum, error) {
	stream.next() // consume 'Enum'
	e := domain.Enum{Values: make([]domain.EnumValue, 0)}

	firstName := stream.next().Value
	if stream.peek().Value == "." {
		stream.next() // consume '.'
		e.Schema = firstName
		e.Name = stream.next().Value
	} else {
		e.Name = firstName
	}

	if err := stream.expect("{"); err != nil {
		return e, err
	}

	for stream.peek().Value != "}" && stream.peek().Type != TokenEOF {
		valName := stream.next().Value
		ev := domain.EnumValue{Name: valName}

		if stream.peek().Value == "[" {
			stream.next()
			for stream.peek().Value != "]" && stream.peek().Type != TokenEOF {
				k := stream.next().Value
				if stream.peek().Value == ":" {
					stream.next()
					v := stream.next().Value
					if strings.EqualFold(k, "note") {
						ev.Note = v
					}
				}
			}
			if stream.peek().Value == "]" {
				stream.next()
			}
		}
		e.Values = append(e.Values, ev)
	}

	if stream.peek().Value == "}" {
		stream.next()
	}

	return e, nil
}

func parseTableGroup(stream *tokenStream) (domain.TableGroup, error) {
	stream.next() // consume 'TableGroup'
	tg := domain.TableGroup{
		Name:  stream.next().Value,
		Items: make([]domain.TableGroupItem, 0),
	}

	if err := stream.expect("{"); err != nil {
		return tg, err
	}

	for stream.peek().Value != "}" && stream.peek().Type != TokenEOF {
		firstName := stream.next().Value
		item := domain.TableGroupItem{}
		if stream.peek().Value == "." {
			stream.next()
			item.Schema = firstName
			item.Table = stream.next().Value
		} else {
			item.Table = firstName
		}
		tg.Items = append(tg.Items, item)
	}

	if stream.peek().Value == "}" {
		stream.next()
	}

	return tg, nil
}

func parseRefBlock(stream *tokenStream) ([]domain.Ref, error) {
	stream.next() // consume 'Ref'
	var refs []domain.Ref

	// Optional ref name e.g. "fk_name:"
	refName := ""
	if stream.peek().Value != ":" && stream.peek().Value != "{" {
		refName = stream.next().Value
		if stream.peek().Value == ":" {
			stream.next()
		}
	} else if stream.peek().Value == ":" {
		stream.next()
	}

	if stream.peek().Value == "{" {
		stream.next() // consume '{'
		for stream.peek().Value != "}" && stream.peek().Type != TokenEOF {
			ref, err := parseSingleRef(stream, refName)
			if err != nil {
				return nil, err
			}
			refs = append(refs, ref)
		}
		if stream.peek().Value == "}" {
			stream.next()
		}
	} else {
		ref, err := parseSingleRef(stream, refName)
		if err != nil {
			return nil, err
		}
		refs = append(refs, ref)
	}

	return refs, nil
}

func parseSingleRef(stream *tokenStream, name string) (domain.Ref, error) {
	ref := domain.Ref{Name: name}

	// Source e.g. "posts.user_id" or "public.posts.(user_id, author_id)"
	srcParts := parseRefEndpoint(stream)
	ref.SourceSchema = srcParts.schema
	ref.SourceTable = srcParts.table
	ref.SourceColumns = srcParts.columns

	// Ref operator > , < , - , <>
	if stream.peek().Value == ">" || stream.peek().Value == "<" || stream.peek().Value == "-" || stream.peek().Value == "<>" {
		ref.Type = domain.RefType(stream.next().Value)
	}

	// Target e.g. "users.id"
	tgtParts := parseRefEndpoint(stream)
	ref.TargetSchema = tgtParts.schema
	ref.TargetTable = tgtParts.table
	ref.TargetColumns = tgtParts.columns

	// Optional settings [delete: cascade, update: no action]
	if stream.peek().Value == "[" {
		stream.next()
		for stream.peek().Value != "]" && stream.peek().Type != TokenEOF {
			k := strings.ToLower(stream.next().Value)
			if stream.peek().Value == ":" {
				stream.next()
				v := strings.ToLower(stream.next().Value)
				if k == "delete" {
					ref.OnDelete = v
				} else if k == "update" {
					ref.OnUpdate = v
				}
			}
			if stream.peek().Value == "," {
				stream.next()
			}
		}
		if stream.peek().Value == "]" {
			stream.next()
		}
	}

	return ref, nil
}

type refEndpoint struct {
	schema  string
	table   string
	columns []string
}

func parseRefEndpoint(stream *tokenStream) refEndpoint {
	ep := refEndpoint{columns: make([]string, 0)}
	parts := make([]string, 0)

	for stream.peek().Value != "." && stream.peek().Value != ">" && stream.peek().Value != "<" && stream.peek().Value != "-" && stream.peek().Value != "<>" && stream.peek().Value != "[" && stream.peek().Value != "]" && stream.peek().Value != "}" {
		if stream.peek().Value == "(" {
			stream.next() // consume '('
			for stream.peek().Value != ")" && stream.peek().Type != TokenEOF {
				ep.columns = append(ep.columns, stream.next().Value)
				if stream.peek().Value == "," {
					stream.next()
				}
			}
			if stream.peek().Value == ")" {
				stream.next()
			}
			break
		}

		tokVal := stream.next().Value
		parts = append(parts, tokVal)

		if stream.peek().Value == "." {
			stream.next() // consume '.'
		} else {
			break
		}
	}

	if len(parts) == 3 {
		ep.schema = parts[0]
		ep.table = parts[1]
		if len(ep.columns) == 0 {
			ep.columns = append(ep.columns, parts[2])
		}
	} else if len(parts) == 2 {
		ep.table = parts[0]
		if len(ep.columns) == 0 {
			ep.columns = append(ep.columns, parts[1])
		}
	} else if len(parts) == 1 {
		ep.table = parts[0]
	}

	return ep
}

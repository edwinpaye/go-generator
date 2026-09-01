package dbml

import (
	"unicode"
)

type TokenType int

const (
	TokenEOF TokenType = iota
	TokenError
	TokenIdentifier
	TokenString
	TokenNumber
	TokenSymbol // {, }, [, ], (, ), :, ,, ., =, >, <, -, <>
	TokenKeyword
	TokenComment
)

type Token struct {
	Type  TokenType
	Value string
	Line  int
	Col   int
}

type Lexer struct {
	input []rune
	pos   int
	line  int
	col   int
}

func NewLexer(input string) *Lexer {
	return &Lexer{
		input: []rune(input),
		pos:   0,
		line:  1,
		col:   1,
	}
}

func (l *Lexer) peek() rune {
	if l.pos >= len(l.input) {
		return 0
	}
	return l.input[l.pos]
}

func (l *Lexer) next() rune {
	if l.pos >= len(l.input) {
		return 0
	}
	ch := l.input[l.pos]
	l.pos++
	if ch == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
	return ch
}

func (l *Lexer) Tokenize() []Token {
	var tokens []Token
	for {
		l.skipWhitespaceAndComments()
		if l.peek() == 0 {
			tokens = append(tokens, Token{Type: TokenEOF, Line: l.line, Col: l.col})
			break
		}

		tok := l.nextTok()
		if tok.Type != TokenComment {
			tokens = append(tokens, tok)
		}
	}
	return tokens
}

func (l *Lexer) skipWhitespaceAndComments() {
	for {
		ch := l.peek()
		if ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n' {
			l.next()
			continue
		}

		// Single line comment //
		if ch == '/' && l.peekNext() == '/' {
			l.next() // /
			l.next() // /
			for l.peek() != 0 && l.peek() != '\n' {
				l.next()
			}
			continue
		}

		// Multi-line comment /* ... */
		if ch == '/' && l.peekNext() == '*' {
			l.next() // /
			l.next() // *
			for l.peek() != 0 {
				if l.peek() == '*' && l.peekNext() == '/' {
					l.next()
					l.next()
					break
				}
				l.next()
			}
			continue
		}

		break
	}
}

func (l *Lexer) peekNext() rune {
	if l.pos+1 >= len(l.input) {
		return 0
	}
	return l.input[l.pos+1]
}

func (l *Lexer) nextTok() Token {
	line, col := l.line, l.col
	ch := l.peek()

	// Strings: 'string' or "string" or '''multi line string''' or `string`
	if ch == '\'' || ch == '"' || ch == '`' {
		return l.readString(ch)
	}

	// Double symbol <>
	if ch == '<' && l.peekNext() == '>' {
		l.next()
		l.next()
		return Token{Type: TokenSymbol, Value: "<>", Line: line, Col: col}
	}

	// Symbols
	if isSymbol(ch) {
		l.next()
		return Token{Type: TokenSymbol, Value: string(ch), Line: line, Col: col}
	}

	// Numbers
	if unicode.IsDigit(ch) {
		return l.readNumber()
	}

	// Identifiers or keywords
	if isIdentStart(ch) {
		return l.readIdentifier()
	}

	// Fallback single char token
	l.next()
	return Token{Type: TokenError, Value: string(ch), Line: line, Col: col}
}

func (l *Lexer) readString(quote rune) Token {
	line, col := l.line, l.col
	l.next() // Consume opening quote

	// Check triple quotes '''
	isTriple := false
	if quote == '\'' && l.peek() == '\'' && l.peekNext() == '\'' {
		l.next()
		l.next()
		isTriple = true
	}

	var val []rune
	for {
		ch := l.peek()
		if ch == 0 {
			break
		}
		if isTriple {
			if ch == '\'' && l.peekNext() == '\'' && l.peekAhead(2) == '\'' {
				l.next()
				l.next()
				l.next()
				break
			}
		} else {
			if ch == quote {
				l.next()
				break
			}
			if ch == '\\' {
				l.next() // escape
				ch = l.peek()
			}
		}
		val = append(val, ch)
		l.next()
	}

	return Token{Type: TokenString, Value: string(val), Line: line, Col: col}
}

func (l *Lexer) peekAhead(n int) rune {
	if l.pos+n >= len(l.input) {
		return 0
	}
	return l.input[l.pos+n]
}

func (l *Lexer) readNumber() Token {
	line, col := l.line, l.col
	var val []rune
	for unicode.IsDigit(l.peek()) || l.peek() == '.' {
		val = append(val, l.next())
	}
	return Token{Type: TokenNumber, Value: string(val), Line: line, Col: col}
}

func (l *Lexer) readIdentifier() Token {
	line, col := l.line, l.col
	var val []rune
	for isIdentPart(l.peek()) {
		val = append(val, l.next())
	}
	return Token{Type: TokenIdentifier, Value: string(val), Line: line, Col: col}
}

func isSymbol(ch rune) bool {
	symbols := "{}[];:,.-=><()"
	for _, s := range symbols {
		if ch == s {
			return true
		}
	}
	return false
}

func isIdentStart(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_' || ch == '$' || ch == '@'
}

func isIdentPart(ch rune) bool {
	return unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_' || ch == '-' || ch == '$' || ch == '@'
}

package query

import (
	"strings"
)

// GuardDecision describes the safety verdict for a SQL statement.
type GuardDecision struct {
	Action       string
	NeedsConfirm bool
}

// DestructiveError is returned when a statement requires confirmation.
type DestructiveError struct {
	Decision GuardDecision
}

// Error implements the error interface.
func (e *DestructiveError) Error() string {
	return "destructive " + e.Decision.Action + " statement requires confirmation"
}

// Inspect classifies a SQL statement and decides whether confirmation is needed.
func Inspect(sql string) (*GuardDecision, error) {
	action := classify(sql)
	decision := &GuardDecision{Action: action}
	switch action {
	case "UPDATE", "DELETE":
		decision.NeedsConfirm = !hasWhere(sql)
	case "DROP", "TRUNCATE", "ALTER":
		decision.NeedsConfirm = true
	}
	return decision, nil
}

// classify returns the leading SQL action keyword with comments stripped.
func classify(sql string) string {
	fields := strings.Fields(strings.ToUpper(stripComments(sql)))
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}

// hasWhere reports whether the statement references a WHERE clause.
func hasWhere(sql string) bool {
	for _, field := range strings.Fields(strings.ToUpper(stripComments(sql))) {
		if field == "WHERE" {
			return true
		}
	}
	return false
}

// stripComments removes SQL line and block comments while preserving literals.
func stripComments(sql string) string {
	var b strings.Builder
	b.Grow(len(sql))
	inSingle := false
	inDouble := false
	inLine := false
	inBlock := false
	for i := 0; i < len(sql); i++ {
		c := sql[i]
		next := byte(0)
		if i+1 < len(sql) {
			next = sql[i+1]
		}
		switch {
		case inLine:
			if c == '\n' {
				inLine = false
				b.WriteByte(c)
			}
		case inBlock:
			if c == '*' && next == '/' {
				inBlock = false
				i++
			}
		case inSingle:
			b.WriteByte(c)
			if c == '\'' {
				inSingle = false
			}
		case inDouble:
			b.WriteByte(c)
			if c == '"' {
				inDouble = false
			}
		default:
			if c == '-' && next == '-' {
				inLine = true
				continue
			}
			if c == '/' && next == '*' {
				inBlock = true
				i++
				continue
			}
			if c == '\'' {
				inSingle = true
			}
			if c == '"' {
				inDouble = true
			}
			b.WriteByte(c)
		}
	}
	return b.String()
}

package migration

import "strings"

func splitStatements(sql string) []string {
	var statements []string
	var current strings.Builder

	lines := strings.Split(sql, "\n")

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		if trimmedLine == "" {
			continue
		}

		if strings.HasPrefix(trimmedLine, "--") {
			continue
		}

		if current.Len() > 0 {
			current.WriteString("\n")
		}
		current.WriteString(line)

		if strings.HasSuffix(trimmedLine, ";") {
			stmt := strings.TrimSpace(current.String())
			if stmt != "" && !strings.HasPrefix(stmt, "--") {
				statements = append(statements, stmt)
			}
			current.Reset()
		}
	}

	if current.Len() > 0 {
		stmt := strings.TrimSpace(current.String())
		if stmt != "" && !strings.HasPrefix(stmt, "--") {
			statements = append(statements, stmt)
		}
	}

	return statements
}

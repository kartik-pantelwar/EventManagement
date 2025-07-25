
package sqlparser

func ParseSqlFile(fileContent string) []string {
	var result []string
	content := trimExtraLines(trimSqlComments(fileContent))
	temp := ""
	for i := 0; i < len(content); i++ {
		char := string(content[i])
		if char == ";" {
			result = append(result, temp)
			temp = ""
		} else {
			temp += char
		}
	}
	// Technically if we reached here it means our sql file is not correct since
	// last statement does not end with semicolon but we will still include residue
	// statement so users can be made aware of mistake during migrations
	if temp != "" {
		result = append(result, temp)
	}
	return result
}

func trimSqlComments(file string) string {
	var result string
	inCommentZone := false
	for i := 0; i < len(file)-1; i++ {
		char1 := string(file[i])
		char2 := string(file[i+1])
		if char1 == "-" && char2 == "-" {
			i += 1
			inCommentZone = true
		} else {
			if !inCommentZone {
				result += char1
			} else {
				if char1 == "\n" {
					inCommentZone = false
					result += char1
				}
			}
		}
	}
	return result
}

func trimExtraLines(file string) string {
	var result string
	isEscape := false
	for i := 0; i < len(file); i++ {
		char := string(file[i])
		if char != "\n" {
			if isEscape {
				// This is so that in case a file starts with extra lines we do not necessrily
				// add another extra line
				if result != "" {
					result += "\n"
				}
				isEscape = false
			}
			result += char
		} else {
			isEscape = true
		}
	}
	return result
}

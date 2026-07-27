package main

import "os"

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		RunParser()
		return
	}

	if args[0] == "-cli" {
		cliArgs := args[1:]

		// The bash loop passes the whole command line as one quoted
		// argument (`-cli "$cmd"`), so split it into words ourselves.
		if len(cliArgs) == 1 {
			if split, err := shellSplit(cliArgs[0]); err == nil {
				cliArgs = split
			}
		}

		runCLI(cliArgs)
		return
	}

	// Unrecognized invocation - fall back to the parser.
	RunParser()
}

// shellSplit splits a string into words the way Python's shlex.split()
// does: whitespace-separated, honoring single and double quotes and
// backslash escapes.
func shellSplit(s string) ([]string, error) {
	var words []string
	var cur []rune
	inWord := false

	runes := []rune(s)
	i := 0
	for i < len(runes) {
		c := runes[i]
		switch {
		case c == ' ' || c == '\t' || c == '\n':
			if inWord {
				words = append(words, string(cur))
				cur = cur[:0]
				inWord = false
			}
			i++
		case c == '\'':
			inWord = true
			i++
			for i < len(runes) && runes[i] != '\'' {
				cur = append(cur, runes[i])
				i++
			}
			if i >= len(runes) {
				return nil, errUnterminatedQuote
			}
			i++ // skip closing quote
		case c == '"':
			inWord = true
			i++
			for i < len(runes) && runes[i] != '"' {
				if runes[i] == '\\' && i+1 < len(runes) && (runes[i+1] == '"' || runes[i+1] == '\\') {
					cur = append(cur, runes[i+1])
					i += 2
					continue
				}
				cur = append(cur, runes[i])
				i++
			}
			if i >= len(runes) {
				return nil, errUnterminatedQuote
			}
			i++ // skip closing quote
		case c == '\\' && i+1 < len(runes):
			inWord = true
			cur = append(cur, runes[i+1])
			i += 2
		default:
			inWord = true
			cur = append(cur, c)
			i++
		}
	}
	if inWord {
		words = append(words, string(cur))
	}
	return words, nil
}

type shellSplitError string

func (e shellSplitError) Error() string { return string(e) }

const errUnterminatedQuote = shellSplitError("unterminated quote")
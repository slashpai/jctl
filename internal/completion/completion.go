package completion

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/alecthomas/kong"
)

// Handle serves shell completion requests triggered via COMP_LINE.
// Returns true when a completion request was handled.
func Handle(k *kong.Kong) bool {
	line := os.Getenv("COMP_LINE")
	if line == "" {
		return false
	}

	point := len(line)
	if raw := os.Getenv("COMP_POINT"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n >= 0 && n <= len(line) {
			point = n
		}
	}

	for _, candidate := range Predict(k.Model.Node, line, point) {
		fmt.Fprintln(k.Stdout, candidate)
	}
	return true
}

// Predict returns completion candidates for the given command line and cursor.
func Predict(root *kong.Node, line string, point int) []string {
	if root == nil {
		return nil
	}

	prefix := wordAt(line, point)
	if point > len(line) {
		point = len(line)
	}
	tokens := tokenize(line[:point])
	if len(tokens) == 0 {
		return nil
	}
	tokens = tokens[1:] // drop binary name

	node, ctx := resolveContext(root, tokens, prefix)
	switch ctx.kind {
	case kindCommand:
		return filterPrefix(commandCandidates(node), prefix)
	case kindPositional:
		return filterPrefix(ctx.enumValues, prefix)
	case kindFlag:
		return filterPrefix(flagCandidates(node, true), prefix)
	case kindFlagValue:
		if len(ctx.enumValues) > 0 {
			return filterPrefix(ctx.enumValues, prefix)
		}
		return nil
	default:
		return nil
	}
}

type completionKind int

const (
	kindCommand completionKind = iota
	kindFlag
	kindFlagValue
	kindPositional
)

type completionContext struct {
	kind       completionKind
	enumValues []string
}

func resolveContext(root *kong.Node, tokens []string, prefix string) (*kong.Node, completionContext) {
	node := root
	i := 0

	for i < len(tokens) {
		token := tokens[i]
		if strings.HasPrefix(token, "-") {
			break
		}
		child := matchChild(node, token)
		if child == nil {
			break
		}
		node = child
		i++
	}

	if i == len(tokens) {
		return node, leafContinuationContext(node, 0, prefix)
	}

	flags := allFlags(node)
	positionalIndex := 0
	for i < len(tokens) {
		token := tokens[i]

		if !strings.HasPrefix(token, "-") {
			if i == len(tokens)-1 {
				return node, leafContinuationContext(node, positionalIndex, prefix)
			}
			positionalIndex++
			i++
			continue
		}

		name, _, hasValue := splitFlagToken(token)
		flag := matchFlag(flags, name)
		if flag == nil {
			if i == len(tokens)-1 {
				return node, completionContext{kind: kindFlag}
			}
			break
		}

		if hasValue {
			i++
			continue
		}

		if flag.Value.IsBool() {
			i++
			continue
		}

		if i+1 < len(tokens) {
			i += 2
			continue
		}

		return node, completionContext{
			kind:       kindFlagValue,
			enumValues: enumValues(flag),
		}
	}

	if len(tokens) > 0 && strings.HasPrefix(tokens[len(tokens)-1], "-") {
		return node, completionContext{kind: kindFlag}
	}

	return node, leafContinuationContext(node, positionalIndex, prefix)
}

func leafContinuationContext(node *kong.Node, positionalIndex int, prefix string) completionContext {
	if values := positionalEnumValues(node, positionalIndex); len(values) > 0 {
		return completionContext{kind: kindPositional, enumValues: values}
	}
	if len(node.Children) > 0 {
		return completionContext{kind: kindCommand}
	}
	if prefix == "" && hasFlags(node) {
		return completionContext{kind: kindFlag}
	}
	if positionalIndex < len(node.Positional) && prefix != "" {
		return completionContext{kind: kindPositional}
	}
	if hasFlags(node) {
		return completionContext{kind: kindFlag}
	}
	return completionContext{kind: kindCommand}
}

func hasFlags(node *kong.Node) bool {
	for _, flag := range allFlags(node) {
		if flag != nil && !flag.Hidden {
			return true
		}
	}
	return false
}

func commandCandidates(node *kong.Node) []string {
	if node == nil {
		return nil
	}

	seen := make(map[string]struct{})
	var out []string
	for _, child := range node.Children {
		if child == nil || child.Hidden {
			continue
		}
		addCandidate(&out, seen, child.Name)
		for _, alias := range child.Aliases {
			addCandidate(&out, seen, alias)
		}
	}
	return out
}

func positionalEnumValues(node *kong.Node, index int) []string {
	if node == nil || index >= len(node.Positional) {
		return nil
	}
	positional := node.Positional[index]
	if positional == nil || positional.Tag.Hidden || positional.Enum == "" {
		return nil
	}
	return enumValuesFor(positional)
}

func enumValuesFor(value *kong.Value) []string {
	if value == nil || value.Enum == "" {
		return nil
	}
	values := make([]string, 0, len(value.EnumMap()))
	for candidate := range value.EnumMap() {
		if candidate != "" {
			values = append(values, candidate)
		}
	}
	return values
}

func flagCandidates(node *kong.Node, includeShort bool) []string {
	seen := make(map[string]struct{})
	var out []string
	for _, flag := range allFlags(node) {
		if flag == nil || flag.Hidden {
			continue
		}
		addCandidate(&out, seen, "--"+flag.Name)
		for _, alias := range flag.Aliases {
			addCandidate(&out, seen, "--"+alias)
		}
		if includeShort && flag.Short != 0 {
			addCandidate(&out, seen, "-"+string(flag.Short))
		}
	}
	return out
}

func allFlags(node *kong.Node) []*kong.Flag {
	var out []*kong.Flag
	for _, group := range node.AllFlags(true) {
		out = append(out, group...)
	}
	return out
}

func matchChild(node *kong.Node, name string) *kong.Node {
	for _, child := range node.Children {
		if child == nil || child.Hidden {
			continue
		}
		if child.Name == name {
			return child
		}
		for _, alias := range child.Aliases {
			if alias == name {
				return child
			}
		}
	}
	return nil
}

func matchFlag(flags []*kong.Flag, name string) *kong.Flag {
	name = strings.TrimLeft(name, "-")
	if len(name) == 1 {
		for _, flag := range flags {
			if flag.Short == rune(name[0]) {
				return flag
			}
		}
	}
	for _, flag := range flags {
		if flag.Name == name {
			return flag
		}
		for _, alias := range flag.Aliases {
			if alias == name {
				return flag
			}
		}
	}
	return nil
}

func splitFlagToken(token string) (name string, value string, hasValue bool) {
	if before, after, ok := strings.Cut(token, "="); ok {
		return before, after, true
	}
	return token, "", false
}

func enumValues(flag *kong.Flag) []string {
	if flag == nil {
		return nil
	}
	return enumValuesFor(flag.Value)
}

func addCandidate(out *[]string, seen map[string]struct{}, candidate string) {
	if candidate == "" {
		return
	}
	if _, ok := seen[candidate]; ok {
		return
	}
	seen[candidate] = struct{}{}
	*out = append(*out, candidate)
}

func filterPrefix(candidates []string, prefix string) []string {
	if prefix == "" {
		return candidates
	}
	var out []string
	for _, candidate := range candidates {
		if strings.HasPrefix(candidate, prefix) {
			out = append(out, candidate)
		}
	}
	return out
}

func wordAt(line string, point int) string {
	if point > len(line) {
		point = len(line)
	}
	start := point
	for start > 0 && line[start-1] != ' ' {
		start--
	}
	return line[start:point]
}

func tokenize(line string) []string {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil
	}
	return strings.Fields(line)
}

package jira

import (
	"encoding/json"
	"fmt"
	"strings"
)

type adfNode struct {
	Type    string            `json:"type"`
	Content []adfNode         `json:"content"`
	Text    string            `json:"text"`
	Attrs   map[string]any    `json:"attrs"`
	Marks   []adfMark         `json:"marks"`
}

type adfMark struct {
	Type  string         `json:"type"`
	Attrs map[string]any `json:"attrs"`
}

func RenderADF(rawJSON string) string {
	if rawJSON == "" {
		return ""
	}

	var doc adfNode
	if err := json.Unmarshal([]byte(rawJSON), &doc); err != nil {
		return rawJSON
	}

	var sb strings.Builder
	renderBlock(&sb, doc, 0)
	return strings.TrimSpace(sb.String())
}

func renderBlock(sb *strings.Builder, node adfNode, depth int) {
	switch node.Type {
	case "doc":
		for _, child := range node.Content {
			renderBlock(sb, child, depth)
		}

	case "paragraph":
		for _, child := range node.Content {
			renderInline(sb, child)
		}
		sb.WriteString("\n\n")

	case "heading":
		level := 1
		if v, ok := node.Attrs["level"]; ok {
			switch lv := v.(type) {
			case float64:
				level = int(lv)
			case int:
				level = lv
			}
		}
		prefix := strings.Repeat("#", level) + " "
		sb.WriteString(prefix)
		for _, child := range node.Content {
			renderInline(sb, child)
		}
		sb.WriteString("\n\n")

	case "bulletList":
		for _, item := range node.Content {
			sb.WriteString("• ")
			for _, child := range item.Content {
				renderBlock(sb, child, depth+1)
			}
		}

	case "orderedList":
		for i, item := range node.Content {
			fmt.Fprintf(sb, "%d. ", i+1)
			for _, child := range item.Content {
				renderBlock(sb, child, depth+1)
			}
		}

	case "listItem":
		for _, child := range node.Content {
			renderBlock(sb, child, depth)
		}

	case "codeBlock":
		sb.WriteString("```\n")
		for _, child := range node.Content {
			renderInline(sb, child)
		}
		sb.WriteString("\n```\n\n")

	case "blockquote":
		var inner strings.Builder
		for _, child := range node.Content {
			renderBlock(&inner, child, depth)
		}
		for _, line := range strings.Split(strings.TrimSpace(inner.String()), "\n") {
			sb.WriteString("> ")
			sb.WriteString(line)
			sb.WriteString("\n")
		}
		sb.WriteString("\n")

	case "rule":
		sb.WriteString("────────────────────────────────────────\n\n")

	default:
		for _, child := range node.Content {
			renderBlock(sb, child, depth)
		}
	}
}

func renderInline(sb *strings.Builder, node adfNode) {
	switch node.Type {
	case "text":
		sb.WriteString(node.Text)

	case "hardBreak":
		sb.WriteString("\n")

	case "mention":
		name := ""
		if v, ok := node.Attrs["text"]; ok {
			name, _ = v.(string)
		}
		if name == "" {
			name = "Unknown"
		}
		sb.WriteString("@")
		sb.WriteString(name)

	case "emoji":
		text := ""
		if v, ok := node.Attrs["text"]; ok {
			text, _ = v.(string)
		}
		if text != "" {
			sb.WriteString(text)
		}

	case "inlineCard", "link":
		url := ""
		if v, ok := node.Attrs["url"]; ok {
			url, _ = v.(string)
		}
		if url != "" {
			sb.WriteString(url)
		} else {
			for _, child := range node.Content {
				renderInline(sb, child)
			}
		}
	}
}

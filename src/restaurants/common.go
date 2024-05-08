package restaurants

import (
	"bytes"

	"golang.org/x/net/html"
)

const (
	Pizza = iota
	Lindy
	Kozel
	Erika
)

type Menu struct {
	Restaurant int
	Meals      *[]Meal
}

type Meal struct {
	Name   string
	Price  string
	Dishes []string
}

func collectHTMLText(node *html.Node, buffer *bytes.Buffer) {
	if node.Type == html.TextNode {
		buffer.WriteString(node.Data)
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		collectHTMLText(child, buffer)
	}
}

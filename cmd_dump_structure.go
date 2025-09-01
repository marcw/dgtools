package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/charmbracelet/lipgloss/tree"
	"github.com/marcw/dgtools/internal/discogs"
	"github.com/urfave/cli/v3"
)

// node represents a node in the XML structure tree
type node struct {
	Name     string
	Count    int
	Children map[string]*node
	Attrs    map[string]int // attribute name -> count
}

func (n *node) String() string {
	attrString := ""
	if len(n.Attrs) > 0 {
		attrs := make([]string, 0)
		for k := range n.Attrs {
			attrs = append(attrs, k)
		}
		sort.Strings(attrs)
		attrString = fmt.Sprintf(" [%s]", strings.Join(attrs, ", "))
	}

	return fmt.Sprintf("%s%s", n.Name, attrString)
}

// xmlStructure holds the root of our structure tree
type xmlStructure struct {
	root *node
}

// structureAnalyzer handles the parsing and analysis
type structureAnalyzer struct {
	structure *xmlStructure
	stack     []*node // current path in the tree
}

// NewstructureAnalyzer creates a new analyzer
func NewstructureAnalyzer() *structureAnalyzer {
	return &structureAnalyzer{
		structure: &xmlStructure{
			root: &node{
				Name:     "root",
				Count:    0,
				Children: make(map[string]*node),
				Attrs:    make(map[string]int),
			},
		},
		stack: make([]*node, 0),
	}
}

// ParseXML parses the XML from a reader and builds the structure
func (s *structureAnalyzer) ParseXML(reader io.Reader) error {
	decoder := xml.NewDecoder(reader)
	s.stack = append(s.stack, s.structure.root)

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		switch se := token.(type) {
		case xml.StartElement:
			s.processStartElement(se)
		case xml.EndElement:
			s.processEndElement(se)
		}
	}

	return nil
}

// processStartElement handles XML start elements
func (s *structureAnalyzer) processStartElement(se xml.StartElement) {
	currentParent := s.stack[len(s.stack)-1]
	elementName := se.Name.Local

	// Get or create child node
	var childnode *node
	if existing, exists := currentParent.Children[elementName]; exists {
		childnode = existing
	} else {
		childnode = &node{
			Name:     elementName,
			Count:    0,
			Children: make(map[string]*node),
			Attrs:    make(map[string]int),
		}
		currentParent.Children[elementName] = childnode
	}

	// Increment count for this element
	childnode.Count++

	// Process attributes
	for _, attr := range se.Attr {
		attrName := attr.Name.Local
		childnode.Attrs[attrName]++
	}

	// Push to stack
	s.stack = append(s.stack, childnode)
}

// processEndElement handles XML end elements
func (s *structureAnalyzer) processEndElement(se xml.EndElement) {
	if len(s.stack) > 1 {
		s.stack = s.stack[:len(s.stack)-1]
	}
}

func (s *structureAnalyzer) toTree(node *node) *tree.Tree {
	tree := tree.Root(node.String())

	childNames := make([]string, 0)
	for name := range node.Children {
		childNames = append(childNames, name)
	}
	sort.Strings(childNames)
	for name := range node.Children {
		child := node.Children[name]
		if len(child.Children) > 0 {
			tree.Child(s.toTree(child))
		} else {
			tree.Child(child.String())
		}
	}

	return tree
}

func (s *structureAnalyzer) ToTree() *tree.Tree {
	return s.toTree(s.structure.root)
}

var discogsDumpStructureCmd = &cli.Command{
	Name:  "structure",
	Usage: "Dump the structure of the database",
	Arguments: []cli.Argument{
		&cli.StringArg{
			Name:      "file",
			UsageText: "The file to dump the structure of",
		},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		if cmd.StringArg("file") == "" {
			return fmt.Errorf("file is required")
		}

		analyzer := NewstructureAnalyzer()
		dd, err := discogs.OpenDumpFile(cmd.StringArg("file"))
		if err != nil {
			return err
		}
		defer dd.Close()

		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Start()
		s.Suffix = " Parsing XML..."
		if err := analyzer.ParseXML(dd); err != nil {
			return err
		}
		s.Stop()

		tree := analyzer.ToTree()
		fmt.Println(tree)
		return nil
	},
}

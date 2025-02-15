package view

import (
	"bytes"
)

// NodeSet Table结构体集合
type Tables struct {
	tables []Node
	NodeSetBase
}

func (t *Tables) getList() []Node {
	return t.tables
}

func (t *Tables) ToString() string {
	buf := bytes.Buffer{}
	t.WriteTo(&buf)
	return buf.String()
}

func (t *Tables) WriteTo(buf *bytes.Buffer) {
	for i, table := range t.tables {
		switch table.(type) {
		case *Table:
			table.WriteTo(buf)
		default:
			buf.WriteString("(")
			table.WriteTo(buf)
			buf.WriteString(")")
		}
		if i < len(t.tables)-1 {
			buf.WriteString(", ")
		}
	}
}

func (t *Tables) IsNull() bool {
	if len(t.tables) < 1 {
		return true
	} else {
		return false
	}
}

func (t *Tables) Append(g Node) {
	t.tables = append(t.tables, g)
}

type Table struct {
	NodeBase
	Value string
}

func (t *Table) ToString() string {
	return t.Value
}

func (t *Table) WriteTo(buf *bytes.Buffer) {
	buf.WriteString(t.Value)
}

package nodes

import wf "github.com/Rithvik-C18/go-flow/internal/workflow"

type BaseNode struct {
	Id   string
	Name string
}

type Node interface {
	Execute(wf.IO_Flow)
}

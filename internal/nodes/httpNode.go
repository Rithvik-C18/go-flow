package nodes

import (
	"encoding/json"
	"fmt"

	"github.com/go-resty/resty/v2"

	wf "github.com/Rithvik-C18/go-flow/internal/workflow"
)

type HttpNode struct {
	BaseNode
	url        string
	httpMethod string
}

func (h HttpNode) Execute(io *wf.IO_Flow) {

	client := resty.New()

	switch h.httpMethod {
	case "GET":
		resp, err := client.R().Get(h.url)
		if err != nil {
			io.Logs = err.Error()
		}

		var result map[string]any

		err = json.Unmarshal([]byte(resp.Body()), &result)
		if err != nil {
			fmt.Println("Error parsing JSON:", err)
			return
		}
		io.Output = result
	}

}

func main() {
	http := HttpNode{
		url:        "https://example.com",
		httpMethod: "GET",
	}
	io := wf.IO_Flow{}
	http.Execute(&io)

	// fmt.Println(IO_Flow.output)
}

// N nodes
// DAG graph
// topological sort of the graph
// for loop:

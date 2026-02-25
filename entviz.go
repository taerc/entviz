// Package entviz 为 Facebook 的 Ent ORM 框架提供了一个扩展，用于生成
// schema 图的交互式 HTML 可视化。
//
// 该包与 Ent 的代码生成管道集成，自动创建一个静态 HTML 文件（schema-viz.html），
// 使用 vis-network.js 将数据库 schema 显示为交互式网络图。可视化内容包括：
//   - 实体类型作为节点，包含其字段信息
//   - 实体之间的关系作为带标签的边
//   - 自引用关系使用曲线箭头显示
//
// 使用方法：
//  1. 将扩展添加到 entc 配置中：
//     entc.Generate("./ent", entc.Extensions(&entviz.Extension{}))
//  2. 运行代码生成：go generate ./ent
//  3. 在浏览器中打开生成的 schema-viz.html
//
// 或者使用 GeneratePage() 以编程方式生成 HTML。
package entviz

import (
	"bytes"
	"embed"
	_ "embed"
	"encoding/json"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"

	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
)

type (
	// jsGraph 表示用于可视化的 JSON 可序列化图结构。
	// 包含所有节点（实体）和边（关系），将由 JavaScript 可视化库渲染。
	jsGraph struct {
		Nodes []jsNode `json:"nodes"`
		Edges []jsEdge `json:"edges"`
	}

	// jsNode 表示 schema 图中的单个实体。
	// 每个节点对应一个 Ent 类型，包含其字段定义。
	jsNode struct {
		ID     string    `json:"id"`
		Fields []jsField `json:"fields"`
	}

	// jsEdge 表示 schema 中两个实体之间的关系。
	// 边是有向的，并带有关系名称标签。
	jsEdge struct {
		From  string `json:"from"`
		To    string `json:"to"`
		Label string `json:"label"`
	}

	// jsField 表示实体中的单个字段定义。
	// 包含字段名称和类型，用于在可视化中显示。
	jsField struct {
		Name    string `json:"name"`
		Type    string `json:"type"`
		Comment string `json:"comment"`
	}
)

// toJsGraph 将 Ent 的内部图表示转换为 JSON 可序列化结构。
// 它通过以下方式将 Ent 的 gen.Graph 转换为 jsGraph：
//   - 提取每个节点（实体）及其字段
//   - 为关系创建边，跳过反向边以避免重复
//   - 保留实体名称作为节点 ID，关系名称作为边标签
//
// 参数：
//   - g: 包含 schema 信息的 Ent 生成图
//
// 返回：
//   - jsGraph: 适合 JSON 序列化和可视化的简化图结构
func toJsGraph(g *gen.Graph) jsGraph {
	graph := jsGraph{}
	for _, n := range g.Nodes {
		node := jsNode{ID: n.Name}
		for _, f := range n.Fields {
			node.Fields = append(node.Fields, jsField{
				Name:    f.Name,
				Type:    f.Type.String(),
				Comment: f.Comment(),
			})
		}
		graph.Nodes = append(graph.Nodes, node)
		for _, e := range n.Edges {
			if e.IsInverse() {
				continue
			}
			graph.Edges = append(graph.Edges, jsEdge{
				From:  n.Name,
				To:    e.Type.Name,
				Label: e.Name,
			})
		}

	}
	return graph
}

var (
	//go:embed viz.tmpl
	tmplhtml string
	//go:embed entviz.go.tmpl
	tmplfile string
	//go:embed assets
	assets embed.FS
	viztmpl  = template.Must(template.New("viz").Parse(tmplhtml))
)

type templateData struct {
	FiraCodeCSS    template.CSS
	VisNetworkJS   template.JS
	RandomColorJS  template.JS
	GraphJSON      template.JS
}

// generateHTML 生成包含 schema 可视化的完整 HTML 页面。
// 该函数执行以下步骤：
//   1. 将 Ent 图转换为 JSON 可序列化格式
//   2. 将图数据序列化为 JSON 字符串
//   3. 使用预定义的 HTML 模板（viz.tmpl）渲染最终页面
//
// 参数：
//   - g: 包含 schema 信息的 Ent 生成图
//
// 返回：
//   - []byte: 生成的 HTML 页面字节数组
//   - error: 如果生成过程中发生错误则返回错误
func generateHTML(g *gen.Graph) ([]byte, error) {
	firaCodeCSS, err := fs.ReadFile(assets, "assets/fira_code.css")
	if err != nil {
		return nil, err
	}
	visNetworkJS, err := fs.ReadFile(assets, "assets/vis-network.min.js")
	if err != nil {
		return nil, err
	}
	randomColorJS, err := fs.ReadFile(assets, "assets/randomcolor.min.js")
	if err != nil {
		return nil, err
	}

	graph := toJsGraph(g)
	graphJSON, err := json.Marshal(&graph)
	if err != nil {
		return nil, err
	}

	data := templateData{
		FiraCodeCSS:   template.CSS(firaCodeCSS),
		VisNetworkJS:  template.JS(visNetworkJS),
		RandomColorJS: template.JS(randomColorJS),
		GraphJSON:     template.JS(graphJSON),
	}

	var b bytes.Buffer
	if err := viztmpl.Execute(&b, data); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// VisualizeSchema 是一个 Ent 钩子，用于生成可视化 schema 图的静态 HTML 页面。
// 该钩子在 Ent 代码生成流程中运行：
//   1. 首先调用下一个生成器完成标准代码生成
//   2. 然后生成 schema 可视化 HTML
//   3. 将 HTML 文件写入目标目录（默认为 ent/schema-viz.html）
//
// 参数：
//   - next: 下一个生成器，用于完成标准代码生成
//
// 返回：
//   - gen.Generator: 包装后的生成器，会在标准生成后添加可视化生成步骤
func VisualizeSchema(next gen.Generator) gen.Generator {
	return gen.GenerateFunc(func(g *gen.Graph) error {
		if err := next.Generate(g); err != nil {
			return err
		}
		buf, err := generateHTML(g)
		if err != nil {
			return err
		}
		path := filepath.Join(g.Config.Target, "schema-viz.html")
		return os.WriteFile(path, buf, 0644)
	})
}

// Extension 是 Ent 代码生成器的扩展，用于集成 schema 可视化功能。
// 该扩展实现了 entc.Extension 接口，通过提供钩子和模板来扩展 Ent 的代码生成流程。
//
// 使用方法：
//   entc.Generate("./ent", entc.Extensions(&entviz.Extension{}))
type Extension struct {
	entc.DefaultExtension
}

// Hooks 返回在代码生成过程中执行的钩子列表。
// 该方法返回 VisualizeSchema 钩子，该钩子会在标准代码生成完成后
// 自动生成 schema 可视化的 HTML 页面。
//
// 返回：
//   - []gen.Hook: 包含 VisualizeSchema 钩子的列表
func (Extension) Hooks() []gen.Hook {
	return []gen.Hook{
		VisualizeSchema,
	}
}

// Templates 返回代码生成过程中使用的模板列表。
// 该方法返回 entviz.go.tmpl 模板，该模板用于生成辅助代码。
//
// 返回：
//   - []*gen.Template: 包含 entviz 模板的列表
func (Extension) Templates() []*gen.Template {
	return []*gen.Template{
		gen.MustParse(gen.NewTemplate("entviz").Parse(tmplfile)),
	}
}

// GeneratePage 从指定的 schema 路径生成可视化 HTML 页面。
// 该函数用于独立于代码生成流程之外生成可视化页面，适用于：
//   - 命令行工具（如 entviz 命令）
//   - 运行时动态生成可视化
//   - 测试和调试目的
//
// 参数：
//   - schemaPath: Ent schema 文件所在的目录路径
//   - cfg: Ent 代码生成配置，如果为 nil 则使用默认配置
//
// 返回：
//   - []byte: 生成的 HTML 页面字节数组
//   - error: 如果加载 schema 或生成 HTML 时发生错误则返回错误
func GeneratePage(schemaPath string, cfg *gen.Config) ([]byte, error) {
	g, err := entc.LoadGraph(schemaPath, cfg)
	if err != nil {
		return nil, err
	}
	return generateHTML(g)
}

package datasources

import (
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

func GetTableWriter(tableWidth int) table.Writer {
    t := table.NewWriter()
	t.SetStyle(table.StyleLight)

	if(tableWidth > 0) {
		t.Style().Size = table.SizeOptions{
			WidthMin: tableWidth,
			WidthMax: tableWidth,
		}
	}
	return t
}

func RenderTable(t table.Writer, title string) string {
    if (t.Length() == 0) {
        t.AppendRow([]interface{}{title})
        t.SetColumnConfigs([]table.ColumnConfig{
            {Number: 1, Align: text.AlignLeft},
        })
    } else {
        t.SetTitle(title)
    }
    return t.Render()
}
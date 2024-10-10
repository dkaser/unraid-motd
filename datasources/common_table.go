package datasources

import (
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

type TableConfig interface {
    GetBorder() bool
	GetTableWidth() int
}

func GetTableWriter(c TableConfig) table.Writer {
    t := table.NewWriter()
	t.SetStyle(table.StyleLight)
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, Align: text.AlignRight},
	})

	if(c.GetTableWidth() > 0) {
		t.Style().Size = table.SizeOptions{
			WidthMin: c.GetTableWidth(),
			WidthMax: c.GetTableWidth(),
		}
	}

	t.Style().Options.SeparateColumns = c.GetBorder()
	t.Style().Options.DrawBorder = c.GetBorder()

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
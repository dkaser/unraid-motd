package datasources

import (
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

type TableConfig interface {
	GetBorder() bool
	GetTableWidth() int
}

func GetTableWriter(tableConf TableConfig) table.Writer {
	outputTable := table.NewWriter()
	outputTable.SetStyle(table.StyleLight)
	outputTable.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, Align: text.AlignRight},
	})

	if tableConf.GetTableWidth() > 0 {
		outputTable.Style().Size = table.SizeOptions{
			WidthMin: tableConf.GetTableWidth(),
			WidthMax: tableConf.GetTableWidth(),
		}
	}

	outputTable.Style().Options.SeparateColumns = tableConf.GetBorder()
	outputTable.Style().Options.DrawBorder = tableConf.GetBorder()

	return outputTable
}

func RenderTable(outputTable table.Writer, title string) string {
	if outputTable.Length() == 0 {
		outputTable.AppendRow([]interface{}{title})
		outputTable.SetColumnConfigs([]table.ColumnConfig{
			{Number: 1, Align: text.AlignLeft},
		})
	} else {
		outputTable.SetTitle(title)
	}

	return outputTable.Render()
}

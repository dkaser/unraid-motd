package datasources

import (
	"github.com/jedib0t/go-pretty/v6/table"
    "github.com/jedib0t/go-pretty/v6/text"
  )

func GetTableStyle() (table.Style) {
	return table.Style{
        Name: "myNewStyle",
        Box: table.BoxStyle{
            BottomLeft:       "└",
            BottomRight:      "┘",
            BottomSeparator:  "┴",
            EmptySeparator:   text.RepeatAndTrim(" ", text.RuneWidthWithoutEscSequences("┼")),
            Left:             "│",
            LeftSeparator:    "├",
            MiddleHorizontal: "─",
            MiddleSeparator:  "┼",
            MiddleVertical:   "│",
            PaddingLeft:      " ",
            PaddingRight:     " ",
            PageSeparator:    "\n",
            Right:            "│",
            RightSeparator:   "┤",
            TopLeft:          "┌",
            TopRight:         "┐",
            TopSeparator:     "┬",
            UnfinishedRow:    " ≈",
        },
        Options: table.Options{
            DrawBorder:      true,
            SeparateColumns: true,
            SeparateFooter:  true,
            SeparateHeader:  true,
            SeparateRows:    true,
        },
    }
}
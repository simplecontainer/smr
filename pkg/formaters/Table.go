package formaters

import (
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
)

func SetStyle(table *tablewriter.Table) {
	table.Configure(func(config *tablewriter.Config) {
		config.Row.Padding.Global.Left = tw.PaddingNone.Left
		config.Row.Formatting.AutoWrap = tw.WrapTruncate
		config.Row.ColMaxWidths.Global = 60
		config.Header.Padding.Global.Left = tw.PaddingNone.Left
		config.Header.Formatting.AutoWrap = tw.WrapTruncate
		config.Header.Alignment = tw.CellAlignment{
			Global: tw.AlignLeft,
		}
	})

	table.Options(tablewriter.WithRendition(tw.Rendition{
		Borders: tw.BorderNone,
		Symbols: tw.NewSymbols(tw.StyleDefault),
		Settings: tw.Settings{
			Separators: tw.Separators{BetweenColumns: tw.Off, BetweenRows: tw.Off},
		},
	}))

	// Optional: Customize alignment and padding table.SetAlignment(tablewriter.ALIGN_LEFT)
	//table.SetCenterSeparator("")
	//table.SetColumnSeparator("")
	//table.SetRowSeparator("")
	//table.SetHeaderLine(false)
	//table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	//table.SetTablePadding("  ") // Two spaces padding between columns
}

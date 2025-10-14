package main

import (
	"fmt"
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type Row struct {
	Barcode  *widget.Entry
	SampleID *widget.Entry
}

func main() {
	a := app.New()
	w := a.NewWindow("Barcode Table")
	w.Resize(fyne.NewSize(400, 300))

	// Create the initial 3 rows
	rows := make([]*Row, 0)

	tableContainer := container.NewVBox()

	// Add header
	header := container.NewGridWithColumns(2,
		widget.NewLabelWithStyle("barcode", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("sample_id", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
	)
	tableContainer.Add(header)

	// Function to add a new row
	addRow := func() {
		r := &Row{
			Barcode:  widget.NewEntry(),
			SampleID: widget.NewEntry(),
		}
		rowUI := container.NewGridWithColumns(2, r.Barcode, r.SampleID)
		tableContainer.Add(rowUI)
		rows = append(rows, r)
	}

	// Start with 3 default rows
	for i := 0; i < 3; i++ {
		addRow()
	}

	// Button to add row
	addButton := widget.NewButton("Add Row", func() {
		addRow()
	})

	// Button to delete the last row
	deleteButton := widget.NewButton("Delete Last Row", func() {
		if len(rows) > 0 {
			lastIndex := len(rows) - 1
			tableContainer.Objects = tableContainer.Objects[:len(tableContainer.Objects)-1]
			rows = rows[:lastIndex]
			tableContainer.Refresh()
		}
	})

	// Button to save to TSV file
	saveButton := widget.NewButton("Save to TSV", func() {
		var sb strings.Builder
		// Write header
		sb.WriteString("barcode\tsample_id\n")
		// Write rows
		for _, r := range rows {
			line := fmt.Sprintf("%s\t%s\n", r.Barcode.Text, r.SampleID.Text)
			sb.WriteString(line)
		}

		file, err := os.Create("output.tsv")
		if err != nil {
			fmt.Println("Error creating file:", err)
			return
		}
		defer file.Close()

		_, err = file.WriteString(sb.String())
		if err != nil {
			fmt.Println("Error writing file:", err)
			return
		}

		dialog := widget.NewLabel("Saved to output.tsv")
		tableContainer.Add(dialog)
		tableContainer.Refresh()
	})

	buttons := container.NewHBox(addButton, deleteButton, saveButton)
	mainContent := container.NewBorder(nil, buttons, nil, nil, tableContainer)

	w.SetContent(mainContent)
	w.ShowAndRun()
}

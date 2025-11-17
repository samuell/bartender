package main

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type Row struct {
	Barcode  *widget.Entry
	SampleID *ForwardJumpOnReturnEntry
}

func main() {
	a := app.New()

	w := a.NewWindow("Barcode mapping")
	w.Resize(fyne.NewSize(640, 480))

	rows := make([]*Row, 0)
	rowCounter := 1 // To keep track of the row index for auto-fill

	// Table container
	tableContainer := container.NewVBox()

	// Output path entry above table
	outputPathEntry := NewForwardJumpOnReturnEntry(w.Canvas(), nil)
	outputPathEntry.SetText("barcodesheet.csv") // default path
	outputPathLabel := widget.NewLabel("Output file path:")
	outputPathContainer := container.NewBorder(nil, nil, outputPathLabel, nil, outputPathEntry)

	// Add header
	header := container.NewGridWithColumns(2,
		widget.NewLabelWithStyle("barcode", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("sample_id", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
	)
	tableContainer.Add(header)

	// Function to add a new row
	addRow := func() {
		rowCounter := getLastBarcodeNumber(rows)
		barcodeText := fmt.Sprintf("barcode%02d", rowCounter+1)
		samplePlaceholder := fmt.Sprintf("Sample ID for barcode%02d", rowCounter+1)

		var previous *ForwardJumpOnReturnEntry
		if len(rows) > 0 {
			previous = rows[len(rows)-1].SampleID
		}

		r := &Row{
			Barcode:  widget.NewEntry(),
			SampleID: NewForwardJumpOnReturnEntry(w.Canvas(), previous),
		}
		r.Barcode.SetText(barcodeText)
		r.SampleID.SetPlaceHolder(samplePlaceholder)

		rowUI := container.NewGridWithColumns(2, r.Barcode, r.SampleID)
		tableContainer.Add(rowUI)
		rows = append(rows, r)
	}

	// Initialize with 3 rows
	for i := 0; i < 3; i++ {
		addRow()
	}

	// Buttons
	addButton := widget.NewButton("Add Row", addRow)

	deleteButton := widget.NewButton("Delete Last Row", func() {
		if len(rows) > 0 {
			lastIndex := len(rows) - 1
			tableContainer.Objects = tableContainer.Objects[:len(tableContainer.Objects)-1]
			rows = rows[:lastIndex]
			tableContainer.Refresh()
			rowCounter-- // Decrement counter so next new row continues numbering correctly
		}
	})

	saveButton := widget.NewButton("Save to CSV", func() {
		filePath := outputPathEntry.Text
		if strings.TrimSpace(filePath) == "" {
			filePath = "output.csv"
		}

		var sb strings.Builder
		sb.WriteString("barcode,sample_id\n")
		for _, r := range rows {
			line := fmt.Sprintf("%s,%s\n", r.Barcode.Text, r.SampleID.Text)
			sb.WriteString(line)
		}

		file, err := os.Create(filePath)
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

		// Show a simple confirmation message in the GUI
		msg := widget.NewLabel(fmt.Sprintf("Saved to %s", filePath))

		openFileButton := widget.NewButton("Open file", func() {
			fileUrl, err := url.Parse("file://" + filePath)
			CheckMsg(err, "Could not parse url: "+filePath)
			a.OpenURL(fileUrl)
		})

		showInFolderButton := widget.NewButton("Show in folder", func() {
			var dirPath string
			if len(filePath) > 0 && filePath[0:1] == "/" {
				dirPath = filepath.Dir(filePath)
			} else {
				ex, err := os.Executable()
				CheckMsg(err, "Could not get executable path")
				dirPath = filepath.Dir(ex)
			}

			dirUrl, err := url.Parse("file://" + dirPath)
			CheckMsg(err, "Could not parse url: "+dirPath)
			a.OpenURL(dirUrl)
		})

		// Show a button to open folder

		tableContainer.Add(msg)
		tableContainer.Add(openFileButton)
		tableContainer.Add(showInFolderButton)
		tableContainer.Refresh()
	})

	buttons := container.NewHBox(addButton, deleteButton, saveButton)
	mainContent := container.NewBorder(outputPathContainer, buttons, nil, nil, tableContainer)

	w.SetContent(mainContent)
	w.ShowAndRun()
}

// ForwardJumpOnReturnEntry

func NewForwardJumpOnReturnEntry(canvas fyne.Canvas, previousEntry *ForwardJumpOnReturnEntry) *ForwardJumpOnReturnEntry {
	entry := &ForwardJumpOnReturnEntry{}
	entry.ExtendBaseWidget(entry)
	entry.canvas = canvas
	if previousEntry != nil {
		previousEntry.SetNext(entry)
	}
	return entry
}

type ForwardJumpOnReturnEntry struct {
	widget.Entry
	canvas   fyne.Canvas
	previous *ForwardJumpOnReturnEntry
	next     *ForwardJumpOnReturnEntry
}

func (e *ForwardJumpOnReturnEntry) SetNext(entry *ForwardJumpOnReturnEntry) {
	e.next = entry
}

func (e *ForwardJumpOnReturnEntry) TypedKey(key *fyne.KeyEvent) {
	if key.Name == fyne.KeyReturn {
		if e.next != nil {
			e.next.SetPlaceHolder("Enter Sample ID now!")
			e.canvas.Focus(e.next)
			return
		}
	}

	e.Entry.TypedKey(key)
}

func getLastBarcodeNumber(rows []*Row) int {
	if len(rows) > 0 {
		barcodeText := rows[len(rows)-1].Barcode.Text[7:]
		lastRowIdx, err := strconv.Atoi(barcodeText)
		CheckMsg(err, "Could not convert to int: "+barcodeText)
		return lastRowIdx
	}
	return 0
}

func CheckMsg(err error, message string) {
	if err != nil {
		fmt.Println(message)
		os.Exit(1)
	}
}

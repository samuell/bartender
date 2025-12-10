package main

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func main() {
	a := NewBartenderApp()
	a.window.ShowAndRun()
}

// BartenderApp

func NewBartenderApp() *BartenderApp {
	innerApp := app.New()
	a := &BartenderApp{
		innerApp,
		innerApp.NewWindow("Barcode mapping"),
		[]*Row{},
		0,
		container.NewVBox(),
	}
	a.window.Resize(fyne.NewSize(640, 480))
	// Table container
	a.tableContainer = container.NewVBox()

	// Output path entry above table
	outputPathEntry := widget.NewEntry()
	now := time.Now()
	outputPathEntry.SetText("/data/trana/barcodesheets/16s_ont_" + now.Format("060102") + ".csv") // default path
	outputPathLabel := widget.NewLabel("Output file path:")
	outputPathContainer := container.NewBorder(nil, nil, outputPathLabel, nil, outputPathEntry)

	// Add header
	header := container.NewGridWithColumns(2,
		widget.NewLabelWithStyle("barcode", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("sample_id", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
	)
	a.tableContainer.Add(header)

	// Initialize with 3 rows
	for i := 0; i < 3; i++ {
		a.addRow()
	}

	// Buttons
	addButton := widget.NewButton("Add Row", a.addRow)

	deleteButton := widget.NewButton("Delete Last Row", func() {
		if len(a.rows) > 0 {
			lastIndex := len(a.rows) - 1
			a.tableContainer.Objects = a.tableContainer.Objects[:len(a.tableContainer.Objects)-1]
			a.rows = a.rows[:lastIndex]
			a.tableContainer.Refresh()
			a.rowCounter-- // Decrement counter so next new row continues numbering correctly
		}
	})

	saveButton := widget.NewButton("Save to CSV", func() {
		filePath := outputPathEntry.Text
		if strings.TrimSpace(filePath) == "" {
			filePath = "output.csv"
		}

		var sb strings.Builder
		sb.WriteString("barcode,sample_id\n")
		for _, r := range a.rows {
			line := fmt.Sprintf("%s,%s\n", r.Barcode.Text, r.SampleID.Text)
			sb.WriteString(line)
		}

		file, err := os.Create(filePath)
		if !CheckMsgDialog(err, "Error creating file", a.window) {
			return
		}
		defer file.Close()

		_, err = file.WriteString(sb.String())
		if !CheckMsgDialog(err, "Error writing file", a.window) {
			return
		}

		// Show a simple confirmation message in the GUI
		msg := widget.NewLabel(fmt.Sprintf("Saved to %s", filePath))

		openFileButton := widget.NewButton("Open file", func() {
			fileUrl, err := url.Parse("file://" + filePath)
			if !CheckMsgDialog(err, "Could not parse url: "+filePath, a.window) {
				return
			}
			a.OpenURL(fileUrl)
		})

		showInFolderButton := widget.NewButton("Show in folder", func() {
			var dirPath string
			if len(filePath) > 0 && filePath[0:1] == "/" {
				dirPath = filepath.Dir(filePath)
			} else {
				ex, err := os.Executable()
				if !CheckMsgDialog(err, "Could not get executable path", a.window) {
					return
				}
				dirPath = filepath.Dir(ex)
			}

			dirUrl, err := url.Parse("file://" + dirPath)
			if !CheckMsgDialog(err, "Could not parse url: "+dirPath, a.window) {
				return
			}
			a.OpenURL(dirUrl)
		})

		// Show a button to open folder

		a.tableContainer.Add(msg)
		a.tableContainer.Add(openFileButton)
		a.tableContainer.Add(showInFolderButton)
		a.tableContainer.Refresh()
	})

	buttons := container.NewHBox(addButton, deleteButton, saveButton)
	mainContent := container.NewBorder(outputPathContainer, buttons, nil, nil, a.tableContainer)

	a.window.SetContent(mainContent)
	return a
}

type BartenderApp struct {
	fyne.App
	window         fyne.Window
	rows           []*Row
	rowCounter     int
	tableContainer *fyne.Container
}

type Row struct {
	Barcode  *widget.Entry
	SampleID *ForwardJumpOnReturnEntry
}

func (a *BartenderApp) addRow() {
	a.rowCounter = getLastBarcodeNumber(a.rows)
	barcodeText := fmt.Sprintf("barcode%02d", a.rowCounter+1)
	samplePlaceholder := fmt.Sprintf("Sample ID for barcode%02d", a.rowCounter+1)

	var previous *ForwardJumpOnReturnEntry
	if len(a.rows) > 0 {
		previous = a.rows[len(a.rows)-1].SampleID
	}

	r := &Row{
		Barcode:  widget.NewEntry(),
		SampleID: a.NewForwardJumpOnReturnEntry(a.window.Canvas(), previous),
	}
	r.Barcode.SetText(barcodeText)
	r.SampleID.SetPlaceHolder(samplePlaceholder)

	rowUI := container.NewGridWithColumns(2, r.Barcode, r.SampleID)
	a.tableContainer.Add(rowUI)
	a.rows = append(a.rows, r)
}

// ForwardJumpOnReturnEntry

func (a *BartenderApp) NewForwardJumpOnReturnEntry(canvas fyne.Canvas, previousEntry *ForwardJumpOnReturnEntry) *ForwardJumpOnReturnEntry {
	entry := &ForwardJumpOnReturnEntry{}
	entry.ExtendBaseWidget(entry)
	entry.canvas = canvas
	if previousEntry != nil {
		previousEntry.SetNext(entry)
	}
	entry.buffer.Reset()
	entry.idleDelay = 500 * time.Millisecond
	entry.app = a
	return entry
}

type ForwardJumpOnReturnEntry struct {
	widget.Entry
	app       *BartenderApp
	canvas    fyne.Canvas
	previous  *ForwardJumpOnReturnEntry
	next      *ForwardJumpOnReturnEntry
	buffer    strings.Builder
	mu        sync.Mutex
	idleTimer *time.Timer
	idleDelay time.Duration
}

func (e *ForwardJumpOnReturnEntry) SetNext(entry *ForwardJumpOnReturnEntry) {
	e.next = entry
}

func (e *ForwardJumpOnReturnEntry) TypedRune(r rune) {
	e.mu.Lock()
	e.buffer.WriteRune(r)
	fyne.Do(func() {
		e.Entry.TypedRune(r)
		e.SetText(e.buffer.String())
	})
	e.mu.Unlock()
}

func (e *ForwardJumpOnReturnEntry) TypedKey(key *fyne.KeyEvent) {
	if key.Name == fyne.KeyReturn || key.Name == fyne.KeyEnter || key.Name == fyne.KeyDown {
		e.resetTimerAndProcessBuffer(key)
		return
	}

	e.mu.Lock()
	e.Entry.TypedKey(key)
	if len(key.Name) > 1 {
		e.buffer.Reset()
		e.buffer.WriteString(e.Entry.Text)
	}
	e.mu.Unlock()
}

func (e *ForwardJumpOnReturnEntry) resetTimerAndProcessBuffer(key *fyne.KeyEvent) {
	if e.idleTimer != nil {
		e.idleTimer.Stop()
	}
	if key != nil {
		e.idleTimer = time.AfterFunc(e.idleDelay, func() {
			barcode := e.buffer.String()
			if len(barcode) > 0 {
				e.mu.Lock()
				fyne.Do(func() {
					e.SetText(barcode)
				})
				e.mu.Unlock()
				if e.next != nil {
					fyne.Do(func() {
						e.next.SetPlaceHolder("Enter Sample ID now!")
						e.canvas.Focus(e.next)
					})
				}
			}
		})
	}
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

func CheckMsgDialog(err error, message string, window fyne.Window) bool {
	if err != nil {
		dialog.ShowError(errors.New(fmt.Sprintf("%s:\n%s", message, err.Error())), window)
		return false
	}
	return true
}

// Copyright (C) 2025 Murilo Gomes Julio
// SPDX-License-Identifier: MIT

// Site: https://mugomes.github.io

package mgdialogopenfile

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type FileDialogOpen struct {
	a           fyne.App
	title       string
	exts        []string
	multiSelect bool
	onSelect    func([]string)

	lastDir string
}

var lastDirFile = filepath.Join(os.TempDir(), "mgdialogopenfile_lastdir.txt")

// API EXIGIDA PELO SEU MAIN
func New(a fyne.App, title string, exts []string, multiselect bool, onSelect func([]string)) *FileDialogOpen {
	dlg := &FileDialogOpen{
		a:           a,
		title:       title,
		exts:        exts,
		multiSelect: multiselect,
		onSelect:    onSelect,
	}
	dlg.loadLastDir()
	return dlg
}

func (d *FileDialogOpen) Show() {
	win := d.a.NewWindow(d.title)
	win.Resize(fyne.NewSize(740, 520))
	win.CenterOnScreen()

	dir := d.lastDir
	if dir == "" {
		dir, _ = os.UserHomeDir()
	}

	pathLabel := widget.NewLabel(dir)
	search := widget.NewEntry()
	search.SetPlaceHolder("🔍 Buscar...")

	files := d.listDir(dir)
	filtered := files
	selected := map[int]bool{}

	list := widget.NewList(
		func() int { return len(filtered) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.ListItemID, o fyne.CanvasObject) {
			row := filtered[id]
			name := row.Name()

			if selected[id] {
				o.(*widget.Label).SetText("✔ " + name)
			} else {
				if row.IsDir() {
					o.(*widget.Label).SetText("📁 " + name)
				} else {
					o.(*widget.Label).SetText("📄 " + name)
				}
			}
		},
	)

	list.OnSelected = func(id widget.ListItemID) {
		if id < 0 || id >= len(filtered) {
			return
		}
		f := filtered[id]

		// Abrir diretório
		if f.IsDir() {
			dir = filepath.Join(dir, f.Name())
			pathLabel.SetText(dir)
			files = d.listDir(dir)
			filtered = d.applyFilter(files, search.Text)
			selected = map[int]bool{}
			list.Refresh()
			return
		}

		// Abrir arquivo (somente single-select)
		if !d.multiSelect {
			d.saveLastDir(dir)
			d.onSelect([]string{filepath.Join(dir, f.Name())})
			win.Close()
			return
		}

		// Clique simples
		if f.IsDir() {
			// apenas destaca a pasta
			selected = map[int]bool{id: true}
		} else {
			if d.multiSelect {
				selected[id] = !selected[id]
			} else {
				selected = map[int]bool{id: true}
			}
		}

		list.Refresh()
	}

	// BOTÃO VOLTAR
	btnBack := widget.NewButtonWithIcon("", theme.NavigateBackIcon(), func() {
		parent := filepath.Dir(dir)
		if parent != dir {
			dir = parent
			pathLabel.SetText(dir)
			files = d.listDir(dir)
			filtered = d.applyFilter(files, search.Text)
			selected = map[int]bool{}
			list.Refresh()
		}
	})

	// BOTÃO ABRIR
	btnOpen := widget.NewButtonWithIcon("Abrir", theme.ConfirmIcon(), func() {
		var out []string
		for i := range selected {
			f := filtered[i]
			if !f.IsDir() && selected[i] {
				out = append(out, filepath.Join(dir, f.Name()))
			}
		}

		if len(out) > 0 {
			d.saveLastDir(dir)
			d.onSelect(out)
			win.Close()
		}
	})

	// BUSCA
	search.OnChanged = func(txt string) {
		filtered = d.applyFilter(files, txt)
		selected = map[int]bool{}
		list.Refresh()
	}

	// LAYOUT
	top := container.NewBorder(nil, nil, btnBack, nil,
		container.NewVBox(pathLabel, search),
	)

	bottom := container.NewHBox(btnOpen)

	win.SetContent(
		container.NewBorder(
			top,
			bottom,
			nil, nil,
			list,
		),
	)

	win.Show()
}

//////////////////////////////////////////////////////////////
// FUNÇÕES AUXILIARES
//////////////////////////////////////////////////////////////

func (d *FileDialogOpen) listDir(path string) []fs.FileInfo {
	entries, _ := os.ReadDir(path)

	var list []fs.FileInfo

	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			continue
		}

		if len(d.exts) > 0 && !info.IsDir() {
			ok := false
			for _, ext := range d.exts {
				if strings.EqualFold(filepath.Ext(info.Name()), ext) {
					ok = true
					break
				}
			}
			if !ok {
				continue
			}
		}

		list = append(list, info)
	}

	sort.Slice(list, func(i, j int) bool {
		a, b := list[i], list[j]
		if a.IsDir() != b.IsDir() {
			return a.IsDir()
		}
		return strings.ToLower(a.Name()) < strings.ToLower(b.Name())
	})

	return list
}

func (d *FileDialogOpen) applyFilter(files []fs.FileInfo, query string) []fs.FileInfo {
	if query == "" {
		return files
	}

	q := strings.ToLower(query)
	var out []fs.FileInfo

	for _, f := range files {
		if strings.Contains(strings.ToLower(f.Name()), q) {
			out = append(out, f)
		}
	}

	return out
}

//////////////////////////////////////////////////////////////
// ÚLTIMO DIRETÓRIO
//////////////////////////////////////////////////////////////

func (d *FileDialogOpen) saveLastDir(dir string) {
	os.WriteFile(lastDirFile, []byte(dir), 0644)
}

func (d *FileDialogOpen) loadLastDir() {
	b, err := os.ReadFile(lastDirFile)
	if err == nil {
		d.lastDir = strings.TrimSpace(string(b))
	}
}

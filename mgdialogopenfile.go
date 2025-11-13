// Copyright (C) 2025 Murilo Gomes Julio
// SPDX-License-Identifier: MIT

// Site: https://mugomes.github.io

package mgfileopen

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var lastDirFile = filepath.Join(os.TempDir(), ".mgfileopen_lastdir")

// Callback genérico: se multiSelect = false → len(paths) == 1
type FileSelectCallback func(paths []string)

// Show cria e retorna uma janela seletora de arquivos não modal.
//
// Parâmetros:
//   - a: aplicação Fyne
//   - title: título da janela
//   - exts: extensões aceitas (ex: []string{".jpg", ".png"} ou nil para todos)
//   - multiSelect: true para permitir selecionar múltiplos arquivos
//   - onSelect: callback com os arquivos selecionados
func Show(a fyne.App, title string, exts []string, multiSelect bool, onSelect FileSelectCallback) fyne.Window {
	win := a.NewWindow(title)
	win.Resize(fyne.NewSize(800, 500))
	win.CenterOnScreen()

	dir := loadLastDir()
	if dir == "" {
		dir, _ = os.Getwd()
	}

	pathLabel := widget.NewLabel(dir)
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("🔍 Buscar arquivo...")

	files := listDir(dir, exts)
	filtered := files

	selected := map[int]bool{}

	icon := func(info fs.FileInfo) fyne.Resource {
		if info.IsDir() {
			return theme.FolderIcon()
		}
		return theme.FileIcon()
	}

	list := widget.NewList(
		func() int { return len(filtered) },
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewIcon(nil),
				widget.NewLabel(""),
			)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			ico := o.(*fyne.Container).Objects[0].(*widget.Icon)
			lbl := o.(*fyne.Container).Objects[1].(*widget.Label)
			f := filtered[i]
			ico.SetResource(icon(f))

			// destaca seleção múltipla
			if selected[i] {
				lbl.SetText("✔ " + f.Name())
			} else {
				lbl.SetText(f.Name())
			}
		},
	)

	updateList := func() {
		search := strings.ToLower(searchEntry.Text)
		filtered = []fs.FileInfo{}
		for _, f := range files {
			if search == "" || strings.Contains(strings.ToLower(f.Name()), search) {
				filtered = append(filtered, f)
			}
		}
		list.Refresh()
	}

	searchEntry.OnChanged = func(s string) { updateList() }

	list.OnSelected = func(id widget.ListItemID) {
		f := filtered[id]
		if f.IsDir() {
			// ao clicar em pasta, entra
			dir = filepath.Join(dir, f.Name())
			pathLabel.SetText(dir)
			files = listDir(dir, exts)
			updateList()
			selected = map[int]bool{}
			return
		}

		if multiSelect {
			selected[id] = !selected[id]
			list.Refresh()
		} else {
			selected = map[int]bool{id: true}
			openSelection(win, dir, filtered, selected, onSelect)
		}
	}

	var lastClickTime time.Time
	var lastClickID widget.ListItemID = -1

	list.OnSelected = func(id widget.ListItemID) {
		f := filtered[id]

		// Verifica duplo clique
		now := time.Now()
		if lastClickID == id && now.Sub(lastClickTime) < 400*time.Millisecond {
			// duplo clique detectado
			if f.IsDir() {
				dir = filepath.Join(dir, f.Name())
				pathLabel.SetText(dir)
				files = listDir(dir, exts)
				updateList()
				selected = map[int]bool{}
			} else if !multiSelect {
				saveLastDir(dir)
				onSelect([]string{filepath.Join(dir, f.Name())})
				win.Close()
			}
			lastClickID = -1
			return
		}

		lastClickID = id
		lastClickTime = now

		// Clique simples (seleciona)
		if f.IsDir() {
			// só destaca diretórios, não abre
			selected = map[int]bool{id: true}
		} else {
			if multiSelect {
				selected[id] = !selected[id]
				list.Refresh()
			} else {
				selected = map[int]bool{id: true}
			}
		}
	}

	backBtn := widget.NewButtonWithIcon("", theme.NavigateBackIcon(), func() {
		parent := filepath.Dir(dir)
		if parent != dir {
			dir = parent
			pathLabel.SetText(dir)
			files = listDir(dir, exts)
			updateList()
			selected = map[int]bool{}
		}
	})

	openBtn := widget.NewButtonWithIcon("Abrir", theme.ConfirmIcon(), func() {
		openSelection(win, dir, filtered, selected, onSelect)
	})

	cancelBtn := widget.NewButtonWithIcon("Cancelar", theme.CancelIcon(), func() {
		win.Close()
	})

	top := container.NewBorder(nil, nil, backBtn, nil,
		container.NewVBox(pathLabel, searchEntry),
	)

	bottom := container.NewHBox(openBtn, cancelBtn)
	content := container.NewBorder(top, bottom, nil, nil, list)
	win.SetContent(content)
	return win
}

func openSelection(win fyne.Window, dir string, files []fs.FileInfo, selected map[int]bool, onSelect FileSelectCallback) {
	var paths []string
	for i, ok := range selected {
		if ok {
			f := files[i]
			if !f.IsDir() {
				paths = append(paths, filepath.Join(dir, f.Name()))
			}
		}
	}
	if len(paths) > 0 {
		saveLastDir(dir)
		onSelect(paths)
		win.Close()
	}
}

// listDir lista arquivos e pastas com filtro opcional
func listDir(dir string, exts []string) []fs.FileInfo {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return []fs.FileInfo{}
	}

	var list []fs.FileInfo
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			continue
		}
		if len(exts) > 0 && !info.IsDir() {
			keep := false
			for _, ext := range exts {
				if strings.EqualFold(filepath.Ext(info.Name()), ext) {
					keep = true
					break
				}
			}
			if !keep {
				continue
			}
		}
		list = append(list, info)
	}

	sort.Slice(list, func(i, j int) bool {
		a, b := list[i], list[j]
		if a.IsDir() && !b.IsDir() {
			return true
		}
		if !a.IsDir() && b.IsDir() {
			return false
		}
		return strings.ToLower(a.Name()) < strings.ToLower(b.Name())
	})
	return list
}

func saveLastDir(dir string) {
	_ = os.WriteFile(lastDirFile, []byte(dir), 0644)
}

func loadLastDir() string {
	b, err := os.ReadFile(lastDirFile)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

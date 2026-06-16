package main

import (
	"path/filepath"
	"testing"
)

func TestApplicationMenuContainsWorkspaceFileActions(t *testing.T) {
	app := NewAppWithConfig(filepath.Join(t.TempDir(), "projects.json"))
	menu := app.applicationMenu()
	if len(menu.Items) == 0 {
		t.Fatal("menu has no top-level items")
	}
	fileMenu := menu.Items[0]
	if fileMenu.Label != "文件" {
		t.Fatalf("first menu label = %q, want 文件", fileMenu.Label)
	}
	if fileMenu.SubMenu == nil {
		t.Fatal("file menu has no submenu")
	}
	labels := make([]string, 0, len(fileMenu.SubMenu.Items))
	for _, item := range fileMenu.SubMenu.Items {
		labels = append(labels, item.Label)
	}
	want := []string{"打开项目", "最近打开", "清理最近打开", "关闭"}
	if len(labels) != len(want) {
		t.Fatalf("file menu labels = %#v, want %#v", labels, want)
	}
	for index := range want {
		if labels[index] != want[index] {
			t.Fatalf("file menu labels = %#v, want %#v", labels, want)
		}
	}
}

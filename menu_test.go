package main

import (
	"path/filepath"
	"testing"

	"github.com/wailsapp/wails/v2/pkg/menu"
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

func TestApplicationMenuIncludesMacOSStandardRolesWithWorkspaceFileActions(t *testing.T) {
	app := NewAppWithConfig(filepath.Join(t.TempDir(), "projects.json"))
	appMenu := app.applicationMenuForPlatform("darwin")

	roles := topLevelMenuRoles(appMenu)
	if !roles[menu.AppMenuRole] {
		t.Fatalf("menu roles = %#v, want AppMenuRole", roles)
	}
	if !roles[menu.EditMenuRole] {
		t.Fatalf("menu roles = %#v, want EditMenuRole", roles)
	}
	if !roles[menu.WindowMenuRole] {
		t.Fatalf("menu roles = %#v, want WindowMenuRole", roles)
	}
	if !topLevelMenuLabels(appMenu)["文件"] {
		t.Fatalf("menu labels = %#v, want 文件 menu", topLevelMenuLabels(appMenu))
	}
}

func TestApplicationMenuKeepsExistingTopLevelFileMenuOnNonMacPlatforms(t *testing.T) {
	app := NewAppWithConfig(filepath.Join(t.TempDir(), "projects.json"))
	appMenu := app.applicationMenuForPlatform("linux")

	if len(appMenu.Items) != 1 {
		t.Fatalf("top-level menu item count = %d, want 1", len(appMenu.Items))
	}
	if appMenu.Items[0].Label != "文件" {
		t.Fatalf("top-level menu label = %q, want 文件", appMenu.Items[0].Label)
	}
}

func topLevelMenuRoles(appMenu *menu.Menu) map[menu.Role]bool {
	roles := map[menu.Role]bool{}
	for _, item := range appMenu.Items {
		roles[item.Role] = true
	}
	return roles
}

func topLevelMenuLabels(appMenu *menu.Menu) map[string]bool {
	labels := map[string]bool{}
	for _, item := range appMenu.Items {
		labels[item.Label] = true
	}
	return labels
}

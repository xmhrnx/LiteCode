// LiteCode is a light IDE with dockable panels, a built-in Git workflow and
// a polished, keyboard-first workflow, built on Giode.
package main

import (
	"fmt"
	"image"
	"os"

	"github.com/xmhrnx/giode"

	"litecode/assets"
	"litecode/internal/config"
	"litecode/ui/app"
	"litecode/ui/theme"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "litecode:", err)
		os.Exit(1)
	}
}

func run() error {
	if err := theme.LoadFonts(); err != nil {
		return err
	}
	if err := theme.LoadIcons(); err != nil {
		return err
	}
	giode.SecureStore.Service = "dev.litecode.app"
	pluginDir, err := config.Path("plugins")
	if err != nil {
		return err
	}
	var root *app.Root
	a := giode.New(giode.Config{
		Title:         "LiteCode",
		Size:          image.Pt(1440, 900),
		MinSize:       image.Pt(720, 480),
		Resizable:     true,
		Frameless:     true,
		Center:        true,
		Background:    theme.Current().Editor,
		Font:          theme.UIFont,
		Icon:          assets.AppIcon,
		OnFileDrop:    func(paths []string) { root.HandleDrop(paths) },
		OnFocusChange: func(focused bool) { root.FocusChanged(focused) },
		SingleInstanceLock: &giode.SingleInstanceLock{
			UniqueID: "litecode-6f1d2b7e",
			OnLaunchArgs: func(args []string) {
				if len(args) > 0 {
					root.OpenProject(args[0])
				}
			},
		},
	})
	root = app.NewRoot(a, pluginDir)
	root.Start(os.Args[1:])
	return a.Run(root.View)
}

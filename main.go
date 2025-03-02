package main

import (
	"Luminous/icon"
	"fmt"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
	"os"

	"github.com/getlantern/systray"
)

var mToggle *systray.MenuItem
var quitChan chan struct{}

func main() {
	addToStartup()
	systray.Run(onReady, onExit)
}

func onReady() {
	quitChan = make(chan struct{})
	systray.SetTitle("Luminous")

	mStatus := systray.AddMenuItem(getStatusText(), "Current theme status")
	mStatus.Disable()
	mToggle = systray.AddMenuItem("Switch Theme", "Dark ↔ Light")

	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Close the application")

	updateIconAndTooltip()

	go watchThemeChanges()

	go func() {
		for {
			select {
			case <-mToggle.ClickedCh:
				toggleTheme()
				updateIconAndTooltip()
				mStatus.SetTitle(getStatusText())
			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
}

func onExit() {
	if quitChan != nil {
		close(quitChan)
	}
	fmt.Println("Application is shutting down...")
}

func updateIconAndTooltip() {
	currentTheme, err := getCurrentTheme()
	if err != nil {
		fmt.Println("Error: Could not get theme:", err)
		return
	}

	if currentTheme == 1 {
		systray.SetIcon(icon.Moon)
		systray.SetTooltip("Switch to Dark Theme")
		mToggle.SetTitle("Switch to Dark Theme")
		mToggle.SetIcon(icon.Moon)
	} else {
		systray.SetIcon(icon.Sun)
		systray.SetTooltip("Switch to Light Theme")
		mToggle.SetTitle("Switch to Light Theme")
		mToggle.SetIcon(icon.Sun)
	}
}

func getStatusText() string {
	currentTheme, err := getCurrentTheme()
	if err != nil {
		return "Status: Unknown"
	}
	if currentTheme == 1 {
		return "Status: Light Theme"
	}
	return "Status: Dark Theme"
}

func toggleTheme() {
	currentTheme, err := getCurrentTheme()
	if err != nil {
		fmt.Println("Error: Could not get theme:", err)
		return
	}

	newTheme := 1 - currentTheme

	if err = setTheme(newTheme); err != nil {
		fmt.Println("Error: Could not set theme:", err)
		return
	}

	if err = refreshTheme(); err != nil {
		fmt.Println("Error: Theme refresh failed:", err)
		return
	}
}

func getCurrentTheme() (int, error) {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Themes\Personalize`, registry.QUERY_VALUE)
	if err != nil {
		return -1, err
	}
	defer k.Close()

	val, _, err := k.GetIntegerValue("AppsUseLightTheme")
	if err != nil {
		return -1, err
	}
	return int(val), nil
}

func setTheme(value int) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Themes\Personalize`, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()

	if err := k.SetDWordValue("AppsUseLightTheme", uint32(value)); err != nil {
		return err
	}
	if err := k.SetDWordValue("SystemUsesLightTheme", uint32(value)); err != nil {
		return err
	}
	return nil
}

func refreshTheme() error {
	const WM_SETTINGCHANGE = 0x001A
	const HWND_BROADCAST = 0xffff

	user32 := windows.NewLazySystemDLL("user32.dll")
	sendMessageW := user32.NewProc("SendMessageW")

	ret, _, err := sendMessageW.Call(
		uintptr(HWND_BROADCAST),
		uintptr(WM_SETTINGCHANGE),
		0,
		0,
	)
	if ret == 0 {
		return fmt.Errorf("SendMessageW failed: %v", err)
	}
	return nil
}

func addToStartup() {
	exePath, err := os.Executable()
	if err != nil {
		fmt.Println("Error: Could not get executable path:", err)
		return
	}

	key, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Run`, registry.ALL_ACCESS)
	if err != nil {
		fmt.Println("Error: Could not open registry key:", err)
		return
	}
	defer key.Close()

	appName := "Luminous"
	val, _, err := key.GetStringValue(appName)
	if err == nil && val == exePath {
		fmt.Println("Application already in startup.")
		return
	}

	err = key.SetStringValue(appName, exePath)
	if err != nil {
		fmt.Println("Error: Could not set registry value:", err)
		return
	}

	fmt.Println("Application added to startup successfully")
}

func watchThemeChanges() {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Themes\Personalize`, registry.READ)
	if err != nil {
		fmt.Println("Error: Could not open registry key:", err)
		return
	}
	defer k.Close()

	hKey := windows.Handle(k)

	for {
		err = windows.RegNotifyChangeKeyValue(hKey, true, windows.REG_NOTIFY_CHANGE_LAST_SET, 0, false)
		if err != nil {
			fmt.Println("Error: Registry watch failed:", err)
			return
		}
		updateIconAndTooltip()
	}
}

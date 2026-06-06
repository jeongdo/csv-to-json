package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
)

const (
	smCxScreen = 0
	smCyScreen = 1
)

func getScreenResolution() (int, int) {
	if runtime.GOOS != "windows" {
		return 1920, 1080
	}
	mod := syscall.NewLazyDLL("user32.dll")
	proc := mod.NewProc("GetSystemMetrics")
	width, _, _ := proc.Call(uintptr(smCxScreen))
	height, _, _ := proc.Call(uintptr(smCyScreen))
	return int(width), int(height)
}

func launchAppWindow(url string, appWidth, appHeight int) {
	scrWidth, scrHeight := getScreenResolution()
	posX := (scrWidth - appWidth) / 2
	posY := (scrHeight - appHeight) / 2

	// 독립적인 프로필 경로 생성 (매번 깨끗한 창으로 띄우기 위함)
	profilePath := filepath.Join(os.TempDir(), "CsvToJson_Profile")

	// 브라우저 실행 인자 구성
	appArg := "--app=" + url
	sizeArg := fmt.Sprintf("--window-size=%d,%d", appWidth, appHeight)
	posArg := fmt.Sprintf("--window-position=%d,%d", posX, posY)
	profileArg := "--user-data-dir=" + profilePath
	// 위치 제어 무시 방지: 기존 창의 상태를 가져오지 않도록 함
	noFirstRun := "--no-first-run"

	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		chromePaths := []string{
			os.Getenv("ProgramFiles") + `\Google\Chrome\Application\chrome.exe`,
			os.Getenv("ProgramFiles(x86)") + `\Google\Chrome\Application\chrome.exe`,
			os.Getenv("LocalAppData") + `\Google\Chrome\Application\chrome.exe`,
		}
		edgePaths := []string{
			os.Getenv("ProgramFiles(x86)") + `\Microsoft\Edge\Application\msedge.exe`,
			os.Getenv("ProgramFiles") + `\Microsoft\Edge\Application\msedge.exe`,
		}

		browserLaunched := false
		for _, path := range append(chromePaths, edgePaths...) {
			if _, err := os.Stat(path); err == nil {
				cmd = exec.Command(path, appArg, sizeArg, posArg, profileArg, noFirstRun)
				browserLaunched = true
				break
			}
		}

		if !browserLaunched {
			cmd = exec.Command("cmd", "/c", "start", url)
			cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		}

	case "darwin":
		cmd = exec.Command("open", "-a", "Google Chrome", "--args", appArg, sizeArg, posArg)

	default:
		cmd = exec.Command("xdg-open", url)
	}

	if cmd != nil {
		_ = cmd.Start() // 백그라운드 실행
	}
}

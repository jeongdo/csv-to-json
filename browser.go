// browser.go
package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"syscall"
)

const (
	smCxScreen = 0
	smCyScreen = 1
)

// 화면 해상도 구하기
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

// OS별 브라우저 실행 및 3단계 폴백 시스템
func launchAppWindow(url string, appWidth, appHeight int) {
	scrWidth, scrHeight := getScreenResolution()
	posX := (scrWidth - appWidth) / 2
	posY := (scrHeight - appHeight) / 2

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

		sizeArg := fmt.Sprintf("--window-size=%d,%d", appWidth, appHeight)
		posArg := fmt.Sprintf("--window-position=%d,%d", posX, posY)
		appArg := "--app=" + url

		browserLaunched := false

		// 1단계: 크롬 탐색
		for _, path := range chromePaths {
			if _, err := os.Stat(path); err == nil {
				cmd = exec.Command(path, appArg, sizeArg, posArg)
				browserLaunched = true
				break
			}
		}

		// 2단계: 엣지 탐색
		if !browserLaunched {
			for _, path := range edgePaths {
				if _, err := os.Stat(path); err == nil {
					cmd = exec.Command(path, appArg, sizeArg, posArg)
					browserLaunched = true
					break
				}
			}
		}

		// 3단계: 기본 브라우저 새 탭
		if !browserLaunched {
			cmd = exec.Command("cmd", "/c", "start", url)
			cmd.SysProcAttr = &syscall.SysProcAttr{
				HideWindow: true,
			}
		}

	case "darwin":
		cmd = exec.Command(
			"open",
			"-a",
			"Google Chrome",
			"--args",
			"--app="+url,
		)

	default:
		cmd = exec.Command("xdg-open", url)
	}

	if cmd != nil {
		cmd.Run()
	}
}

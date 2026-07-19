package main

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

//go:embed pi/cbm.md
var piCbmRules string

var (
	piLog          *log.Logger
	ansiEscapeRe = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
)

func init() {
	logDir := filepath.Join(os.Getenv("APPDATA"), "pi-mgr")
	os.MkdirAll(logDir, 0755)
	f, err := os.Create(filepath.Join(logDir, "pi-manage.log"))
	if err == nil {
		piLog = log.New(f, "", log.LstdFlags)
	} else {
		piLog = log.New(os.Stderr, "[pi-manage] ", log.LstdFlags)
	}
}

// runPiCommand executes a pi CLI command locally.
func runPiCommand(args []string, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := newCmd(ctx, "pi", args...)
	output, err := cmd.CombinedOutput()
	outStr := strings.TrimSpace(string(output))
	piLog.Printf("LOCAL: pi %s", strings.Join(args, " "))
	if err != nil {
		piLog.Printf("LOCAL FAIL: %v | %s", err, outStr)
	} else {
		piLog.Printf("LOCAL OK: %s", outStr)
	}
	return outStr, err
}

// runPiCommandSSH executes a pi CLI command on a remote host via SSH.
func runPiCommandSSH(sshAddress string, args []string, timeout time.Duration) (string, error) {
	user, host, port, err := parseSSHAddress(sshAddress)
	if err != nil {
		return "", fmt.Errorf("SSH 地址格式错误: %s", err.Error())
	}
	piLog.Printf(">> ssh %s@%s -p %s: pi %s", user, host, port, strings.Join(args, " "))

	// Use -t and bash -i to force interactive session (loads .bashrc fully)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	piArgs := strings.Join(args, " ")
	cmd := newCmd(ctx, "ssh",
		"-p", port,
		"-t",
		"-o", "ConnectTimeout=10",
		"-o", "BatchMode=yes",
		fmt.Sprintf("%s@%s", user, host),
		"bash -i -c 'pi "+piArgs+"'",
	)
	output, err := cmd.CombinedOutput()
	outStr := strings.TrimSpace(string(output))
	// Strip ANSI escape codes (colors, bold, dim) from terminal output
	outStr = ansiEscapeRe.ReplaceAllString(outStr, "")
	// Strip SSH connection close message that appears with -t flag
	lines := strings.Split(outStr, "\n")
	var clean []string
	for _, line := range lines {
		if !strings.HasPrefix(line, "Connection to ") {
			clean = append(clean, line)
		}
	}
	outStr = strings.TrimSpace(strings.Join(clean, "\n"))
	if err != nil {
		piLog.Printf("<< FAIL: %v | %s", err, outStr)
	} else {
		piLog.Printf("<< OK: %s", outStr)
	}
	return outStr, err
}

// GetPiVersion returns the installed pi version.
func (a *App) GetPiVersion() (string, error) {
	out, err := runPiCommand([]string{"--version"}, 10*time.Second)
	if err != nil {
		return "", fmt.Errorf("未检测到 Pi，请确认已安装 pi")
	}
	return out, nil
}

// GetPiPackages returns the raw output of "pi list".
func (a *App) GetPiPackages() (string, error) {
	out, err := runPiCommand([]string{"list"}, 10*time.Second)
	if err != nil {
		return "", fmt.Errorf("获取插件列表失败")
	}
	return out, nil
}

// UpdatePiSelf upgrades pi itself via "pi update --self".
func (a *App) UpdatePiSelf() (string, error) {
	out, err := runPiCommand([]string{"update", "--self"}, 120*time.Second)
	return checkResult(out, err, "升级 Pi 失败")
}

// UpdateAllPiPackages upgrades all installed packages via "pi update --extensions".
func (a *App) UpdateAllPiPackages() (string, error) {
	out, err := runPiCommand([]string{"update", "--extensions"}, 120*time.Second)
	return checkResult(out, err, "升级插件失败")
}

// UpdatePiPackage upgrades a single package via "pi update <source>".
func (a *App) UpdatePiPackage(source string) (string, error) {
	out, err := runPiCommand([]string{"update", source}, 120*time.Second)
	return checkResult(out, err, "升级插件失败")
}

// RemovePiPackage removes a package via "pi remove <source>".
func (a *App) RemovePiPackage(source string) (string, error) {
	out, err := runPiCommand([]string{"remove", source}, 30*time.Second)
	return checkResult(out, err, "删除插件失败")
}

// --- Remote variants (via SSH) ---

// GetRemotePiVersion returns pi version on a remote host.
func (a *App) GetRemotePiVersion(sshAddress string) (string, error) {
	out, err := runPiCommandSSH(sshAddress, []string{"--version"}, 10*time.Second)
	if err != nil {
		msg := out
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("远程 Pi 连接失败: %s", msg)
	}
	return out, nil
}

// GetRemotePiPackages returns plugin list on a remote host.
func (a *App) GetRemotePiPackages(sshAddress string) (string, error) {
	out, err := runPiCommandSSH(sshAddress, []string{"list"}, 10*time.Second)
	if err != nil {
		msg := out
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("获取远程插件列表失败: %s", msg)
	}
	return out, nil
}

// UpdateRemotePiSelf upgrades pi on a remote host.
func (a *App) UpdateRemotePiSelf(sshAddress string) (string, error) {
	out, err := runPiCommandSSH(sshAddress, []string{"update", "--self"}, 120*time.Second)
	return checkResult(out, err, "远程升级 Pi 失败")
}

// UpdateRemoteAllPiPackages upgrades all plugins on a remote host.
func (a *App) UpdateRemoteAllPiPackages(sshAddress string) (string, error) {
	out, err := runPiCommandSSH(sshAddress, []string{"update", "--extensions"}, 120*time.Second)
	return checkResult(out, err, "远程升级插件失败")
}

// UpdateRemotePiPackage upgrades a single plugin on a remote host.
func (a *App) UpdateRemotePiPackage(sshAddress string, source string) (string, error) {
	out, err := runPiCommandSSH(sshAddress, []string{"update", source}, 120*time.Second)
	return checkResult(out, err, "远程升级插件失败")
}

// RemoveRemotePiPackage removes a plugin on a remote host.
func (a *App) RemoveRemotePiPackage(sshAddress string, source string) (string, error) {
	out, err := runPiCommandSSH(sshAddress, []string{"remove", source}, 30*time.Second)
	return checkResult(out, err, "远程删除插件失败")
}

// InstallPiPackage installs a plugin via "pi install <source>".
func (a *App) InstallPiPackage(source string) (string, error) {
	out, err := runPiCommand([]string{"install", source}, 120*time.Second)
	return checkResult(out, err, "安装插件失败")
}

// InstallRemotePiPackage installs a plugin on a remote host.
func (a *App) InstallRemotePiPackage(sshAddress string, source string) (string, error) {
	out, err := runPiCommandSSH(sshAddress, []string{"install", source}, 120*time.Second)
	return checkResult(out, err, "远程安装插件失败")
}

// GetCbmRules returns the CBM usage rules from pi/cbm.md.
func (a *App) GetCbmRules() (string, error) {
	return piCbmRules, nil
}
func checkResult(out string, err error, failMsg string) (string, error) {
	if err != nil {
		if out != "" {
			return out, errors.New(out)
		}
		return "", fmt.Errorf("%s: %s", failMsg, err.Error())
	}
	return out, nil
}

package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// SyncItemStatus represents the sync result for a single item.
type SyncItemStatus struct {
	Name    string `json:"name"`    // "settings.json" / "models.json" / "prompts" / "skills"
	Status  string `json:"status"`  // "success" / "skipped" / "failed"
	Message string `json:"message"` // details
}

// SyncResult represents the overall sync result.
type SyncResult struct {
	Overall string          `json:"overall"` // "success" / "partial" / "failed"
	Items   []SyncItemStatus `json:"items"`
}

// SSHConnectionResult represents the result of an SSH connection test.
type SSHConnectionResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// parseSSHAddress parses user@host[:port] format.
// Returns user, host, port (defaults to "22").
func parseSSHAddress(address string) (string, string, string, error) {
	address = strings.TrimSpace(address)
	if address == "" {
		return "", "", "", fmt.Errorf("SSH 地址不能为空")
	}

	parts := strings.SplitN(address, "@", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", "", fmt.Errorf("SSH 地址格式应为 user@host[:port]")
	}

	user := parts[0]
	hostPort := parts[1]

	port := "22"
	if idx := strings.LastIndex(hostPort, ":"); idx >= 0 {
		port = hostPort[idx+1:]
		hostPort = hostPort[:idx]
		if port == "" {
			return "", "", "", fmt.Errorf("SSH 地址格式错误：端口号不能为空")
		}
	}

	if hostPort == "" {
		return "", "", "", fmt.Errorf("SSH 地址格式错误：主机名不能为空")
	}

	return user, hostPort, port, nil
}

// TestSSHConnection tests SSH connectivity to the given address.
func (a *App) TestSSHConnection(address string) SSHConnectionResult {
	user, host, port, err := parseSSHAddress(address)
	if err != nil {
		return SSHConnectionResult{false, fmt.Sprintf("地址格式错误: %s", err.Error())}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cmd := newCmd(ctx, "ssh",
		"-p", port,
		"-o", "ConnectTimeout=10",
		"-o", "BatchMode=yes",
		fmt.Sprintf("%s@%s", user, host),
		"exit",
	)

	output, err := cmd.CombinedOutput()

	if err != nil {
		msg := strings.TrimSpace(string(output))

		switch {
		case strings.Contains(msg, "Connection timed out"),
			strings.Contains(msg, "连接超时"):
			return SSHConnectionResult{false, "SSH 连接超时，请检查主机地址和网络连接"}
		case strings.Contains(msg, "Connection refused"):
			return SSHConnectionResult{false, "SSH 连接被拒绝，请检查目标主机 SSH 服务是否运行"}
		case strings.Contains(msg, "Host key verification failed"):
			return SSHConnectionResult{false, "主机密钥验证失败，请先在终端中手动连接以确认主机指纹"}
		case strings.Contains(msg, "Permission denied"),
			strings.Contains(msg, "Authentication failed"):
			return SSHConnectionResult{false, "SSH 认证失败，请检查 SSH 密钥配置"}
		case strings.Contains(msg, "No route to host"):
			return SSHConnectionResult{false, "无法到达目标主机，请检查网络连接和地址"}
		case strings.Contains(msg, "Name or service not known"),
			strings.Contains(msg, "Temporary failure in name resolution"):
			return SSHConnectionResult{false, "主机名解析失败，请检查 SSH 地址中的主机名"}
		default:
			if len(msg) > 120 {
				msg = msg[:120] + "..."
			}
			if msg == "" {
				msg = err.Error()
			}
			return SSHConnectionResult{false, fmt.Sprintf("SSH 连接失败: %s", msg)}
		}
	}

	return SSHConnectionResult{true, "SSH 连接成功"}
}

// newCmd creates an exec.CommandContext with hidden window (no flashing terminal).
func newCmd(ctx context.Context, name string, arg ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, name, arg...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd
}

// sshExec runs a command on a remote host via SSH. Returns combined output and error.
func sshExec(user, host, port, cmdStr string, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := newCmd(ctx, "ssh",
		"-p", port,
		"-o", "ConnectTimeout=10",
		"-o", "BatchMode=yes",
		fmt.Sprintf("%s@%s", user, host),
		cmdStr,
	)

	output, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(output)), err
}

// scpCopy copies a local file or directory to a remote destination using scp.
func scpCopy(user, host, port, localPath, remoteDest string, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := newCmd(ctx, "scp",
		"-r", // recursive: works for both files and directories
		"-P", port,
		"-o", "ConnectTimeout=10",
		"-o", "BatchMode=yes",
		localPath,
		fmt.Sprintf("%s@%s:%s", user, host, remoteDest),
	)

	output, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(output)), err
}

// syncDirViaSCP syncs a local directory to a remote location with mirror semantics:
//  1. scp -r the entire directory to a remote temp location
//  2. ssh to atomically replace the target (rm -rf + mv)
//
// This achieves the same effect as rsync --delete without requiring rsync on the client.
func syncDirViaSCP(user, host, port, localDir, remoteTarget, tempSuffix string, itemName string) (string, string) {
	remoteTemp := "~/.pi/_sync_" + tempSuffix

	// Step 1: scp -r the entire directory to a remote temp location
	out, err := scpCopy(user, host, port, localDir, remoteTemp, 60*time.Second)
	if err != nil {
		msg := out
		if msg == "" {
			msg = err.Error()
		}
		return "failed", msg
	}

	// Step 2: atomically replace the target on the remote side
	// rm the existing target, then mv the temp copy into place
	replaceCmd := fmt.Sprintf("rm -rf %s && mv %s %s", remoteTarget, remoteTemp, remoteTarget)
	out, err = sshExec(user, host, port, replaceCmd, 15*time.Second)
	if err != nil {
		msg := out
		if msg == "" {
			msg = err.Error()
		}
		return "failed", msg
	}

	return "success", "已同步"
}

// SyncPiConfig syncs the local pi configuration to the remote machine via SSH.
// It performs pre-checks, then transfers each item. Each item is independent.
func (a *App) SyncPiConfig(address string) SyncResult {
	items := []SyncItemStatus{}
	successCount := 0
	failedCount := 0

	addResult := func(name, status, message string) {
		items = append(items, SyncItemStatus{Name: name, Status: status, Message: message})
		switch status {
		case "success":
			successCount++
		case "failed":
			failedCount++
		}
	}

	// Parse address
	user, host, port, err := parseSSHAddress(address)
	if err != nil {
		addResult("", "failed", fmt.Sprintf("地址解析失败: %s", err.Error()))
		return SyncResult{Overall: "failed", Items: items}
	}

	// Pre-check SSH connectivity (AC-09, AC-10: fail fast without file transfer)
	if _, err := sshExec(user, host, port, "exit", 15*time.Second); err != nil {
		addResult("", "failed", "SSH 连接失败，无法开始同步")
		return SyncResult{Overall: "failed", Items: items}
	}

	// Pre-create remote directories (AC-05, AC-06)
	// All pi config files live under ~/.pi/agent/
	if _, err := sshExec(user, host, port, "mkdir -p ~/.pi/agent ~/.pi/agent/prompts ~/.pi/agent/skills", 15*time.Second); err != nil {
		addResult("", "failed", "远程目录创建失败")
		return SyncResult{Overall: "failed", Items: items}
	}

	homeDir, _ := os.UserHomeDir()
	piDir := filepath.Join(homeDir, ".pi")

	// ---- Item 1: settings.json (scp) ----
	// Pi stores settings.json at ~/.pi/agent/settings.json (not ~/.pi/settings.json)
	settingsSrc := filepath.Join(piDir, "agent", "settings.json")
	if _, err := os.Stat(settingsSrc); os.IsNotExist(err) {
		addResult("settings.json", "skipped", "本地文件不存在，已跳过")
	} else {
		out, err := scpCopy(user, host, port, settingsSrc, "~/.pi/agent/settings.json", 30*time.Second)
		if err != nil {
			msg := out
			if msg == "" {
				msg = err.Error()
			}
			addResult("settings.json", "failed", msg)
		} else {
			addResult("settings.json", "success", "已同步")
		}
	}

	// ---- Item 2: models.json (scp) ----
	modelsSrc := filepath.Join(piDir, "agent", "models.json")
	if _, err := os.Stat(modelsSrc); os.IsNotExist(err) {
		addResult("models.json", "skipped", "本地文件不存在，已跳过")
	} else {
		out, err := scpCopy(user, host, port, modelsSrc, "~/.pi/agent/models.json", 30*time.Second)
		if err != nil {
			msg := out
			if msg == "" {
				msg = err.Error()
			}
			addResult("models.json", "failed", msg)
		} else {
			addResult("models.json", "success", "已同步")
		}
	}

	// ---- Item 3: prompts/ (scp -r + ssh atomic replace) ----
	// Pi stores prompts at ~/.pi/agent/prompts/ (not ~/.pi/prompts/)
	promptsSrc := filepath.Join(piDir, "agent", "prompts")
	if _, err := os.Stat(promptsSrc); os.IsNotExist(err) {
		addResult("prompts", "skipped", "本地目录不存在，已跳过")
	} else {
		status, msg := syncDirViaSCP(user, host, port, promptsSrc, "~/.pi/agent/prompts", "prompts", "prompts")
		addResult("prompts", status, msg)
	}

	// ---- Item 4: skills/ (scp -r + ssh atomic replace) ----
	skillsSrc := filepath.Join(piDir, "agent", "skills")
	if _, err := os.Stat(skillsSrc); os.IsNotExist(err) {
		addResult("skills", "skipped", "本地目录不存在，已跳过")
	} else {
		status, msg := syncDirViaSCP(user, host, port, skillsSrc, "~/.pi/agent/skills", "skills", "skills")
		addResult("skills", status, msg)
	}

	// Determine overall status (AC-12: all failed → overall "failed")
	overall := "success"
	if failedCount > 0 && successCount == 0 {
		overall = "failed"
	} else if failedCount > 0 {
		overall = "partial"
	}

	return SyncResult{Overall: overall, Items: items}
}

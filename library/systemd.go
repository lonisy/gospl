package library

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type SystemdUnit struct {
	Unit    UnitSection    `yaml:"Unit"`
	Service ServiceSection `yaml:"Service"`
	Install InstallSection `yaml:"Install"`
}

type UnitSection struct {
	Description   string `yaml:"Description"`
	Documentation string `yaml:"Documentation"`
	After         string `yaml:"After"`
}

type ServiceSection struct {
	Type             string   `yaml:"Type"`
	Restart          string   `yaml:"Restart"`
	RestartSec       string   `yaml:"RestartSec"`
	Environment      []string `yaml:"Environment"`
	WorkingDirectory string   `yaml:"WorkingDirectory"`
	ExecStart        string   `yaml:"ExecStart"`
	ExecReload       string   `yaml:"ExecReload"`
	ExecStop         string   `yaml:"ExecStop"`
	PrivateTmp       bool     `yaml:"PrivateTmp"`
}

type InstallSection struct {
	WantedBy string `yaml:"WantedBy"`
}

type SystemdService struct {
	unit SystemdUnit
}

var Systemd SystemdService

func init() {
	// 获取
	etcdEndpoints := os.Getenv("ETCD_ENDPOINTS")
	adEtcdEndpoints := os.Getenv("AD_ETCD_ENDPOINTS")

	Systemd.unit = SystemdUnit{
		Unit: UnitSection{
			Description:   "My Application",
			Documentation: "https://xxx/README.md",
			After:         "network.target",
		},
		Service: ServiceSection{
			Type:       "simple",
			Restart:    "always",
			RestartSec: "3s",
			Environment: []string{
				fmt.Sprintf("APP_ENV=%s", os.Getenv("APP_ENV")),
				fmt.Sprintf("ETCD_ENDPOINTS=%s", etcdEndpoints),
				fmt.Sprintf("AD_ETCD_ENDPOINTS=%s", adEtcdEndpoints),
			},
			WorkingDirectory: "/data/app/myapp",
			ExecStart:        "/usr/local/bin/myapp",
			ExecReload:       "/bin/kill -s HUP $MAINPID",
			ExecStop:         "/bin/kill -s SIGTERM $MAINPID",
			PrivateTmp:       true,
		},
		Install: InstallSection{
			WantedBy: "multi-user.target",
		},
	}
}

func GenerateServiceFile(unit SystemdUnit, filePath string) error {
	var sb strings.Builder

	// 写入 [Unit] 部分
	sb.WriteString("[Unit]\n")
	sb.WriteString(fmt.Sprintf("Description=%s\n", unit.Unit.Description))
	sb.WriteString(fmt.Sprintf("Documentation=%s\n", unit.Unit.Documentation))
	sb.WriteString(fmt.Sprintf("After=%s\n\n", unit.Unit.After))

	// 写入 [Service] 部分
	sb.WriteString("[Service]\n")
	sb.WriteString(fmt.Sprintf("Type=%s\n", unit.Service.Type))
	sb.WriteString(fmt.Sprintf("Restart=%s\n", unit.Service.Restart))
	sb.WriteString(fmt.Sprintf("RestartSec=%s\n", unit.Service.RestartSec))
	for _, env := range unit.Service.Environment {
		sb.WriteString(fmt.Sprintf("Environment=%s\n", env))
	}
	sb.WriteString(fmt.Sprintf("WorkingDirectory=%s\n", unit.Service.WorkingDirectory))
	sb.WriteString(fmt.Sprintf("ExecStart=%s\n", unit.Service.ExecStart))
	sb.WriteString(fmt.Sprintf("ExecReload=%s\n", unit.Service.ExecReload))
	sb.WriteString(fmt.Sprintf("ExecStop=%s\n", unit.Service.ExecStop))
	sb.WriteString(fmt.Sprintf("PrivateTmp=%t\n\n", unit.Service.PrivateTmp))

	// 写入 [Install] 部分
	sb.WriteString("[Install]\n")
	sb.WriteString(fmt.Sprintf("WantedBy=%s\n", unit.Install.WantedBy))

	// 将服务文件写入指定路径
	return os.WriteFile(filePath, []byte(sb.String()), 0644)
}

func isSystemdSupported() bool {
	// 检查 systemd 的二进制文件是否存在
	paths := []string{
		"/usr/lib/systemd/systemd",
		"/bin/systemd",
		"/sbin/systemd",
		"/lib/systemd/systemd",
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}
	out, err := exec.Command("ps", "-p", "1", "-o", "comm=").Output()
	if err != nil {
		return false
	}
	if strings.TrimSpace(string(out)) == "systemd" {
		return true
	}
	return false
}

func (s *SystemdService) Register(serviceName, description, documentation string) error {
	fmt.Println("isSystemdSupported:", isSystemdSupported())
	if isSystemdSupported() == false {
		fmt.Println("Systemd is not supported.")
		return nil
	}
	exists, err := s.Exists(serviceName)
	if err != nil {
		fmt.Printf("Error: checking service existence: %v\n", err)
		return nil
	}

	if exists {
		fmt.Printf("Service: Service %s exists.\n", serviceName)
		return nil
	} else {
		fmt.Printf("Service: Service %s does not exist.\n", serviceName)
	}
	exePath, err := os.Executable()
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}

	s.unit.Unit.Description = description
	s.unit.Unit.Documentation = documentation

	exeDir := filepath.Dir(exePath)
	fmt.Println("Executable Path:", exePath)
	s.unit.Service.ExecStart = exePath
	fmt.Println("Executable Directory:", exeDir)
	s.unit.Service.WorkingDirectory = exeDir
	fmt.Println("Executable Name:", serviceName)
	// /usr/lib/systemd/system/ 多
	// /etc/systemd/system/ 少
	serviceFile := fmt.Sprintf("/etc/systemd/system/%s.service", serviceName)
	fmt.Println("Service File:", serviceFile)
	if err := GenerateServiceFile(s.unit, serviceFile); err != nil {
		fmt.Println("Error:", err)
		return err
	}
	fmt.Println("Service File: Generated")

	cmd := exec.Command("systemctl", "daemon-reload")
	err = cmd.Run()
	if err != nil {
		return err
	}

	cmd = exec.Command("systemctl", "enable", serviceName)
	err = cmd.Run()
	if err != nil {
		return err
	}

	cmd = exec.Command("systemctl", "start", serviceName)
	return cmd.Run()
}

func (s SystemdService) Start(serviceName string) error {
	if isSystemdSupported() == false {
		fmt.Println("Systemd is not supported.")
		return nil
	}
	exists, err := s.Exists(serviceName)
	if err != nil {
		fmt.Printf("Error: checking service existence: %v\n", err)
		return nil
	}

	if exists {
		fmt.Printf("Service: Service %s exists.\n", serviceName)
	} else {
		fmt.Printf("Service: Service %s does not exist.\n", serviceName)
		return nil
	}
	// Start the service
	cmd := exec.Command("systemctl", "start", serviceName)
	if err := cmd.Run(); err != nil {
		fmt.Println("Error starting service:", err)
		return err
	} else {
		fmt.Println("Service started.")
	}
	return nil
}

func (s SystemdService) Stop(serviceName string) error {
	if isSystemdSupported() == false {
		fmt.Println("Systemd is not supported.")
		return nil
	}
	exists, err := s.Exists(serviceName)
	if err != nil {
		fmt.Printf("Error: checking service existence: %v\n", err)
		return nil
	}
	if exists {
		fmt.Printf("Service: Service %s exists.\n", serviceName)
	} else {
		fmt.Printf("Service: Service %s does not exist.\n", serviceName)
		return nil
	}

	active, err := s.IsServiceActive(serviceName)
	if err != nil {
		fmt.Printf("Error checking service status: %v\n", err)
		return err
	}
	if active {
		fmt.Printf("Service %s is running.\n", serviceName)
		// Stop the service
		cmd := exec.Command("systemctl", "stop", serviceName)
		if err := cmd.Run(); err != nil {
			fmt.Println("Error stopping service:", err)
			return err
		} else {
			fmt.Println("Service stopped.")
		}
	} else {
		fmt.Printf("Service %s is not running.\n", serviceName)
	}
	return nil
}

func (s SystemdService) Remove(serviceName string) error {
	if isSystemdSupported() == false {
		fmt.Println("Systemd is not supported.")
		return nil
	}
	cmd := exec.Command("systemctl", "stop", serviceName)
	err := cmd.Run()
	if err != nil {
		return err
	}
	cmd = exec.Command("systemctl", "disable", serviceName)
	err = cmd.Run()
	if err != nil {
		return err
	}
	servicePath := fmt.Sprintf("/etc/systemd/system/%s.service", serviceName)
	err = os.Remove(servicePath)
	if err != nil {
		return err
	}
	cmd = exec.Command("systemctl", "daemon-reload")
	return cmd.Run()
}

func (s SystemdService) IsServiceActive(serviceName string) (bool, error) {
	cmd := exec.Command("systemctl", "is-active", serviceName)
	output, err := cmd.Output()
	if err != nil {
		return false, err // Return error if the command fails
	}
	status := strings.TrimSpace(string(output)) // Remove any trailing whitespace
	return status == "active", nil              // Return true if status is "active"
}

func (s SystemdService) IsServiceExists(serviceName string) (bool, error) {
	cmd := exec.Command("systemctl", "status", serviceName)
	err := cmd.Run()
	if err != nil {
		// If the command fails, check if the error is due to service not existing
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() == 3 {
				// Exit code 3 means the service does not exist
				return false, nil
			}
		}
		return false, err
	}
	return true, nil
}

func (s SystemdService) Exists(serviceName string) (bool, error) {
	serviceName = serviceName + ".service"
	cmd := exec.Command("systemctl", "list-unit-files", "--type=service")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return false, fmt.Errorf("failed to execute command: %v", err)
	}
	if strings.Contains(out.String(), serviceName) {
		return true, nil
	}
	return false, nil
}

func (s SystemdService) Status(serviceName string) error {
	if isSystemdSupported() == false {
		fmt.Println("Systemd is not supported.")
		return nil
	}
	cmd := exec.Command("systemctl", "status", serviceName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode := exitError.ExitCode()
			if exitCode == 3 {
				fmt.Println("Service is not running (inactive).")
			} else {
				fmt.Printf("Error getting service status: %v\n", err)
			}
		}
	} else {
		fmt.Println("Service is running")
	}
	fmt.Println(string(output))
	return nil
}

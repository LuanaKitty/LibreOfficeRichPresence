package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/hugolgst/rich-go/client"
)

// ID do aplicativo Discord (https://discord.com/developers/applications)
const ClientID = "1441894223765835798"

type AppInfo struct {
	Name string
	Icon string
	Desc string
}

var apps = map[string]AppInfo{
	"soffice.bin": {Name: "LibreOffice", Icon: "libreoffice", Desc: "LibreOffice"},
	"swriter":     {Name: "Writer", Icon: "writer", Desc: "Documento de Texto"},
	"scalc":       {Name: "Calc", Icon: "calc", Desc: "Planilha"},
	"simpress":    {Name: "Impress", Icon: "impress", Desc: "Apresentação"},
	"sdraw":       {Name: "Draw", Icon: "draw", Desc: "Desenho"},
	"sbase":       {Name: "Base", Icon: "base", Desc: "Banco de Dados"},
	"smath":       {Name: "Math", Icon: "math", Desc: "Fórmula"},
}

type WindowInfo struct {
	Title   string
	Process string
	PID     string
}

func runCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func getActiveWindowInfo() (*WindowInfo, error) {
	// ID da janela ativa
	winID, err := runCommand("xdotool", "getactivewindow")
	if err != nil {
		return nil, err
	}

	//nome da janela
	winName, err := runCommand("xdotool", "getwindowname", winID)
	if err != nil {
		return nil, err
	}

	//PID do processo
	winPID, err := runCommand("xdotool", "getwindowpid", winID)
	if err != nil {
		return nil, err
	}

	//nome do processo
	procName, err := runCommand("ps", "-p", winPID, "-o", "comm=")
	if err != nil {
		return nil, err
	}

	return &WindowInfo{
		Title:   winName,
		Process: procName,
		PID:     winPID,
	}, nil
}

func detectLibreOfficeApp(windowTitle string) string {
	titleLower := strings.ToLower(windowTitle)

	switch {
	case strings.Contains(titleLower, "writer") ||
		strings.Contains(titleLower, ".odt") ||
		strings.Contains(titleLower, ".doc"):
		return "swriter"
	case strings.Contains(titleLower, "calc") ||
		strings.Contains(titleLower, ".ods") ||
		strings.Contains(titleLower, ".xls"):
		return "scalc"
	case strings.Contains(titleLower, "impress") ||
		strings.Contains(titleLower, ".odp") ||
		strings.Contains(titleLower, ".ppt"):
		return "simpress"
	case strings.Contains(titleLower, "draw") ||
		strings.Contains(titleLower, ".odg"):
		return "sdraw"
	case strings.Contains(titleLower, "base") ||
		strings.Contains(titleLower, ".odb"):
		return "sbase"
	case strings.Contains(titleLower, "math") ||
		strings.Contains(titleLower, ".odf"):
		return "smath"
	default:
		return "soffice.bin"
	}
}

func extractFilename(windowTitle string) string {
	patterns := []string{
		`^(.*?)\s*[-–—]\s*LibreOffice\s+\w+$`,
		`^(.*?)\s*[-–—]\s*LibreOffice.*$`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(windowTitle)
		if len(matches) > 1 {
			return strings.TrimSpace(matches[1])
		}
	}

	return windowTitle
}

func truncate(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen-3] + "..."
	}
	return s
}

func isLibreOfficeProcess(processName string) bool {
	lower := strings.ToLower(processName)
	return strings.Contains(lower, "soffice") ||
		strings.Contains(lower, "libreoffice")
}

func main() {
	fmt.Println("UwU Iniciando Discord Rich Presence para LibreOffice...")

	// Conecta ao Discord
	err := client.Login(ClientID)
	if err != nil {
		fmt.Printf("XnX Erro ao conectar ao Discord: %v\n", err)
		fmt.Println("   Certifique-se de que o Discord está aberto!!!.")
		os.Exit(1)
	}
	fmt.Println("0u0 Conectado ao Discord!")

	// Configura handler para encerramento gracioso
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	startTime := time.Now()
	var lastState string
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Verifica imediatamente na primeira execução
	checkAndUpdate := func() {
		windowInfo, err := getActiveWindowInfo()
		if err != nil {
			return
		}

		if isLibreOfficeProcess(windowInfo.Process) {
			appKey := detectLibreOfficeApp(windowInfo.Title)
			appInfo, exists := apps[appKey]
			if !exists {
				appInfo = apps["soffice.bin"]
			}

			filename := extractFilename(windowInfo.Title)
			state := fmt.Sprintf("%s:%s", appKey, filename)

			// Só atualiza se mudou
			if state != lastState {
				fmt.Printf(">.> Detectado: %s - %s\n", appInfo.Name, filename)

				displayFilename := truncate(filename, 50)
				stateText := fmt.Sprintf("XP Editando: %s", displayFilename)

				err := client.SetActivity(client.Activity{
					State:      stateText,
					Details:    appInfo.Desc,
					LargeImage: appInfo.Icon,
					LargeText:  fmt.Sprintf("LibreOffice %s", appInfo.Name),
					SmallImage: "libreoffice",
					SmallText:  "LibreOffice",
					Timestamps: &client.Timestamps{
						Start: &startTime,
					},
				})

				if err != nil {
					fmt.Printf("XnX Erro ao atualizar status: %v\n", err)
				}
				lastState = state
			}
		} else {
			// LibreOffice não está em foco
			if lastState != "" {
				fmt.Println("TwT  LibreOffice não está em foco")
				client.SetActivity(client.Activity{})
				lastState = ""
			}
		}
	}

	// Executa verificação inicial
	checkAndUpdate()

	// Loop principal
	for {
		select {
		case <-ticker.C:
			checkAndUpdate()
		case <-sigChan:
			fmt.Println("\n Encerrando...:3")
			client.Logout()
			os.Exit(0)
		}
	}
}

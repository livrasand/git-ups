package cli

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/livrasand/git-ups/internal/jobs"
)

var version = "0.1.0"

func Run(args []string) int {
	if len(args) == 0 {
		printUsage()
		return 0
	}

	switch args[0] {
	case "install":
		return cmdInstall()
	case "clone":
		return cmdClone(args[1:])
	case "fetch":
		return cmdGitJob("fetch", args[1:])
	case "pull":
		return cmdGitJob("pull", args[1:])
	case "push":
		return cmdGitJob("push", args[1:])
	case "jobs":
		return cmdJobs(args[1:])
	case "watch":
		return cmdWatch(args[1:])
	case "pause":
		return cmdPause(args[1:])
	case "resume":
		return cmdResume(args[1:])
	case "cancel":
		return cmdCancel(args[1:])
	case "run":
		return cmdRun(args[1:])
	case "help", "-h", "--help":
		printUsage()
		return 0
	case "version", "--version", "-v":
		fmt.Printf("git-ups %s\n", version)
		return 0
	default:
		return passthrough(args)
	}
}

func passthrough(args []string) int {
	cmd := exec.Command("git", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return ee.ExitCode()
		}
		fmt.Fprintf(os.Stderr, "git-ups: %v\n", err)
		return 1
	}
	return 0
}

func dataDir() string {
	if v := os.Getenv("GITGOST_HOME"); v != "" {
		return v
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ".gitgost"
	}
	return filepath.Join(home, ".gitgost")
}

func dbPath() string {
	return filepath.Join(dataDir(), "gitgost.db")
}

func openStore() (*jobs.Store, error) {
	return jobs.Open(dbPath())
}

func cmdInstall() int {
	s, err := openStore()
	if err != nil {
		fmt.Fprintf(os.Stderr, "git-ups: installation failed: %v\n", err)
		return 1
	}
	s.Close()

	if err := selfInstall(); err != nil {
		fmt.Fprintf(os.Stderr, "aviso: %v\n", err)
	}

	fmt.Println("git ups install: OK")
	fmt.Printf("Job queue: %s\n", dbPath())
	fmt.Println("Uso: git ups clone <url>  |  git ups jobs  |  git ups watch <id>")
	return 0
}

// installedNames son los nombres de subcomando git que expone el binario:
// installedNames are the command names the binary is installed under: `git ups`
// (git-ups) and `git gost` (git-gost, kept for compatibility). Both run the same binary.
var installedNames = []string{"git-gost", "git-ups"}

func selfInstall() error {
	allOnPath := true
	for _, name := range installedNames {
		if _, err := exec.LookPath(name); err != nil {
			allOnPath = false
			break
		}
	}
	if allOnPath {
		return nil
	}

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("locate running binary: %w", err)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("locate home directory: %w", err)
	}

	binDir := filepath.Join(home, ".local", "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		return fmt.Errorf("create %s: %w", binDir, err)
	}

	for _, name := range installedNames {
		binName := name
		if runtime.GOOS == "windows" {
			binName += ".exe"
		}
		target := filepath.Join(binDir, binName)

		if exe == target {
			continue
		}
		if err := copyFile(exe, target); err != nil {
			return fmt.Errorf("copy binary to %s: %w", target, err)
		}
		if err := os.Chmod(target, 0o755); err != nil {
			return fmt.Errorf("set permissions on %s: %w", target, err)
		}
		fmt.Printf("%s installed at %s\n", name, target)
	}

	if runtime.GOOS == "windows" {
		fmt.Printf("Add %s to your PATH to use 'git ups ...'.\n", binDir)
		return nil
	}
	return addToPath(binDir, home)
}

func addToPath(dir, home string) error {
	shell := filepath.Base(os.Getenv("SHELL"))
	var rc, line string
	switch shell {
	case "fish":
		rc = filepath.Join(home, ".config", "fish", "config.fish")
		line = fmt.Sprintf("fish_add_path %q", dir)
	case "bash":
		rc = filepath.Join(home, ".bashrc")
		if _, err := os.Stat(filepath.Join(home, ".bash_profile")); err == nil {
			rc = filepath.Join(home, ".bash_profile")
		}
		line = fmt.Sprintf(`export PATH=%q:$PATH`, dir)
	default:
		rc = filepath.Join(home, ".zshrc")
		line = fmt.Sprintf(`export PATH=%q:$PATH`, dir)
	}

	data, err := os.ReadFile(rc)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if strings.Contains(string(data), dir) {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(rc), 0o755); err != nil {
		return fmt.Errorf("create %s: %w", filepath.Dir(rc), err)
	}
	f, err := os.OpenFile(rc, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := fmt.Fprintf(f, "\n# git-ups\n%s\n", line); err != nil {
		return err
	}
	fmt.Printf("Added %s to PATH in %s (reload your shell or run: source %s)\n", dir, rc, rc)
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}

func cmdClone(args []string) int {
	foreground, rest, err := parseFlags(args)
	if err != nil {
		return 1
	}
	if len(rest) == 0 {
		fmt.Fprintln(os.Stderr, "usage: git ups clone <url> [directory] [-f]")
		return 1
	}
	url, target := rest[0], ""
	if len(rest) > 1 {
		target = rest[1]
	}
	if target == "" {
		target = defaultCloneDir(url)
	}

	base, err := ServerBase()
	if err != nil {
		fmt.Fprintf(os.Stderr, "git-ups: %v\n", err)
		return 1
	}
	rewritten, err := RewriteURL(base, url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "git-ups: %v\n", err)
		return 1
	}

	s, err := openStore()
	if err != nil {
		fmt.Fprintf(os.Stderr, "git-ups: %v\n", err)
		return 1
	}
	defer s.Close()

	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}
	id, err := s.Create(&jobs.Job{
		Operation: "clone",
		URL:       rewritten,
		Origin:    url,
		Target:    target,
		CWD:       cwd,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "git-ups: create job: %v\n", err)
		return 1
	}

	if foreground {
		return runForeground(s, id)
	}
	return launchBackground(s, id)
}

func defaultCloneDir(raw string) string {
	u, err := parseRepoURL(raw)
	if err != nil {
		return "repo"
	}
	parts := splitPath(u.Path)
	if len(parts) == 0 {
		return "repo"
	}
	return strings.TrimSuffix(parts[len(parts)-1], ".git")
}

func cmdGitJob(op string, args []string) int {
	foreground, rest, err := parseFlags(args)
	if err != nil {
		return 1
	}

	s, err := openStore()
	if err != nil {
		fmt.Fprintf(os.Stderr, "git-ups: %v\n", err)
		return 1
	}
	defer s.Close()

	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}
	id, err := s.Create(&jobs.Job{
		Operation: op,
		Args:      rest,
		CWD:       cwd,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "git-ups: create job: %v\n", err)
		return 1
	}

	if foreground {
		return runForeground(s, id)
	}
	return launchBackground(s, id)
}

func parseFlags(args []string) (foreground bool, rest []string, err error) {
	for _, a := range args {
		switch a {
		case "-f", "--foreground":
			foreground = true
		default:
			rest = append(rest, a)
		}
	}
	return foreground, rest, nil
}

func runForeground(s *jobs.Store, id int64) int {
	if err := jobs.Run(s, id); err != nil {
		job, gerr := s.Get(id)
		if gerr == nil && job.Error != "" {
			fmt.Fprintf(os.Stderr, "git-ups: %s\n", job.Error)
		} else {
			fmt.Fprintf(os.Stderr, "git-ups: %v\n", err)
		}
		return 1
	}
	fmt.Printf("Job %d: completed\n", id)
	return 0
}

func launchBackground(s *jobs.Store, id int64) int {
	pid, err := startBackground(id)
	if err != nil {
		_ = s.SetState(id, jobs.StateFailed)
		_ = s.SetError(id, err.Error())
		fmt.Fprintf(os.Stderr, "git-ups: %v\n", err)
		return 1
	}
	_ = s.SetPID(id, pid)
	fmt.Printf("Job created.\n\nID: %d\nRunning in background...\n\nUse 'git ups watch %d' to follow progress.\n", id, id)
	return 0
}

func cmdRun(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "internal usage: git ups run <id>")
		return 1
	}
	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "git-ups: invalid id: %s\n", args[0])
		return 1
	}
	s, err := openStore()
	if err != nil {
		fmt.Fprintf(os.Stderr, "git-ups: %v\n", err)
		return 1
	}
	defer s.Close()
	if err := jobs.Run(s, id); err != nil {
		fmt.Fprintf(os.Stderr, "git-ups: %v\n", err)
		return 1
	}
	return 0
}

func cmdJobs(args []string) int {
	s, err := openStore()
	if err != nil {
		fmt.Fprintf(os.Stderr, "git-ups: %v\n", err)
		return 1
	}
	defer s.Close()

	limit := 50
	for _, a := range args {
		if n, err := strconv.Atoi(a); err == nil && n > 0 {
			limit = n
		}
	}

	list, err := s.List(limit)
	if err != nil {
		fmt.Fprintf(os.Stderr, "git-ups: %v\n", err)
		return 1
	}

	fmt.Printf("%-6s %-10s %-10s %s\n", "ID", "STATUS", "OPERATION", "TARGET")
	for _, j := range list {
		target := j.Target
		if target == "" {
			if j.URL != "" {
				target = j.URL
			} else {
				target = "."
			}
		}
		fmt.Printf("%-6d %-10s %-10s %s\n", j.ID, strings.ToUpper(j.State), j.Operation, target)
	}
	return 0
}

func cmdWatch(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: git ups watch <id>")
		return 1
	}
	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "git-ups: invalid id: %s\n", args[0])
		return 1
	}
	s, err := openStore()
	if err != nil {
		fmt.Fprintf(os.Stderr, "git-ups: %v\n", err)
		return 1
	}
	defer s.Close()

	last := ""
	for {
		job, err := s.Get(id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "git-ups: %v\n", err)
			return 1
		}
		if job.Progress != last {
			fmt.Printf("[%s] %s\n", strings.ToUpper(job.State), job.Progress)
			last = job.Progress
		}
		switch job.State {
		case jobs.StateCompleted:
			fmt.Printf("[COMPLETED] Job %d finished\n", id)
			return 0
		case jobs.StateFailed:
			fmt.Printf("[FAILED] %s\n", job.Error)
			return 0
		case jobs.StateCancelled:
			fmt.Printf("[CANCELLED] Job %d cancelled\n", id)
			return 0
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func cmdPause(args []string) int {
	return signalCommand("pause", args, func(j *jobs.Job, s *jobs.Store) (int, error) {
		switch j.State {
		case jobs.StateQueued, jobs.StateRetrying:
			return 0, s.SetState(j.ID, jobs.StatePaused)
		case jobs.StateRunning:
			if err := signalJob(j.PID, sigStop); err != nil {
				return 0, err
			}
			return 0, s.SetState(j.ID, jobs.StatePaused)
		default:
			return 0, fmt.Errorf("job %d cannot be paused (state: %s)", j.ID, j.State)
		}
	})
}

func cmdResume(args []string) int {
	return signalCommand("resume", args, func(j *jobs.Job, s *jobs.Store) (int, error) {
		if j.State != jobs.StatePaused {
			return 0, fmt.Errorf("job %d is not paused (state: %s)", j.ID, j.State)
		}
		if j.PID > 0 {
			if err := signalJob(j.PID, sigCont); err != nil {
				return 0, err
			}
			return 0, s.SetState(j.ID, jobs.StateRunning)
		}
		pid, err := startBackground(j.ID)
		if err != nil {
			return 0, err
		}
		_ = s.SetPID(j.ID, pid)
		return 0, s.SetState(j.ID, jobs.StateQueued)
	})
}

func cmdCancel(args []string) int {
	return signalCommand("cancel", args, func(j *jobs.Job, s *jobs.Store) (int, error) {
		switch j.State {
		case jobs.StateQueued, jobs.StateRetrying, jobs.StatePaused:
			_ = s.SetError(j.ID, "cancelled")
			return 0, s.SetState(j.ID, jobs.StateCancelled)
		case jobs.StateRunning:
			_ = signalJob(j.PID, sigTerm)
			_ = s.SetError(j.ID, "cancelled")
			return 0, s.SetState(j.ID, jobs.StateCancelled)
		default:
			return 0, fmt.Errorf("job %d already finished (%s)", j.ID, j.State)
		}
	})
}

func signalCommand(name string, args []string, action func(*jobs.Job, *jobs.Store) (int, error)) int {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "usage: git ups %s <id>\n", name)
		return 1
	}
	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "git-ups: invalid id: %s\n", args[0])
		return 1
	}
	s, err := openStore()
	if err != nil {
		fmt.Fprintf(os.Stderr, "git-ups: %v\n", err)
		return 1
	}
	defer s.Close()

	j, err := s.Get(id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "git-ups: %v\n", err)
		return 1
	}
	if _, err := action(j, s); err != nil {
		fmt.Fprintf(os.Stderr, "git-ups: %v\n", err)
		return 1
	}
	fmt.Printf("Job %d: %s\n", id, name)
	return 0
}

func printUsage() {
	fmt.Print(`git-ups: a resilient network layer for Git (Phase 0).

Usage:
  git ups clone <url> [directory]    Clone via the gitGost server (background by default)
  git ups fetch [args...]            Create a fetch job in the current repo
  git ups pull [args...]             Create a pull job in the current repo
  git ups push [args...]             Create a push job in the current repo
  git ups jobs [n]                   List the last n jobs (default 50)
  git ups watch <id>                 Follow a job's progress
  git ups pause <id>                 Pause a job
  git ups resume <id>                Resume a job
  git ups cancel <id>                Cancel a job
  git ups install                    Prepare the client environment
  git ups version                    Show version

Any other command (status, log, branch, ...) is passed through to Git.

Opciones de jobs:
  -f, --foreground   Ejecutar en primer plano mostrando el progreso en vivo

Variables de entorno:
  GITGOST_SERVER   URL base del servidor gitGost (por defecto https://gitgost.livrasand.com)
  GITGOST_HOME     Directorio de datos (por defecto ~/.gitgost)
`)
}

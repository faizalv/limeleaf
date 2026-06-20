package limeleaf

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

func isFirstRun(dataDir string) bool {
	_, err := os.Stat(filepath.Join(dataDir, "PG_VERSION"))
	return os.IsNotExist(err)
}

func initDB(ctx context.Context, binDir, dataDir, username string, logger *log.Logger) error {
	cmd := exec.CommandContext(ctx,
		filepath.Join(binDir, "bin", "initdb"),
		"-D", dataDir,
		"--no-locale",
		"--encoding=UTF8",
		"--auth=trust",
		"--username="+username,
	)
	cmd.Env = pgEnv(binDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("initdb: %w\n%s", err, output)
	}
	logger.Printf("initdb completed: %s", dataDir)
	return nil
}

func startPostgres(ctx context.Context, binDir, dataDir string, logger *log.Logger) error {
	logFile := filepath.Join(dataDir, "postgres.log")
	cmd := exec.CommandContext(ctx,
		filepath.Join(binDir, "bin", "pg_ctl"),
		"start",
		"-D", dataDir,
		"-l", logFile,
		"-w",
	)
	cmd.Env = pgEnv(binDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("pg_ctl start: %w\n%s", err, output)
	}
	logger.Printf("postgres started on datadir %s", dataDir)
	return nil
}

func stopPostgres(binDir, dataDir string, logger *log.Logger) error {
	cmd := exec.Command(
		filepath.Join(binDir, "bin", "pg_ctl"),
		"stop",
		"-D", dataDir,
		"-m", "fast",
	)
	cmd.Env = pgEnv(binDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("pg_ctl stop: %w\n%s", err, output)
	}
	logger.Printf("postgres stopped: %s", dataDir)
	return nil
}

func cleanStalePid(dataDir string) error {
	pidFile := filepath.Join(dataDir, "postmaster.pid")
	data, err := os.ReadFile(pidFile)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("reading pid file: %w", err)
	}

	lines := strings.SplitN(string(data), "\n", 2)
	if len(lines) == 0 {
		return nil
	}

	pid, err := strconv.Atoi(strings.TrimSpace(lines[0]))
	if err != nil {
		return os.Remove(pidFile)
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return os.Remove(pidFile)
	}

	if err := proc.Signal(syscall.Signal(0)); err != nil {
		os.Remove(pidFile)
		return nil
	}

	return fmt.Errorf("postgres process %d is still running on datadir %s", pid, dataDir)
}

func runSetupSQL(ctx context.Context, binDir string, port int, username, database, sql string) error {
	cmd := exec.CommandContext(ctx,
		filepath.Join(binDir, "bin", "psql"),
		"-h", "127.0.0.1",
		"-p", strconv.Itoa(port),
		"-U", username,
		"-d", database,
		"-c", sql,
	)
	cmd.Env = pgEnv(binDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("psql: %w\n%s", err, output)
	}
	return nil
}

func pgEnv(binDir string) []string {
	env := os.Environ()
	env = append(env,
		"LD_LIBRARY_PATH="+filepath.Join(binDir, "lib"),
		"DYLD_LIBRARY_PATH="+filepath.Join(binDir, "lib"),
	)
	return env
}

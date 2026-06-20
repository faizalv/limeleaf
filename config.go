package limeleaf

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func writePostgresConf(dataDir string, port int, settings map[string]string) error {
	defaults := map[string]string{
		"listen_addresses":        "'127.0.0.1'",
		"port":                    fmt.Sprintf("%d", port),
		"unix_socket_directories": "''",
		"shared_buffers":          "'128MB'",
		"max_connections":         "20",
		"log_destination":         "'stderr'",
		"logging_collector":       "off",
		"fsync":                   "on",
		"synchronous_commit":      "on",
	}

	for k, v := range settings {
		defaults[k] = v
	}

	keys := make([]string, 0, len(defaults))
	for k := range defaults {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	for _, k := range keys {
		fmt.Fprintf(&b, "%s = %s\n", k, defaults[k])
	}

	return os.WriteFile(filepath.Join(dataDir, "postgresql.conf"), []byte(b.String()), 0600)
}

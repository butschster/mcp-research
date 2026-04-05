package config

import (
	"flag"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Transport string `yaml:"transport"`
	MCPPort   int    `yaml:"mcp_port"`
	WebPort   int    `yaml:"web_port"`
	DBPath    string `yaml:"db"`
	LogLevel  string `yaml:"log_level"`
	Version   bool   `yaml:"-"`
}

// Load reads config from config file (if present), then env vars, then CLI flags.
// Priority: flags > env > yaml > defaults.
func Load() Config {
	cfg := Config{
		Transport: "stdio",
		MCPPort:   8081,
		WebPort:   8088,
		LogLevel:  "info",
	}

	// Pre-parse --config flag (needs to be read before other flags)
	configPath := "config.yaml"
	if p := os.Getenv("MCP_RESEARCH_CONFIG"); p != "" {
		configPath = p
	}
	for i, arg := range os.Args[1:] {
		if arg == "--config" && i+1 < len(os.Args)-1 {
			configPath = os.Args[i+2]
			break
		}
		if len(arg) > 9 && arg[:9] == "--config=" {
			configPath = arg[9:]
			break
		}
	}

	// 1. Load from config file if it exists
	if data, err := os.ReadFile(configPath); err == nil {
		_ = yaml.Unmarshal(data, &cfg)
	}

	// 2. Override with env vars
	if v := os.Getenv("MCP_RESEARCH_TRANSPORT"); v != "" {
		cfg.Transport = v
	}
	if v := os.Getenv("MCP_RESEARCH_DB"); v != "" {
		cfg.DBPath = v
	}
	if v := os.Getenv("MCP_RESEARCH_LOG_LEVEL"); v != "" {
		cfg.LogLevel = v
	}

	// 3. Override with CLI flags (highest priority)
	var configFlag string
	flag.StringVar(&configFlag, "config", configPath, "path to config.yaml")
	flag.StringVar(&cfg.Transport, "transport", cfg.Transport, "transport: stdio or sse")
	flag.IntVar(&cfg.MCPPort, "mcp-port", cfg.MCPPort, "MCP SSE port")
	flag.IntVar(&cfg.WebPort, "web-port", cfg.WebPort, "Web/API port")
	flag.StringVar(&cfg.DBPath, "db", cfg.DBPath, "SQLite database path (empty = in-memory)")
	flag.StringVar(&cfg.LogLevel, "log-level", cfg.LogLevel, "log level: debug, info, warn, error")
	flag.BoolVar(&cfg.Version, "version", false, "print version and exit")

	flag.Parse()
	return cfg
}

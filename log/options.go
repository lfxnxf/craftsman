package log

// log
type Log struct {
	Level   string `toml:"level"`
	Rotate  string `toml:"rotate"`
	LogPath string `toml:"logpath"`
}

package inits

import (
	"fmt"
	"testing"
	"time"
)

type cfg struct {
	Age                int       `toml:"age"`
	Cats               []string  `toml:"cats"`
	Pi                 float64   `toml:"pi"`
	Perfection         []int     `toml:"perfection"`
	DOB                time.Time `toml:"dob"`
	StsAccessKeyID     string    `toml:"STS_ACCESS_KEY_ID"`
	StsAccessKeySecret string    `toml:"STS_ACCESS_KEY_SECRET"`
	StsIndexName       string    `toml:"STS_INDEX_NAME"`
}

func TestNewServer(t *testing.T) {
	var c cfg

	s := NewServer(&c)

	fmt.Println(s)
}
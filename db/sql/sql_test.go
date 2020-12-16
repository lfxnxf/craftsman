package sql

import (
	"fmt"
	"os"
	"testing"

)

type groupConfig struct {
	Databases []struct {
		Name   string   `toml:"name"`
		Master string   `toml:"master"`
		Slaves []string `toml:"slaves"`
	} `toml:"database"`
}

type AllServerPolicy struct {
	ServerName      string `gorm:"server_name"`
	ServerType      int    `gorm:"server_type"`
	ServerURL       string `gorm:"server_url"`
	SecureServerURL string `gorm:"secure_server_url"`
}

type User struct {
	UID      int    `json:"uid"`
	UserName string `json:"user_name"`
	GroupId  int    `json:"group_id"`
}

var groupManager *GroupManager

func setUp() {
	config := "./sql_config.toml"
	var gc groupConfig
	err := tomlconfig.ParseTomlConfig(config, &gc)
	if err != nil {
		os.Exit(1)
	}
	groupManager = newGroupManager()
	for _, d := range gc.Databases {
		g, err := NewGroup(SQLGroupConfig{
			Name:   d.Name,
			Master: d.Master,
			Slaves: d.Slaves,
		})
		if err != nil {
			os.Exit(1)
		}
		err = groupManager.Add(d.Name, g)
		if err != nil {
			os.Exit(1)
		}
	}
}

func TestGroupCreate(t *testing.T) {
	setUp()
	if m := groupManager.Get("test1").Master(); m == nil {

	}
	groupManager.Get("test1").Slave()
	groupManager.Get("test2").Master()
	groupManager.Get("test2").Slave()
}

func TestQuery(t *testing.T) {
	setUp()
	var all []User
	err := groupManager.Get("test").Master().Table("user").Find(&all).Error
	if err != nil {
		os.Exit(1)
	}
	fmt.Println(all)
}

func TestPartition(t *testing.T) {
	setUp()
	var all []*AllServerPolicy
	p := func() (bool, string, string) {
		return true, "test1", "all_server_policy2"
	}
	err := groupManager.PartitionBy(p).Find(&all).Error
	if err != nil {
		os.Exit(1)
	}
}

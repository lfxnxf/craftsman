package inits

import (
	toml "github.com/tiantianjianbao/craftsman/config/toml"
)

type MetaConfig struct {
	meta *toml.MetaData
}

var (
	Integer   = "Integer"
	Float     = "Float"
	Datetime  = "Datetime"
	String    = "String"
	Bool      = "Bool"
	Array     = "Array"
	Hash      = "Hash"
	ArrayHash = "ArrayHash"
)

func ParseTomlString(data string, v interface{}) error {
	_, err := toml.Decode(data, v)
	if err != nil {
		return err
	}
	return nil
}

func ParseTomlConfig(filepath string, v interface{}) error {
	_, err := toml.DecodeFile(filepath, v)
	if err != nil {
		return err
	}
	return nil
}

func NewConfig(filepath string, v interface{}) (*toml.Config, error) {
	metaData, err := toml.DecodeFile(filepath, v)
	if err != nil {
		return nil, err
	}
	return &toml.Config{
		Meta: metaData,
	}, nil
}

func NewTomlConfig(filepath string) (*toml.Config, error) {
	var v interface{}
	metaData, err := toml.DecodeFile(filepath, &v)
	if err != nil {
		return nil, err
	}

	return &toml.Config{
		Meta: metaData,
	}, nil
}

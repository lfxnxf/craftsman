package toml

type Config struct {
	Meta MetaData
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
	_, err := Decode(data, v)
	if err != nil {
		return err
	}
	return nil
}

func ParseTomlConfig(filepath string, v interface{}) error {
	_, err := DecodeFile(filepath, v)

	if err != nil {
		return err
	}
	return nil
}

func NewConfig(filepath string, v interface{}) (*Config, error) {
	metaData, err := DecodeFile(filepath, v)
	if err != nil {
		return nil, err
	}
	return &Config{
		Meta: metaData,
	}, nil
}

func NewTomlConfig(filepath string) (*Config, error) {
	var v interface{}
	metaData, err := DecodeFile(filepath, &v)
	if err != nil {
		return nil, err
	}

	return &Config{
		Meta: metaData,
	}, nil
}

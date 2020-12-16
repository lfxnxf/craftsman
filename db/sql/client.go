package sql

func InitSQLClient(sqlConfig []SQLGroupConfig) error {
	if len(sqlConfig) > 0 {
		for _, d := range sqlConfig {
			g, err := NewGroup(d, nil)
			if err != nil {
				return err
			}
			err = SQLGroupManager.Add(d.Name, g)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

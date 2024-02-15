package database

import (
	"encoding/json"
	"os"
)

type genesis struct {
	Balances map[Account]uint `json:"balances"`
}

func loadGenesis(path string) (genesis, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return genesis{}, err
	}

	var loadedGenesis genesis
	if err = json.Unmarshal(content, &loadedGenesis); err != nil {
		return genesis{}, err
	}

	return loadedGenesis, nil
}

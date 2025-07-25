package migrate

import (
	"bufio"
	"errors"
	"os"
	"strconv"
)

func (m *Migrate) getVersion() (int, error) {
	path := m.path + "/migrate.log"
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			_, err = os.Create(path)
			if err != nil {
				return -1, err
			}
			return -1, err
		}
		return -1, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	content := ""
	for scanner.Scan() {
		content += scanner.Text()
	}

	if content == "" {
		return -1, err
	}

	version, err := strconv.Atoi(string(content))
	if err != nil {
		return -1, err
	}
	return version, nil
}

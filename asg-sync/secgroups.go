package main

import (
	"encoding/json"
	"io/ioutil"
	"strings"

	cfclient "github.com/cloudfoundry-community/go-cfclient"
)

type SecGroup struct {
	Name  string
	Rules []cfclient.SecGroupRule
}

const separator = ":"

func ReadSecGroupFolder(folder string) (map[string]SecGroup, error) {
	var secGroups map[string]SecGroup
	secGroups = make(map[string]SecGroup)
	files, err := ioutil.ReadDir(folder)
	if err != nil {
		return secGroups, err
	}

	for _, file := range files {
		if !file.IsDir() {
			var secGroup SecGroup
			secGroup.Name = strings.TrimSuffix(file.Name(), ".json")
			secGroupBytes, err := ioutil.ReadFile(folder + "/" + file.Name())
			if err != nil {
				return secGroups, err
			}
			err = json.Unmarshal(secGroupBytes, &secGroup.Rules)
			if err != nil {
				return secGroups, err
			}
			secGroups[secGroup.Name] = secGroup
		}
	}
	return secGroups, nil
}

func (s *SecGroup) IsGlobal() bool {
	splitString := strings.Split(s.Name, separator)
	return len(splitString) == 1
}

func (s *SecGroup) Space() string {
	splitString := strings.Split(s.Name, separator)
	if len(splitString) > 1 {
		return splitString[1]
	}
	return ""
}

func (s *SecGroup) Org() string {
	splitString := strings.Split(s.Name, separator)
	return splitString[0]
}

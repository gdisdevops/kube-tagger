package util

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"strings"
)

type MapFlag struct {
	Items map[string]string
}

func (mFlag *MapFlag) String() string {
	return ""
}

func (mFlag *MapFlag) Set(value string) error {
	if mFlag.Items == nil {
		mFlag.Items = map[string]string{}
	}

	log.Infof("Set value %s", value)
	tagString := strings.Split(value, "=")
	if len(tagString) != 2 {
		return fmt.Errorf("invalid tag format, expected key=value")
	}

	key := tagString[0]
	val := tagString[1]
	currentList := mFlag.Items
	currentList[key] = val

	return nil
}

package dsnet

import (
	"io/ioutil"
	"encoding/json"
	"fmt"
)

func Add(hostname string, owner string, description string, publicKey string) {
	raw, err := ioutil.ReadFile(CONFIG_FILE)
	check(err)
	conf := DsnetConfig{}
	err = json.Unmarshal(raw, &conf)
	check(err)

	fmt.Printf("%+v\n", conf)
}

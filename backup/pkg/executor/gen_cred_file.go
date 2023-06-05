package executor

import (
	"io/ioutil"
	"strconv"

	"github.com/mitchellh/go-homedir"
)

func genCredFile(host string, port int, database string, user string, password string) error {
	cred := host + ":" + strconv.Itoa(port) + ":" + database + ":" + user + ":" + password

	homedir, err := homedir.Dir()
	if err != nil {
		return err
	}

	return ioutil.WriteFile(homedir+"/"+".pgpass", []byte(cred), 0600)
}

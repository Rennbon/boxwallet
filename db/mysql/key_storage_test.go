package mysql_test

import (
	"path"
	"testing"

	"github.com/Rennbon/boxwallet/bcconfig"
	mysqlConf "github.com/Rennbon/boxwallet/bcconfig/mysql"
	mysqlDB "github.com/Rennbon/boxwallet/db/mysql"
	"github.com/Rennbon/boxwallet/mock"
)

func TestNewKeyStorage(t *testing.T) {
	path1 := path.Join(mock.Gopath, mock.ProjectDir, mock.MySqlConfPath)
	provide, err := bcconfig.FromConfigString(path1, mock.YmlExt)
	if err != nil {
		t.Fail()
		t.Error(err)
		return
	}
	conf, err := mysqlConf.DecodeConfig(provide)
	if err != nil {
		t.Fail()
		t.Error(err)
		return
	}
	ksdb := mysqlDB.NewKeyStorage(conf.KeyStorage)
	if !ksdb.HasTable(&mysqlDB.Key{}) {
		ksdb.CreateTable(&mysqlDB.Key{})
	}
	t.Log(ksdb)
}

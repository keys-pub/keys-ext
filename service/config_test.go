package service_test

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/keys-pub/keys"
)

func testPath() string {
	return filepath.Join(os.TempDir(), fmt.Sprintf("%s.sdb", keys.RandFileName()))
}

// func TestConfigServiceNotOpen(t *testing.T) {
// 	db := sdb.New()
// 	service := config.NewService(db)
// 	_, err := service.Get(context.TODO(), &config.GetRequest{Key: "/encrypt"})
// 	require.EqualError(t, err, "db not open")
// }

// func TestConfigService(t *testing.T) {
// 	db := sdb.New()
// 	service := config.NewService(db)

// 	dbPath := testPath()
// 	key := keys.Rand32()
// 	err := db.OpenAtPath(context.TODO(), dbPath, key)
// 	require.NoError(t, err)
// 	defer func() {
// 		db.Close()
// 		_ = os.RemoveAll(dbPath)
// 	}()

// 	g, err := service.Get(context.TODO(), &config.GetRequest{Key: "/encrypt"})
// 	require.NoError(t, err)
// 	require.Nil(t, g.Value)

// 	val := &config.Encrypt{
// 		Recipients:      []string{"gabriel@github"},
// 		Sender:          "gabriel@echo",
// 		AddToRecipients: true,
// 		Sign:            true,
// 	}
// 	any, err := anypb.New(val)
// 	require.NoError(t, err)
// 	_, err = service.Set(context.TODO(), &config.SetRequest{
// 		Key:   "/encrypt",
// 		Value: any,
// 	})
// 	require.NoError(t, err)

// 	g, err = service.Get(context.TODO(), &config.GetRequest{Key: "/encrypt"})
// 	require.NoError(t, err)
// 	require.NotNil(t, g.Value)
// 	var out config.Encrypt
// 	err = any.UnmarshalTo(&out)
// 	require.NoError(t, err)
// 	require.Equal(t, val.Recipients, out.Recipients)
// 	require.Equal(t, val.Sender, out.Sender)
// 	require.Equal(t, val.AddToRecipients, out.AddToRecipients)
// 	require.Equal(t, val.Sign, out.Sign)
// }

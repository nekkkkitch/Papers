package db

import (
	"aus/internal/pkg/models"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDB(t *testing.T) {
	cfg := Config{Host: "db", Port: "5432", User: "user", Password: "123", DBName: "papersdb"}
	db, _ := New(&cfg)
	defer db.Close()

	t.Log("Testing db in process...")
	{
		testID := 0
		testUser := models.User{Login: "test", Password: "fakehash"}
		t.Logf("\tTest %d:\tAdding user %v", testID, testUser)
		_, err := db.AddUser(testUser)
		require.NoError(t, err, "Cannot add user %v")
	}
}

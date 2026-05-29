package repository

import (
	"testing"

	"github.com/javaup/flashsale-system/internal/config"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestInitDatabases_NilConfig(t *testing.T) {
	dm, err := InitDatabases(nil)
	assert.Error(t, err)
	assert.Nil(t, dm)
	assert.Contains(t, err.Error(), "config is nil")
}

func TestInitDatabases_EmptySources(t *testing.T) {
	dm, err := InitDatabases(&config.DatabaseConfig{
		Driver:  "mysql",
		Sources: []config.DataSourceConfig{},
	})
	assert.Error(t, err)
	assert.Nil(t, dm)
	assert.Contains(t, err.Error(), "no database sources")
}

func TestDatabaseManager_GetDB_NotFound(t *testing.T) {
	dm := &DatabaseManager{Sources: map[string]*gorm.DB{}}
	assert.Nil(t, dm.GetDB("nonexistent"))
}

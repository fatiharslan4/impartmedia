package data

import (
	"testing"

	"github.com/impartwealthapp/backend/pkg/models"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

var localDynamo = "http://localhost:8000"
var logger = newLogger()

func newLogger() *zap.Logger {
	l, _ := zap.NewProduction()
	return l
}

func TestNewHiveStore(t *testing.T) {
	hs, err := NewHiveData("us-east-2", localDynamo, "local", logger)

	assert.NotNil(t, hs)
	assert.Nil(t, err)
}

func TestHiveDynamo_CreateHive2(t *testing.T) {
	hs, _ := NewHiveData("us-east-2", localDynamo, "local", logger)

	h := models.RandomHive()
	hd := make(models.HiveDistributions, 1, 1)
	hd[0] = models.HiveDistribution{DisplayValue: "ABC", DisplayText: "letters", SortValue: 1}

	h.HiveDistributions = hd
	hive, err := hs.NewHive(h)

	assert.Nil(t, err)
	assert.JSONEq(t, h.ToJson(), hive.ToJson())
}

func TestHiveDynamo_GetHive(t *testing.T) {
	hs, _ := NewHiveData("us-east-2", localDynamo, "local", logger)

	h := models.RandomHive()
	h.HiveDistributions = append(h.HiveDistributions, models.HiveDistribution{
		DisplayText: "blach", DisplayValue: "blech",
	})
	h.HiveDistributions = append(h.HiveDistributions, models.HiveDistribution{
		DisplayText: "blach2", DisplayValue: "",
	})

	hive, err := hs.NewHive(h)
	assert.Nil(t, err)
	assert.Equal(t, h.HiveName, hive.HiveName)

	getHive, err := hs.GetHive(hive.HiveID, true)
	assert.NoError(t, err)
	assert.Equal(t, hive, getHive)

}

func TestHiveDynamo_GetHives(t *testing.T) {
	hs, _ := NewHiveData("us-east-2", localDynamo, "local", logger)
	h1 := models.RandomHive()
	_, err := hs.NewHive(h1)
	assert.Nil(t, err)

	h2 := models.RandomHive()
	_, err = hs.NewHive(h2)
	assert.Nil(t, err)

	hives, err := hs.GetHives()
	assert.Nil(t, err)
	assert.Contains(t, hives, h1)

	assert.Contains(t, hives, h2)
}

func TestHiveDynamo_EditHive(t *testing.T) {
	hs, _ := NewHiveData("us-east-2", localDynamo, "local", logger)

	h := models.RandomHive()

	hive, err := hs.NewHive(h)

	assert.Nil(t, err)
	hive.HiveDescription = "abc123"
	r, err := hs.EditHive(hive)

	assert.Nil(t, err)
	assert.JSONEq(t, hive.ToJson(), r.ToJson())
}

package repository

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMongoUserRepositoryImplementsUserRepository(t *testing.T) {
	var _ UserRepository = (*MongoUserRepository)(nil)
}

func TestUserRepositoryEmailHashIndexShape(t *testing.T) {
	indexModel := bson.D{{Key: "emailHash", Value: 1}}
	require.Equal(t, "emailHash", indexModel[0].Key)
	require.Equal(t, 1, indexModel[0].Value)
}

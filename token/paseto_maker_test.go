package token

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/joekings2k/logistics-eta/util"
	"github.com/stretchr/testify/require"
)


func TestPasetoMaker(t *testing.T) {
	maker, err := NewPasetoMaker(util.RandomString(32))
	require.NoError(t, err)
	userID, err := uuid.NewRandom()
	require.NoError(t, err)
	duration := time.Minute
	issuedAt := time.Now()
	expiredAt := issuedAt.Add(duration)

	token, err := maker.CreateToken(userID, duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	payload, err := maker.VerifyToken(token)
	require.NoError(t, err)
	require.NotEmpty(t, payload)
	
	require.NotZero(t, payload.ID)
	require.Equal(t, userID, payload.UserID)
	require.WithinDuration(t, issuedAt, payload.IssuedAt, time.Second)
	require.WithinDuration(t, expiredAt, payload.ExpiredAt, time.Second)
	
}


func TestExpiredPasetoToken(t *testing.T) {
	maker, err := NewPasetoMaker(util.RandomString(32))
	require.NoError(t, err)
	userID, err := uuid.NewRandom()
	require.NoError(t, err)
	token, err := maker.CreateToken(userID, -time.Minute)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	payload, err := maker.VerifyToken(token)
	require.Error(t, err)
	require.EqualError(t, err, ErrExpiredToken.Error())
	require.Nil(t, payload)
}


func TestPasetoSecretKeyLength(t *testing.T){
	_, err := NewPasetoMaker(util.RandomString(10))
	require.Error(t, err)
}



func TestInvalidToken(t *testing.T) {
	maker, err := NewPasetoMaker(util.RandomString(32))
	require.NoError(t, err)
	userID := uuid.New()
	duration := time.Minute
	
	token, err := maker.CreateToken(userID, duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	payload, err := maker.VerifyToken("invalidToken")
	require.Error(t, err)
	require.EqualError(t, err, ErrInvalidToken.Error())
	require.Nil(t, payload)
}


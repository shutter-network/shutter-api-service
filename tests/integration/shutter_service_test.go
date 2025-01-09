package integration

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"time"

	httpError "github.com/shutter-network/shutter-service-api/internal/error"
	"github.com/shutter-network/shutter/shlib/shcrypto"

	cryptorand "crypto/rand"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shutter-network/shutter-service-api/internal/service"
	"github.com/shutter-network/shutter-service-api/internal/usecase"
)

var msg = []byte("please hide this message")

func (s *TestShutterService) TestRequestDecryptionKeyBeforeTimestampReached() {
	address := crypto.PubkeyToAddress(*s.config.PublicKey).Hex()
	identityPrefix, err := generateRandomBytes(32)
	s.Require().NoError(err)
	identityPrefixStringified := hex.EncodeToString(identityPrefix)

	query := fmt.Sprintf("?address=%s&identityPrefix=%s", address, identityPrefixStringified)
	url := "/api/get_data_for_encryption" + query

	recorder := httptest.NewRecorder()
	s.router.ServeHTTP(recorder, httptest.NewRequest("GET", url, nil))

	s.Require().Equal(http.StatusOK, recorder.Code)

	body, err := io.ReadAll(recorder.Body)
	s.Require().NoError(err)

	var dataForEncryptionResponse map[string]usecase.GetDataForEncryptionResponse
	err = json.Unmarshal(body, &dataForEncryptionResponse)
	s.Require().NoError(err)

	res := dataForEncryptionResponse["message"]
	s.Require().GreaterOrEqual(res.Eon, uint64(1))
	s.Require().NotNil(res.EonKey)
	s.Require().NotNil(res.Identity)
	s.Require().NotNil(res.IdentityPrefix)

	identityStringified := res.Identity

	decryptionTimestamp := time.Now().Add(1 * time.Hour).Unix()
	reqBody := service.RegisterIdentityRequest{
		DecryptionTimestamp: uint64(decryptionTimestamp),
		IdentityPrefix:      identityPrefixStringified,
	}

	jsonData, err := json.Marshal(reqBody)
	s.Require().NoError(err)
	url = "/api/register_identity"

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	s.Require().NoError(err)

	req.Header.Set("Content-Type", "application/json")

	recorder = httptest.NewRecorder()

	s.router.ServeHTTP(recorder, req)
	s.Require().Equal(http.StatusOK, recorder.Code)

	query = fmt.Sprintf("?identity=%s", identityStringified)
	url = "/api/get_decryption_key" + query

	recorder = httptest.NewRecorder()
	s.router.ServeHTTP(recorder, httptest.NewRequest("GET", url, nil))

	s.Require().Equal(http.StatusBadRequest, recorder.Code)

	body, err = io.ReadAll(recorder.Body)
	s.Require().NoError(err)

	var errorResponse httpError.Http
	err = json.Unmarshal(body, &errorResponse)
	s.Require().NoError(err)

	s.Require().Equal(errorResponse.Description, "timestamp not reached yet, decryption key requested too early")
}

func (s *TestShutterService) TestRequestDecryptionKeyAfterTimestampReached() {
	address := crypto.PubkeyToAddress(*s.config.PublicKey).Hex()
	identityPrefix, err := generateRandomBytes(32)
	s.Require().NoError(err)
	identityPrefixStringified := hex.EncodeToString(identityPrefix)

	query := fmt.Sprintf("?address=%s&identityPrefix=%s", address, identityPrefixStringified)
	url := "/api/get_data_for_encryption" + query

	recorder := httptest.NewRecorder()
	s.router.ServeHTTP(recorder, httptest.NewRequest("GET", url, nil))

	s.Require().Equal(http.StatusOK, recorder.Code)

	body, err := io.ReadAll(recorder.Body)
	s.Require().NoError(err)

	var dataForEncryptionResponse map[string]usecase.GetDataForEncryptionResponse
	err = json.Unmarshal(body, &dataForEncryptionResponse)
	s.Require().NoError(err)

	res := dataForEncryptionResponse["message"]
	s.Require().GreaterOrEqual(res.Eon, uint64(1))
	s.Require().NotNil(res.EonKey)
	s.Require().NotNil(res.Identity)
	s.Require().NotNil(res.IdentityPrefix)

	identityStringified := res.Identity

	identityBytes, err := hex.DecodeString(res.Identity)
	s.Require().NoError(err)

	eonKeyBytes, err := hex.DecodeString(res.EonKey)
	s.Require().NoError(err)

	var identity *shcrypto.EpochID
	identity.Unmarshal(identityBytes)

	var eonPublicKey *shcrypto.EonPublicKey
	eonPublicKey.Unmarshal(eonKeyBytes)

	sigma, err := shcrypto.RandomSigma(cryptorand.Reader)
	s.Require().NoError(err)

	encryptedTransaction := shcrypto.Encrypt(msg, eonPublicKey, identity, sigma)

	decryptionTimestamp := time.Now().Add(2 * time.Second).Unix()
	reqBody := service.RegisterIdentityRequest{
		DecryptionTimestamp: uint64(decryptionTimestamp),
		IdentityPrefix:      identityPrefixStringified,
	}

	jsonData, err := json.Marshal(reqBody)
	s.Require().NoError(err)
	url = "/api/register_identity"

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	s.Require().NoError(err)

	req.Header.Set("Content-Type", "application/json")

	recorder = httptest.NewRecorder()

	s.router.ServeHTTP(recorder, req)
	s.Require().Equal(http.StatusOK, recorder.Code)

	time.Sleep(10 * time.Second)

	query = fmt.Sprintf("?identity=%s", identityStringified)
	url = "/api/get_decryption_key" + query

	recorder = httptest.NewRecorder()
	s.router.ServeHTTP(recorder, httptest.NewRequest("GET", url, nil))

	s.Require().Equal(http.StatusOK, recorder.Code)

	body, err = io.ReadAll(recorder.Body)
	s.Require().NoError(err)

	var decryptionKeyResponse map[string]usecase.GetDecryptionKeyResponse
	err = json.Unmarshal(body, &decryptionKeyResponse)
	s.Require().NoError(err)

	decryptionKeyStringified := decryptionKeyResponse["message"].DecryptionKey
	s.Require().NotEmpty(decryptionKeyStringified)

	var decryptionKey *shcrypto.EpochSecretKey
	decryptionKeyBytes, err := hex.DecodeString(decryptionKeyStringified)
	s.Require().NoError(err)
	err = decryptionKey.Unmarshal(decryptionKeyBytes)
	s.Require().NoError(err)

	decryptedMessage, err := encryptedTransaction.Decrypt(decryptionKey)
	s.Require().NoError(err)
	s.Require().Equal(msg, decryptedMessage)
}

func (s *TestShutterService) TestRequestDecryptionKeyForUnregisteredIdentity() {
	address := crypto.PubkeyToAddress(*s.config.PublicKey).Hex()
	identityPrefix, err := generateRandomBytes(32)
	s.Require().NoError(err)
	identityPrefixStringified := hex.EncodeToString(identityPrefix)

	query := fmt.Sprintf("?address=%s&identityPrefix=%s", address, identityPrefixStringified)
	url := "/api/get_data_for_encryption" + query

	recorder := httptest.NewRecorder()
	s.router.ServeHTTP(recorder, httptest.NewRequest("GET", url, nil))

	s.Require().Equal(http.StatusOK, recorder.Code)

	body, err := io.ReadAll(recorder.Body)
	s.Require().NoError(err)

	var dataForEncryptionResponse map[string]usecase.GetDataForEncryptionResponse
	err = json.Unmarshal(body, &dataForEncryptionResponse)
	s.Require().NoError(err)

	res := dataForEncryptionResponse["message"]
	s.Require().GreaterOrEqual(res.Eon, uint64(1))
	s.Require().NotNil(res.EonKey)
	s.Require().NotNil(res.Identity)
	s.Require().NotNil(res.IdentityPrefix)

	identityStringified := res.Identity

	query = fmt.Sprintf("?identity=%s", identityStringified)
	url = "/api/get_decryption_key" + query

	recorder = httptest.NewRecorder()
	s.router.ServeHTTP(recorder, httptest.NewRequest("GET", url, nil))

	s.Require().Equal(http.StatusBadRequest, recorder.Code)

	body, err = io.ReadAll(recorder.Body)
	s.Require().NoError(err)

	var errorResponse httpError.Http
	err = json.Unmarshal(body, &errorResponse)
	s.Require().NoError(err)

	s.Require().Equal(errorResponse.Description, "identity has not been registerd yet")
}

func (s *TestShutterService) TestRequestDecryptCommitmentAfterTimestampReached() {
	address := crypto.PubkeyToAddress(*s.config.PublicKey).Hex()
	identityPrefix, err := generateRandomBytes(32)
	s.Require().NoError(err)
	identityPrefixStringified := hex.EncodeToString(identityPrefix)

	query := fmt.Sprintf("?address=%s&identityPrefix=%s", address, identityPrefixStringified)
	url := "/api/get_data_for_encryption" + query

	recorder := httptest.NewRecorder()
	s.router.ServeHTTP(recorder, httptest.NewRequest("GET", url, nil))

	s.Require().Equal(http.StatusOK, recorder.Code)

	body, err := io.ReadAll(recorder.Body)
	s.Require().NoError(err)

	var dataForEncryptionResponse map[string]usecase.GetDataForEncryptionResponse
	err = json.Unmarshal(body, &dataForEncryptionResponse)
	s.Require().NoError(err)

	res := dataForEncryptionResponse["message"]
	s.Require().GreaterOrEqual(res.Eon, uint64(1))
	s.Require().NotNil(res.EonKey)
	s.Require().NotNil(res.Identity)
	s.Require().NotNil(res.IdentityPrefix)

	identityStringified := res.Identity

	identityBytes, err := hex.DecodeString(res.Identity)
	s.Require().NoError(err)

	eonKeyBytes, err := hex.DecodeString(res.EonKey)
	s.Require().NoError(err)

	var identity *shcrypto.EpochID
	identity.Unmarshal(identityBytes)

	var eonPublicKey *shcrypto.EonPublicKey
	eonPublicKey.Unmarshal(eonKeyBytes)

	sigma, err := shcrypto.RandomSigma(cryptorand.Reader)
	s.Require().NoError(err)

	encryptedCommitment := shcrypto.Encrypt(msg, eonPublicKey, identity, sigma)
	encrypedCommitmentBytes := encryptedCommitment.Marshal()
	encryptedCommitmentStringified := hex.EncodeToString(encrypedCommitmentBytes)

	decryptionTimestamp := time.Now().Add(2 * time.Second).Unix()
	reqBody := service.RegisterIdentityRequest{
		DecryptionTimestamp: uint64(decryptionTimestamp),
		IdentityPrefix:      identityPrefixStringified,
	}

	jsonData, err := json.Marshal(reqBody)
	s.Require().NoError(err)
	url = "/api/register_identity"

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	s.Require().NoError(err)

	req.Header.Set("Content-Type", "application/json")

	recorder = httptest.NewRecorder()

	s.router.ServeHTTP(recorder, req)
	s.Require().Equal(http.StatusOK, recorder.Code)

	time.Sleep(10 * time.Second)

	query = fmt.Sprintf("?identity=%s&encryptedCommitment=%s", identityStringified, encryptedCommitmentStringified)
	url = "/api/decrypt_commitment" + query

	recorder = httptest.NewRecorder()
	s.router.ServeHTTP(recorder, httptest.NewRequest("GET", url, nil))

	s.Require().Equal(http.StatusOK, recorder.Code)

	body, err = io.ReadAll(recorder.Body)
	s.Require().NoError(err)

	var decryptionKeyResponse map[string][]byte
	err = json.Unmarshal(body, &decryptionKeyResponse)
	s.Require().NoError(err)

	decryptionKeyBytes := decryptionKeyResponse["message"]

	s.Require().NotEmpty(decryptionKeyBytes)

	var decryptionKey *shcrypto.EpochSecretKey
	err = decryptionKey.Unmarshal(decryptionKeyBytes)
	s.Require().NoError(err)

	decryptedMessage, err := encryptedCommitment.Decrypt(decryptionKey)
	s.Require().NoError(err)
	s.Require().Equal(msg, decryptedMessage)
}

package integration

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shutter-network/shutter-service-api/internal/usecase"
	"github.com/stretchr/testify/assert"
)

func (s *TestShutterService) TestGetDataForEncryption() {
	address := crypto.PubkeyToAddress(*s.config.PublicKey).Hex()
	identityPrefix, err := generateRandomBytes(32)
	s.Require().NoError(err)
	identityPrefixStringified := hex.EncodeToString(identityPrefix)

	query := fmt.Sprintf("?address=%s&identityPrefix=%s", address, identityPrefixStringified)
	url := s.testServer.URL + "/api/get_data_for_encryption" + query

	resp, err := http.Get(url)
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	s.Require().Equal(http.StatusOK, resp.StatusCode)

	var response usecase.GetDataForEncryptionResponse
	err = json.NewDecoder(resp.Body).Decode(&response)

	s.Require().NoError(err)
	s.Require().Greater(response.Eon, uint64(1))
	s.Require().NotNil(response.EonKey)
	s.Require().NotNil(response.Identity)
	s.Require().NotNil(response.IdentityPrefix)
}

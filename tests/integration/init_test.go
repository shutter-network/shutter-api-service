package integration

import (
	"context"
	"net/http/httptest"
	"os"
	"testing"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
	"github.com/shutter-network/shutter-service-api/common"
	"github.com/shutter-network/shutter-service-api/common/database"
	"github.com/shutter-network/shutter-service-api/internal/data"
	"github.com/shutter-network/shutter-service-api/internal/router"
	"github.com/stretchr/testify/suite"
)

type TestShutterService struct {
	suite.Suite
	db         *pgxpool.Pool
	dbQuery    *data.Queries
	router     *gin.Engine
	config     *common.Config
	ethClient  *ethclient.Client
	contract   *common.Contract
	testServer *httptest.Server
}

func (s *TestShutterService) SetupSuite() {
	ctx := context.Background()
	var err error
	dbURL := database.GetDBURL()
	s.db, err = database.NewDB(ctx, dbURL)
	s.Require().NoError(err)

	s.dbQuery = data.New(s.db)
	signingKey, err := crypto.HexToECDSA(os.Getenv("SIGNING_KEY"))
	if err != nil {
		log.Err(err).Msg("failed to parse signing key")
	}

	keyperHTTPUrl := os.Getenv("KEYPER_HTTP_URL")

	s.config, err = common.NewConfig(keyperHTTPUrl, signingKey)
	if err != nil {
		log.Err(err).Msg("unable to parse keyper http url")
		return
	}

	rpc_url := os.Getenv("RPC_URL")
	s.ethClient, err = ethclient.Dial(rpc_url)
	if err != nil {
		log.Err(err).Msg("failed to initialize rpc client")
		return
	}

	shutterRegistryContractAddressStringified := os.Getenv("SHUTTER_REGISTRY_CONTRACT_ADDRESS")
	shutterRegistryContractAddress := ethCommon.HexToAddress(shutterRegistryContractAddressStringified)

	keyBroadcastContractAddressStringified := os.Getenv("KEY_BROADCAST_CONTRACT_ADDRESS")
	keyBroadcastContractAddress := ethCommon.HexToAddress(keyBroadcastContractAddressStringified)

	keyperSetManagerContractAddressStringified := os.Getenv("KEYPER_SET_MANAGER_CONTRACT_ADDRESS")
	keyperSetManagerContractAddress := ethCommon.HexToAddress(keyperSetManagerContractAddressStringified)

	s.contract, err = common.NewContract(s.ethClient, shutterRegistryContractAddress, keyperSetManagerContractAddress, keyBroadcastContractAddress)
	if err != nil {
		log.Err(err).Msg("failed to instantiate shutter contracts")
		return
	}
	s.router = router.NewRouter(s.db, s.contract, s.ethClient, s.config)
	s.testServer = httptest.NewServer(s.router)
}

func TestShutterServiceSuite(t *testing.T) {
	suite.Run(t, new(TestShutterService))
}

func (s *TestShutterService) TearDownSuite() {
	s.testServer.Close()
	s.db.Close()
}

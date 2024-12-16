package router

import (
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shutter-network/shutter-service-api/common"
	"github.com/shutter-network/shutter-service-api/internal/middleware"
	"github.com/shutter-network/shutter-service-api/internal/service"
)

func NewRouter(
	db *pgxpool.Pool,
	contract *common.Contract,
	ethClient *ethclient.Client,
	config *common.Config,
) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.Use(middleware.ErrorHandler())

	cryptoService := service.NewCryptoService(db, contract, ethClient, config)
	api := router.Group("/api")
	{
		api.GET("/get_decryption_key", cryptoService.GetDecryptionKey)
		api.GET("/get_data_for_encryption", cryptoService.GetDataForEncryption)
		api.POST("/register_identity", cryptoService.RegisterIdentity)
		api.GET("/decrypt-commitment", cryptoService.DecryptCommitment)
	}
	return router
}

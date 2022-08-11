package gateway

import (
	"net/http"

	"github.com/Ionian-Web3-Storage/ionian-client/node"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

const httpStatusInternalError = 600

var allClients []*node.Client

func MustServeLocal(nodes []string) {
	if len(nodes) == 0 {
		logrus.Fatal("storage nodes not configured")
	}

	for _, url := range nodes {
		client, err := node.NewClient(url)
		if err != nil {
			logrus.WithError(err).WithField("url", url).Fatal("Failed to connect storage node")
		}
		allClients = append(allClients, client)
	}

	server := http.Server{
		Addr:    "127.0.0.1:6789",
		Handler: newLocalRouter(),
	}

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		logrus.WithError(err).Fatal("Failed to serve API")
	}
}

func newLocalRouter() *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	if logrus.IsLevelEnabled(logrus.DebugLevel) {
		router.Use(gin.Logger())
	}

	localApi := router.Group("/local")
	localApi.GET("/nodes", wrap(listNodes))
	localApi.GET("/file", wrap(getLocalFileInfo))
	localApi.GET("/status", wrap(getFileStatus))
	localApi.POST("/upload", wrap(uploadLocalFile))
	localApi.POST("/download", wrap(downloadFileLocal))

	return router
}

func wrap(controller func(*gin.Context) (interface{}, error)) func(*gin.Context) {
	return func(c *gin.Context) {
		result, err := controller(c)
		if err != nil {
			switch e := err.(type) {
			case *BusinessError:
				c.JSON(http.StatusOK, e)
			case validator.ValidationErrors: // binding error
				c.JSON(http.StatusOK, ErrValidation.WithData(e.Error()))
			default: // internal server error
				c.JSON(httpStatusInternalError, ErrInternalServer.WithData(err.Error()))
			}
		} else if result == nil {
			c.JSON(http.StatusOK, ErrNil)
		} else {
			c.JSON(http.StatusOK, ErrNil.WithData(result))
		}
	}
}
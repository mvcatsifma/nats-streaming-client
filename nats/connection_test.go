package nats

import (
	"github.com/mvcatsifma/nats-streaming-client/types"
	stand "github.com/nats-io/nats-streaming-server/server"
	"github.com/nats-io/stan.go"
	"github.com/stretchr/testify/suite"
	"log"
	"testing"
)

var testClusterId = "test-cluster-id"
var testClientId = "test-client-id"

func TestConnectorSuite(t *testing.T) {
	suite.Run(t, new(ConnectorSuite))
}

func (suite *ConnectorSuite) Test_WhenConnectedCanGetConnection() {
	_ : suite.manager.Open()

	connection := suite.manager.GetConn()

	suite.NotNil(connection)

	_, ok := connection.(stan.Conn)
	suite.True(ok)
}

func (suite *ConnectorSuite) Test_Open_AlreadyOpened_ReturnsError() {
	_ : suite.manager.Open()
	err := suite.manager.Open()

	suite.Nil(err)
}

func (suite *ConnectorSuite) Test_Open_SendsConnected() {
	sub1 := suite.manager.SubscribeToStatusChanges()
	sub2 := suite.manager.SubscribeToStatusChanges()

	// ignore initial status update
	<-sub1
	<-sub2

	_ = suite.manager.Open()

	currentStatus := types.CONNECTED
	suite.Equal(currentStatus, suite.manager.Status)
	suite.Equal(currentStatus, <-sub1)
	suite.Equal(currentStatus, <-sub2)
}

func (suite *ConnectorSuite) Test_ShouldCloseSlowSubscriber() {
	sub1 := suite.manager.SubscribeToStatusChanges()
	sub2 := suite.manager.SubscribeToStatusChanges()

	// do not read from sub2
	s1 := <-sub1
	suite.Equal(types.NOT_CONNECTED, s1)

	_ = suite.manager.Open()

	s2 := <-sub1
	suite.Equal(types.CONNECTED, s2)

	<- sub2
	_, ok := <- sub2
	suite.False(ok)
}

func (suite *ConnectorSuite) Test_WhenCreated_SendsNotConnected() {
	sub := suite.manager.SubscribeToStatusChanges()

	currentStatus := <-sub

	expectedStatus := types.NOT_CONNECTED
	suite.Equal(expectedStatus, currentStatus)
	suite.Equal(expectedStatus, suite.manager.Status)
}

func (suite *ConnectorSuite) Test_ConnectionLost_SendsLost() {
	sub := suite.manager.SubscribeToStatusChanges()
	// ignore initial status update
	<-sub

	_ = suite.manager.Open()
	// ignore connected status
	<-sub

	suite.server.Shutdown()

	currentStatus := <- sub

	expectedStatus := types.LOST
	suite.Equal(expectedStatus, suite.manager.Status)
	suite.Equal(expectedStatus, currentStatus)
}

func (suite *ConnectorSuite) Test_Close_SendsNotConnected() {
	sub := suite.manager.SubscribeToStatusChanges()
	// ignore initial status update
	<- sub

	_ = suite.manager.Open()
	// ignore connected status
	<- sub

	err := suite.manager.Close()
	currentStatus := <- sub

	suite.Nil(err)
	expectedStatus := types.NOT_CONNECTED
	suite.Equal(expectedStatus, suite.manager.Status)
	suite.Equal(expectedStatus, currentStatus)
	suite.Nil(suite.manager.Conn)
}

func (suite *ConnectorSuite) Test_Close_ReturnsErrorIfNotConnected() {
	err := suite.manager.Close()

	suite.NotNil(err)
	suite.IsType(&types.NotConnectedError{}, err)
}

func (suite *ConnectorSuite) Test_Open_FailsIfInvalidClientId(){
	conn := NewConnectionManager(testClusterId, "") // client id is required
	sub := conn.SubscribeToStatusChanges()

	err := conn.Open()
	currentStatus := <- sub

	suite.NotNil(err)
	expectedStatus := types.NOT_CONNECTED
	suite.Equal(expectedStatus, suite.manager.Status)
	suite.Equal(expectedStatus, currentStatus)
	suite.Nil(suite.manager.Conn)
}

type ConnectorSuite struct {
	suite.Suite
	server  *stand.StanServer
	manager *manager
}

func (suite *ConnectorSuite) SetupTest() {
	if suite.server != nil && suite.server.State() == stand.Standalone {
		suite.server.Shutdown()
	}
	sOpts := stand.GetDefaultOptions()
	sOpts.ID = testClusterId
	s, err := stand.RunServerWithOpts(sOpts, &stand.DefaultNatsServerOptions)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.server = s
	suite.manager = NewConnectionManager(testClusterId, testClientId)
}

func (suite *ConnectorSuite) TearDownTest() {
	err := suite.manager.Close()
	if err != nil {
		log.Println(err)
	}
}

func (suite *ConnectorSuite) SetupSuite() {
}

func (suite *ConnectorSuite) TearDownSuite() {
}

func (suite *ConnectorSuite) BeforeTest(suiteName, testName string) {
}

func (suite *ConnectorSuite) AfterTest(suiteName, testName string) {
}

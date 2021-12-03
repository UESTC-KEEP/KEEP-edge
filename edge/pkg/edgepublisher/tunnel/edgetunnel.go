package edgetunnel

import (
	"crypto/tls"
	"fmt"
	"github.com/gorilla/websocket"
	"keep/constants"
	"keep/edge/pkg/common/modules"
	"keep/edge/pkg/edgepublisher/tunnel/cert"
	beehiveContext "keep/pkg/util/core/context"
	"keep/pkg/util/core/model"
	"keep/pkg/util/loggerv1.0.1"
	"net/http"
	"net/url"
	"time"
)

type edgeTunnel struct {
	hostnameOverride string
	nodeIP           string
	reconnectChan    chan struct{}
}

var session *tunnelSession
var sessionConnected bool

func newEdgeTunnel(hostnameOverride, nodeIP string) *edgeTunnel {
	return &edgeTunnel{
		hostnameOverride: hostnameOverride,
		nodeIP:           nodeIP,
		reconnectChan:    make(chan struct{}),
	}
}

func (e *edgeTunnel) start() {
	serverURL := url.URL{
		Scheme: "wss",
		Host:   fmt.Sprintf("%s:%d", constants.DefaultKeepCloudIP, constants.DefaultWebSocketPort),
		Path:   constants.DefaultWebSocketUrl,
	}

	certManager := cert.NewCertManager(e.hostnameOverride)
	certManager.Start()

	clientCert, err := tls.LoadX509KeyPair(constants.DefaultCertFile, constants.DefaultKeyFile)
	if err != nil {
		logger.Info("Failed to load x509 key pair: ", err, "try again")
		time.Sleep(10 * time.Second)
		clientCert, err = tls.LoadX509KeyPair(constants.DefaultCertFile, constants.DefaultKeyFile)
	}
	if err != nil {
		logger.Fatal("Failed to load x509 key pair: ", err, "Exiting...")
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{clientCert},
	}

	for {
		select {
		case <-beehiveContext.Done():
			return
		default:
		}
		var err error
		session, err = e.tlsClientConnect(serverURL, tlsConfig)
		if err != nil {
			logger.Error("connect failed: ", err)
			time.Sleep(5 * time.Second)
			continue
		}
		sessionConnected = true

		go session.startPing(e.reconnectChan)
		go session.routeToEdge(e.reconnectChan)

		<-e.reconnectChan
		sessionConnected = false
		session.Close()
		logger.Warn("connection broken, reconnecting...")
		time.Sleep(5 * time.Second)

		//清空reconnectChan
	clean:
		for {
			select {
			case <-e.reconnectChan:
			default:
				break clean
			}
		}
	}
}

func (e *edgeTunnel) tlsClientConnect(url url.URL, tlsConfig *tls.Config) (*tunnelSession, error) {
	logger.Info("Start a new tunnel connection")

	dial := websocket.Dialer{
		TLSClientConfig:  tlsConfig,
		HandshakeTimeout: time.Duration(30) * time.Second,
	}
	header := http.Header{}
	header.Add(constants.SessionKeyHostNameOverride, e.hostnameOverride)
	header.Add(constants.SessionKeyInternalIP, e.nodeIP)

	con, _, err := dial.Dial(url.String(), header)
	if err != nil {
		return nil, err
	}

	session := NewTunnelSession(con)
	return session, nil
}

func StartEdgeTunnel(nodeName, nodeIP string) {
	edget := newEdgeTunnel(nodeName, nodeIP)
	edget.start()
}

func WriteToCloud(msg *model.Message) {
	for i := 0; i < 5 && !sessionConnected; i++ {
		logger.Info("session not connected, waiting")
		time.Sleep(3 * time.Second)
	}
	if !sessionConnected {
		msgToEdgeTwin := model.NewMessage("")
		msgToEdgeTwin.SetResourceOperation(msg.GetResource(), "")
		_, err := beehiveContext.SendSync(modules.EdgeTwinGroup, *msgToEdgeTwin, time.Second)
		if err != nil {
			logger.Error("send message to edge twin error: ", err)
		}
	}
	err := session.Tunnel.WriteMessage(msg)
	if err != nil {
		msgToEdgeTwin := model.NewMessage("")
		msgToEdgeTwin.SetResourceOperation(msg.GetResource(), "")
		_, err := beehiveContext.SendSync(modules.EdgeTwinGroup, *msgToEdgeTwin, time.Second)
		if err != nil {
			logger.Error("send message to edge twin error: ", err)
		}
	}

}

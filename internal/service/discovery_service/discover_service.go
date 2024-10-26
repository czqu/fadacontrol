package discovery_service

import (
	"fadacontrol/internal/base/logger"
	"fadacontrol/internal/entity"
	"fadacontrol/internal/schema"
	"fadacontrol/pkg/goroutine"
	"fadacontrol/pkg/utils"
	"fadacontrol/pkg/utils/cache"
	"fmt"
	"gorm.io/gorm"
	"net"
	"os"
	"sync"
	"time"
)

type DiscoverService struct {
	_db          *gorm.DB
	config       entity.DiscoverConfig
	ipFail       cache.Cache[string, int]
	ipAlwaysFail cache.Cache[string, int]
	port         int
	ipFailRetry  time.Duration
	udpDone      chan int
	hostname     string
	ListenConn   *net.UDPConn
	StartLock    sync.Mutex
	StopLock     sync.Mutex
	RestartLock  sync.Mutex
}

const udpSendInterval = 2 * time.Second
const udpMaxTryTime = 10
const connTimeout = 5 * time.Second

func NewDiscoverService(db *gorm.DB) *DiscoverService {
	d := DiscoverService{
		_db: db, config: entity.DiscoverConfig{},
		port: 4084, hostname: "",
		ipFail:       cache.NewSyncMapMemCache[string, int](4 * 1024),
		ipAlwaysFail: cache.NewSyncMapMemCache[string, int](4 * 1024),
		ipFailRetry:  30 * time.Second,
	}
	d.ipFail.StartAutoClean(d.ipFailRetry / 2)
	d.ipAlwaysFail.StartAutoClean(1 * time.Minute)
	return &d
}

func (d *DiscoverService) GetDiscoverConfig() (*schema.DiscoverSchema, error) {
	var config entity.DiscoverConfig
	err := d._db.First(&config).Error
	if err != nil {
		return nil, err
	}
	return &schema.DiscoverSchema{Enabled: config.Enabled}, err
}

func (d *DiscoverService) PatchDiscoverServiceConfig(content map[string]interface{}) error {

	var config entity.DiscoverConfig
	if err := d._db.First(&config).Error; err != nil {
		return err
	}
	if err := d._db.Model(&config).Updates(content).Error; err != nil {
		return err
	}
	return nil

}

func (d *DiscoverService) listenAndSend(port int) {
	defer func() {
		logger.Info("The UDP listen service is stopped")
	}()
	addr := net.UDPAddr{
		Port: port,
		IP:   net.IPv4zero,
	}
	for {
		conn, err := net.ListenUDP("udp", &addr)
		if err != nil {
			logger.Warn("Error listening:", err.Error())
			return
		}
		d.ListenConn = conn
		defer conn.Close()

		logger.Info("Listening on port", port)
		buffer := make([]byte, 1024)
		for {
			n, remoteAddr, err := conn.ReadFromUDP(buffer)
			if err != nil {
				if isClosedConnError(err) {
					logger.Warn("Udp Connection closed:")
					return
				}
				logger.Warn("Error reading from UDP:", err)
				continue
			}

			logger.Debugf("Received message from %s: %s", remoteAddr, string(buffer[:n]))

			err = conn.SetWriteDeadline(time.Now().Add(connTimeout))
			if err != nil {
				fmt.Println("SetWriteDeadline failed:", err)
				break
			}
			_, err = conn.WriteToUDP([]byte(d.hostname), remoteAddr)

			if err != nil {
				logger.Warn("Error sending response:", err)
			} else {
				logger.Warnf("Sent 'hello' to %s", remoteAddr)
			}
		}
	}

}
func (d *DiscoverService) StopService() error {
	if !d.StopLock.TryLock() {
		return nil
	}
	defer d.StopLock.Unlock()

	if d.ListenConn != nil {
		d.ListenConn.Close()
	}
	d.udpDone <- 1
	close(d.udpDone)
	logger.Info("The UDP service service is stopped")
	return nil
}
func (d *DiscoverService) StartBroadcast() {
	logger.Info("The UDP broadcast service is launched")
	var err error
	d.hostname, err = os.Hostname()
	if err != nil {
		logger.Warn("Error getting hostname:", err)
		d.StopService()
		return
	}

	goroutine.RecoverGO(func() {
		defer func() {
			logger.Info("The UDP broadcast service is stopped")
		}()
		logger.Debug("Sending UDP Broadcast ")
		for {
			select {
			case <-d.udpDone:
				return
			case <-time.After(udpSendInterval):
				d.udpBroadcast()
			}
		}
	})

}

func (d *DiscoverService) GetValidInterface(t utils.AddressType) []utils.Interface {

	interfaces, err := utils.GetValidInterface(utils.UNSET)
	if err != nil {
		logger.Error("Error getting interface list:", err)
		return []utils.Interface{}
	}
	return interfaces
}
func (d *DiscoverService) udpBroadcast() {
	d.sendUdp(net.IPv4bcast)
	interfaces := d.GetValidInterface(utils.IPV4)
	for _, iface := range interfaces {
		for _, ipnet := range iface.IPAddresses {
			d.sendUdp(ipnet)
		}
	}

}
func (d *DiscoverService) sendUdp(ip net.IP) {

	//only broadcast ipv4
	if ip.To4() == nil {
		return
	}
	ip = ip.To4()
	tryTimes, ok := d.ipFail.Get(ip.String())
	if !ok {
		tryTimes = 0
	}
	if tryTimes >= udpMaxTryTime {
		exists := d.ipAlwaysFail.Exists(ip.String())
		if !exists {
			d.ipAlwaysFail.SetWithTTL(ip.String(), 1, 1*time.Hour)
			logger.Warnf("The ip %s is not available,will reduce try time", ip.String())
		}

		return
	}
	var broadcastIP net.IP
	var lddr *net.UDPAddr

	if !ip.Equal(net.IPv4bcast) {
		subnet := ip.DefaultMask()
		broadcastIP = make(net.IP, len(ip))
		for i := range broadcastIP {
			broadcastIP[i] = ip[i] | ^subnet[i]
		}
		lddr = &net.UDPAddr{
			IP:   ip,
			Port: 0,
		}
	} else {
		broadcastIP = ip
		lddr = nil
	}

	conn, err := net.DialUDP("udp", lddr, &net.UDPAddr{
		IP:   broadcastIP,
		Port: d.port,
	})
	if err != nil {

		t, _ := d.ipFail.Get(ip.String())

		t = t + 1
		if d.ipAlwaysFail.Exists(ip.String()) {
			d.ipFail.SetWithTTL(ip.String(), udpMaxTryTime, d.ipFailRetry)
		} else {
			d.ipFail.SetWithTTL(ip.String(), t, d.ipFailRetry)
			logger.Warn(err, "Will retry", 10-t, "more times")

		}
		logger.Debug(err)
		if lddr != nil {
			logger.Debugf(lddr.String())
		}

		return
	}
	defer func(conn *net.UDPConn) {
		if conn != nil {
			err := conn.Close()
			if err != nil {
				logger.Error(err)
			}
		}
	}(conn)

	err = conn.SetWriteDeadline(time.Now().Add(connTimeout))
	if err != nil {
		fmt.Println("SetWriteDeadline failed:", err)
		return
	}

	_, err = conn.Write([]byte(d.hostname))
	if err != nil {

		t, _ := d.ipFail.Get(ip.String())
		t = t + 1
		if d.ipAlwaysFail.Exists(ip.String()) {
			d.ipFail.SetWithTTL(ip.String(), udpMaxTryTime, d.ipFailRetry)
		} else {
			d.ipFail.SetWithTTL(ip.String(), t, d.ipFailRetry)
			logger.Warn(err, "Will retry", 10-tryTimes, "more times")
		}

		logger.Debug(err)
		return
	}

}
func (d *DiscoverService) readConfig() {

	if err := d._db.First(&d.config).Error; err != nil {
		logger.Errorf("failed to find database: %v", err)
	}
}
func (d *DiscoverService) StartService() {
	if !d.StartLock.TryLock() {
		return
	}
	defer d.StartLock.Unlock()
	d.udpDone = make(chan int)
	d.readConfig()
	if d.config.Enabled == false {
		return
	}

	logger.Info("starting discovery service")
	d.StartBroadcast()
	goroutine.RecoverGO(func() {
		d.listenAndSend(4085)
	})
}
func (d *DiscoverService) RestartService() error {
	if !d.RestartLock.TryLock() {
		return nil
	}
	defer d.RestartLock.Unlock()
	d.StopService()
	d.StartService()
	return nil

}
func isClosedConnError(err error) bool {
	netErr, ok := err.(*net.OpError)
	return ok && netErr.Err.Error() == "use of closed network connection"
}

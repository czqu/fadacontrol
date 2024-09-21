package discovery_service

import (
	"fadacontrol/internal/base/logger"
	"fadacontrol/pkg/utils"
	"net"
	"os"
	"time"
)

var ipFail map[string]int

const port = 4084

var hostname = ""

type NetInterface struct {
	MACAddr       string
	InterfaceName string
	IPAddresses   []string
}

var udpStopFlag = false

func listenAndSend(port int) {
	addr := net.UDPAddr{
		Port: port,
		IP:   net.IPv4zero,
	}
	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		logger.Error("Error listening:", err.Error())
		return
	}
	defer conn.Close()
	logger.Info("Listening on port", port)
	buffer := make([]byte, 1024)
	for {
		// 读取 UDP 消息
		n, remoteAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			logger.Warn("Error reading from UDP:", err)
			continue
		}
		logger.Debug("Received message from %s: %s", remoteAddr, string(buffer[:n]))

		_, err = conn.WriteToUDP([]byte(hostname), remoteAddr)

		if err != nil {
			logger.Warn("Error sending response:", err)
		} else {
			logger.Warn("Sent 'hello' to %s", remoteAddr)
		}
	}

}
func StopBroadcast() error {
	logger.Info("The UDP broadcast service is stopped")
	udpStopFlag = true
	return nil
}
func StartBroadcast() {
	if udpStopFlag {
		return
	}
	logger.Info("The UDP broadcast service is launched")
	var err error
	hostname, err = os.Hostname()
	if err != nil {
		logger.Warn("Error getting hostname:", err)
		StopBroadcast()
		return
	}
	//go listenAndSend(4085)
	go func() {
		logger.Debug("Sending UDP Broadcast ")
		for {

			if udpStopFlag {
				return
			}
			udpBroadcast()
			time.Sleep(2 * time.Second)
		}
	}()

}
func udpBroadcast() {
	sendUdp(net.IPv4bcast)
	ipFail = make(map[string]int)
	interfaces, err := utils.GetValidInterface(utils.UNSET)
	if err != nil {
		logger.Error("Error getting interface list:", err)
		return
	}
	for _, iface := range interfaces {
		for _, ipnet := range iface.IPAddresses {
			sendUdp(ipnet)
		}
	}

}
func sendUdp(ip net.IP) {

	//only broadcast ipv4
	if ip.To4() == nil {
		return
	}
	ip = ip.To4()
	tryTimes := ipFail[ip.String()]
	if tryTimes >= 10 {
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
		Port: port,
	})
	if err != nil {
		ipFail[ip.String()] += 1
		logger.Warn(err, "Will retry", 10-tryTimes, "more times")
		logger.Debug(err)
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

	_, err = conn.Write([]byte(hostname))
	if err != nil {
		ipFail[ip.String()] += 1
		logger.Warn(err, "Will retry", 10-tryTimes, "more times")
		logger.Debug(err)
		return
	}

}

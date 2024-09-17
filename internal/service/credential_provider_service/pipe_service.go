package credential_provider_service

import (
	"bytes"
	"fadacontrol/internal/base/exception"
	"fadacontrol/internal/base/logger"
	"fadacontrol/internal/entity"
	"fadacontrol/pkg/sys"
	"github.com/skip2/go-qrcode"
	"golang.org/x/image/bmp"
	"golang.org/x/image/draw"
	"image"
	"image/color"
	"io"
	"net"
)

type CredentialProviderService struct {
	pipe net.Conn
}

func NewCredentialProviderService() *CredentialProviderService {
	return &CredentialProviderService{}
}

const (
	pipePrefix    = `\\.\pipe\fc.pipe.`
	pipeCacheSize = 4 * 1024
)
const DataPipeName = pipePrefix + "v1.data.4k"

type pipeSendStatus struct {
	err    *exception.Exception
	packet *entity.PipePacket
}

var resp = make(chan pipeSendStatus)

func (p *CredentialProviderService) SetQrCode(contents string, size, borderSize int) error {
	qr, err := qrcode.New(contents, qrcode.Highest)

	if err != nil {
		return err
	}
	qr.DisableBorder = true
	qrImage := qr.Image(size)
	newSize := size + 2*borderSize
	// 创建一个新的 RGBA 图像
	borderedImage := image.NewRGBA(image.Rect(0, 0, newSize, newSize))
	white := color.RGBA{255, 255, 255, 255}
	draw.Draw(borderedImage, borderedImage.Bounds(), &image.Uniform{white}, image.Point{}, draw.Src)
	draw.Draw(borderedImage, qrImage.Bounds().Add(image.Point{borderSize, borderSize}), qrImage, image.Point{}, draw.Over)

	var buf bytes.Buffer

	// 将图像编码为 BMP 格式并写入缓冲区
	if err := bmp.Encode(&buf, borderedImage); err != nil {
		return err
	}
	err = p.SetFieldBitmap(buf.Bytes())
	if err != nil {
		return err
	}

	return nil
}
func (p *CredentialProviderService) Connect() {
	go func() {
		err := sys.ListenNamedPipeWithHandler(DataPipeName, p.PipeHandler)
		if err != nil {
			logger.Error(err.Error())
		}
	}()

}

func (p *CredentialProviderService) SetFieldBitmap(data []byte) *exception.Exception {
	if len(data) == 0 {
		logger.Debug("can not be zero")
		return exception.ErrParameterError
	}
	var packet entity.PipePacket
	packet.Tpe = entity.SetFieldBitmap
	packet.Size = uint32(len(data))
	packet.Data = data
	logger.Debug("set field bitmap")
	ret := p.SendData(&packet)
	logger.Debug("set field bitmap success")
	return ret

}
func (p *CredentialProviderService) SetCommandLinkText(text string) *exception.Exception {
	if len(text) == 0 {
		return exception.ErrParameterError
	}
	var packet entity.PipePacket
	packet.Tpe = entity.SetCommandClickText
	packet.Size = uint32(len(text))
	packet.Data = []byte(text)
	ret := p.SendData(&packet)
	return ret
}
func (p *CredentialProviderService) SendData(packet *entity.PipePacket) *exception.Exception {
	if p.pipe == nil {
		logger.Debug("pipe is nil")
		return exception.ErrSystemUnknownException
	}
	packetData, err := packet.Pack()
	if err != nil {
		logger.Debug("err")
		return exception.ErrSystemUnknownException
	}

	logger.Debugf("write data")
	if _, err := p.pipe.Write(packetData); err != nil {
		logger.Error("write data err")
		return exception.ErrSystemUnknownException
	}

	logger.Debugf("write data %d ", len(packetData))
	ret := p.getResp()
	logger.Debug("get resp ok")
	return ret
}
func (p *CredentialProviderService) getResp() *exception.Exception {
	select {
	case ret := <-resp:
		return ret.err
		//case <-time.After(time.Second * 1):
		//	return exception.ErrSystemUnknownException
	}
}

func (p *CredentialProviderService) PipeHandler(conn net.Conn) {
	//go func() {
	//	time.Sleep(5 * time.Second)
	//	p.pipe = conn
	//	logger.Debug("connect pipe")
	//	err := p.SetQrCode("test", 256)
	//	if err != nil {
	//		logger.Error(err.Error())
	//	}
	//	err = p.SetCommandLinkText("test")
	//	if err != nil {
	//		logger.Error(err.Error())
	//	}
	//}()

	defer conn.Close()
	p.pipe = conn
	for {
		var packet entity.PipePacket
		logger.Debug("read pipe")
		err := packet.Unpack(conn)

		if err == io.EOF {
			logger.Info("pipe has been closed ")
			return
		}
		if err != nil {
			logger.Error("pipe recv err")
			return
		}
		switch packet.Tpe {
		case entity.Resp:
			logger.Debugf("recv resp")
			if len(packet.Data) != 1 {
				logger.Debug("read err")
				resp <- pipeSendStatus{exception.ErrSystemUnknownException, nil}
				break
			}
			code := packet.Data[0]
			resp <- pipeSendStatus{exception.GetErrorByCode(int(code)), &packet}
			logger.Debug("over")
			break
		case entity.CommandClicked:
			logger.Debugf("receive clicked command")
			text := "{\n    \"hostname\": \"test_my_image\",\n    \"client_id\": \"12345789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz\"\n}\n"
			go p.SetQrCode(text, 256, 5)

			break
			logger.Debug("over")
		default:
			logger.Debug("read err")
			break
		}
	}

}

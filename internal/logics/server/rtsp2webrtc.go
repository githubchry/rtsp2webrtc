package server

import (
	"errors"
	"github.com/deepch/vdk/codec/h264parser"
	"github.com/pion/webrtc/v2"
	"github.com/pion/webrtc/v2/pkg/media"
	"log"
	"math/rand"
	"rtsp2webrtc/internal/logics/protos"
	"strings"
	"time"

	"github.com/deepch/vdk/format/rtsp"
)

func GetCodecInfo(req *protos.GetCodecInfoReq) (rsp *protos.GetCodecInfoRsp) {
	rsp = &protos.GetCodecInfoRsp{Status: &protos.Status{}}

	rtsp.DebugRtsp = true

	log.Println("connect", req.Url)

	// 连接到rtsp流地址
	session, err := rtsp.Dial(req.Url)
	if err != nil {
		rsp.Status.Code = -1
		rsp.Status.Message = "连接url错误:" + err.Error()
		return
	}

	session.RtpKeepAliveTimeout = 10 * time.Second

	// 获取rtsp流编码类型
	codec, err := session.Streams()
	if err != nil {
		rsp.Status.Code = -2
		rsp.Status.Message = "rtsp会话错误:" + err.Error()
		return
	}

	// 打印rtsp流编码类型
	for i, data := range codec {
		log.Printf("rtsp codec%d: %s", i, data.Type().String())
	}

	session.Close()

	return
}

func ExchangeSdp(req *protos.ExchangeSdpReq) (rsp *protos.ExchangeSdpRsp) {
	rsp = &protos.ExchangeSdpRsp{Status: &protos.Status{}}

	log.Println("req sdp len: ", len(req.Sdp))

	offer := webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  req.Sdp,
	}

	api, err := NewMediaEngineAPI(offer)
	if err != nil {
		rsp.Status.Code = -2
		rsp.Status.Message = "初始化媒体引擎异常:" + err.Error()
		return
	}

	pc, videoTrack, err := NewPC(api, offer)
	if err != nil {
		rsp.Status.Code = -3
		rsp.Status.Message = "初始化PeerConnection:" + err.Error()
		return
	}

	rsp.Sdp = pc.LocalDescription().SDP
	log.Println("rsp sdp len: ", len(rsp.Sdp))

	pc.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		log.Printf("Connection State has changed => %s \n", connectionState.String())
		switch connectionState {
		case webrtc.ICEConnectionStateConnected:
			// 连接成功 起一条线程发送音视频数据到对端
			log.Println("双向连接成功")
			go SendRtspStream(videoTrack, req.Url)
		case webrtc.ICEConnectionStateDisconnected:
			// 连接断开, 停止线程
		}

	})

	return
}

var payloadType uint8

func NewMediaEngineAPI(offer webrtc.SessionDescription) (api *webrtc.API, err error) {

	mediaEngine := webrtc.MediaEngine{}

	err = mediaEngine.PopulateFromSDP(offer)
	if err != nil {
		log.Println("PopulateFromSDP error", err)
		return
	}

	for _, videoCodec := range mediaEngine.GetCodecsByKind(webrtc.RTPCodecTypeVideo) {
		if videoCodec.Name == "H264" && strings.Contains(videoCodec.SDPFmtpLine, "packetization-mode=1") {
			payloadType = videoCodec.PayloadType
			break
		}
	}

	if payloadType == 0 {
		log.Println("Remote peer does not support H264")
		err = errors.New("Remote peer does not support H264")
		return
	}

	if payloadType != 126 {
		log.Println("Video might not work with codec", payloadType)
	}

	log.Println("Work payloadType", payloadType)

	api = webrtc.NewAPI(webrtc.WithMediaEngine(mediaEngine))

	return
}

func NewPC(api *webrtc.API, offer webrtc.SessionDescription) (peerConnection *webrtc.PeerConnection, videoTrack *webrtc.Track, err error) {

	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	// 这里会浪费掉600ms左右 测试过把config置空也没有效果
	peerConnection, err = api.NewPeerConnection(config)
	if err != nil {
		log.Println("NewPeerConnection error", err)
		return
	}

	//ADD KeepAlive Timer
	timer1 := time.NewTimer(time.Second * 2)
	peerConnection.OnDataChannel(func(d *webrtc.DataChannel) {
		// Register text message handling
		d.OnMessage(func(msg webrtc.DataChannelMessage) {
			log.Printf("Message from DataChannel '%s': '%s'\n", d.Label(), string(msg.Data))
			timer1.Reset(2 * time.Second)
		})
	})

	//ADD Video Track
	videoTrack, err = peerConnection.NewTrack(payloadType, rand.Uint32(), "video", "lable_chry_video")
	if err != nil {
		log.Fatalln("NewTrack", err)
	}
	_, err = peerConnection.AddTransceiverFromTrack(videoTrack,
		webrtc.RtpTransceiverInit{
			Direction: webrtc.RTPTransceiverDirectionSendonly,
		},
	)
	if err != nil {
		log.Println("AddTransceiverFromTrack error", err)
		return
	}
	_, err = peerConnection.AddTrack(videoTrack)
	if err != nil {
		log.Println("AddTrack error", err)
		return
	}

	//ADD Audio Track

	// Set sdp
	if err = peerConnection.SetRemoteDescription(offer); err != nil {
		log.Println("SetRemoteDescription error", err, offer.SDP)
		return
	}
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		log.Println("CreateAnswer error", err)
		return
	}

	if err = peerConnection.SetLocalDescription(answer); err != nil {
		log.Println("SetLocalDescription error", err)
		return
	}

	return
}

func SendRtspStream(videoTrack *webrtc.Track, url string) {

	log.Println(url)

	// 连接到rtsp流地址
	session, err := rtsp.Dial(url)
	if err != nil {
		log.Println("连接url错误:" + err.Error())
		return
	}

	session.RtpKeepAliveTimeout = 10 * time.Second

	// 获取rtsp流编码类型
	codec, err := session.Streams()
	if err != nil {
		log.Println("rtsp会话错误:" + err.Error())
		return
	}

	// 打印rtsp流编码类型
	var sps []byte
	var pps []byte
	for i, data := range codec {
		log.Printf("rtsp codec%d: %s", i, data.Type().String())
		if data.Type().IsVideo() {
			sps = data.(h264parser.CodecData).SPS()
			pps = data.(h264parser.CodecData).PPS()
		}
	}

	if len(sps) <= 0 || len(pps) <= 0 {
		log.Printf("sps len: %d, pps len: %d\n", len(sps), len(pps))
	}

	var Vpre time.Duration
	var start bool

	for {
		pck, err := session.ReadPacket()
		if err != nil {
			log.Println("读rtsp数据错误:", err)
			break
		}

		if pck.IsKeyFrame {
			start = true
		}

		if !start {
			continue
		}

		// pck.Data前面4个字节为帧数据长度, 即等于 (len(pck.Data)-4)
		if pck.IsKeyFrame {
			pck.Data = append([]byte("\000\000\000\001"+string(sps)+"\000\000\000\001"+string(pps)+"\000\000\000\001"), pck.Data[4:]...)
		} else {
			// 经测试可直接发送pck.Data[4:]出去, 不需要再前面加00 00 00 01
			pck.Data = append([]byte("\000\000\000\001"), pck.Data[4:]...)
		}

		var Vts time.Duration
		if pck.Idx == 0 && videoTrack != nil {
			if Vpre != 0 {
				Vts = pck.Time - Vpre
			}
			samples := uint32(90000 / 1000 * Vts.Milliseconds())
			err := videoTrack.WriteSample(media.Sample{Data: pck.Data, Samples: samples})
			if err != nil {
				return
			}
			Vpre = pck.Time
		}
		//else if pck.Idx == 1 && audioTrack != nil {
		//    err := audioTrack.WriteSample(media.Sample{Data: pck.Data, Samples: uint32(len(pck.Data))})
		//    if err != nil {
		//        return
		//    }
		//}
	}

	err = session.Close()
	if err != nil {
		log.Println("session Close error", err)
	}
	log.Println("session Close")
}

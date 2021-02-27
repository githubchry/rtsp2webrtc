package controller

import (
	"context"
	"rtsp2webrtc/internal/logics/protos"
	"rtsp2webrtc/internal/logics/server"
)

func GetCodecInfoHandler(ctx context.Context, req *protos.GetCodecInfoReq) (rsp *protos.GetCodecInfoRsp) {

	rsp = &protos.GetCodecInfoRsp{Status: &protos.Status{}}

	// 校验参数
	if len(req.Url) <= 0 {
		rsp.Status.Code = -2
		rsp.Status.Message = "url参数异常"
		return
	}

	// 调用真正的api
	return server.GetCodecInfo(req)
}

func ExchangeSdpHandler(ctx context.Context, req *protos.ExchangeSdpReq) (rsp *protos.ExchangeSdpRsp) {

	rsp = &protos.ExchangeSdpRsp{Status: &protos.Status{}}

	// 校验参数
	if len(req.Sdp) <= 0 {
		rsp.Status.Code = -2
		rsp.Status.Message = "sdp参数异常"
		return
	}

	// 调用真正的api
	return server.ExchangeSdp(req)
}

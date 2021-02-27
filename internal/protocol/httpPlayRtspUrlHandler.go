package protocol

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"rtsp2webrtc/internal/controller"
	"rtsp2webrtc/internal/logics/protos"
)

func HTTPGetCodecInfoHandler(w http.ResponseWriter, r *http.Request) {

	req := &protos.GetCodecInfoReq{}
	rsp := &protos.GetCodecInfoRsp{Status: &protos.Status{}}
	//把protobuf二进制数据转成logics.UserLoginReq结构体
	data, _ := ioutil.ReadAll(r.Body)
	if err := json.Unmarshal(data, req); err != nil {
		log.Println("Failed to parse json:", err)
		rsp.Status.Code = -1
		rsp.Status.Message = "请求格式异常"
		return
	}

	rsp = controller.GetCodecInfoHandler(r.Context(), req)
	data, _ = json.Marshal(&rsp)
	w.Write(data)
}

func HTTPExchangeSdpHandler(w http.ResponseWriter, r *http.Request) {

	req := &protos.ExchangeSdpReq{}
	rsp := &protos.ExchangeSdpRsp{Status: &protos.Status{}}
	//把protobuf二进制数据转成logics.UserLoginReq结构体
	data, _ := ioutil.ReadAll(r.Body)
	if err := json.Unmarshal(data, req); err != nil {
		log.Println("Failed to parse json:", err)
		rsp.Status.Code = -1
		rsp.Status.Message = "请求格式异常"
		return
	}

	rsp = controller.ExchangeSdpHandler(r.Context(), req)
	data, _ = json.Marshal(&rsp)
	w.Write(data)
}

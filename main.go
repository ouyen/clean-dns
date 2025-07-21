package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/miekg/dns"
)

// 简单的DoH上游地址，可以选择 Cloudflare 或 Google
//const dohURL = "https://dns.google/dns-query"

//const dohURL = "https://cloudflare-dns.com/dns-query"

const dohURL = "https://dns.alidns.com/dns-query"

func handleDNSRequest(w dns.ResponseWriter, req *dns.Msg) {
	// 打包DNS请求为二进制
	packedReq, err := req.Pack()
	if err != nil {
		log.Printf("Pack error: %v", err)
		dns.HandleFailed(w, req)
		return
	}

	// 通过HTTP POST发送到DoH服务器
	httpResp, err := http.Post(
		dohURL,
		"application/dns-message",
		bytes.NewReader(packedReq),
	)
	if err != nil {
		log.Printf("HTTP POST error: %v", err)
		dns.HandleFailed(w, req)
		return
	}
	defer httpResp.Body.Close()

	// 读取并解包返回的二进制DNS响应
	respData, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		log.Printf("ReadAll error: %v", err)
		dns.HandleFailed(w, req)
		return
	}

	respMsg := &dns.Msg{}
	if err := respMsg.Unpack(respData); err != nil {
		log.Printf("Unpack error: %v", err)
		dns.HandleFailed(w, req)
		return
	}

	// 回复DNS客户端
	w.WriteMsg(respMsg)
}

func main() {
	// 监听本地 1053 UDP 端口
	addr := ":1053"
	server := &dns.Server{Addr: addr, Net: "udp"}

	dns.HandleFunc(".", handleDNSRequest)

	fmt.Printf("Listening on %s for DNS requests (UDP)...\n", addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

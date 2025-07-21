package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/miekg/dns"
)

// 简单的DoH上游地址，可以选择 Cloudflare 或 Google
//const dohURL = "https://dns.google/dns-query"

//const dohURL = "https://cloudflare-dns.com/dns-query"

//const dohURL = "https://dns.alidns.com/dns-query"

func handleDNSRequest(w dns.ResponseWriter, req *dns.Msg) {
	// 打包DNS请求为二进制
	failedResp := new(dns.Msg)
	failedResp.SetRcode(req, dns.RcodeServerFailure)
	packedReq, err := req.Pack()
	if err != nil {
		log.Printf("Pack error: %v", err)
		_ = w.WriteMsg(failedResp)
		return
	}

	// 通过HTTP POST发送到DoH服务器

	dohURL := "https://104.16.248.249/dns-query"

	// 手动创建 http.Request，设置 Host header
	dohreq, err := http.NewRequest("POST", dohURL, bytes.NewReader(packedReq))
	if err != nil {
		log.Printf("NewRequest error: %v", err)
		_ = w.WriteMsg(failedResp)
		return
	}
	dohreq.Header.Set("Content-Type", "application/dns-message")
	dohreq.Header.Set("accept", "application/dns-json")
	dohreq.Host = "cloudflare-dns.com"

	// 跳过证书校验（等效于 curl -k）
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Do(dohreq)
	if err != nil {
		log.Printf("HTTP POST error: %v", err)
		_ = w.WriteMsg(failedResp)
		return
	}
	defer func() { _ = resp.Body.Close() }()

	// 读取并解包返回的二进制DNS响应
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("ReadAll error: %v", err)
		_ = w.WriteMsg(failedResp)
		return
	}

	respMsg := &dns.Msg{}
	if err := respMsg.Unpack(respData); err != nil {
		log.Printf("Unpack error: %v", err)
		_ = w.WriteMsg(failedResp)
		return
	}

	// 回复DNS客户端
	_ = w.WriteMsg(respMsg)
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

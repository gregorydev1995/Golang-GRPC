package main

import (
	"github.com/unknwon/goconfig"
	"golang.org/x/net/context"
	"gome/api"
	"google.golang.org/grpc"
	"log"
)

// main 方法实现对 gRPC 接口的请求
func main() {
	// 为了调用服务方法，我们首先创建一个 gRPC channel 和服务器交互。
	// 可以使用 DialOptions 在 grpc.Dial 中设置授权认证（如， TLS，GCE认证，JWT认证）
	config, err := goconfig.LoadConfigFile("config.ini")
	host, _ := config.GetValue("grpc", "host")
	port, _ := config.GetValue("grpc", "port")
	conn, err := grpc.Dial(host+":"+port, grpc.WithInsecure())
	if err != nil {
		log.Fatalln("Can't connect: " + host + ":" + port)
	}
	defer conn.Close()
	// 一旦 gRPC channel 建立起来，我们需要一个客户端 存根 去执行 RPC
	client := api.NewOrderClient(conn)

	// 0-buy,1-sale
	order := api.OrderRequest{Uuid: "1", Oid: "4", Symbol: "eth2usdt", Transaction: 1, Price: 0.5, Volume: 2}
	// 调用简单 RPC
	resp, err := client.DoOrder(context.Background(), &order)
	// 如果调用没有返回错误，那么我们就可以从服务器返回的第一个返回值中读到响应信息
	if err != nil {
		log.Fatalln("Do Format error:" + err.Error())
	}
	log.Println(resp.Message)
}

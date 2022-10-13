package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/Eric0329/grpc-go/greet/proto"
)

var addr string
var port int

func init() {
	flag.StringVar(&addr, "h", "0.0.0.0", "Server IP")
	flag.IntVar(&port, "p", 50051, "Server PORT")
}

func rpcHello(c pb.HellowServiceClient) {
	resp, err := c.SayHello(context.Background(), &pb.HelloReq{
		Greeting: "pass msg from rpcHello()",
	})

	if nil != err {
		log.Println("error to call gRPC SayHello", err)
	} else {
		log.Println("reply msg from server:", resp.Reply)
	}
}

func rpcHelloPlus(c pb.HellowServiceClient) {
	stream, err := c.SayHelloManyTimes(
		context.Background(),
		&pb.HelloReq{Greeting: "pass msg from rpcHelloPlus()"})

	if nil != err {
		log.Fatalln(err)
	}

	for {
		msg, err := stream.Recv()

		if io.EOF == err {
			log.Println("read the end of resp")
			break
		}

		if nil != err {
			log.Fatalln("error from reading response")
			break
		}

		log.Println(msg)
	}
}

func rpcDownload(c pb.HellowServiceClient) {
	stream, err := c.Download(context.Background(), &emptypb.Empty{})
	if nil != err {
		log.Fatalln(err)
	}

	//-- create a file to save the download bytes with dynamic suffix filename
	f, err := os.CreateTemp("/home/eric/", "download*.tmp")
	if err != nil {
		log.Fatalln(err)
	}

	//-- call Recv to get the file chunk (httpBody.Data) until the EOF
	for {
		httpBody, err := stream.Recv()

		if nil != err {
			if io.EOF == err {
				log.Println("download is completed")

				if err = f.Close(); nil != err {
					log.Fatalln(err)
				} else {
					log.Println("save file into: ", f.Name())
				}

				break
			}

			log.Fatal(err)
			break
		}

		//-- write the download chunk to the file
		_, err = f.Write(httpBody.Data)
		if nil != err {
			log.Fatalln(err)
			break
		}
	}

}

func rpcSum(c pb.CalculatorServiceClient) {
	resp, err := c.Sum(context.Background(),
		&pb.CalcReq{
			N1: 1,
			N2: 2,
		},
	)

	if nil != err {
		log.Fatalln("unable to rpc Sum(): ", err)
	} else {
		log.Println(resp.Sum)
	}

}

func main() {
	flag.Parse()

	addrPort := fmt.Sprintf("%v:%v", addr, port)
	//	conn, err := grpc.Dial(addrPort)
	conn, err := grpc.Dial(addrPort, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if nil != err {
		log.Fatalf("cann't connect to server %v, %v\n", addrPort, err)
	}

	defer conn.Close()

	//	c_calc := pb.NewCalculatorServiceClient(conn)
	//	rpcSum(c_calc)

	c_hello := pb.NewHellowServiceClient(conn)
	rpcDownload(c_hello)
	// rpcHelloPlus(c_hello)

}

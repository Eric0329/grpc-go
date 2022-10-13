package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"

	"google.golang.org/genproto/googleapis/api/httpbody"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/Eric0329/grpc-go/greet/proto"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

type Server struct {
	pb.HellowServiceServer
	pb.CalculatorServiceServer
}

func (s *Server) SayHello(ctx context.Context, in *pb.HelloReq) (*pb.HelloResp, error) {
	log.Println("SayHello is invoked with ", in)

	return &pb.HelloResp{
		Reply: "the msg comes from client: " + in.Greeting,
	}, nil
}

func (s *Server) SayHelloManyTimes(in *pb.HelloReq, stream pb.HellowService_SayHelloManyTimesServer) error {
	log.Println("SayHelloManyTimes is invoked with ", in)

	for i := 0; i < 10; i++ {
		stream.Send(
			&pb.HelloResp{Reply: strconv.Itoa(i)})
	}

	return nil
}

func (s *Server) Download(in *empty.Empty, stream pb.HellowService_DownloadServer) error {
	bufSize := 64 * 1024 //-- 64K, tweak this as desired
	f, err := os.Open("/home/eric/Downloads/elasticsearch-8.4.3-linux-x86_64.tar.gz")

	if err != nil {
		fmt.Println(err)
		return err
	}

	defer f.Close()

	buff := make([]byte, bufSize)
	for {
		bytesRead, err := f.Read(buff)
		if err != nil {
			if err == io.EOF {
				log.Println(f.Name(), " transfer is complete.")
				break
			}

			log.Fatalln(err)
			return err
		}

		chunk := &httpbody.HttpBody{
			Data: buff[:bytesRead],
		}

		err = stream.Send(chunk)
		if err != nil {
			log.Fatalln("error while sending chunk:", err)
			return err
		}
	}

	return nil
}

var (
	addr           string
	grpcPort       int
	grpcGWHttpPort int
)

func init() {
	flag.StringVar(&addr, "h", "0.0.0.0", "Server listen IP")
	flag.IntVar(&grpcPort, "p", 50051, "gRPC Server listen PORT")
	flag.IntVar(&grpcGWHttpPort, "hp", 50080, "gRPC Gateway Http Server listen PORT")
}

func grpcGateWayHTTPServer() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Register gRPC server endpoint
	// Note: Make sure the gRPC server is running properly and accessible
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err := pb.RegisterHellowServiceHandlerFromEndpoint(ctx, mux, fmt.Sprintf("%v:%v", addr, grpcPort), opts)
	if err != nil {
		log.Fatalln(err)
		return err
	}

	log.Println("gRPC Gateway HTTP server is going to serve on ", fmt.Sprintf("%v:%v", addr, grpcGWHttpPort))

	// Start HTTP server (and proxy calls to gRPC server endpoint)
	err = http.ListenAndServe(fmt.Sprintf("%v:%v", addr, grpcGWHttpPort), mux)
	if nil != err {
		log.Fatalln(err)
		return err
	}

	return nil
}

func grpcServer() error {
	lis, err := net.Listen("tcp", fmt.Sprintf("%v:%v", addr, grpcPort))
	if nil != err {
		log.Fatalf("error to listen %v", err)
		return err
	}

	defer lis.Close()

	log.Printf("gRPC server is going to serve on %v:%v\n", addr, grpcPort)

	s := grpc.NewServer()
	pb.RegisterHellowServiceServer(s, &Server{})
	//	pb.RegisterCalculatorServiceServer(s, &Server{})

	if err = s.Serve(lis); nil != err {
		log.Fatalf("failed to server: %v", err)
		return err
	}

	defer s.Stop()

	return nil
}

func main() {
	flag.Parse()

	go grpcGateWayHTTPServer()
	grpcServer()
}

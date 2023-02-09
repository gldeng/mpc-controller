package main

import (
	"context"
	"flag"
	"github.com/avalido/mpc-controller/core/mpc"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/tests/mocks/mpc_server/services"
	"google.golang.org/grpc"
	"net"
)

type server struct {
	Log logger.Logger
	K   *services.KeyGenerator
	S   *services.Signer
	P   *services.Provider
	mpc.UnimplementedMpcServer
}

func (s *server) Keygen(ctx context.Context, in *mpc.KeygenRequest) (*mpc.KeygenResponse, error) {
	in_ := services.KeygenInput{in.RequestId, in.ParticipantPublicKeys, int(in.Threshold)}
	err := s.K.Keygen(&in_)
	if err != nil {
		s.Log.Error("keygen got an error", []logger.Field{{"reqId", in.RequestId}, {"error", err}}...)
	}
	return &mpc.KeygenResponse{RequestId: in.RequestId}, err
}

func (s *server) Sign(ctx context.Context, in *mpc.SignRequest) (*mpc.SignResponse, error) {
	in_ := services.SignInput{in.RequestId, in.PublicKey, in.ParticipantPublicKeys, in.Hash}
	err := s.S.Sign(&in_)
	if err != nil {
		s.Log.Error("sign got an error", []logger.Field{{"reqId", in.RequestId}, {"error", err}}...)
	}
	return &mpc.SignResponse{RequestId: in.RequestId}, err
}

func (s *server) CheckResult(ctx context.Context, in *mpc.CheckResultRequest) (*mpc.CheckResultResponse, error) {
	res_ := mpc.CheckResultResponse{}
	res_.RequestId = in.RequestId

	in_ := services.ResultInput{in.RequestId}
	res, err := s.P.Result(&in_)
	if err != nil {
		res_.RequestStatus = mpc.CheckResultResponse_ERROR
		s.Log.Error("result got an error", []logger.Field{{"reqId", in.RequestId}, {"error", err}}...)
		return &res_, err
	}

	res_.Result = res.Result
	res_.RequestStatus = status(res.RequestStatus)
	res_.RequestType = typ(res.RequestType)
	return &res_, nil
}

func status(s services.RequestStatus) mpc.CheckResultResponse_REQUEST_STATUS {
	switch s {
	case services.StatusReceived:
		return mpc.CheckResultResponse_RECEIVED
	case services.StatusProcessing:
		return mpc.CheckResultResponse_PROCESSING
	case services.StatusDone:
		return mpc.CheckResultResponse_DONE
	case services.StatusOfflineStageDone:
		return mpc.CheckResultResponse_PROCESSING
	case services.StatusError:
		return mpc.CheckResultResponse_ERROR
	default:
		return mpc.CheckResultResponse_UNKNOWN_STATUS
	}
}

func typ(t services.RequestType) mpc.CheckResultResponse_REQUEST_TYPE {
	switch t {
	case services.TypeKeygen:
		return mpc.CheckResultResponse_KEYGEN
	case services.TypeSign:
		return mpc.CheckResultResponse_SIGN
	default:
		return mpc.CheckResultResponse_UNKNOWN_TYPE
	}
}

func main() {
	var listenAddr = flag.String("addr", ":9000", "address to listen")
	var participants = flag.Int("p", 7, "number of participants in the group")
	var threshold = flag.Int("t", 4, "number of the group threshold")
	flag.Parse()

	logger.DevMode = true
	logger.UseConsoleEncoder = true
	myLogger := logger.Default()

	lis, err := net.Listen("tcp", *listenAddr)
	if err != nil {
		myLogger.Fatal("failed to listen", []logger.Field{{"error", err}}...)
	}
	s := grpc.NewServer()
	mpc.RegisterMpcServer(s, &server{Log: myLogger, K: &services.KeyGenerator{myLogger, *participants}, S: &services.Signer{myLogger, *threshold}, P: &services.Provider{}})
	myLogger.Info("server listening", []logger.Field{{"address", *listenAddr}}...)
	if err := s.Serve(lis); err != nil {
		myLogger.Fatal("failed to serve", []logger.Field{{"error", err}}...)
	}
}

package api

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"

	"github.com/AccumulateNetwork/metrics-api/global"

	"go.neonxp.dev/jsonrpc2/rpc"
	"go.neonxp.dev/jsonrpc2/transport"
)

type Server struct {
	ctx    context.Context
	cancel context.CancelFunc
	r      *rpc.RpcServer
}

func StartAPI(port int) error {

	r := rpc.New(
		rpc.WithTransport(&transport.HTTP{Bind: ":" + strconv.Itoa(port), CORSOrigin: "*"}), // HTTP transport
	)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	s := &Server{
		ctx:    ctx,
		cancel: cancel,
		r:      r,
	}

	s.r.Register("staking", rpc.H(s.Staking))

	if err := s.r.Run(ctx); err != nil {
		log.Fatal(err)
	}

	return nil
}

func (s *Server) Staking(ctx context.Context, _ *NoArgs) (interface{}, error) {
	return &global.StakingRecords, nil
}

type NoArgs struct {
}

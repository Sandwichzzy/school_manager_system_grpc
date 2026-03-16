package handlers

import (
	"context"

	"github.com/Sandwichzzy/school_manager_system_grpc/internals/repositories/mongodb"
	pb "github.com/Sandwichzzy/school_manager_system_grpc/proto/gen"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) AddExecs(ctx context.Context, req *pb.Execs) (*pb.Execs, error) {
	for _, exec := range req.GetExecs() {
		if exec.Id != "" {
			return nil, status.Error(codes.InvalidArgument, "request is incorrect: non-empty ID field are not allowed.")
		}
	}
	addedExecs, err := mongodb.AddExecsToDb(ctx, req.GetExecs())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.Execs{Execs: addedExecs}, nil
}

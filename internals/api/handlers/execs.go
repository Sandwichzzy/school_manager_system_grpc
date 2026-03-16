package handlers

import (
	"context"

	"github.com/Sandwichzzy/school_manager_system_grpc/internals/models"
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

func (s *Server) GetExecs(ctx context.Context, req *pb.GetExecsRequest) (*pb.Execs, error) {
	// Filtering,getting the filters from the request,another function
	filter, err := buildFilter(req.Exec, &models.Exec{})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	// Sorting,getting the sort options from the request,another function
	sortOptions := buildSortOptions(req.GetSortBy())
	// Access the database to fetch data,another function

	execs, err := mongodb.GetExecsFromDb(ctx, sortOptions, filter)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.Execs{Execs: execs}, nil
}

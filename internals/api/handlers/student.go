package handlers

import (
	"context"

	"github.com/Sandwichzzy/school_manager_system_grpc/internals/models"
	"github.com/Sandwichzzy/school_manager_system_grpc/internals/repositories/mongodb"
	pb "github.com/Sandwichzzy/school_manager_system_grpc/proto/gen"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) AddStudents(ctx context.Context, req *pb.Students) (*pb.Students, error) {
	for _, student := range req.GetStudents() {
		if student.Id != "" {
			return nil, status.Error(codes.InvalidArgument, "request is incorrect: non-empty ID field are not allowed.")
		}
	}
	addedStudents, err := mongodb.AddStudentsToDb(ctx, req.GetStudents())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.Students{Students: addedStudents}, nil
}

func (s *Server) GetStudents(ctx context.Context, req *pb.GetStudentsRequest) (*pb.Students, error) {
	// Filtering,getting the filters from the request,another function
	filter, err := buildFilter(req.Student, &models.Student{})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	// Sorting,getting the sort options from the request,another function
	sortOptions := buildSortOptions(req.GetSortBy())
	// Access the database to fetch data,another function

	pageNumber := req.GetPageNumber()
	pageSize := req.GetPageSize()
	if pageNumber < 1 {
		pageNumber = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	students, err := mongodb.GetStudentsFromDb(ctx, sortOptions, filter, pageNumber, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.Students{Students: students}, nil
}

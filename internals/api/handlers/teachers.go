package handlers

import (
	"context"
	"errors"
	"fmt"

	"github.com/Sandwichzzy/school_manager_system_grpc/internals/models"
	"github.com/Sandwichzzy/school_manager_system_grpc/internals/repositories/mongodb"
	"github.com/Sandwichzzy/school_manager_system_grpc/pkg/utils"
	pb "github.com/Sandwichzzy/school_manager_system_grpc/proto/gen"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) AddTeachers(ctx context.Context, req *pb.Teachers) (*pb.Teachers, error) {
	for _, teacher := range req.GetTeachers() {
		if teacher.Id != "" {
			return nil, status.Error(codes.InvalidArgument, "request is incorrect: non-empty ID field are not allowed.")
		}
	}
	addedTeachers, err := mongodb.AddTeachersToDb(ctx, req.GetTeachers())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.Teachers{Teachers: addedTeachers}, nil
}

func (s *Server) GetTeachers(ctx context.Context, req *pb.GetTeachersRequest) (*pb.Teachers, error) {
	// Filtering,getting the filters from the request,another function
	filter, err := buildFilter(req.Teacher, &models.Teacher{})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	// Sorting,getting the sort options from the request,another function
	sortOptions := buildSortOptions(req.GetSortBy())
	// Access the database to fetch data,another function

	teachers, err := mongodb.GetTeachersFromDB(ctx, sortOptions, filter)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.Teachers{Teachers: teachers}, nil
}

func (s *Server) UpdateTeachers(ctx context.Context, req *pb.Teachers) (*pb.Teachers, error) {
	updatedTeachers, err := mongodb.ModifyTeachersInDB(ctx, req.Teachers)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.Teachers{Teachers: updatedTeachers}, nil

}

func (s *Server) DeleteTeachers(ctx context.Context, req *pb.TeacherIds) (*pb.DeleteTeachersConfirmation, error) {
	ids := req.GetIds()
	if len(ids) == 0 {
		return nil, utils.ErrorHandler(errors.New("no teacher ids provided"), "no teacher ids provided")
	}
	var teacherIdsToDelete []string
	for _, v := range ids {
		teacherIdsToDelete = append(teacherIdsToDelete, v.Id)
	}

	client, err := mongodb.CreateMongoClient()
	if err != nil {
		return nil, utils.ErrorHandler(err, "internal error")
	}
	defer client.Disconnect(ctx)

	objectIds := make([]primitive.ObjectID, len(teacherIdsToDelete))
	for i, id := range teacherIdsToDelete {
		objectId, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return nil, utils.ErrorHandler(err, fmt.Sprintf("incorrect id: %v", id))
		}
		objectIds[i] = objectId
	}
	filter := bson.M{"_id": bson.M{"$in": objectIds}}

	result, err := client.Database("school").Collection("teachers").DeleteMany(ctx, filter)
	if err != nil {
		return nil, utils.ErrorHandler(err, "internal error")
	}
	if result.DeletedCount == 0 {
		return nil, utils.ErrorHandler(err, "no teachers were deleted. Ids do not exist")
	}

	deletedIds := make([]string, result.DeletedCount)
	for i, id := range objectIds {
		deletedIds[i] = id.Hex()
	}

	return &pb.DeleteTeachersConfirmation{
		Status:     "Teachers successfully deleted",
		DeletedIds: deletedIds,
	}, nil
}

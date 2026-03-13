package handlers

import pb "github.com/Sandwichzzy/school_manager_system_grpc/proto/gen"

type Server struct {
	pb.UnimplementedStudentsServiceServer
	pb.UnimplementedExecsServiceServer
	pb.UnimplementedTeachersServiceServer
}

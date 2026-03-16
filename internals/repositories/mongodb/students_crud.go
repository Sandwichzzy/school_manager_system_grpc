package mongodb

import (
	"context"
	"fmt"

	"github.com/Sandwichzzy/school_manager_system_grpc/internals/models"
	"github.com/Sandwichzzy/school_manager_system_grpc/pkg/utils"
	pb "github.com/Sandwichzzy/school_manager_system_grpc/proto/gen"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func AddStudentsToDb(ctx context.Context, studentsFromReq []*pb.Student) ([]*pb.Student, error) {
	client, err := CreateMongoClient()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Failed to create MongoDB client:")
	}
	defer client.Disconnect(ctx)

	// 准备一个切片，用于存放转换后的学生模型（指针），长度与请求中的学生数量一致
	newStudents := make([]*models.Student, len(studentsFromReq))
	// 遍历请求中的每个 protobuf 学生对象
	for i, pbStudent := range studentsFromReq {
		newStudents[i] = mapPbStudentToModelStudent(pbStudent)
	}
	var addedStudents []*pb.Student
	for _, student := range newStudents {
		result, err := client.Database("school").Collection("students").InsertOne(ctx, student)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Failed to insert student in database:")
		}
		objectId, ok := result.InsertedID.(primitive.ObjectID)
		if ok {
			student.Id = objectId.Hex()
		}
		fmt.Println(objectId)

		pbStudent := mapModelStudentToPb(*student)
		addedStudents = append(addedStudents, pbStudent)
	}
	return addedStudents, nil
}

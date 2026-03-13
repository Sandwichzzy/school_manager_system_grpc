package mongodb

import (
	"context"
	"fmt"
	"reflect"

	pb "github.com/Sandwichzzy/school_manager_system_grpc/proto/gen"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/Sandwichzzy/school_manager_system_grpc/internals/models"
	"github.com/Sandwichzzy/school_manager_system_grpc/pkg/utils"
)

func AddTeachersToDb(ctx context.Context, teachersFromReq []*pb.Teacher) ([]*pb.Teacher, error) {
	client, err := CreateMongoClient()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Failed to create MongoDB client:")
	}
	defer client.Disconnect(ctx)

	// 准备一个切片，用于存放转换后的教师模型（指针），长度与请求中的教师数量一致
	newTeachers := make([]*models.Teacher, len(teachersFromReq))
	// 遍历请求中的每个 protobuf 教师对象
	for i, pbTeacher := range teachersFromReq {
		newTeachers[i] = mapPbTeacherToModelTeacher(pbTeacher)
	}
	var addedTeachers []*pb.Teacher
	for _, teacher := range newTeachers {
		result, err := client.Database("school").Collection("teachers").InsertOne(ctx, teacher)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Failed to insert teacher in database:")
		}
		objectId, ok := result.InsertedID.(primitive.ObjectID)
		if ok {
			teacher.ID = objectId.Hex()
		}
		fmt.Println(objectId)

		pbTeacher := mapModelTeacherToPb(teacher)
		addedTeachers = append(addedTeachers, pbTeacher)
	}
	return addedTeachers, nil
}

func mapModelTeacherToPb(teacher *models.Teacher) *pb.Teacher {
	pbTeacher := &pb.Teacher{}
	modelVal := reflect.ValueOf(*teacher)
	pbVal := reflect.ValueOf(pbTeacher).Elem()

	for i := 0; i < modelVal.NumField(); i++ {
		modelField := modelVal.Field(i)
		modelFieldType := modelVal.Type().Field(i)
		// pbFieldType := pbVal.Type().Field(i)

		pbField := pbVal.FieldByName(modelFieldType.Name)
		if pbField.IsValid() && pbField.CanSet() {
			pbField.Set(modelField)
		}
	}
	return pbTeacher
}

func mapPbTeacherToModelTeacher(pbTeacher *pb.Teacher) *models.Teacher {
	modelTeacher := models.Teacher{}
	// 使用反射获取 protobuf 教师对象的可反射值（假设 pbTeacher 是指针，调用 Elem() 获取其指向的值）
	pbVal := reflect.ValueOf(pbTeacher).Elem()
	// 获取 modelTeacher 的可设置反射值（传递指针以便能够修改字段）
	modelVal := reflect.ValueOf(&modelTeacher).Elem()
	for i := 0; i < pbVal.NumField(); i++ {
		// 获取当前字段的反射值和字段名
		pbField := pbVal.Field(i)
		fieldName := pbVal.Type().Field(i).Name
		// 根据字段名从 modelTeacher 中查找对应的字段
		modelField := modelVal.FieldByName(fieldName)
		// 如果 modelField 存在且可设置，则将 protobuf 字段的值赋给它
		if modelField.IsValid() && modelField.CanSet() {
			modelField.Set(pbField)
		}
	}

	return &modelTeacher
}

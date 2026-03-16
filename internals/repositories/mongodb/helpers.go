package mongodb

import (
	"context"
	"reflect"

	"github.com/Sandwichzzy/school_manager_system_grpc/internals/models"
	"github.com/Sandwichzzy/school_manager_system_grpc/pkg/utils"
	pb "github.com/Sandwichzzy/school_manager_system_grpc/proto/gen"
	"go.mongodb.org/mongo-driver/mongo"
)

func decodeEntities[T any, M any](ctx context.Context, cursor *mongo.Cursor, newEntity func() *T, newModel func() *M) ([]*T, error) {
	var entities []*T
	for cursor.Next(ctx) {
		model := newModel()
		err := cursor.Decode(&model)
		if err != nil {
			return nil, utils.ErrorHandler(err, "internal error")
		}
		entity := newEntity()
		modelVal := reflect.ValueOf(model).Elem()
		pbVal := reflect.ValueOf(entity).Elem()

		for i := 0; i < modelVal.NumField(); i++ {
			modelField := modelVal.Field(i)
			modelFieldName := modelVal.Type().Field(i).Name

			pbField := pbVal.FieldByName(modelFieldName)
			if pbField.IsValid() && pbField.CanSet() {
				pbField.Set(modelField)
			}
		}
		entities = append(entities, entity)
	}
	err := cursor.Err()
	if err != nil {
		return nil, utils.ErrorHandler(err, "internal error")
	}
	return entities, nil
}

func mapModelTeacherToPb(teacherModel models.Teacher) *pb.Teacher {
	return mapModelToPb(teacherModel, func() *pb.Teacher { return &pb.Teacher{} })
}

// func mapModelStudentToPb(studentModel models.Student) *pb.Student {
// 	return mapModelToPb(studentModel, func() *pb.Student { return &pb.Student{} })
// }

// func mapModelExecToPb(execModel models.Exec) *pb.Exec {
// 	return mapModelToPb(execModel, func() *pb.Exec { return &pb.Exec{} })
// }

func mapModelToPb[M any, P any](model M, newPb func() *P) *P {
	pbStruct := newPb()
	modelVal := reflect.ValueOf(model)
	pbVal := reflect.ValueOf(pbStruct).Elem()

	for i := 0; i < modelVal.NumField(); i++ {
		modelField := modelVal.Field(i)
		modelFieldType := modelVal.Type().Field(i)
		// pbFieldType := pbVal.Type().Field(i)

		pbField := pbVal.FieldByName(modelFieldType.Name)
		if pbField.IsValid() && pbField.CanSet() {
			pbField.Set(modelField)
		}
	}
	return pbStruct
}

func mapPbTeacherToModelTeacher(pbTeacher *pb.Teacher) *models.Teacher {
	return mapPbToModel(pbTeacher, func() *models.Teacher { return &models.Teacher{} })
}

// func mapPbStudentToModelStudent(pbStudent *pb.Student) *models.Student {
// 	return mapPbToModel(pbStudent, func() *models.Student { return &models.Student{} })
// }

// func mapPbExecToModelExec(pbExec *pb.Exec) *models.Exec {
// 	return mapPbToModel(pbExec, func() *models.Exec { return &models.Exec{} })
// }

func mapPbToModel[P any, M any](pbStruct P, newModel func() *M) *M {
	modelStruct := newModel()
	// 使用反射获取 protobuf 教师对象的可反射值（假设 pbTeacher 是指针，调用 Elem() 获取其指向的值）
	pbVal := reflect.ValueOf(pbStruct).Elem()
	// 获取 modelTeacher 的可设置反射值（传递指针以便能够修改字段）
	modelVal := reflect.ValueOf(&modelStruct).Elem()
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

	return modelStruct
}

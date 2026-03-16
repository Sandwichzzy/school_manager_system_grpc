package mongodb

import (
	"context"
	"errors"
	"fmt"

	pb "github.com/Sandwichzzy/school_manager_system_grpc/proto/gen"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

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
			teacher.Id = objectId.Hex()
		}
		fmt.Println(objectId)

		pbTeacher := mapModelTeacherToPb(*teacher)
		addedTeachers = append(addedTeachers, pbTeacher)
	}
	return addedTeachers, nil
}

func GetTeachersFromDB(ctx context.Context, sortOptions bson.D, filter bson.M) ([]*pb.Teacher, error) {
	client, err := CreateMongoClient()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Internal Error")
	}
	defer client.Disconnect(ctx)
	coll := client.Database("school").Collection("teachers")
	var cursor *mongo.Cursor
	if len(sortOptions) < 1 {
		cursor, err = coll.Find(ctx, filter)
	} else {
		cursor, err = coll.Find(ctx, filter, options.Find().SetSort(sortOptions))
	}

	if err != nil {
		return nil, utils.ErrorHandler(err, "Internal error")
	}
	defer cursor.Close(ctx)

	teachers, err := decodeEntities(ctx, cursor, func() *pb.Teacher { return &pb.Teacher{} }, func() *models.Teacher { return &models.Teacher{} })
	if err != nil {
		return nil, err
	}
	return teachers, nil
}

func ModifyTeachersInDB(ctx context.Context, pbTeachers []*pb.Teacher) ([]*pb.Teacher, error) {
	client, err := CreateMongoClient()
	if err != nil {
		return nil, utils.ErrorHandler(err, "internal error")
	}
	defer client.Disconnect(ctx)

	var updatedTeachers []*pb.Teacher
	for _, teacher := range pbTeachers {
		if teacher.Id == "" {
			return nil, utils.ErrorHandler(errors.New("id cannot be blank"), "id cannnot be blank!")
		}
		modelTeacher := mapPbTeacherToModelTeacher(teacher)
		objId, err := primitive.ObjectIDFromHex(teacher.Id)
		if err != nil {
			return nil, utils.ErrorHandler(err, "invalid id")
		}

		//convert modelTeacher to BSON document
		modelDoc, err := bson.Marshal(modelTeacher)
		if err != nil {
			return nil, utils.ErrorHandler(err, "internal error")
		}

		//bson.Unmarshal 将刚刚编码的二进制数据解码到 updateDoc 中，即把结构体字段和值映射到一个 map 里。
		var updateDoc bson.M
		err = bson.Unmarshal(modelDoc, &updateDoc)
		if err != nil {
			return nil, utils.ErrorHandler(err, "internal error")
		}

		// remove the _id field from the update document
		delete(updateDoc, "_id")
		_, err = client.Database("school").Collection("teachers").UpdateOne(ctx, bson.M{"_id": objId}, bson.M{"$set": updateDoc})
		if err != nil {
			return nil, utils.ErrorHandler(err, fmt.Sprintln("error updatding teacher id:", teacher.Id))
		}
		updatedTeacher := mapModelTeacherToPb(*modelTeacher)

		updatedTeachers = append(updatedTeachers, updatedTeacher)
	}
	return updatedTeachers, nil
}

package repo

import (
	"context"
	"log"
	"os"

	"follower.xws.com/model"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type FollowersRepo struct {
	driver neo4j.DriverWithContext
	logger *log.Logger
}

func New(logger *log.Logger) (*FollowersRepo, error) {
	uri := os.Getenv("NEO4J_DB")
	user := os.Getenv("NEO4J_USERNAME")
	pass := os.Getenv("NEO4J_PASS")
	auth := neo4j.BasicAuth(user, pass, "")

	driver, err := neo4j.NewDriverWithContext(uri, auth)

	if err != nil {
		logger.Panic(err)
		return nil, err
	}

	return &FollowersRepo{
		driver: driver,
		logger: logger,
	}, nil
}

func (mr *FollowersRepo) CheckConnection() {
	ctx := context.Background()
	err := mr.driver.VerifyConnectivity(ctx)
	if err != nil {
		mr.logger.Panic(err)
		return
	}
	// Print Neo4J server address
	mr.logger.Printf(`Neo4J server address: %s`, mr.driver.Target().Host)
}

func (mr *FollowersRepo) CloseDriverConnection(ctx context.Context) {
	mr.driver.Close(ctx)
}

func (mr *FollowersRepo) SaveFollowing(user *model.User, userToFollow *model.User) error {
	ctx := context.Background()
	session := mr.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)
	mr.SaveUser(user)
	mr.SaveUser(userToFollow)
	_, err := session.ExecuteWrite(ctx,
		func(transaction neo4j.ManagedTransaction) (any, error) {
			result, err := transaction.Run(ctx,
				"MATCH (a:User), (b:User) WHERE a.username = $userUsername AND b.username = $followUsername CREATE (a)-[r: IS_FOLLOWING]->(b) RETURN type(r)",
				map[string]any{"userUsername": user.Username, "followUsername": userToFollow.Username})
			if err != nil {
				return nil, err
			}
			if result.Next(ctx) {
				return result.Record().Values[0], nil
			}
			return nil, result.Err()
		})
	if err != nil {
		mr.logger.Println("Error inserting following:", err)
		return err
	}
	return nil
}

func (mr *FollowersRepo) SaveUser(user *model.User) (bool, error) {
	userInDatabase, err := mr.ReadUser(user.UserId)
	if (userInDatabase == model.User{}) {
		err = mr.WriteUserToDatabase(user)
		if err != nil {
			return false, err
		}
		return true, nil
	}
	if err != nil {
		return false, err
	}
	return false, nil
}

func (mr *FollowersRepo) WriteUserToDatabase(user *model.User) error {
	ctx := context.Background()
	session := mr.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"}) //baza podataka na koju se povezujem
	defer session.Close(ctx)
	newUser, err := session.ExecuteWrite(ctx,
		func(transaction neo4j.ManagedTransaction) (any, error) {
			result, err := transaction.Run(ctx,
				"CREATE (u:User) SET u.userId = $userId, u.username = $username, u.profileImage = $profileImage RETURN u.username + ', from node ' + id(u)",
				map[string]any{"userId": user.UserId, "username": user.Username, "profileImage": user.ProfileImage})
			if err != nil {
				return nil, err
			}

			if result.Next(ctx) {
				return result.Record().Values[0], nil
			}

			return nil, result.Err()
		})
	if err != nil {
		mr.logger.Println("Error inserting Person:", err)
		return err
	}
	mr.logger.Println(newUser.(string))
	return nil
}

func (mr *FollowersRepo) ReadUser(userId string) (model.User, error) {
	ctx := context.Background()
	session := mr.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)
	user, err := session.ExecuteRead(ctx,
		func(transaction neo4j.ManagedTransaction) (any, error) {
			result, err := transaction.Run(ctx,
				"MATCH (u {userId: $userId}) RETURN u.userId, u.username, u.profileImage",
				map[string]any{"userId": userId})
			if err != nil {
				return nil, err
			}

			if result.Next(ctx) {
				return result.Record().Values, nil
			}

			return nil, result.Err()
		})
	if err != nil {
		mr.logger.Println("Error reading user:", err)
		return model.User{}, err
	}
	if user == nil {
		return model.User{}, nil
	}
	var id, username, profileImage string
	for _, value := range user.([]interface{}) {
		if val, ok := value.(string); ok {
			if id == "" {
				id = val
			} else if username == "" {
				username = val
			} else if profileImage == "" {
				profileImage = val
			}
		}
	}
	userFromDatabase := model.User{
		UserId:       id,
		Username:     username,
		ProfileImage: profileImage,
	}

	return userFromDatabase, nil
}

func (mr *FollowersRepo) GetFollowingsForUser(userId string) (model.Users, error) {
	ctx := context.Background()
	session := mr.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	userResults, err := session.ExecuteRead(ctx,
		func(transaction neo4j.ManagedTransaction) (any, error) {
			result, err := transaction.Run(ctx,
				`match (n:User)<-[r:IS_FOLLOWING]-(p:User) where p.userId = $userId return n.userId as id, n.username as username, n.profileImage as pImage`,
				map[string]any{"userId": userId})
			if err != nil {
				return nil, err
			}

			var users model.Users
			for result.Next(ctx) {
				record := result.Record()
				id, _ := record.Get("id")
				username, _ := record.Get("username")
				pImage, _ := record.Get("pImage")
				users = append(users, &model.User{
					UserId:       id.(string),
					Username:     username.(string),
					ProfileImage: pImage.(string),
				})
			}
			return users, nil
		})
	if err != nil {
		mr.logger.Println("Error querying search:", err)
		return nil, err
	}
	return userResults.(model.Users), nil
}

func (mr *FollowersRepo) GetFollowersForUser(userId string) (model.Users, error) {
	ctx := context.Background()
	session := mr.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	userResults, err := session.ExecuteRead(ctx,
		func(transaction neo4j.ManagedTransaction) (any, error) {
			result, err := transaction.Run(ctx,
				`match (n:User)<-[r:IS_FOLLOWING]-(p:User) where n.userId = $userId return p.userId as id, p.username as username, p.profileImage as pImage`,
				map[string]any{"userId": userId})
			if err != nil {
				return nil, err
			}

			var users model.Users
			for result.Next(ctx) {
				record := result.Record()
				id, _ := record.Get("id")
				username, _ := record.Get("username")
				pImage, _ := record.Get("pImage")
				users = append(users, &model.User{
					UserId:       id.(string),
					Username:     username.(string),
					ProfileImage: pImage.(string),
				})
			}
			return users, nil
		})
	if err != nil {
		mr.logger.Println("Error querying search:", err)
		return nil, err
	}
	return userResults.(model.Users), nil
}

func (mr *FollowersRepo) DeleteFollowing(userId string, userToUnfollowId string) error {
	ctx := context.Background()
	session := mr.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)
	_, err := session.ExecuteWrite(ctx,
		func(transaction neo4j.ManagedTransaction) (any, error) {
			result, err := transaction.Run(ctx,
				"MATCH (:User {userId: $userId})-[f:IS_FOLLOWING]->(:User {userId: $userToUnfollowId}) DELETE f",
				map[string]any{"userId": userId, "userToUnfollowId": userToUnfollowId})
			if err != nil {
				return nil, err
			}

			if result.Next(ctx) {
				return result.Record().Values[0], nil
			}

			return nil, result.Err()
		})
	if err != nil {
		mr.logger.Println("Error inserting Person:", err)
		return err
	}
	return nil
}

func (mr *FollowersRepo) GetRecommendationsForUser(userId string) (model.Users, error) {
	ctx := context.Background()
	session := mr.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	userResults, err := session.ExecuteRead(ctx,
		func(transaction neo4j.ManagedTransaction) (any, error) {
			result, err := transaction.Run(ctx,
				`match (n:User), (n)-[:IS_FOLLOWING]->(u:User)-[:IS_FOLLOWING]->(ru:User) where not (n)-[:IS_FOLLOWING]->(ru) and n.userId = $userId return distinct ru.userId as id, ru.username as username, ru.profileImage as pImage`,
				map[string]any{"userId": userId})
			if err != nil {
				return nil, err
			}

			var users model.Users
			for result.Next(ctx) {
				record := result.Record()
				id, _ := record.Get("id")
				username, _ := record.Get("username")
				pImage, _ := record.Get("pImage")
				users = append(users, &model.User{
					UserId:       id.(string),
					Username:     username.(string),
					ProfileImage: pImage.(string),
				})
			}
			return users, nil
		})
	if err != nil {
		mr.logger.Println("Error querying search:", err)
		return nil, err
	}
	return userResults.(model.Users), nil
}

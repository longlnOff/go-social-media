package db

import (
	"context"
	"log"
	"strconv"
	"math/rand"
	"github.com/longln/go-social-media/internal/store"
)

func Seed(store store.Storage) {
	ctx := context.Background()

	users := generateUsers(100)
	for _, user := range users {
		err := store.Users.Create(ctx, user)
		if err != nil {
			log.Println("Error creating user:", err)
			return
		}
	}

	posts := generatePosts(100, users)
	for _, post := range posts {
		err := store.Posts.Create(ctx, post)
		if err != nil {
			log.Println("Error creating post:", err)
			return
		}
	}

	comments := generateComments(500, users, posts)
	for _, comment := range comments {
		err := store.Comments.Create(ctx, comment)
		if err != nil {
			log.Println("Error creating comment:", err)
			return
		}
	}

	log.Println("Seeding complete")
}


func generateUsers(n int) []*store.User {
	users := make([]*store.User, n)
	for i := 0; i < n; i++ {
		users[i] = &store.User{
			Username: "user" + strconv.Itoa(i),
			Email:    "user" + strconv.Itoa(i) + "@example.com",
			Password: "password",
		}
	}
	return users
}

func generatePosts(n int, users []*store.User) []*store.Post {
	posts := make([]*store.Post, n)
	for i := 0; i < n; i++ {
		posts[i] = &store.Post{
			Title:   "Post " + strconv.Itoa(i),
			Content: "Content of post " + strconv.Itoa(i),
			UserID:  users[i].ID,
			User:    *users[i],
			Tags: []string{"tag1", "tag2"},
		}
	}
	return posts
}

func generateComments(n int, users []*store.User, posts []*store.Post) []*store.Comment {
	comments := make([]*store.Comment, n)
	for i := 0; i < n; i++ {
		comments[i] = &store.Comment{
			Content: "Comment " + strconv.Itoa(i),
			UserID:  users[rand.Intn(len(users))].ID,
			PostID:  posts[rand.Intn(len(posts))].ID,
		}
	}
	return comments
}
package repository

import (
	"path/filepath"
	"testing"

	"sn-backend/internal/db/sqlite"
	"sn-backend/internal/model"
)

// newTestRepo gives each test its own migrated database file.
func newTestRepo(t *testing.T) *Repository {
	t.Helper()

	path := filepath.Join(t.TempDir(), "test.db")
	if err := sqlite.Migrate(path); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	db, err := sqlite.Open(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	return New(db)
}

func addUser(t *testing.T, repo *Repository, email string) *model.User {
	t.Helper()

	user := &model.User{Email: email, Password: "x", FirstName: "Test", LastName: "User", DateOfBirth: "1995-01-01"}
	if err := repo.CreateUser(user); err != nil {
		t.Fatalf("create user: %v", err)
	}
	return user
}

func canSee(t *testing.T, repo *Repository, viewerID, postID int64) bool {
	t.Helper()

	visible, err := repo.CanViewPost(viewerID, postID)
	if err != nil {
		t.Fatalf("can view post: %v", err)
	}
	return visible
}

// A "private" post must reach the followers its author picked, and nobody else.
func TestPrivatePostReachesOnlyTheChosenFollowers(t *testing.T) {
	repo := newTestRepo(t)

	author := addUser(t, repo, "author@example.com")
	chosen := addUser(t, repo, "chosen@example.com")
	other := addUser(t, repo, "other@example.com")
	stranger := addUser(t, repo, "stranger@example.com")

	for _, follower := range []*model.User{chosen, other} {
		if _, err := repo.CreateFollowRequest(follower.ID, author.ID, model.FollowAccepted); err != nil {
			t.Fatalf("follow: %v", err)
		}
	}

	post := &model.Post{AuthorID: author.ID, Content: "secret", Privacy: model.PostSelected}
	if err := repo.CreatePost(post); err != nil {
		t.Fatalf("create post: %v", err)
	}

	// Before an audience is set the post reaches its author alone.
	if canSee(t, repo, chosen.ID, post.ID) {
		t.Error("a private post with no audience should reach nobody else")
	}

	if err := repo.SetPostAudience(post.ID, author.ID, []int64{chosen.ID, stranger.ID}); err != nil {
		t.Fatalf("set audience: %v", err)
	}

	if !canSee(t, repo, author.ID, post.ID) {
		t.Error("the author should always see their own post")
	}
	if !canSee(t, repo, chosen.ID, post.ID) {
		t.Error("the chosen follower should see the post")
	}
	if canSee(t, repo, other.ID, post.ID) {
		t.Error("a follower who was not chosen should not see the post")
	}
	// stranger was asked for but does not follow the author, so it was dropped.
	if canSee(t, repo, stranger.ID, post.ID) {
		t.Error("a user who does not follow the author cannot be added to the audience")
	}
}

func TestSetPostAudienceRefusesAnotherAuthor(t *testing.T) {
	repo := newTestRepo(t)

	author := addUser(t, repo, "author@example.com")
	intruder := addUser(t, repo, "intruder@example.com")

	post := &model.Post{AuthorID: author.ID, Content: "secret", Privacy: model.PostSelected}
	if err := repo.CreatePost(post); err != nil {
		t.Fatalf("create post: %v", err)
	}

	if err := repo.SetPostAudience(post.ID, intruder.ID, []int64{intruder.ID}); err != ErrNotOwner {
		t.Fatalf("want ErrNotOwner, got %v", err)
	}
}

// An "almost private" post is for accepted followers only.
func TestAlmostPrivatePostReachesFollowersOnly(t *testing.T) {
	repo := newTestRepo(t)

	author := addUser(t, repo, "author@example.com")
	follower := addUser(t, repo, "follower@example.com")
	pending := addUser(t, repo, "pending@example.com")

	if _, err := repo.CreateFollowRequest(follower.ID, author.ID, model.FollowAccepted); err != nil {
		t.Fatalf("follow: %v", err)
	}
	if _, err := repo.CreateFollowRequest(pending.ID, author.ID, model.FollowPending); err != nil {
		t.Fatalf("follow: %v", err)
	}

	post := &model.Post{AuthorID: author.ID, Content: "for my followers", Privacy: model.PostFollowersOnly}
	if err := repo.CreatePost(post); err != nil {
		t.Fatalf("create post: %v", err)
	}

	if !canSee(t, repo, follower.ID, post.ID) {
		t.Error("an accepted follower should see the post")
	}
	if canSee(t, repo, pending.ID, post.ID) {
		t.Error("a pending follower should not see the post yet")
	}
}

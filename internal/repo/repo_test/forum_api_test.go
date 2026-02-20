/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package repo_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/apache/answer/internal/base/handler"
	"github.com/apache/answer/internal/base/middleware"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/controller"
	"github.com/apache/answer/internal/entity"
	authrepo "github.com/apache/answer/internal/repo/auth"
	forumrepo "github.com/apache/answer/internal/repo/forum"
	"github.com/apache/answer/internal/repo/unique"
	authservice "github.com/apache/answer/internal/service/auth"
	forumservice "github.com/apache/answer/internal/service/forum"
	"github.com/gin-gonic/gin"
	pmerrors "github.com/segmentfault/pacman/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type forumAPIResp struct {
	Code   int             `json:"code"`
	Reason string          `json:"reason"`
	Data   json.RawMessage `json:"data"`
}

type createPostResp struct {
	ID string `json:"id"`
}

type createCategoryResp struct {
	ID string `json:"id"`
}

type mergeJobResp struct {
	ID                string `json:"id"`
	Status            string `json:"status"`
	AppliedRevisionID string `json:"applied_revision_id"`
}

type mergeJobDetailResp struct {
	Job struct {
		ID                string `json:"id"`
		Status            string `json:"status"`
		ReviewerID        string `json:"reviewer_id"`
		AppliedRevisionID string `json:"applied_revision_id"`
	} `json:"job"`
	PostRefs []struct {
		PostID string `json:"post_id"`
	} `json:"post_refs"`
}

type wikiRevisionIDResp struct {
	ID string `json:"id"`
}

type topicPostsListResp struct {
	List []struct {
		ID         string `json:"id"`
		MergeState string `json:"merge_state"`
		ArchivedAt string `json:"archived_at"`
	} `json:"list"`
	Total int `json:"total"`
}

func authed(userID string, roleID int, h gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("ctxUuidKey", &entity.UserCacheInfo{
			UserID:      userID,
			RoleID:      roleID,
			UserStatus:  entity.UserStatusAvailable,
			EmailStatus: entity.EmailStatusAvailable,
		})
		h(c)
	}
}

func requireAuth(userID string, roleID int, h gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("Authorization") == "" {
			handler.HandleResponse(c, pmerrors.Unauthorized(reason.UnauthorizedError), nil)
			return
		}
		authed(userID, roleID, h)(c)
	}
}

func Test_forumAPI_ReplyVoteSolvedFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx := context.TODO()

	repo := forumrepo.NewForumRepo(testDataSource, unique.NewUniqueIDRepo(testDataSource))
	service := forumservice.NewForumService(repo, nil)
	fc := controller.NewForumController(service)

	r := gin.New()
	r.POST("/api/v1/topics/:id/posts", authed("1", 1, fc.CreateTopicPost))
	r.POST("/api/v1/topics/:id/votes", authed("1", 1, fc.VoteTopic))
	r.POST("/api/v1/posts/:id/votes", authed("1", 1, fc.VotePost))
	r.POST("/api/v1/topics/:id/solution", authed("1", 1, fc.SetTopicSolution))

	_, topic := createTopicFixture(t, repo)

	postID := createTopicPostByAPI(t, r, topic.ID, "reply from api integration test")
	t.Cleanup(func() {
		_, _ = testDataSource.DB.Context(ctx).ID(postID).Delete(&entity.Post{})
		_, _ = testDataSource.DB.Context(ctx).Where("post_id = ?", postID).Delete(&entity.PostVote{})
	})

	postVoteByAPI(t, r, postID, 1)
	topicVoteByAPI(t, r, topic.ID, 1)
	setSolvedByAPI(t, r, topic.ID, postID)

	t.Cleanup(func() {
		_, _ = testDataSource.DB.Context(ctx).Where("topic_id = ?", topic.ID).Delete(&entity.TopicVote{})
		_, _ = testDataSource.DB.Context(ctx).Where("topic_id = ?", topic.ID).Delete(&entity.TopicSolution{})
	})

	topicAfter, exist, err := repo.GetTopic(ctx, topic.ID)
	require.NoError(t, err)
	require.True(t, exist)
	assert.Equal(t, 1, topicAfter.PostCount)
	assert.Equal(t, 1, topicAfter.VoteCount)
	assert.Equal(t, postID, topicAfter.SolvedPostID)

	postAfter, exist, err := repo.GetPost(ctx, postID)
	require.NoError(t, err)
	require.True(t, exist)
	assert.Equal(t, 1, postAfter.VoteCount)
}

func Test_forumAPI_Unauthorized_ForVoteAndSolved(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx := context.TODO()

	repo := forumrepo.NewForumRepo(testDataSource, unique.NewUniqueIDRepo(testDataSource))
	service := forumservice.NewForumService(repo, nil)
	fc := controller.NewForumController(service)

	r := gin.New()
	r.POST("/api/v1/topics/:id/votes", requireAuth("1", 1, fc.VoteTopic))
	r.POST("/api/v1/posts/:id/votes", requireAuth("1", 1, fc.VotePost))
	r.POST("/api/v1/topics/:id/solution", requireAuth("1", 1, fc.SetTopicSolution))

	_, topic := createTopicFixture(t, repo)
	post := &entity.Post{
		TopicID:    topic.ID,
		UserID:     "1",
		Original:   "reply for unauthorized test",
		Parsed:     "reply for unauthorized test",
		MergeState: entity.PostMergeStateActive,
		Status:     1,
	}
	require.NoError(t, repo.AddPost(ctx, post))
	t.Cleanup(func() {
		_, _ = testDataSource.DB.Context(ctx).ID(post.ID).Delete(&entity.Post{})
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/posts/"+post.ID+"/votes", bytes.NewReader([]byte(`{"value":1}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusUnauthorized, w.Code, w.Body.String())

	resp := &forumAPIResp{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), resp))
	assert.Equal(t, http.StatusUnauthorized, resp.Code)
	assert.Equal(t, reason.UnauthorizedError, resp.Reason)

	var postVoteCount int64
	postVoteCount, err := testDataSource.DB.Context(ctx).Where("post_id = ?", post.ID).Count(&entity.PostVote{})
	require.NoError(t, err)
	assert.Equal(t, int64(0), postVoteCount)

	req = httptest.NewRequest(http.MethodPost, "/api/v1/topics/"+topic.ID+"/solution", bytes.NewReader([]byte(fmt.Sprintf(`{"post_id":%q}`, post.ID))))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusUnauthorized, w.Code, w.Body.String())

	var solutionCount int64
	solutionCount, err = testDataSource.DB.Context(ctx).Where("topic_id = ?", topic.ID).Count(&entity.TopicSolution{})
	require.NoError(t, err)
	assert.Equal(t, int64(0), solutionCount)
}

func Test_forumAPI_Forbidden_CreateCategory_WhenUserNotModeratorAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := forumrepo.NewForumRepo(testDataSource, unique.NewUniqueIDRepo(testDataSource))
	service := forumservice.NewForumService(repo, nil)
	fc := controller.NewForumController(service)

	r := gin.New()
	r.POST("/api/v1/categories", requireAuth("1", 1, fc.CreateCategory))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/categories", bytes.NewReader([]byte(`{
		"slug":"forbidden-category-test",
		"name":"Forbidden Category Test",
		"description":"forbidden test"
	}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusForbidden, w.Code, w.Body.String())

	resp := &forumAPIResp{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), resp))
	assert.Equal(t, http.StatusForbidden, resp.Code)
	assert.Equal(t, reason.ForbiddenError, resp.Reason)
}

func Test_forumAPI_WithAuthMiddleware_RequiresTokenAndAllowsValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx := context.TODO()

	repo := forumrepo.NewForumRepo(testDataSource, unique.NewUniqueIDRepo(testDataSource))
	service := forumservice.NewForumService(repo, nil)
	fc := controller.NewForumController(service)

	authRepo := authrepo.NewAuthRepo(testDataSource)
	authSvc := authservice.NewAuthService(authRepo, nil)
	authMW := middleware.NewAuthUserMiddleware(authSvc, nil)

	r := gin.New()
	secured := r.Group("/api/v1")
	secured.Use(authMW.MustAuthAndAccountAvailable())
	secured.POST("/topics/:id/votes", fc.VoteTopic)

	_, topic := createTopicFixture(t, repo)

	// Missing token should be blocked by middleware.
	req := httptest.NewRequest(http.MethodPost, "/api/v1/topics/"+topic.ID+"/votes", bytes.NewReader([]byte(`{"value":1}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusUnauthorized, w.Code, w.Body.String())

	resp := &forumAPIResp{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), resp))
	assert.Equal(t, reason.UnauthorizedError, resp.Reason)

	// Real token in cache should pass middleware and write DB.
	token := issueAccessTokenForTest(t, authSvc, "1", 1, entity.UserStatusAvailable, entity.EmailStatusAvailable)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/topics/"+topic.ID+"/votes", bytes.NewReader([]byte(`{"value":1}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	var topicVotes []*entity.TopicVote
	require.NoError(t, testDataSource.DB.Context(ctx).Where("topic_id = ?", topic.ID).Find(&topicVotes))
	require.Len(t, topicVotes, 1)
	assert.Equal(t, "1", topicVotes[0].UserID)
	assert.Equal(t, 1, topicVotes[0].Value)
	t.Cleanup(func() {
		_, _ = testDataSource.DB.Context(ctx).ID(topicVotes[0].ID).Delete(&entity.TopicVote{})
	})
}

func Test_forumAPI_WithAuthMiddleware_CreateCategoryRoleCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx := context.TODO()

	repo := forumrepo.NewForumRepo(testDataSource, unique.NewUniqueIDRepo(testDataSource))
	service := forumservice.NewForumService(repo, nil)
	fc := controller.NewForumController(service)

	authRepo := authrepo.NewAuthRepo(testDataSource)
	authSvc := authservice.NewAuthService(authRepo, nil)
	authMW := middleware.NewAuthUserMiddleware(authSvc, nil)

	r := gin.New()
	secured := r.Group("/api/v1")
	secured.Use(authMW.MustAuthAndAccountAvailable())
	secured.POST("/categories", fc.CreateCategory)

	normalToken := issueAccessTokenForTest(t, authSvc, "1", 1, entity.UserStatusAvailable, entity.EmailStatusAvailable)
	adminToken := issueAccessTokenForTest(t, authSvc, "1", 2, entity.UserStatusAvailable, entity.EmailStatusAvailable)
	payload := []byte(fmt.Sprintf(`{
		"slug":"auth-mw-category-%d",
		"name":"Auth MW Category",
		"description":"middleware role test"
	}`, time.Now().UnixNano()))

	// Role=User should be forbidden by controller role check.
	req := httptest.NewRequest(http.MethodPost, "/api/v1/categories", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+normalToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusForbidden, w.Code, w.Body.String())
	resp := &forumAPIResp{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), resp))
	assert.Equal(t, reason.ForbiddenError, resp.Reason)

	// Role=Admin should pass and create category.
	req = httptest.NewRequest(http.MethodPost, "/api/v1/categories", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), resp))
	assert.Contains(t, resp.Reason, "success")
	created := &createCategoryResp{}
	require.NoError(t, json.Unmarshal(resp.Data, created))
	require.NotEmpty(t, created.ID)
	t.Cleanup(func() {
		_, _ = testDataSource.DB.Context(ctx).ID(created.ID).Delete(&entity.Category{})
	})
}

func Test_forumAPI_WithAuthMiddleware_PostVoteAndSolved(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx := context.TODO()

	repo := forumrepo.NewForumRepo(testDataSource, unique.NewUniqueIDRepo(testDataSource))
	service := forumservice.NewForumService(repo, nil)
	fc := controller.NewForumController(service)

	authRepo := authrepo.NewAuthRepo(testDataSource)
	authSvc := authservice.NewAuthService(authRepo, nil)
	authMW := middleware.NewAuthUserMiddleware(authSvc, nil)

	r := gin.New()
	secured := r.Group("/api/v1")
	secured.Use(authMW.MustAuthAndAccountAvailable())
	secured.POST("/posts/:id/votes", fc.VotePost)
	secured.POST("/topics/:id/solution", fc.SetTopicSolution)

	_, topic := createTopicFixture(t, repo)
	post := &entity.Post{
		TopicID:    topic.ID,
		UserID:     "1",
		Original:   "reply for auth middleware post vote and solved",
		Parsed:     "reply for auth middleware post vote and solved",
		MergeState: entity.PostMergeStateActive,
		Status:     1,
	}
	require.NoError(t, repo.AddPost(ctx, post))
	t.Cleanup(func() {
		_, _ = testDataSource.DB.Context(ctx).ID(post.ID).Delete(&entity.Post{})
		_, _ = testDataSource.DB.Context(ctx).Where("post_id = ?", post.ID).Delete(&entity.PostVote{})
		_, _ = testDataSource.DB.Context(ctx).Where("topic_id = ?", topic.ID).Delete(&entity.TopicSolution{})
	})

	// Missing token should be blocked by middleware.
	req := httptest.NewRequest(http.MethodPost, "/api/v1/posts/"+post.ID+"/votes", bytes.NewReader([]byte(`{"value":1}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusUnauthorized, w.Code, w.Body.String())

	// Real token in cache should pass middleware and write DB for both APIs.
	token := issueAccessTokenForTest(t, authSvc, "1", 1, entity.UserStatusAvailable, entity.EmailStatusAvailable)

	req = httptest.NewRequest(http.MethodPost, "/api/v1/posts/"+post.ID+"/votes", bytes.NewReader([]byte(`{"value":1}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	req = httptest.NewRequest(http.MethodPost, "/api/v1/topics/"+topic.ID+"/solution", bytes.NewReader([]byte(fmt.Sprintf(`{"post_id":%q}`, post.ID))))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	votes := make([]*entity.PostVote, 0)
	require.NoError(t, testDataSource.DB.Context(ctx).Where("post_id = ?", post.ID).Find(&votes))
	require.Len(t, votes, 1)
	assert.Equal(t, "1", votes[0].UserID)
	assert.Equal(t, 1, votes[0].Value)

	solution := &entity.TopicSolution{}
	exist, err := testDataSource.DB.Context(ctx).Where("topic_id = ?", topic.ID).Get(solution)
	require.NoError(t, err)
	require.True(t, exist)
	assert.Equal(t, post.ID, solution.PostID)
	assert.Equal(t, "1", solution.SetByUserID)
}

func Test_forumAPI_MergeWorkflow_EndToEndAndIdempotent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx := context.TODO()

	repo := forumrepo.NewForumRepo(testDataSource, unique.NewUniqueIDRepo(testDataSource))
	service := forumservice.NewForumService(repo, nil)
	fc := controller.NewForumController(service)

	r := gin.New()
	r.POST("/api/v1/topics/:id/merge-jobs", authed("1", 1, fc.CreateMergeJob))
	r.GET("/api/v1/topics/:id/merge-jobs/:jobId", fc.GetMergeJob)
	r.POST("/api/v1/topics/:id/merge-jobs/:jobId/apply", authed("1", 2, fc.ApplyMergeJob))
	r.GET("/api/v1/topics/:id/wiki", fc.GetTopicWiki)
	r.GET("/api/v1/topics/:id/wiki/revisions", fc.ListTopicWikiRevisions)
	r.GET("/api/v1/topics/:id/contributors", fc.ListTopicContributors)
	r.GET("/api/v1/topics/:id/posts", fc.ListTopicPosts)

	_, topic := createTopicFixture(t, repo)
	postOne := &entity.Post{
		TopicID:    topic.ID,
		UserID:     "1",
		Original:   "merge candidate one",
		Parsed:     "merge candidate one",
		MergeState: entity.PostMergeStateActive,
		Status:     1,
	}
	postTwo := &entity.Post{
		TopicID:    topic.ID,
		UserID:     "2",
		Original:   "merge candidate two",
		Parsed:     "merge candidate two",
		MergeState: entity.PostMergeStateActive,
		Status:     1,
	}
	require.NoError(t, repo.AddPost(ctx, postOne))
	require.NoError(t, repo.AddPost(ctx, postTwo))
	t.Cleanup(func() {
		_, _ = testDataSource.DB.Context(ctx).Where("topic_id = ?", topic.ID).Delete(&entity.ContributionCredit{})
		_, _ = testDataSource.DB.Context(ctx).In("post_id", []string{postOne.ID, postTwo.ID}).Delete(&entity.MergeJobPostRef{})
		_, _ = testDataSource.DB.Context(ctx).Where("topic_id = ?", topic.ID).Delete(&entity.MergeJob{})
		_, _ = testDataSource.DB.Context(ctx).Where("topic_id = ?", topic.ID).Delete(&entity.WikiRevision{})
		_, _ = testDataSource.DB.Context(ctx).In("id", []string{postOne.ID, postTwo.ID}).Delete(&entity.Post{})
	})

	createJobPayload := []byte(fmt.Sprintf(`{"post_ids":[%q,%q],"summary":"merge two replies"}`, postOne.ID, postTwo.ID))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/topics/"+topic.ID+"/merge-jobs", bytes.NewReader(createJobPayload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	createJobResp := mustDecodeForumData[mergeJobResp](t, w.Body.Bytes())
	require.NotEmpty(t, createJobResp.ID)
	assert.Equal(t, entity.MergeJobStatusPending, createJobResp.Status)

	req = httptest.NewRequest(http.MethodGet, "/api/v1/topics/"+topic.ID+"/merge-jobs/"+createJobResp.ID, nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	jobDetailResp := mustDecodeForumData[mergeJobDetailResp](t, w.Body.Bytes())
	assert.Equal(t, createJobResp.ID, jobDetailResp.Job.ID)
	assert.Equal(t, entity.MergeJobStatusPending, jobDetailResp.Job.Status)
	require.Len(t, jobDetailResp.PostRefs, 2)

	applyPayload := []byte(`{
		"title":"Merged canonical answer",
		"summary":"merge reviewed",
		"document":"This revision merges two replies into canonical docs.",
		"contribution_weight":2
	}`)
	req = httptest.NewRequest(
		http.MethodPost,
		"/api/v1/topics/"+topic.ID+"/merge-jobs/"+createJobResp.ID+"/apply",
		bytes.NewReader(applyPayload),
	)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	appliedRevision := mustDecodeForumData[wikiRevisionIDResp](t, w.Body.Bytes())
	require.NotEmpty(t, appliedRevision.ID)

	topicAfter, exist, err := repo.GetTopic(ctx, topic.ID)
	require.NoError(t, err)
	require.True(t, exist)
	assert.Equal(t, appliedRevision.ID, topicAfter.CurrentWikiRevisionID)

	req = httptest.NewRequest(http.MethodGet, "/api/v1/topics/"+topic.ID+"/wiki", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	currentWikiResp := mustDecodeForumData[wikiRevisionIDResp](t, w.Body.Bytes())
	assert.Equal(t, appliedRevision.ID, currentWikiResp.ID)

	req = httptest.NewRequest(http.MethodGet, "/api/v1/topics/"+topic.ID+"/wiki/revisions", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	revisions := mustDecodeForumData[[]wikiRevisionIDResp](t, w.Body.Bytes())
	require.NotEmpty(t, revisions)
	foundRevision := false
	for _, revision := range revisions {
		if revision.ID == appliedRevision.ID {
			foundRevision = true
			break
		}
	}
	assert.True(t, foundRevision)

	req = httptest.NewRequest(http.MethodGet, "/api/v1/topics/"+topic.ID+"/posts", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	topicPostsResp := mustDecodeForumData[topicPostsListResp](t, w.Body.Bytes())
	require.GreaterOrEqual(t, topicPostsResp.Total, 2)
	archivedStateByPost := make(map[string]bool, len(topicPostsResp.List))
	for _, item := range topicPostsResp.List {
		archivedStateByPost[item.ID] = item.MergeState == entity.PostMergeStateArchived && item.ArchivedAt != ""
	}
	assert.True(t, archivedStateByPost[postOne.ID])
	assert.True(t, archivedStateByPost[postTwo.ID])

	req = httptest.NewRequest(http.MethodGet, "/api/v1/topics/"+topic.ID+"/contributors", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	contributors := mustDecodeForumData[[]map[string]any](t, w.Body.Bytes())
	contributorWeight := map[string]int{}
	for _, item := range contributors {
		userID := toStringSafe(
			item["user_id"],
			toStringSafe(item["userID"], toStringSafe(item["UserID"], "")),
		)
		weight := toIntSafe(item["weight"], toIntSafe(item["Weight"], 0))
		contributorWeight[userID] = weight
	}
	assert.Equal(t, 2, contributorWeight["1"])
	assert.Equal(t, 2, contributorWeight["2"])

	// Apply the same merge job again; should be idempotent and return the same revision.
	req = httptest.NewRequest(
		http.MethodPost,
		"/api/v1/topics/"+topic.ID+"/merge-jobs/"+createJobResp.ID+"/apply",
		bytes.NewReader(applyPayload),
	)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	appliedRevisionAgain := mustDecodeForumData[wikiRevisionIDResp](t, w.Body.Bytes())
	assert.Equal(t, appliedRevision.ID, appliedRevisionAgain.ID)

	mergeJobAfter, refsAfter, exist, err := repo.GetMergeJob(ctx, createJobResp.ID)
	require.NoError(t, err)
	require.True(t, exist)
	assert.Equal(t, entity.MergeJobStatusApplied, mergeJobAfter.Status)
	assert.Equal(t, appliedRevision.ID, mergeJobAfter.AppliedRevisionID)
	require.Len(t, refsAfter, 2)

	var creditCount int64
	creditCount, err = testDataSource.DB.Context(ctx).Where("revision_id = ?", appliedRevision.ID).Count(&entity.ContributionCredit{})
	require.NoError(t, err)
	assert.Equal(t, int64(2), creditCount)
}

func issueAccessTokenForTest(
	t *testing.T,
	authSvc *authservice.AuthService,
	userID string,
	roleID int,
	userStatus int,
	emailStatus int,
) string {
	t.Helper()
	accessToken, _, err := authSvc.SetUserCacheInfo(context.TODO(), &entity.UserCacheInfo{
		UserID:      userID,
		UserStatus:  userStatus,
		EmailStatus: emailStatus,
		RoleID:      roleID,
	})
	require.NoError(t, err)
	require.NotEmpty(t, accessToken)
	return accessToken
}

func createTopicPostByAPI(t *testing.T, r *gin.Engine, topicID, text string) string {
	t.Helper()
	payload := []byte(fmt.Sprintf(`{"original_text":%q}`, text))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/topics/"+topicID+"/posts", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	resp := &forumAPIResp{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), resp))
	require.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Reason, "success")

	post := &createPostResp{}
	require.NoError(t, json.Unmarshal(resp.Data, post))
	require.NotEmpty(t, post.ID)
	return post.ID
}

func postVoteByAPI(t *testing.T, r *gin.Engine, postID string, value int) {
	t.Helper()
	payload := []byte(fmt.Sprintf(`{"value":%d}`, value))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/posts/"+postID+"/votes", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
}

func topicVoteByAPI(t *testing.T, r *gin.Engine, topicID string, value int) {
	t.Helper()
	payload := []byte(fmt.Sprintf(`{"value":%d}`, value))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/topics/"+topicID+"/votes", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
}

func setSolvedByAPI(t *testing.T, r *gin.Engine, topicID, postID string) {
	t.Helper()
	payload := []byte(fmt.Sprintf(`{"post_id":%q}`, postID))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/topics/"+topicID+"/solution", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
}

func mustDecodeForumData[T any](t *testing.T, body []byte) T {
	t.Helper()
	resp := &forumAPIResp{}
	require.NoError(t, json.Unmarshal(body, resp))
	require.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Reason, "success")

	var data T
	require.NoError(t, json.Unmarshal(resp.Data, &data))
	return data
}

func toStringSafe(value any, fallback string) string {
	switch v := value.(type) {
	case string:
		return v
	case float64:
		return fmt.Sprintf("%.0f", v)
	default:
		return fallback
	}
}

func toIntSafe(value any, fallback int) int {
	switch v := value.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	default:
		return fallback
	}
}

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
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/repo/forum"
	"github.com/apache/answer/internal/repo/unique"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newForumRepoForTest() *forum.ForumRepo {
	return forum.NewForumRepo(testDataSource, unique.NewUniqueIDRepo(testDataSource))
}

func createTopicFixture(t *testing.T, repo *forum.ForumRepo) (*entity.Category, *entity.Topic) {
	t.Helper()
	ctx := context.TODO()
	suffix := time.Now().UnixNano()

	category := &entity.Category{
		CreatorID:   "1",
		Slug:        fmt.Sprintf("forum-test-%d", suffix),
		Name:        fmt.Sprintf("Forum Test %d", suffix),
		Description: "forum regression test",
		Status:      1,
	}
	require.NoError(t, repo.AddCategory(ctx, category))

	topic := &entity.Topic{
		CategoryID:    category.ID,
		UserID:        "1",
		Title:         fmt.Sprintf("Topic %d", suffix),
		TopicKind:     entity.TopicKindDiscussion,
		IsWikiEnabled: true,
		Status:        entity.TopicStatusAvailable,
	}
	require.NoError(t, repo.AddTopic(ctx, topic))

	t.Cleanup(func() {
		_, _ = testDataSource.DB.Context(ctx).ID(topic.ID).Delete(&entity.Topic{})
		_, _ = testDataSource.DB.Context(ctx).ID(category.ID).Delete(&entity.Category{})
	})
	return category, topic
}

func Test_forumRepo_ReplyVoteSolvedWorkflow(t *testing.T) {
	ctx := context.TODO()
	repo := newForumRepoForTest()
	_, topic := createTopicFixture(t, repo)

	reply := &entity.Post{
		TopicID:    topic.ID,
		UserID:     "1",
		Original:   "first reply",
		Parsed:     "first reply",
		MergeState: entity.PostMergeStateActive,
		Status:     1,
	}
	require.NoError(t, repo.AddPost(ctx, reply))
	t.Cleanup(func() {
		_, _ = testDataSource.DB.Context(ctx).ID(reply.ID).Delete(&entity.Post{})
	})

	// topic vote, post vote and mark solved are the critical workflow mutations.
	require.NoError(t, repo.UpsertTopicVote(ctx, topic.ID, "1", 1))
	require.NoError(t, repo.UpsertPostVote(ctx, reply.ID, "1", 1))
	require.NoError(t, repo.UpsertTopicSolution(ctx, topic.ID, reply.ID, "1"))

	topicAfter, exist, err := repo.GetTopic(ctx, topic.ID)
	require.NoError(t, err)
	require.True(t, exist)
	assert.Equal(t, 1, topicAfter.PostCount)
	assert.Equal(t, 1, topicAfter.VoteCount)
	assert.Equal(t, reply.ID, topicAfter.SolvedPostID)

	postAfter, exist, err := repo.GetPost(ctx, reply.ID)
	require.NoError(t, err)
	require.True(t, exist)
	assert.Equal(t, 1, postAfter.VoteCount)

	var topicVotes []*entity.TopicVote
	require.NoError(t, testDataSource.DB.Context(ctx).Where("topic_id = ?", topic.ID).Find(&topicVotes))
	require.Len(t, topicVotes, 1)
	assert.Equal(t, 1, topicVotes[0].Value)
	t.Cleanup(func() {
		_, _ = testDataSource.DB.Context(ctx).ID(topicVotes[0].ID).Delete(&entity.TopicVote{})
	})

	var postVotes []*entity.PostVote
	require.NoError(t, testDataSource.DB.Context(ctx).Where("post_id = ?", reply.ID).Find(&postVotes))
	require.Len(t, postVotes, 1)
	assert.Equal(t, 1, postVotes[0].Value)
	t.Cleanup(func() {
		_, _ = testDataSource.DB.Context(ctx).ID(postVotes[0].ID).Delete(&entity.PostVote{})
	})

	solution := &entity.TopicSolution{}
	exist, err = testDataSource.DB.Context(ctx).Where("topic_id = ?", topic.ID).Get(solution)
	require.NoError(t, err)
	require.True(t, exist)
	assert.Equal(t, reply.ID, solution.PostID)
	t.Cleanup(func() {
		_, _ = testDataSource.DB.Context(ctx).ID(solution.ID).Delete(&entity.TopicSolution{})
	})
}

func Test_forumRepo_AddPostConcurrent_NoSQLiteBusy(t *testing.T) {
	ctx := context.TODO()
	repo := newForumRepoForTest()
	_, topic := createTopicFixture(t, repo)

	const workers = 20
	errCh := make(chan error, workers)
	postIDs := make(chan string, workers)
	wg := sync.WaitGroup{}
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func(i int) {
			defer wg.Done()
			post := &entity.Post{
				TopicID:    topic.ID,
				UserID:     "1",
				Original:   fmt.Sprintf("reply-%d", i),
				Parsed:     fmt.Sprintf("reply-%d", i),
				MergeState: entity.PostMergeStateActive,
				Status:     1,
			}
			if err := repo.AddPost(ctx, post); err != nil {
				errCh <- err
				return
			}
			postIDs <- post.ID
		}(i)
	}
	wg.Wait()
	close(errCh)
	close(postIDs)

	for err := range errCh {
		require.NoError(t, err)
	}

	createdIDs := make([]string, 0, workers)
	for id := range postIDs {
		createdIDs = append(createdIDs, id)
	}
	require.Len(t, createdIDs, workers)

	topicAfter, exist, err := repo.GetTopic(ctx, topic.ID)
	require.NoError(t, err)
	require.True(t, exist)
	assert.Equal(t, workers, topicAfter.PostCount)

	t.Cleanup(func() {
		_, _ = testDataSource.DB.Context(ctx).In("id", createdIDs).Delete(&entity.Post{})
	})
}

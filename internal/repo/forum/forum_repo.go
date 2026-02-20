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

package forum

import (
	"context"
	"time"

	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/service/unique"
	"github.com/apache/answer/pkg/uid"
	"github.com/segmentfault/pacman/errors"
	"xorm.io/xorm"
)

type ContributorStat struct {
	UserID string `xorm:"user_id"`
	Weight int    `xorm:"weight"`
}

type TopicPostView struct {
	ID                string     `json:"id" xorm:"id"`
	TopicID           string     `json:"topic_id" xorm:"topic_id"`
	UserID            string     `json:"user_id" xorm:"user_id"`
	OriginalText      string     `json:"original_text" xorm:"original_text"`
	ParsedText        string     `json:"parsed_text" xorm:"parsed_text"`
	MergeState        string     `json:"merge_state" xorm:"merge_state"`
	ArchivedAt        *time.Time `json:"archived_at" xorm:"archived_at"`
	VoteCount         int        `json:"vote_count" xorm:"vote_count"`
	Status            int        `json:"status" xorm:"status"`
	CreatedAt         time.Time  `json:"created_at" xorm:"created_at"`
	AuthorUsername    string     `json:"author_username" xorm:"author_username"`
	AuthorDisplayName string     `json:"author_display_name" xorm:"author_display_name"`
}

type ForumRepo struct {
	data         *data.Data
	uniqueIDRepo unique.UniqueIDRepo
}

func NewForumRepo(data *data.Data, uniqueIDRepo unique.UniqueIDRepo) *ForumRepo {
	return &ForumRepo{
		data:         data,
		uniqueIDRepo: uniqueIDRepo,
	}
}

func (r *ForumRepo) genID(ctx context.Context, table string) (string, error) {
	id, err := r.uniqueIDRepo.GenUniqueIDStr(ctx, table)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (r *ForumRepo) AddCategory(ctx context.Context, category *entity.Category) error {
	id, err := r.genID(ctx, category.TableName())
	if err != nil {
		return err
	}
	category.ID = id
	if _, err = r.data.DB.Context(ctx).Insert(category); err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return nil
}

func (r *ForumRepo) GetCategory(ctx context.Context, categoryID string) (*entity.Category, bool, error) {
	category := &entity.Category{ID: uid.DeShortID(categoryID)}
	exist, err := r.data.DB.Context(ctx).Get(category)
	if err != nil {
		return nil, false, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return category, exist, nil
}

func (r *ForumRepo) ListCategories(ctx context.Context, page, pageSize int) ([]*entity.Category, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	categories := make([]*entity.Category, 0)
	total, err := r.data.DB.Context(ctx).Where("status = ?", 1).Count(&entity.Category{})
	if err != nil {
		return nil, 0, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	if err := r.data.DB.Context(ctx).
		Where("status = ?", 1).
		Desc("created_at").
		Limit(pageSize, (page-1)*pageSize).
		Find(&categories); err != nil {
		return nil, 0, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return categories, total, nil
}

func (r *ForumRepo) AddTopic(ctx context.Context, topic *entity.Topic) error {
	id, err := r.genID(ctx, topic.TableName())
	if err != nil {
		return err
	}
	topic.ID = id
	if _, err = r.data.DB.Context(ctx).Insert(topic); err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return nil
}

func (r *ForumRepo) GetTopic(ctx context.Context, topicID string) (*entity.Topic, bool, error) {
	topic := &entity.Topic{ID: uid.DeShortID(topicID)}
	exist, err := r.data.DB.Context(ctx).Get(topic)
	if err != nil {
		return nil, false, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return topic, exist, nil
}

func (r *ForumRepo) UpdateTopic(ctx context.Context, topic *entity.Topic, cols ...string) error {
	topic.ID = uid.DeShortID(topic.ID)
	if len(cols) == 0 {
		_, err := r.data.DB.Context(ctx).ID(topic.ID).Update(topic)
		if err != nil {
			return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
		}
		return nil
	}
	_, err := r.data.DB.Context(ctx).ID(topic.ID).Cols(cols...).Update(topic)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return nil
}

func (r *ForumRepo) ListTopicsByCategory(ctx context.Context, categoryID string, page, pageSize int) ([]*entity.Topic, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	topics := make([]*entity.Topic, 0)
	total, err := r.data.DB.Context(ctx).Where("category_id = ?", uid.DeShortID(categoryID)).Count(&entity.Topic{})
	if err != nil {
		return nil, 0, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	if err := r.data.DB.Context(ctx).
		Where("category_id = ?", uid.DeShortID(categoryID)).
		Desc("created_at").
		Limit(pageSize, (page-1)*pageSize).
		Find(&topics); err != nil {
		return nil, 0, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return topics, total, nil
}

func (r *ForumRepo) AddPost(ctx context.Context, post *entity.Post) error {
	postID, err := r.genID(ctx, post.TableName())
	if err != nil {
		return err
	}
	post.ID = postID

	_, err = r.data.DB.Transaction(func(session *xorm.Session) (any, error) {
		session = session.Context(ctx)
		if _, err := session.Insert(post); err != nil {
			return nil, err
		}
		if _, err := session.ID(uid.DeShortID(post.TopicID)).Incr("post_count", 1).Cols("last_post_id").Update(
			&entity.Topic{LastPostID: post.ID},
		); err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return nil
}

func (r *ForumRepo) GetPost(ctx context.Context, postID string) (*entity.Post, bool, error) {
	post := &entity.Post{ID: uid.DeShortID(postID)}
	exist, err := r.data.DB.Context(ctx).Get(post)
	if err != nil {
		return nil, false, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return post, exist, nil
}

func (r *ForumRepo) GetPostsByIDs(ctx context.Context, topicID string, postIDs []string) ([]*entity.Post, error) {
	ids := make([]string, 0, len(postIDs))
	for _, id := range postIDs {
		ids = append(ids, uid.DeShortID(id))
	}
	posts := make([]*entity.Post, 0)
	if len(ids) == 0 {
		return posts, nil
	}
	if err := r.data.DB.Context(ctx).Where("topic_id = ?", uid.DeShortID(topicID)).In("id", ids).Find(&posts); err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return posts, nil
}

func (r *ForumRepo) ListTopicPosts(ctx context.Context, topicID string, page, pageSize int) ([]*TopicPostView, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 30
	}
	if pageSize > 100 {
		pageSize = 100
	}

	topicID = uid.DeShortID(topicID)
	total, err := r.data.DB.Context(ctx).Where("topic_id = ?", topicID).Count(&entity.Post{})
	if err != nil {
		return nil, 0, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}

	posts := make([]*TopicPostView, 0)
	query := `
SELECT
	p.id,
	p.topic_id,
	p.user_id,
	p.original_text,
	p.parsed_text,
	p.merge_state,
	p.archived_at,
	p.vote_count,
	p.status,
	p.created_at,
	u.username AS author_username,
	u.display_name AS author_display_name
FROM posts AS p
LEFT JOIN user AS u ON u.id = p.user_id
WHERE p.topic_id = ?
ORDER BY p.created_at ASC
LIMIT ? OFFSET ?`
	if err := r.data.DB.Context(ctx).SQL(query, topicID, pageSize, (page-1)*pageSize).Find(&posts); err != nil {
		return nil, 0, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return posts, total, nil
}

func (r *ForumRepo) AddWikiRevision(ctx context.Context, revision *entity.WikiRevision) error {
	id, err := r.genID(ctx, revision.TableName())
	if err != nil {
		return err
	}
	revision.ID = id
	if _, err = r.data.DB.Context(ctx).Insert(revision); err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return nil
}

func (r *ForumRepo) ListWikiRevisions(ctx context.Context, topicID string) ([]*entity.WikiRevision, error) {
	revisions := make([]*entity.WikiRevision, 0)
	if err := r.data.DB.Context(ctx).Where("topic_id = ?", uid.DeShortID(topicID)).Desc("created_at").Find(&revisions); err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return revisions, nil
}

func (r *ForumRepo) GetWikiRevision(ctx context.Context, revisionID string) (*entity.WikiRevision, bool, error) {
	revision := &entity.WikiRevision{ID: uid.DeShortID(revisionID)}
	exist, err := r.data.DB.Context(ctx).Get(revision)
	if err != nil {
		return nil, false, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return revision, exist, nil
}

func (r *ForumRepo) AddMergeJob(ctx context.Context, job *entity.MergeJob, postIDs []string) error {
	id, err := r.genID(ctx, job.TableName())
	if err != nil {
		return err
	}
	job.ID = id
	refIDs := make([]string, 0, len(postIDs))
	for range postIDs {
		refID, err := r.genID(ctx, entity.MergeJobPostRef{}.TableName())
		if err != nil {
			return err
		}
		refIDs = append(refIDs, refID)
	}

	_, err = r.data.DB.Transaction(func(session *xorm.Session) (any, error) {
		session = session.Context(ctx)
		if _, err := session.Insert(job); err != nil {
			return nil, err
		}

		for i, rawPostID := range postIDs {
			ref := &entity.MergeJobPostRef{
				ID:         refIDs[i],
				MergeJobID: job.ID,
				PostID:     uid.DeShortID(rawPostID),
			}
			if _, err := session.Insert(ref); err != nil {
				return nil, err
			}
		}
		return nil, nil
	})
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return nil
}

func (r *ForumRepo) GetMergeJob(ctx context.Context, jobID string) (*entity.MergeJob, []*entity.MergeJobPostRef, bool, error) {
	job := &entity.MergeJob{ID: uid.DeShortID(jobID)}
	exist, err := r.data.DB.Context(ctx).Get(job)
	if err != nil {
		return nil, nil, false, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	if !exist {
		return nil, nil, false, nil
	}
	refs := make([]*entity.MergeJobPostRef, 0)
	if err := r.data.DB.Context(ctx).Where("merge_job_id = ?", job.ID).Find(&refs); err != nil {
		return nil, nil, false, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return job, refs, true, nil
}

func (r *ForumRepo) ListMergeJobsByTopic(
	ctx context.Context,
	topicID string,
	status string,
	page, pageSize int,
) ([]*entity.MergeJob, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	topicID = uid.DeShortID(topicID)
	buildBaseQuery := func() *xorm.Session {
		session := r.data.DB.Context(ctx).Where("topic_id = ?", topicID)
		if status != "" {
			session = session.And("status = ?", status)
		}
		return session
	}

	total, err := buildBaseQuery().Count(&entity.MergeJob{})
	if err != nil {
		return nil, 0, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	jobs := make([]*entity.MergeJob, 0)
	if err := buildBaseQuery().
		Desc("created_at").
		Limit(pageSize, (page-1)*pageSize).
		Find(&jobs); err != nil {
		return nil, 0, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return jobs, total, nil
}

func (r *ForumRepo) UpdateMergeJob(ctx context.Context, job *entity.MergeJob, cols ...string) error {
	job.ID = uid.DeShortID(job.ID)
	_, err := r.data.DB.Context(ctx).ID(job.ID).Cols(cols...).Update(job)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return nil
}

func (r *ForumRepo) ArchivePosts(ctx context.Context, postIDs []string) error {
	ids := make([]string, 0, len(postIDs))
	for _, id := range postIDs {
		ids = append(ids, uid.DeShortID(id))
	}
	if len(ids) == 0 {
		return nil
	}
	now := time.Now()
	_, err := r.data.DB.Context(ctx).In("id", ids).Cols("merge_state", "archived_at").Update(
		&entity.Post{
			MergeState: entity.PostMergeStateArchived,
			ArchivedAt: &now,
		},
	)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return nil
}

func (r *ForumRepo) AddContributionCredits(ctx context.Context, credits []*entity.ContributionCredit) error {
	if len(credits) == 0 {
		return nil
	}
	for _, credit := range credits {
		id, err := r.genID(ctx, credit.TableName())
		if err != nil {
			return err
		}
		credit.ID = id
	}
	if _, err := r.data.DB.Context(ctx).Insert(credits); err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return nil
}

func (r *ForumRepo) ListContributorsByTopic(ctx context.Context, topicID string) ([]*ContributorStat, error) {
	stats := make([]*ContributorStat, 0)
	query := "SELECT user_id, SUM(weight) AS weight FROM contribution_credits WHERE topic_id = ? GROUP BY user_id ORDER BY weight DESC"
	if err := r.data.DB.Context(ctx).SQL(query, uid.DeShortID(topicID)).Find(&stats); err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return stats, nil
}

func (r *ForumRepo) AddDocLink(ctx context.Context, link *entity.DocLink) error {
	id, err := r.genID(ctx, link.TableName())
	if err != nil {
		return err
	}
	link.ID = id
	if _, err = r.data.DB.Context(ctx).Insert(link); err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return nil
}

func (r *ForumRepo) ListDocLinksBySources(ctx context.Context, sourceTopicIDs []string) ([]*entity.DocLink, error) {
	if len(sourceTopicIDs) == 0 {
		return []*entity.DocLink{}, nil
	}
	ids := make([]string, 0, len(sourceTopicIDs))
	for _, id := range sourceTopicIDs {
		ids = append(ids, uid.DeShortID(id))
	}
	links := make([]*entity.DocLink, 0)
	if err := r.data.DB.Context(ctx).In("source_topic_id", ids).Find(&links); err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return links, nil
}

func (r *ForumRepo) UpsertTopicSolution(ctx context.Context, topicID, postID, userID string) error {
	topicID = uid.DeShortID(topicID)
	postID = uid.DeShortID(postID)
	newID, err := r.genID(ctx, entity.TopicSolution{}.TableName())
	if err != nil {
		return err
	}

	_, err = r.data.DB.Transaction(func(session *xorm.Session) (any, error) {
		session = session.Context(ctx)
		existing := &entity.TopicSolution{}
		exist, err := session.Where("topic_id = ?", topicID).Get(existing)
		if err != nil {
			return nil, err
		}
		if exist {
			existing.PostID = postID
			existing.SetByUserID = userID
			if _, err := session.ID(existing.ID).Cols("post_id", "set_by_user_id").Update(existing); err != nil {
				return nil, err
			}
		} else {
			existing.ID = newID
			existing.TopicID = topicID
			existing.PostID = postID
			existing.SetByUserID = userID
			if _, err := session.Insert(existing); err != nil {
				return nil, err
			}
		}
		if _, err := session.ID(topicID).Cols("solved_post_id").Update(&entity.Topic{SolvedPostID: postID}); err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return nil
}

func (r *ForumRepo) UpsertTopicVote(ctx context.Context, topicID, userID string, value int) error {
	return r.upsertVoteWithCounter(
		ctx,
		&entity.TopicVote{TopicID: uid.DeShortID(topicID), UserID: userID, Value: value},
		func(session *xorm.Session, delta int) error {
			if delta == 0 {
				return nil
			}
			_, err := session.ID(uid.DeShortID(topicID)).Incr("vote_count", delta).Update(&entity.Topic{})
			return err
		},
	)
}

func (r *ForumRepo) UpsertPostVote(ctx context.Context, postID, userID string, value int) error {
	return r.upsertVoteWithCounter(
		ctx,
		&entity.PostVote{PostID: uid.DeShortID(postID), UserID: userID, Value: value},
		func(session *xorm.Session, delta int) error {
			if delta == 0 {
				return nil
			}
			_, err := session.ID(uid.DeShortID(postID)).Incr("vote_count", delta).Update(&entity.Post{})
			return err
		},
	)
}

func (r *ForumRepo) upsertVoteWithCounter(
	ctx context.Context,
	payload any,
	updateCounter func(session *xorm.Session, delta int) error,
) error {
	var newID string
	var err error
	switch vote := payload.(type) {
	case *entity.TopicVote:
		newID, err = r.genID(ctx, vote.TableName())
	case *entity.PostVote:
		newID, err = r.genID(ctx, vote.TableName())
	}
	if err != nil {
		return err
	}

	_, err = r.data.DB.Transaction(func(session *xorm.Session) (any, error) {
		session = session.Context(ctx)
		switch vote := payload.(type) {
		case *entity.TopicVote:
			existing := &entity.TopicVote{}
			exist, err := session.Where("topic_id = ? AND user_id = ?", vote.TopicID, vote.UserID).Get(existing)
			if err != nil {
				return nil, err
			}
			delta := vote.Value
			if exist {
				delta = vote.Value - existing.Value
				existing.Value = vote.Value
				if _, err := session.ID(existing.ID).Cols("value").Update(existing); err != nil {
					return nil, err
				}
			} else {
				vote.ID = newID
				if _, err := session.Insert(vote); err != nil {
					return nil, err
				}
			}
			return nil, updateCounter(session, delta)
		case *entity.PostVote:
			existing := &entity.PostVote{}
			exist, err := session.Where("post_id = ? AND user_id = ?", vote.PostID, vote.UserID).Get(existing)
			if err != nil {
				return nil, err
			}
			delta := vote.Value
			if exist {
				delta = vote.Value - existing.Value
				existing.Value = vote.Value
				if _, err := session.ID(existing.ID).Cols("value").Update(existing); err != nil {
					return nil, err
				}
			} else {
				vote.ID = newID
				if _, err := session.Insert(vote); err != nil {
					return nil, err
				}
			}
			return nil, updateCounter(session, delta)
		default:
			return nil, nil
		}
	})
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return nil
}

func (r *ForumRepo) ListPlatformConfigs(ctx context.Context, prefix string) ([]*entity.Config, error) {
	configs := make([]*entity.Config, 0)
	session := r.data.DB.Context(ctx).OrderBy("id ASC")
	if prefix != "" {
		session = session.Where("`key` like ?", prefix+"%")
	}
	if err := session.Find(&configs); err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return configs, nil
}

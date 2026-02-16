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

func (r *ForumRepo) AddBoard(ctx context.Context, board *entity.Board) error {
	id, err := r.genID(ctx, board.TableName())
	if err != nil {
		return err
	}
	board.ID = id
	if _, err = r.data.DB.Context(ctx).Insert(board); err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return nil
}

func (r *ForumRepo) GetBoard(ctx context.Context, boardID string) (*entity.Board, bool, error) {
	board := &entity.Board{ID: uid.DeShortID(boardID)}
	exist, err := r.data.DB.Context(ctx).Get(board)
	if err != nil {
		return nil, false, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return board, exist, nil
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

func (r *ForumRepo) ListTopicsByBoard(ctx context.Context, boardID string, page, pageSize int) ([]*entity.Topic, int64, error) {
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
	total, err := r.data.DB.Context(ctx).Where("board_id = ?", uid.DeShortID(boardID)).Count(&entity.Topic{})
	if err != nil {
		return nil, 0, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	if err := r.data.DB.Context(ctx).
		Where("board_id = ?", uid.DeShortID(boardID)).
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

	_, err = r.data.DB.Transaction(func(session *xorm.Session) (any, error) {
		session = session.Context(ctx)
		if _, err := session.Insert(job); err != nil {
			return nil, err
		}

		for _, rawPostID := range postIDs {
			ref := &entity.MergeJobPostRef{
				MergeJobID: job.ID,
				PostID:     uid.DeShortID(rawPostID),
			}
			refID, err := r.genID(ctx, ref.TableName())
			if err != nil {
				return nil, err
			}
			ref.ID = refID
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
	userID = uid.DeShortID(userID)

	_, err := r.data.DB.Transaction(func(session *xorm.Session) (any, error) {
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
			id, err := r.genID(ctx, existing.TableName())
			if err != nil {
				return nil, err
			}
			existing.ID = id
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
		&entity.TopicVote{TopicID: uid.DeShortID(topicID), UserID: uid.DeShortID(userID), Value: value},
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
		&entity.PostVote{PostID: uid.DeShortID(postID), UserID: uid.DeShortID(userID), Value: value},
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
	_, err := r.data.DB.Transaction(func(session *xorm.Session) (any, error) {
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
				id, err := r.genID(ctx, vote.TableName())
				if err != nil {
					return nil, err
				}
				vote.ID = id
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
				id, err := r.genID(ctx, vote.TableName())
				if err != nil {
					return nil, err
				}
				vote.ID = id
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

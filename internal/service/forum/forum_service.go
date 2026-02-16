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
	"sort"
	"time"

	domainforum "github.com/apache/answer/internal/domain/forum"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/entity"
	forumrepo "github.com/apache/answer/internal/repo/forum"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/plugin_common"
	"github.com/apache/answer/pkg/uid"
	"github.com/apache/answer/plugin"
	"github.com/segmentfault/pacman/errors"
)

type ForumService struct {
	forumRepo            *forumrepo.ForumRepo
	pluginCommonService  *plugin_common.PluginCommonService
}

func NewForumService(
	forumRepo *forumrepo.ForumRepo,
	pluginCommonService *plugin_common.PluginCommonService,
) *ForumService {
	return &ForumService{
		forumRepo:           forumRepo,
		pluginCommonService: pluginCommonService,
	}
}

func (s *ForumService) CreateBoard(ctx context.Context, req *schema.CreateBoardReq) (*entity.Board, error) {
	board := &entity.Board{
		CreatorID:   req.CreatorID,
		Slug:        req.Slug,
		Name:        req.Name,
		Description: req.Description,
		Status:      1,
	}
	if err := s.forumRepo.AddBoard(ctx, board); err != nil {
		return nil, err
	}
	return board, nil
}

func (s *ForumService) CreateTopic(ctx context.Context, req *schema.CreateTopicReq) (*entity.Topic, error) {
	if _, exist, err := s.forumRepo.GetBoard(ctx, req.BoardID); err != nil {
		return nil, err
	} else if !exist {
		return nil, errors.NotFound(reason.ObjectNotFound)
	}

	topic := &entity.Topic{
		BoardID:       uid.DeShortID(req.BoardID),
		UserID:        req.UserID,
		Title:         req.Title,
		TopicKind:     req.TopicKind,
		IsWikiEnabled: req.IsWikiEnabled,
		Status:        entity.TopicStatusAvailable,
	}
	if err := s.forumRepo.AddTopic(ctx, topic); err != nil {
		return nil, err
	}
	return topic, nil
}

func (s *ForumService) ListTopicsByBoard(ctx context.Context, boardID string, req *schema.TopicListReq) (
	topics []*entity.Topic, total int64, err error,
) {
	return s.forumRepo.ListTopicsByBoard(ctx, boardID, req.Page, req.PageSize)
}

func (s *ForumService) CreatePost(ctx context.Context, topicID string, req *schema.CreatePostReq) (*entity.Post, error) {
	topic, exist, err := s.forumRepo.GetTopic(ctx, topicID)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, errors.NotFound(reason.ObjectNotFound)
	}
	if topic.Status != entity.TopicStatusAvailable {
		return nil, errors.Forbidden(reason.StatusInvalid)
	}

	post := &entity.Post{
		TopicID:    uid.DeShortID(topicID),
		UserID:     req.UserID,
		Original:   req.OriginalText,
		Parsed:     req.OriginalText,
		MergeState: entity.PostMergeStateActive,
		Status:     1,
	}
	if err := s.forumRepo.AddPost(ctx, post); err != nil {
		return nil, err
	}
	return post, nil
}

func (s *ForumService) GetTopicWiki(ctx context.Context, topicID string) (*entity.WikiRevision, error) {
	topic, exist, err := s.forumRepo.GetTopic(ctx, topicID)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, errors.NotFound(reason.ObjectNotFound)
	}
	if topic.CurrentWikiRevisionID == "" || topic.CurrentWikiRevisionID == "0" {
		return nil, nil
	}
	revision, exist, err := s.forumRepo.GetWikiRevision(ctx, topic.CurrentWikiRevisionID)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, nil
	}
	return revision, nil
}

func (s *ForumService) CreateWikiRevision(ctx context.Context, topicID string, req *schema.CreateWikiRevisionReq) (*entity.WikiRevision, error) {
	topic, exist, err := s.forumRepo.GetTopic(ctx, topicID)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, errors.NotFound(reason.ObjectNotFound)
	}
	if !topic.IsWikiEnabled {
		return nil, errors.Forbidden(reason.ForbiddenError)
	}

	revision := &entity.WikiRevision{
		TopicID:          uid.DeShortID(topicID),
		EditorID:         req.EditorID,
		Title:            req.Title,
		Document:         req.Document,
		Summary:          req.Summary,
		ParentRevisionID: topic.CurrentWikiRevisionID,
	}
	if err := s.forumRepo.AddWikiRevision(ctx, revision); err != nil {
		return nil, err
	}

	aggregate := &domainforum.TopicAggregate{
		ID:                    topic.ID,
		CurrentWikiRevisionID: topic.CurrentWikiRevisionID,
	}
	if err := aggregate.ApplyWikiRevision(revision.ID); err != nil {
		return nil, errors.BadRequest(reason.RequestFormatError).WithError(err)
	}
	topic.CurrentWikiRevisionID = aggregate.CurrentWikiRevisionID
	if err := s.forumRepo.UpdateTopic(ctx, topic, "current_wiki_revision_id"); err != nil {
		return nil, err
	}
	return revision, nil
}

func (s *ForumService) ListWikiRevisions(ctx context.Context, topicID string) ([]*entity.WikiRevision, error) {
	if _, exist, err := s.forumRepo.GetTopic(ctx, topicID); err != nil {
		return nil, err
	} else if !exist {
		return nil, errors.NotFound(reason.ObjectNotFound)
	}
	return s.forumRepo.ListWikiRevisions(ctx, topicID)
}

func (s *ForumService) CreateMergeJob(ctx context.Context, topicID string, req *schema.CreateMergeJobReq) (*entity.MergeJob, error) {
	if _, exist, err := s.forumRepo.GetTopic(ctx, topicID); err != nil {
		return nil, err
	} else if !exist {
		return nil, errors.NotFound(reason.ObjectNotFound)
	}
	posts, err := s.forumRepo.GetPostsByIDs(ctx, topicID, req.PostIDs)
	if err != nil {
		return nil, err
	}
	if len(posts) == 0 || len(posts) != len(req.PostIDs) {
		return nil, errors.BadRequest(reason.ObjectNotFound)
	}

	job := &entity.MergeJob{
		TopicID:   uid.DeShortID(topicID),
		CreatorID: req.CreatorID,
		Status:    entity.MergeJobStatusPending,
		Summary:   req.Summary,
	}
	if err := s.forumRepo.AddMergeJob(ctx, job, req.PostIDs); err != nil {
		return nil, err
	}
	return job, nil
}

func (s *ForumService) GetMergeJob(ctx context.Context, topicID, jobID string) (*entity.MergeJob, []*entity.MergeJobPostRef, error) {
	job, refs, exist, err := s.forumRepo.GetMergeJob(ctx, jobID)
	if err != nil {
		return nil, nil, err
	}
	if !exist || job.TopicID != uid.DeShortID(topicID) {
		return nil, nil, errors.NotFound(reason.ObjectNotFound)
	}
	return job, refs, nil
}

func (s *ForumService) ApplyMergeJob(ctx context.Context, topicID, jobID string, req *schema.ApplyMergeJobReq) (*entity.WikiRevision, error) {
	topic, exist, err := s.forumRepo.GetTopic(ctx, topicID)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, errors.NotFound(reason.ObjectNotFound)
	}
	job, refs, exist, err := s.forumRepo.GetMergeJob(ctx, jobID)
	if err != nil {
		return nil, err
	}
	if !exist || job.TopicID != uid.DeShortID(topicID) {
		return nil, errors.NotFound(reason.ObjectNotFound)
	}

	if job.Status == entity.MergeJobStatusApplied && job.AppliedRevisionID != "" && job.AppliedRevisionID != "0" {
		revision, has, err := s.forumRepo.GetWikiRevision(ctx, job.AppliedRevisionID)
		if err != nil {
			return nil, err
		}
		if has {
			return revision, nil
		}
	}

	aggregate := &domainforum.MergeJobAggregate{
		ID:                job.ID,
		Status:            domainforum.MergeJobStatus(job.Status),
		AppliedRevisionID: job.AppliedRevisionID,
	}
	if aggregate.Status == domainforum.MergeJobPending {
		if err := aggregate.MarkReviewed(); err != nil {
			return nil, errors.BadRequest(reason.StatusInvalid).WithError(err)
		}
		job.Status = string(aggregate.Status)
		job.ReviewerID = req.ReviewerID
		if err := s.forumRepo.UpdateMergeJob(ctx, job, "status", "reviewer_id"); err != nil {
			return nil, err
		}
	}

	revision := &entity.WikiRevision{
		TopicID:          uid.DeShortID(topicID),
		EditorID:         req.OperatorID,
		Title:            req.Title,
		Document:         req.Document,
		Summary:          req.Summary,
		ParentRevisionID: topic.CurrentWikiRevisionID,
	}
	if err := s.forumRepo.AddWikiRevision(ctx, revision); err != nil {
		return nil, err
	}
	if err := aggregate.Apply(revision.ID); err != nil {
		return nil, errors.BadRequest(reason.StatusInvalid).WithError(err)
	}

	// Topic invariant: only one current wiki revision can be active at a time.
	topicAggregate := &domainforum.TopicAggregate{
		ID:                    topic.ID,
		CurrentWikiRevisionID: topic.CurrentWikiRevisionID,
	}
	if err := topicAggregate.ApplyWikiRevision(revision.ID); err != nil {
		return nil, errors.BadRequest(reason.RequestFormatError).WithError(err)
	}
	topic.CurrentWikiRevisionID = topicAggregate.CurrentWikiRevisionID
	if err := s.forumRepo.UpdateTopic(ctx, topic, "current_wiki_revision_id"); err != nil {
		return nil, err
	}

	postIDs := make([]string, 0, len(refs))
	for _, ref := range refs {
		postIDs = append(postIDs, ref.PostID)
	}
	if err := s.forumRepo.ArchivePosts(ctx, postIDs); err != nil {
		return nil, err
	}

	posts, err := s.forumRepo.GetPostsByIDs(ctx, topicID, postIDs)
	if err != nil {
		return nil, err
	}
	weight := req.ContributionWeight
	if weight <= 0 {
		weight = 1
	}
	credits := make([]*entity.ContributionCredit, 0, len(posts))
	for _, post := range posts {
		credits = append(credits, &entity.ContributionCredit{
			TopicID:    uid.DeShortID(topicID),
			RevisionID: revision.ID,
			UserID:     post.UserID,
			Weight:     weight,
		})
	}
	if err := s.forumRepo.AddContributionCredits(ctx, credits); err != nil {
		return nil, err
	}

	now := time.Now()
	job.Status = string(aggregate.Status)
	job.AppliedRevisionID = revision.ID
	job.AppliedAt = &now
	job.ReviewerID = req.ReviewerID
	if err := s.forumRepo.UpdateMergeJob(ctx, job, "status", "applied_revision_id", "applied_at", "reviewer_id"); err != nil {
		return nil, err
	}
	return revision, nil
}

func (s *ForumService) ListContributorsByTopic(ctx context.Context, topicID string) ([]*forumrepo.ContributorStat, error) {
	if _, exist, err := s.forumRepo.GetTopic(ctx, topicID); err != nil {
		return nil, err
	} else if !exist {
		return nil, errors.NotFound(reason.ObjectNotFound)
	}
	return s.forumRepo.ListContributorsByTopic(ctx, topicID)
}

func (s *ForumService) AddDocLink(ctx context.Context, req *schema.CreateDocLinkReq) (*entity.DocLink, error) {
	if req.LinkType == "" {
		req.LinkType = entity.DocLinkTypeRelated
	}
	if _, exist, err := s.forumRepo.GetTopic(ctx, req.SourceTopicID); err != nil {
		return nil, err
	} else if !exist {
		return nil, errors.NotFound(reason.ObjectNotFound)
	}
	if _, exist, err := s.forumRepo.GetTopic(ctx, req.TargetTopicID); err != nil {
		return nil, err
	} else if !exist {
		return nil, errors.NotFound(reason.ObjectNotFound)
	}

	link := &entity.DocLink{
		SourceTopicID: uid.DeShortID(req.SourceTopicID),
		TargetTopicID: uid.DeShortID(req.TargetTopicID),
		LinkType:      req.LinkType,
	}
	if err := s.forumRepo.AddDocLink(ctx, link); err != nil {
		return nil, err
	}
	return link, nil
}

type DocGraph struct {
	Nodes []string          `json:"nodes"`
	Edges []*entity.DocLink `json:"edges"`
}

func (s *ForumService) GetDocGraph(ctx context.Context, req *schema.GetDocGraphReq) (*DocGraph, error) {
	if req.Depth <= 0 {
		req.Depth = 2
	}
	visited := map[string]struct{}{req.RootTopicID: {}}
	currentLayer := []string{req.RootTopicID}
	edges := make([]*entity.DocLink, 0)

	for i := 0; i < req.Depth; i++ {
		if len(currentLayer) == 0 {
			break
		}
		layerLinks, err := s.forumRepo.ListDocLinksBySources(ctx, currentLayer)
		if err != nil {
			return nil, err
		}
		nextLayer := make([]string, 0)
		for _, link := range layerLinks {
			edges = append(edges, link)
			if _, ok := visited[link.TargetTopicID]; !ok {
				visited[link.TargetTopicID] = struct{}{}
				nextLayer = append(nextLayer, link.TargetTopicID)
			}
		}
		currentLayer = nextLayer
	}

	nodes := make([]string, 0, len(visited))
	for id := range visited {
		nodes = append(nodes, id)
	}
	sort.Strings(nodes)
	return &DocGraph{Nodes: nodes, Edges: edges}, nil
}

func (s *ForumService) SetTopicSolution(ctx context.Context, topicID string, req *schema.SetTopicSolutionReq) error {
	_, exist, err := s.forumRepo.GetPost(ctx, req.PostID)
	if err != nil {
		return err
	}
	if !exist {
		return errors.NotFound(reason.ObjectNotFound)
	}
	return s.forumRepo.UpsertTopicSolution(ctx, topicID, req.PostID, req.UserID)
}

func (s *ForumService) VotePost(ctx context.Context, postID string, req *schema.ForumVoteReq) error {
	_, exist, err := s.forumRepo.GetPost(ctx, postID)
	if err != nil {
		return err
	}
	if !exist {
		return errors.NotFound(reason.ObjectNotFound)
	}
	return s.forumRepo.UpsertPostVote(ctx, postID, req.UserID, req.Value)
}

func (s *ForumService) VoteTopic(ctx context.Context, topicID string, req *schema.ForumVoteReq) error {
	_, exist, err := s.forumRepo.GetTopic(ctx, topicID)
	if err != nil {
		return err
	}
	if !exist {
		return errors.NotFound(reason.ObjectNotFound)
	}
	return s.forumRepo.UpsertTopicVote(ctx, topicID, req.UserID, req.Value)
}

func (s *ForumService) GetPlatformPlugins(ctx context.Context) ([]*schema.GetAllPluginStatusResp, error) {
	resp := make([]*schema.GetAllPluginStatusResp, 0)
	err := plugin.CallBase(func(base plugin.Base) error {
		info := base.Info()
		resp = append(resp, &schema.GetAllPluginStatusResp{
			SlugName: info.SlugName,
			Enabled:  plugin.StatusManager.IsEnabled(info.SlugName),
		})
		return nil
	})
	if err != nil {
		return nil, errors.InternalServer(reason.UnknownError).WithError(err)
	}
	return resp, nil
}

func (s *ForumService) UpdatePlatformPluginConfig(ctx context.Context, pluginSlug string, req *schema.PlatformPluginConfigReq) error {
	updateReq := &schema.UpdatePluginConfigReq{
		PluginSlugName: pluginSlug,
		ConfigFields:   req.ConfigFields,
	}
	return s.pluginCommonService.UpdatePluginConfig(ctx, updateReq)
}

func (s *ForumService) ListPlatformConfig(ctx context.Context) (map[string]string, error) {
	configs, err := s.forumRepo.ListPlatformConfigs(ctx, "forum.")
	if err != nil {
		return nil, err
	}
	result := make(map[string]string, len(configs))
	for _, cfg := range configs {
		result[cfg.Key] = cfg.Value
	}
	pluginConfigs, err := s.forumRepo.ListPlatformConfigs(ctx, "plugin.status")
	if err != nil {
		return nil, err
	}
	for _, cfg := range pluginConfigs {
		result[cfg.Key] = cfg.Value
	}
	return result, nil
}

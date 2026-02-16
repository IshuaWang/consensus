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

package controller

import (
	"github.com/apache/answer/internal/base/handler"
	"github.com/apache/answer/internal/base/middleware"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/forum"
	"github.com/gin-gonic/gin"
	"github.com/segmentfault/pacman/errors"
)

type ForumController struct {
	forumService *forum.ForumService
}

func NewForumController(forumService *forum.ForumService) *ForumController {
	return &ForumController{forumService: forumService}
}

func (fc *ForumController) CreateBoard(ctx *gin.Context) {
	if !middleware.GetUserIsAdminModerator(ctx) {
		handler.HandleResponse(ctx, errors.Forbidden(reason.ForbiddenError), nil)
		return
	}
	req := &schema.CreateBoardReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.CreatorID = middleware.GetLoginUserIDFromContext(ctx)
	board, err := fc.forumService.CreateBoard(ctx, req)
	handler.HandleResponse(ctx, err, board)
}

func (fc *ForumController) ListBoardTopics(ctx *gin.Context) {
	req := &schema.TopicListReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	boardID := ctx.Param("id")
	topics, total, err := fc.forumService.ListTopicsByBoard(ctx, boardID, req)
	handler.HandleResponse(ctx, err, gin.H{
		"list":  topics,
		"total": total,
	})
}

func (fc *ForumController) CreateTopic(ctx *gin.Context) {
	req := &schema.CreateTopicReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	topic, err := fc.forumService.CreateTopic(ctx, req)
	handler.HandleResponse(ctx, err, topic)
}

func (fc *ForumController) CreateTopicPost(ctx *gin.Context) {
	req := &schema.CreatePostReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	post, err := fc.forumService.CreatePost(ctx, ctx.Param("id"), req)
	handler.HandleResponse(ctx, err, post)
}

func (fc *ForumController) GetTopicWiki(ctx *gin.Context) {
	revision, err := fc.forumService.GetTopicWiki(ctx, ctx.Param("id"))
	handler.HandleResponse(ctx, err, revision)
}

func (fc *ForumController) CreateTopicWikiRevision(ctx *gin.Context) {
	req := &schema.CreateWikiRevisionReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.EditorID = middleware.GetLoginUserIDFromContext(ctx)
	revision, err := fc.forumService.CreateWikiRevision(ctx, ctx.Param("id"), req)
	handler.HandleResponse(ctx, err, revision)
}

func (fc *ForumController) ListTopicWikiRevisions(ctx *gin.Context) {
	revisions, err := fc.forumService.ListWikiRevisions(ctx, ctx.Param("id"))
	handler.HandleResponse(ctx, err, revisions)
}

func (fc *ForumController) CreateMergeJob(ctx *gin.Context) {
	req := &schema.CreateMergeJobReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.CreatorID = middleware.GetLoginUserIDFromContext(ctx)
	mergeJob, err := fc.forumService.CreateMergeJob(ctx, ctx.Param("id"), req)
	handler.HandleResponse(ctx, err, mergeJob)
}

func (fc *ForumController) GetMergeJob(ctx *gin.Context) {
	mergeJob, refs, err := fc.forumService.GetMergeJob(ctx, ctx.Param("id"), ctx.Param("jobId"))
	handler.HandleResponse(ctx, err, gin.H{
		"job":       mergeJob,
		"post_refs": refs,
	})
}

func (fc *ForumController) ApplyMergeJob(ctx *gin.Context) {
	if !middleware.GetUserIsAdminModerator(ctx) {
		handler.HandleResponse(ctx, errors.Forbidden(reason.ForbiddenError), nil)
		return
	}
	req := &schema.ApplyMergeJobReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	userID := middleware.GetLoginUserIDFromContext(ctx)
	req.ReviewerID = userID
	req.OperatorID = userID
	revision, err := fc.forumService.ApplyMergeJob(ctx, ctx.Param("id"), ctx.Param("jobId"), req)
	handler.HandleResponse(ctx, err, revision)
}

func (fc *ForumController) ListTopicContributors(ctx *gin.Context) {
	contributors, err := fc.forumService.ListContributorsByTopic(ctx, ctx.Param("id"))
	handler.HandleResponse(ctx, err, contributors)
}

func (fc *ForumController) CreateDocLink(ctx *gin.Context) {
	req := &schema.CreateDocLinkReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	link, err := fc.forumService.AddDocLink(ctx, req)
	handler.HandleResponse(ctx, err, link)
}

func (fc *ForumController) GetDocGraph(ctx *gin.Context) {
	req := &schema.GetDocGraphReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	graph, err := fc.forumService.GetDocGraph(ctx, req)
	handler.HandleResponse(ctx, err, graph)
}

func (fc *ForumController) SetTopicSolution(ctx *gin.Context) {
	req := &schema.SetTopicSolutionReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	err := fc.forumService.SetTopicSolution(ctx, ctx.Param("id"), req)
	handler.HandleResponse(ctx, err, nil)
}

func (fc *ForumController) VotePost(ctx *gin.Context) {
	req := &schema.ForumVoteReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	err := fc.forumService.VotePost(ctx, ctx.Param("id"), req)
	handler.HandleResponse(ctx, err, nil)
}

func (fc *ForumController) VoteTopic(ctx *gin.Context) {
	req := &schema.ForumVoteReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	err := fc.forumService.VoteTopic(ctx, ctx.Param("id"), req)
	handler.HandleResponse(ctx, err, nil)
}

func (fc *ForumController) GetPlatformPlugins(ctx *gin.Context) {
	plugins, err := fc.forumService.GetPlatformPlugins(ctx)
	handler.HandleResponse(ctx, err, plugins)
}

func (fc *ForumController) UpdatePlatformPluginConfig(ctx *gin.Context) {
	if !middleware.GetUserIsAdminModerator(ctx) {
		handler.HandleResponse(ctx, errors.Forbidden(reason.ForbiddenError), nil)
		return
	}
	req := &schema.PlatformPluginConfigReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	err := fc.forumService.UpdatePlatformPluginConfig(ctx, ctx.Param("id"), req)
	handler.HandleResponse(ctx, err, nil)
}

func (fc *ForumController) GetPlatformConfig(ctx *gin.Context) {
	config, err := fc.forumService.ListPlatformConfig(ctx)
	handler.HandleResponse(ctx, err, config)
}

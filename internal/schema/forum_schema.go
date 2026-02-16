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

package schema

type CreateBoardReq struct {
	Slug        string `validate:"required,gt=1,lte=100" json:"slug"`
	Name        string `validate:"required,gt=1,lte=120" json:"name"`
	Description string `validate:"omitempty,lte=500" json:"description"`
	CreatorID   string `json:"-"`
}

type CreateTopicReq struct {
	BoardID       string `validate:"required" json:"board_id"`
	Title         string `validate:"required,gt=1,lte=180" json:"title"`
	TopicKind     string `validate:"required,oneof=discussion knowledge" json:"topic_kind"`
	IsWikiEnabled bool   `json:"is_wiki_enabled"`
	UserID        string `json:"-"`
}

type CreatePostReq struct {
	OriginalText string `validate:"required,notblank,gte=2,lte=20000" json:"original_text"`
	UserID       string `json:"-"`
}

type CreateWikiRevisionReq struct {
	Title         string   `validate:"required,gt=1,lte=180" json:"title"`
	Document      string   `validate:"required,notblank,gte=2,lte=200000" json:"document"`
	Summary       string   `validate:"omitempty,lte=500" json:"summary"`
	SourcePostIDs []string `json:"source_post_ids"`
	EditorID      string   `json:"-"`
}

type CreateMergeJobReq struct {
	PostIDs   []string `validate:"required,min=1" json:"post_ids"`
	Summary   string   `validate:"omitempty,lte=500" json:"summary"`
	CreatorID string   `json:"-"`
}

type ApplyMergeJobReq struct {
	Title       string `validate:"required,gt=1,lte=180" json:"title"`
	Document    string `validate:"required,notblank,gte=2,lte=200000" json:"document"`
	Summary     string `validate:"omitempty,lte=500" json:"summary"`
	ReviewerID  string `json:"-"`
	OperatorID  string `json:"-"`
	ContributionWeight int `json:"contribution_weight"`
}

type CreateDocLinkReq struct {
	SourceTopicID string `validate:"required" json:"source_topic_id"`
	TargetTopicID string `validate:"required" json:"target_topic_id"`
	LinkType      string `validate:"omitempty,oneof=related" json:"link_type"`
}

type SetTopicSolutionReq struct {
	PostID string `validate:"required" json:"post_id"`
	UserID string `json:"-"`
}

type VoteReq struct {
	Value  int    `validate:"required,oneof=-1 1" json:"value"`
	UserID string `json:"-"`
}

type TopicListReq struct {
	Page     int `validate:"omitempty,min=1" form:"page"`
	PageSize int `validate:"omitempty,min=1,max=100" form:"page_size"`
}

type GetDocGraphReq struct {
	RootTopicID string `validate:"required" form:"root_topic_id"`
	Depth       int    `validate:"omitempty,min=1,max=5" form:"depth"`
}

type PlatformPluginConfigReq struct {
	ConfigFields map[string]any `json:"config_fields"`
}


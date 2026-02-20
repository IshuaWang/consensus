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

package constant

const (
	QuestionObjectType   = "question"
	AnswerObjectType     = "answer"
	TagObjectType        = "tag"
	UserObjectType       = "user"
	CollectionObjectType = "collection"
	CommentObjectType    = "comment"
	ReportObjectType     = "report"
	BadgeObjectType      = "badge"
	BadgeAwardObjectType = "badge_award"
	CategoryObjectType   = "categories"
	TopicObjectType      = "topics"
	PostObjectType       = "posts"
	WikiRevisionType     = "wiki_revisions"
	MergeJobObjectType   = "merge_jobs"
	MergeRefObjectType   = "merge_job_post_refs"
	ContributionType     = "contribution_credits"
	DocLinkObjectType    = "doc_links"
	TopicVoteObjectType  = "topic_votes"
	PostVoteObjectType   = "post_votes"
	TopicSolutionType    = "topic_solutions"
)

var (
	ObjectTypeStrMapping = map[string]int{
		QuestionObjectType:   1,
		AnswerObjectType:     2,
		TagObjectType:        3,
		UserObjectType:       4,
		CollectionObjectType: 6,
		CommentObjectType:    7,
		ReportObjectType:     8,
		BadgeObjectType:      9,
		BadgeAwardObjectType: 10,
		CategoryObjectType:   11,
		TopicObjectType:      12,
		PostObjectType:       13,
		WikiRevisionType:     14,
		MergeJobObjectType:   15,
		MergeRefObjectType:   16,
		ContributionType:     17,
		DocLinkObjectType:    18,
		TopicVoteObjectType:  19,
		PostVoteObjectType:   20,
		TopicSolutionType:    21,
	}

	ObjectTypeNumberMapping = map[int]string{
		1:  QuestionObjectType,
		2:  AnswerObjectType,
		3:  TagObjectType,
		4:  UserObjectType,
		6:  CollectionObjectType,
		7:  CommentObjectType,
		8:  ReportObjectType,
		9:  BadgeObjectType,
		10: BadgeAwardObjectType,
		11: CategoryObjectType,
		12: TopicObjectType,
		13: PostObjectType,
		14: WikiRevisionType,
		15: MergeJobObjectType,
		16: MergeRefObjectType,
		17: ContributionType,
		18: DocLinkObjectType,
		19: TopicVoteObjectType,
		20: PostVoteObjectType,
		21: TopicSolutionType,
	}
)

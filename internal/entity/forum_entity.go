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

package entity

import "time"

const (
	TopicKindDiscussion = "discussion"
	TopicKindKnowledge  = "knowledge"

	TopicStatusAvailable = "available"
	TopicStatusClosed    = "closed"

	PostMergeStateActive   = "active"
	PostMergeStateArchived = "archived"

	MergeJobStatusPending  = "pending"
	MergeJobStatusReviewed = "reviewed"
	MergeJobStatusApplied  = "applied"

	DocLinkTypeRelated = "related"
)

type Board struct {
	ID          string    `xorm:"not null pk BIGINT(20) id"`
	CreatedAt   time.Time `xorm:"not null default CURRENT_TIMESTAMP TIMESTAMP created_at"`
	UpdatedAt   time.Time `xorm:"updated TIMESTAMP"`
	CreatorID   string    `xorm:"not null default 0 BIGINT(20) creator_id"`
	Slug        string    `xorm:"not null default '' unique VARCHAR(100) slug"`
	Name        string    `xorm:"not null default '' VARCHAR(120) name"`
	Description string    `xorm:"not null default '' VARCHAR(500) description"`
	Status      int       `xorm:"not null default 1 INT(11) status"`
}

func (Board) TableName() string {
	return "boards"
}

type Topic struct {
	ID                    string    `xorm:"not null pk BIGINT(20) id"`
	CreatedAt             time.Time `xorm:"not null default CURRENT_TIMESTAMP TIMESTAMP created_at"`
	UpdatedAt             time.Time `xorm:"updated TIMESTAMP"`
	BoardID               string    `xorm:"not null default 0 BIGINT(20) INDEX board_id"`
	UserID                string    `xorm:"not null default 0 BIGINT(20) INDEX user_id"`
	Title                 string    `xorm:"not null default '' VARCHAR(180) title"`
	TopicKind             string    `xorm:"not null default 'discussion' VARCHAR(30) topic_kind"`
	IsWikiEnabled         bool      `xorm:"not null default false BOOL is_wiki_enabled"`
	CurrentWikiRevisionID string    `xorm:"not null default 0 BIGINT(20) current_wiki_revision_id"`
	SolvedPostID          string    `xorm:"not null default 0 BIGINT(20) solved_post_id"`
	Status                string    `xorm:"not null default 'available' VARCHAR(30) status"`
	PostCount             int       `xorm:"not null default 0 INT(11) post_count"`
	VoteCount             int       `xorm:"not null default 0 INT(11) vote_count"`
	LastPostID            string    `xorm:"not null default 0 BIGINT(20) last_post_id"`
}

func (Topic) TableName() string {
	return "topics"
}

type Post struct {
	ID         string     `xorm:"not null pk BIGINT(20) id"`
	CreatedAt  time.Time  `xorm:"not null default CURRENT_TIMESTAMP TIMESTAMP created_at"`
	UpdatedAt  time.Time  `xorm:"updated TIMESTAMP"`
	TopicID    string     `xorm:"not null default 0 BIGINT(20) INDEX topic_id"`
	UserID     string     `xorm:"not null default 0 BIGINT(20) INDEX user_id"`
	Original   string     `xorm:"not null MEDIUMTEXT original_text"`
	Parsed     string     `xorm:"not null MEDIUMTEXT parsed_text"`
	MergeState string     `xorm:"not null default 'active' VARCHAR(30) merge_state"`
	ArchivedAt *time.Time `xorm:"TIMESTAMP archived_at"`
	VoteCount  int        `xorm:"not null default 0 INT(11) vote_count"`
	Status     int        `xorm:"not null default 1 INT(11) status"`
}

func (Post) TableName() string {
	return "posts"
}

type WikiRevision struct {
	ID               string    `xorm:"not null pk BIGINT(20) id"`
	CreatedAt        time.Time `xorm:"not null default CURRENT_TIMESTAMP TIMESTAMP created_at"`
	UpdatedAt        time.Time `xorm:"updated TIMESTAMP"`
	TopicID          string    `xorm:"not null default 0 BIGINT(20) INDEX topic_id"`
	EditorID         string    `xorm:"not null default 0 BIGINT(20) INDEX editor_id"`
	Title            string    `xorm:"not null default '' VARCHAR(180) title"`
	Document         string    `xorm:"not null MEDIUMTEXT document"`
	Summary          string    `xorm:"not null default '' VARCHAR(500) summary"`
	ParentRevisionID string    `xorm:"not null default 0 BIGINT(20) parent_revision_id"`
}

func (WikiRevision) TableName() string {
	return "wiki_revisions"
}

type MergeJob struct {
	ID                string     `xorm:"not null pk BIGINT(20) id"`
	CreatedAt         time.Time  `xorm:"not null default CURRENT_TIMESTAMP TIMESTAMP created_at"`
	UpdatedAt         time.Time  `xorm:"updated TIMESTAMP"`
	TopicID           string     `xorm:"not null default 0 BIGINT(20) INDEX topic_id"`
	CreatorID         string     `xorm:"not null default 0 BIGINT(20) creator_id"`
	ReviewerID        string     `xorm:"not null default 0 BIGINT(20) reviewer_id"`
	Status            string     `xorm:"not null default 'pending' VARCHAR(30) status"`
	Summary           string     `xorm:"not null default '' VARCHAR(500) summary"`
	AppliedRevisionID string     `xorm:"not null default 0 BIGINT(20) applied_revision_id"`
	AppliedAt         *time.Time `xorm:"TIMESTAMP applied_at"`
}

func (MergeJob) TableName() string {
	return "merge_jobs"
}

type MergeJobPostRef struct {
	ID         string    `xorm:"not null pk BIGINT(20) id"`
	CreatedAt  time.Time `xorm:"not null default CURRENT_TIMESTAMP TIMESTAMP created_at"`
	UpdatedAt  time.Time `xorm:"updated TIMESTAMP"`
	MergeJobID string    `xorm:"not null default 0 BIGINT(20) INDEX merge_job_id"`
	PostID     string    `xorm:"not null default 0 BIGINT(20) INDEX post_id"`
}

func (MergeJobPostRef) TableName() string {
	return "merge_job_post_refs"
}

type ContributionCredit struct {
	ID         string    `xorm:"not null pk BIGINT(20) id"`
	CreatedAt  time.Time `xorm:"not null default CURRENT_TIMESTAMP TIMESTAMP created_at"`
	UpdatedAt  time.Time `xorm:"updated TIMESTAMP"`
	TopicID    string    `xorm:"not null default 0 BIGINT(20) INDEX topic_id"`
	RevisionID string    `xorm:"not null default 0 BIGINT(20) INDEX revision_id"`
	UserID     string    `xorm:"not null default 0 BIGINT(20) INDEX user_id"`
	Weight     int       `xorm:"not null default 1 INT(11) weight"`
}

func (ContributionCredit) TableName() string {
	return "contribution_credits"
}

type DocLink struct {
	ID            string    `xorm:"not null pk BIGINT(20) id"`
	CreatedAt     time.Time `xorm:"not null default CURRENT_TIMESTAMP TIMESTAMP created_at"`
	UpdatedAt     time.Time `xorm:"updated TIMESTAMP"`
	SourceTopicID string    `xorm:"not null default 0 BIGINT(20) INDEX source_topic_id"`
	TargetTopicID string    `xorm:"not null default 0 BIGINT(20) INDEX target_topic_id"`
	LinkType      string    `xorm:"not null default 'related' VARCHAR(30) link_type"`
}

func (DocLink) TableName() string {
	return "doc_links"
}

type TopicVote struct {
	ID        string    `xorm:"not null pk BIGINT(20) id"`
	CreatedAt time.Time `xorm:"not null default CURRENT_TIMESTAMP TIMESTAMP created_at"`
	UpdatedAt time.Time `xorm:"updated TIMESTAMP"`
	TopicID   string    `xorm:"not null default 0 BIGINT(20) unique(vote_topic_user) topic_id"`
	UserID    string    `xorm:"not null default 0 BIGINT(20) unique(vote_topic_user) user_id"`
	Value     int       `xorm:"not null default 0 INT(11) value"`
}

func (TopicVote) TableName() string {
	return "topic_votes"
}

type PostVote struct {
	ID        string    `xorm:"not null pk BIGINT(20) id"`
	CreatedAt time.Time `xorm:"not null default CURRENT_TIMESTAMP TIMESTAMP created_at"`
	UpdatedAt time.Time `xorm:"updated TIMESTAMP"`
	PostID    string    `xorm:"not null default 0 BIGINT(20) unique(vote_post_user) post_id"`
	UserID    string    `xorm:"not null default 0 BIGINT(20) unique(vote_post_user) user_id"`
	Value     int       `xorm:"not null default 0 INT(11) value"`
}

func (PostVote) TableName() string {
	return "post_votes"
}

type TopicSolution struct {
	ID          string    `xorm:"not null pk BIGINT(20) id"`
	CreatedAt   time.Time `xorm:"not null default CURRENT_TIMESTAMP TIMESTAMP created_at"`
	UpdatedAt   time.Time `xorm:"updated TIMESTAMP"`
	TopicID     string    `xorm:"not null default 0 BIGINT(20) unique topic_id"`
	PostID      string    `xorm:"not null default 0 BIGINT(20) post_id"`
	SetByUserID string    `xorm:"not null default 0 BIGINT(20) set_by_user_id"`
}

func (TopicSolution) TableName() string {
	return "topic_solutions"
}


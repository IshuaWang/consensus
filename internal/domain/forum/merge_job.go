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

type MergeJobStatus string

const (
	MergeJobPending  MergeJobStatus = "pending"
	MergeJobReviewed MergeJobStatus = "reviewed"
	MergeJobApplied  MergeJobStatus = "applied"
)

// MergeJobAggregate enforces pending -> reviewed -> applied state transitions.
type MergeJobAggregate struct {
	ID                string
	Status            MergeJobStatus
	AppliedRevisionID string
}

func (m *MergeJobAggregate) MarkReviewed() error {
	if m.Status != MergeJobPending {
		return ErrMergeStatusTransition
	}
	m.Status = MergeJobReviewed
	return nil
}

func (m *MergeJobAggregate) Apply(revisionID string) error {
	if revisionID == "" {
		return ErrTopicRevisionRequired
	}
	if m.Status == MergeJobApplied {
		if m.AppliedRevisionID == revisionID {
			return nil
		}
		return ErrMergeAlreadyApplied
	}
	if m.Status != MergeJobReviewed {
		return ErrMergeStatusTransition
	}
	m.Status = MergeJobApplied
	m.AppliedRevisionID = revisionID
	return nil
}


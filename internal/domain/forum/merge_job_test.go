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

import "testing"

func TestMergeJobAggregateTransition(t *testing.T) {
	job := &MergeJobAggregate{ID: "m1", Status: MergeJobPending}

	if err := job.MarkReviewed(); err != nil {
		t.Fatalf("mark reviewed failed: %v", err)
	}
	if job.Status != MergeJobReviewed {
		t.Fatalf("expected reviewed status, got %s", job.Status)
	}

	if err := job.Apply("r1"); err != nil {
		t.Fatalf("apply failed: %v", err)
	}
	if job.Status != MergeJobApplied {
		t.Fatalf("expected applied status, got %s", job.Status)
	}
	if job.AppliedRevisionID != "r1" {
		t.Fatalf("expected applied revision r1, got %s", job.AppliedRevisionID)
	}
}

func TestMergeJobAggregateApplyIsIdempotent(t *testing.T) {
	job := &MergeJobAggregate{ID: "m1", Status: MergeJobApplied, AppliedRevisionID: "r1"}
	if err := job.Apply("r1"); err != nil {
		t.Fatalf("expected idempotent apply, got %v", err)
	}
}

func TestMergeJobAggregateApplyWithAnotherRevisionFails(t *testing.T) {
	job := &MergeJobAggregate{ID: "m1", Status: MergeJobApplied, AppliedRevisionID: "r1"}
	if err := job.Apply("r2"); err == nil {
		t.Fatal("expected conflict error on re-apply with different revision")
	}
}


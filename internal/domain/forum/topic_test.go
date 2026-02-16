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

func TestTopicAggregateApplyWikiRevision(t *testing.T) {
	topic := &TopicAggregate{ID: "t1"}

	if err := topic.ApplyWikiRevision("r1"); err != nil {
		t.Fatalf("apply first revision failed: %v", err)
	}
	if topic.CurrentWikiRevisionID != "r1" {
		t.Fatalf("unexpected current revision: %s", topic.CurrentWikiRevisionID)
	}

	if err := topic.ApplyWikiRevision("r2"); err != nil {
		t.Fatalf("apply next revision failed: %v", err)
	}
	if topic.CurrentWikiRevisionID != "r2" {
		t.Fatalf("unexpected current revision after overwrite: %s", topic.CurrentWikiRevisionID)
	}
}

func TestTopicAggregateApplyWikiRevisionRequiresID(t *testing.T) {
	topic := &TopicAggregate{ID: "t1"}
	if err := topic.ApplyWikiRevision(""); err == nil {
		t.Fatal("expected error when revision id is empty")
	}
}


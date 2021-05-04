// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package blobref

import (
	"testing"

	"cloud.google.com/go/datastore"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestBlobRef_LoadKey(t *testing.T) {
	key := uuid.MustParse("d13c289c-8845-485f-b582-c87342d5dade")
	blob := new(BlobRef)
	assert.NoError(t, blob.LoadKey(datastore.NameKey("blob", key.String(), nil)))
	assert.Equal(t, key, blob.Key)
}

func TestBlobRef_Save(t *testing.T) {
	const (
		size       = int64(123)
		objectName = "object name"
		store      = "store"
		record     = "record"
	)

	blob := BlobRef{
		Size:      size,
		Status:    StatusInitializing,
		StoreKey:  store,
		RecordKey: record,
	}
	expected := []datastore.Property{
		{
			Name:  "Size",
			Value: size,
		},
		{
			Name:  "Status",
			Value: int64(StatusInitializing),
		},
		{
			Name:  "StoreKey",
			Value: store,
		},
		{
			Name:  "RecordKey",
			Value: record,
		},
	}
	actual, err := blob.Save()
	assert.NoError(t, err)
	if assert.NotNil(t, actual) {
		assert.Equal(t, expected, actual[:len(expected)])
		assert.Equal(t, 5, len(actual))
		assert.Equal(t, "Timestamps", actual[4].Name)
	}
}

func TestBlobRef_Load(t *testing.T) {
	const (
		size       = int64(123)
		objectName = "object name"
		store      = "store"
		record     = "record"
	)
	properties := []datastore.Property{
		{
			Name:  "Size",
			Value: size,
		},
		{
			Name:  "Status",
			Value: int64(StatusReady),
		},
		{
			Name:  "StoreKey",
			Value: store,
		},
		{
			Name:  "RecordKey",
			Value: record,
		},
	}
	expected := &BlobRef{
		Size:      123,
		Status:    StatusReady,
		StoreKey:  store,
		RecordKey: record,
	}
	actual := new(BlobRef)
	err := actual.Load(properties)
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}

func newInitBlob(t *testing.T) *BlobRef {
	const (
		size   = int64(4)
		name   = "abc"
		store  = "store"
		record = "record"
	)

	// Initialize
	blob := NewBlobRef(size, store, record)
	if blob == nil {
		t.Fatal("NewBlobRef returned nil.")
	}
	assert.NotEqual(t, uuid.Nil, blob.Key)
	assert.Equal(t, size, blob.Size)
	assert.Equal(t, StatusInitializing, blob.Status)
	assert.Equal(t, store, blob.StoreKey)
	assert.Equal(t, record, blob.RecordKey)
	assert.NotEmpty(t, blob.Timestamps.CreatedAt)
	assert.NotEmpty(t, blob.Timestamps.UpdatedAt)
	assert.NotEmpty(t, blob.Timestamps.Signature)
	return blob
}

func TestBlobRef_LifeCycle(t *testing.T) {
	blob := newInitBlob(t)

	// Mark for deletion
	assert.NoError(t, blob.MarkForDeletion())
	assert.Equal(t, StatusPendingDeletion, blob.Status)

	// Start over
	blob = newInitBlob(t)

	// Ready
	assert.NoError(t, blob.Ready())
	assert.Equal(t, StatusReady, blob.Status)

	// Invalid transitions
	assert.Error(t, blob.Ready())

	// Mark for deletion
	assert.NoError(t, blob.MarkForDeletion())
	assert.Equal(t, StatusPendingDeletion, blob.Status)

	// Invalid transitions
	assert.Error(t, blob.MarkForDeletion())
	assert.Error(t, blob.Ready())
}

func TestBlobRef_Fail(t *testing.T) {
	blob := new(BlobRef)

	// Fail should work for BlobStatusUnknown too.
	assert.NoError(t, blob.Fail())

	blob = newInitBlob(t)
	assert.NoError(t, blob.Fail())

	blob.Status = StatusInitializing
	assert.NoError(t, blob.Fail())

	blob.Status = StatusPendingDeletion
	assert.NoError(t, blob.Fail())

	blob.Status = StatusReady
	assert.NoError(t, blob.Fail())

	blob.Status = StatusError
	assert.NoError(t, blob.Fail())
}

func TestBlobRef_GetObjectPath(t *testing.T) {
	blob := NewBlobRef(0, "", "")

	assert.Equal(t, blob.Key.String(), blob.ObjectPath())
}

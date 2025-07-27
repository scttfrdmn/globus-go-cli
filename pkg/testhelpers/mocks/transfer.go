// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package mocks

import (
	"context"
	"time"
)

// These are simplified mock types to reduce dependencies

// FileEntry represents a file entry in a directory listing
type FileEntry struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	Size         int64  `json:"size"`
	LastModified string `json:"last_modified"`
	Permissions  string `json:"permissions,omitempty"`
	User         string `json:"user,omitempty"`
	Group        string `json:"group,omitempty"`
}

// ListDirectoryResponse represents a directory listing response
type ListDirectoryResponse struct {
	Path string      `json:"path"`
	Data []FileEntry `json:"data"`
}

// ListEndpointsOptions represents options for listing endpoints
type ListEndpointsOptions struct {
	Limit          int    `json:"limit,omitempty"`
	FilterOwnerID  string `json:"filter_owner_id,omitempty"`
	FilterScope    string `json:"filter_scope,omitempty"`
	FilterFullText string `json:"filter_fulltext,omitempty"`
}

// ListDirectoryOptions represents options for listing a directory
type ListDirectoryOptions struct {
	EndpointID string `json:"endpoint_id"`
	Path       string `json:"path"`
	ShowHidden bool   `json:"show_hidden,omitempty"`
}

// CreateDirectoryOptions represents options for creating a directory
type CreateDirectoryOptions struct {
	EndpointID string `json:"endpoint_id"`
	Path       string `json:"path"`
}

// Endpoint represents a transfer endpoint
type Endpoint struct {
	ID               string `json:"id"`
	DisplayName      string `json:"display_name"`
	OwnerString      string `json:"owner_string"`
	Description      string `json:"description"`
	Activated        bool   `json:"activated"`
	GCPConnected     bool   `json:"gcp_connected"`
	DefaultDirectory string `json:"default_directory"`
	Organization     string `json:"organization"`
	Department       string `json:"department"`
	ContactEmail     string `json:"contact_email"`
}

// Task represents a transfer task
type Task struct {
	TaskID                string     `json:"task_id"`
	Status                string     `json:"status"`
	Type                  string     `json:"type"`
	Label                 string     `json:"label"`
	SourceEndpointID      string     `json:"source_endpoint_id"`
	SourceEndpointDisplay string     `json:"source_endpoint_display"`
	DestinationEndpointID string     `json:"destination_endpoint_id"`
	DestEndpointDisplay   string     `json:"destination_endpoint_display"`
	RequestTime           time.Time  `json:"request_time"`
	CompletionTime        *time.Time `json:"completion_time"`
	FilesTransferred      int        `json:"files_transferred"`
	FilesSkipped          int        `json:"files_skipped"`
	BytesTransferred      int64      `json:"bytes_transferred"`
	BytesSkipped          int64      `json:"bytes_skipped"`
	Subtasks              int        `json:"subtasks"`
	SyncLevel             int        `json:"sync_level"`
	VerifyChecksum        bool       `json:"verify_checksum"`
}

// EndpointList represents a list of endpoints
type EndpointList struct {
	Data []Endpoint `json:"data"`
}

// TaskList represents a list of tasks
type TaskList struct {
	Data []Task `json:"data"`
}

// TaskResponse represents a response from submitting a task
type TaskResponse struct {
	TaskID string `json:"task_id"`
}

// DeleteItem represents an item to delete
type DeleteItem struct {
	DataType string `json:"data_type"`
	Path     string `json:"path"`
}

// DeleteTaskRequest represents a request to delete items
type DeleteTaskRequest struct {
	DataType   string       `json:"data_type"`
	EndpointID string       `json:"endpoint_id"`
	Items      []DeleteItem `json:"items"`
}

// OperationResult represents the result of an operation
type OperationResult struct {
	Code string `json:"code"`
}

// EndpointError represents an error with an endpoint
type EndpointError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Error implements the error interface
func (e *EndpointError) Error() string {
	return e.Message
}

// MockTransferClient implements a mock transfer client for testing
type MockTransferClient struct {
	// Function fields for mocking responses
	ListEndpointsFunc  func(ctx context.Context, options *ListEndpointsOptions) (*EndpointList, error)
	GetTaskFunc        func(ctx context.Context, taskID string) (*Task, error)
	SubmitTransferFunc func(ctx context.Context, sourceEndpointID, sourcePath,
		destEndpointID, destPath, label string,
		options map[string]interface{}) (*TaskResponse, error)
	ListDirectoryFunc    func(ctx context.Context, options *ListDirectoryOptions) (*ListDirectoryResponse, error)
	CreateDirectoryFunc  func(ctx context.Context, options *CreateDirectoryOptions) error
	CancelTaskFunc       func(ctx context.Context, taskID string) (*OperationResult, error)
	GetEndpointFunc      func(ctx context.Context, endpointID string) (*Endpoint, error)
	ListTasksFunc        func(ctx context.Context, options *ListTasksOptions) (*TaskList, error)
	CreateDeleteTaskFunc func(ctx context.Context, request *DeleteTaskRequest) (*TaskResponse, error)
}

// ListTasksOptions represents options for listing tasks
type ListTasksOptions struct {
	Limit        int    `json:"limit,omitempty"`
	FilterStatus string `json:"filter_status,omitempty"`
}

// ListEndpoints implements the mock transfer client interface
func (m *MockTransferClient) ListEndpoints(ctx context.Context, options *ListEndpointsOptions) (*EndpointList, error) {
	if m.ListEndpointsFunc != nil {
		return m.ListEndpointsFunc(ctx, options)
	}
	return &EndpointList{}, nil
}

// GetTask implements the mock transfer client interface
func (m *MockTransferClient) GetTask(ctx context.Context, taskID string) (*Task, error) {
	if m.GetTaskFunc != nil {
		return m.GetTaskFunc(ctx, taskID)
	}
	return &Task{}, nil
}

// SubmitTransfer implements the mock transfer client interface
func (m *MockTransferClient) SubmitTransfer(ctx context.Context, sourceEndpointID, sourcePath,
	destEndpointID, destPath, label string,
	options map[string]interface{}) (*TaskResponse, error) {
	if m.SubmitTransferFunc != nil {
		return m.SubmitTransferFunc(ctx, sourceEndpointID, sourcePath, destEndpointID, destPath, label, options)
	}
	return &TaskResponse{TaskID: "mock-task-id"}, nil
}

// ListDirectory implements the mock transfer client interface
func (m *MockTransferClient) ListDirectory(ctx context.Context, options *ListDirectoryOptions) (*ListDirectoryResponse, error) {
	if m.ListDirectoryFunc != nil {
		return m.ListDirectoryFunc(ctx, options)
	}
	return &ListDirectoryResponse{}, nil
}

// CreateDirectory implements the mock transfer client interface
func (m *MockTransferClient) CreateDirectory(ctx context.Context, options *CreateDirectoryOptions) error {
	if m.CreateDirectoryFunc != nil {
		return m.CreateDirectoryFunc(ctx, options)
	}
	return nil
}

// CancelTask implements the mock transfer client interface
func (m *MockTransferClient) CancelTask(ctx context.Context, taskID string) (*OperationResult, error) {
	if m.CancelTaskFunc != nil {
		return m.CancelTaskFunc(ctx, taskID)
	}
	return &OperationResult{Code: "Canceled"}, nil
}

// GetEndpoint implements the mock transfer client interface
func (m *MockTransferClient) GetEndpoint(ctx context.Context, endpointID string) (*Endpoint, error) {
	if m.GetEndpointFunc != nil {
		return m.GetEndpointFunc(ctx, endpointID)
	}
	return &Endpoint{ID: endpointID}, nil
}

// ListTasks implements the mock transfer client interface
func (m *MockTransferClient) ListTasks(ctx context.Context, options *ListTasksOptions) (*TaskList, error) {
	if m.ListTasksFunc != nil {
		return m.ListTasksFunc(ctx, options)
	}
	return &TaskList{}, nil
}

// CreateDeleteTask implements the mock transfer client interface
func (m *MockTransferClient) CreateDeleteTask(ctx context.Context, request *DeleteTaskRequest) (*TaskResponse, error) {
	if m.CreateDeleteTaskFunc != nil {
		return m.CreateDeleteTaskFunc(ctx, request)
	}
	return &TaskResponse{TaskID: "mock-delete-task-id"}, nil
}

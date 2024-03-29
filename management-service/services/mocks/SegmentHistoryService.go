// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import (
	models "github.com/caraml-dev/xp/management-service/models"
	pagination "github.com/caraml-dev/xp/management-service/pagination"
	mock "github.com/stretchr/testify/mock"

	services "github.com/caraml-dev/xp/management-service/services"
)

// SegmentHistoryService is an autogenerated mock type for the SegmentHistoryService type
type SegmentHistoryService struct {
	mock.Mock
}

// CreateSegmentHistory provides a mock function with given fields: _a0
func (_m *SegmentHistoryService) CreateSegmentHistory(_a0 *models.Segment) (*models.SegmentHistory, error) {
	ret := _m.Called(_a0)

	var r0 *models.SegmentHistory
	if rf, ok := ret.Get(0).(func(*models.Segment) *models.SegmentHistory); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.SegmentHistory)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*models.Segment) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeleteSegmentHistory provides a mock function with given fields: segmentId
func (_m *SegmentHistoryService) DeleteSegmentHistory(segmentId int64) error {
	ret := _m.Called(segmentId)

	var r0 error
	if rf, ok := ret.Get(0).(func(int64) error); ok {
		r0 = rf(segmentId)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetDBRecord provides a mock function with given fields: segmentId, version
func (_m *SegmentHistoryService) GetDBRecord(segmentId models.ID, version int64) (*models.SegmentHistory, error) {
	ret := _m.Called(segmentId, version)

	var r0 *models.SegmentHistory
	if rf, ok := ret.Get(0).(func(models.ID, int64) *models.SegmentHistory); ok {
		r0 = rf(segmentId, version)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.SegmentHistory)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(models.ID, int64) error); ok {
		r1 = rf(segmentId, version)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetSegmentHistory provides a mock function with given fields: segmentId, version
func (_m *SegmentHistoryService) GetSegmentHistory(segmentId int64, version int64) (*models.SegmentHistory, error) {
	ret := _m.Called(segmentId, version)

	var r0 *models.SegmentHistory
	if rf, ok := ret.Get(0).(func(int64, int64) *models.SegmentHistory); ok {
		r0 = rf(segmentId, version)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.SegmentHistory)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(int64, int64) error); ok {
		r1 = rf(segmentId, version)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListSegmentHistory provides a mock function with given fields: segmentId, params
func (_m *SegmentHistoryService) ListSegmentHistory(segmentId int64, params services.ListSegmentHistoryParams) ([]*models.SegmentHistory, *pagination.Paging, error) {
	ret := _m.Called(segmentId, params)

	var r0 []*models.SegmentHistory
	if rf, ok := ret.Get(0).(func(int64, services.ListSegmentHistoryParams) []*models.SegmentHistory); ok {
		r0 = rf(segmentId, params)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*models.SegmentHistory)
		}
	}

	var r1 *pagination.Paging
	if rf, ok := ret.Get(1).(func(int64, services.ListSegmentHistoryParams) *pagination.Paging); ok {
		r1 = rf(segmentId, params)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*pagination.Paging)
		}
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(int64, services.ListSegmentHistoryParams) error); ok {
		r2 = rf(segmentId, params)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

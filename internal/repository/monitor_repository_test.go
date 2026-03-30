package repository

import "testing"

func TestMongoMonitorRepositoryImplementsMonitorRepository(t *testing.T) {
	var _ MonitorRepository = (*MongoMonitorRepository)(nil)
}

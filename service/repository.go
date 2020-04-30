package service

type Repository interface {
	StorageInfo(service, destination string) (path, types string, err error)
}

type mockRepository struct {
}

func (mr *mockRepository) StorageInfo(service, destination string) (path, types string, err error) {
	return service, "image/jpeg, image/jpg", nil
}

func NewMockRepository() Repository {
	return &mockRepository{}
}

//var engine, err = xorm.NewEngine(driverName, dataSourceName)

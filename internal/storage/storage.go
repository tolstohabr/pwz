package storage

import (
	"encoding/json"
	"os"

	"PWZ1.0/internal/models"
	"PWZ1.0/internal/models/domainErrors"
)

type Storage interface {
	SaveOrder(order models.Order) error
	GetOrder(id uint64) (models.Order, error)
	DeleteOrder(id uint64) error
	ListOrders() ([]models.Order, error)
	UpdateOrder(order models.Order) error
}

type FileStorage struct {
	filePath string
}

func NewFileStorage(path string) *FileStorage {
	return &FileStorage{filePath: path}
}

func (fs *FileStorage) load() ([]models.Order, error) {
	file, err := os.Open(fs.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []models.Order{}, nil
		}
		return nil, err
	}
	defer file.Close()

	var orders []models.Order
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&orders)
	if err != nil {
		return nil, err
	}
	return orders, nil
}

func (fs *FileStorage) save(orders []models.Order) error {
	file, err := os.Create(fs.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(orders)
}

func (fs *FileStorage) SaveOrder(order models.Order) error {
	orders, err := fs.load()
	if err != nil {
		return err
	}

	for _, o := range orders {
		if o.ID == order.ID {
			return domainErrors.ErrDuplicateOrder
		}
	}

	orders = append(orders, order)
	return fs.save(orders)
}

func (fs *FileStorage) UpdateOrder(order models.Order) error {
	orders, err := fs.load()
	if err != nil {
		return err
	}
	for i, ord := range orders {
		if ord.ID == order.ID {
			orders[i] = order
			break
		}
	}
	return fs.save(orders)
}

func (fs *FileStorage) GetOrder(id uint64) (models.Order, error) {
	orders, err := fs.load()
	if err != nil {
		return models.Order{}, err
	}

	for _, o := range orders {
		if o.ID == id {
			return o, nil
		}
	}

	return models.Order{}, domainErrors.ErrOrderNotFound
}

func (fs *FileStorage) DeleteOrder(id uint64) error {
	orders, err := fs.load()
	if err != nil {
		return err
	}

	updated := []models.Order{}
	found := false

	for _, o := range orders {
		if o.ID == id {
			found = true
			continue
		}
		updated = append(updated, o)
	}

	if !found {
		return domainErrors.ErrOrderNotFound
	}

	return fs.save(updated)
}

func (fs *FileStorage) ListOrders() ([]models.Order, error) {
	return fs.load()
}

package service

import (
	"bytes"
	"fmt"
	"github.com/google/uuid"
	"os"
	"sync"
)

// ImageStore is a interface to store laptop images
type ImageStore interface {
	Save(laptopID string, imageType string, imageData bytes.Buffer) (string, error)
}

// DiskImageStore stores images on disk and its info on memory
type DiskImageStore struct {
	mutex       sync.RWMutex
	imageFolder string
	image       map[string]*ImageInfo
}

// ImageInfo contains information of the laptop image
type ImageInfo struct {
	LaptopID string
	Type     string
	Path     string
}

// NewDiskImageStore returns a new DiskImageStore
func NewDiskImageStore(imageFolder string) *DiskImageStore {

	return &DiskImageStore{
		imageFolder: imageFolder,
		image:       make(map[string]*ImageInfo),
	}
}

func (store *DiskImageStore) Save(laptopID string, imageType string, imageData bytes.Buffer) (string, error) {
	// Random generate image ID
	imageID, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("cannot generate image id: %w", err)
	}

	// Create image path
	imagePath := fmt.Sprintf("%s/%s%s", store.imageFolder, imageID, imageType)

	// Create image path and Write data into file
	file, err := os.Create(imagePath)
	if err != nil {
		return "", fmt.Errorf("cannot create image file: %w", err)
	}
	_, err = imageData.WriteTo(file)
	if err != nil {
		return "", fmt.Errorf("cannot write image to file: %w", err)
	}

	store.mutex.RLock()
	defer store.mutex.RUnlock()
	store.image[imageID.String()] = &ImageInfo{
		LaptopID: laptopID,
		Type:     imageType,
		Path:     imagePath,
	}

	return imageID.String(), nil
}

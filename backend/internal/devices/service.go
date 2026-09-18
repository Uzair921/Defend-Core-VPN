package devices

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

type Service struct {
	repo         *Repository
	serverPubKey string
}

func NewService(repo *Repository, serverPubKey string) *Service {
	return &Service{
		repo:         repo,
		serverPubKey: serverPubKey,
	}
}

type DeviceConfig struct {
	DeviceID     string `json:"device_id"`
	DeviceName   string `json:"device_name"`
	ServerHost   string `json:"server_host"`
	ServerPort   int    `json:"server_port"`
	AssignedIP   string `json:"assigned_ip"`
	PublicKey    string `json:"public_key"`
	PrivateKey   string `json:"private_key"`
	ServerPubKey string `json:"server_public_key"`
	DNS          string `json:"dns"`
}

func (s *Service) Create(ctx context.Context, userID uuid.UUID, req CreateRequest, serverHost string, serverPort int) (*CreateResponse, error) {
	// Generate keypair
	kp, err := GenerateKeyPair()
	if err != nil {
		return nil, err
	}

	privateKeyB64 := EncodeKey(kp.PrivateKey)
	publicKeyB64 := EncodeKey(kp.PublicKey)

	// Store device with public key
	device, err := s.repo.Create(ctx, userID, req.Name, publicKeyB64, req.Platform)
	if err != nil {
		return nil, err
	}

	// Build config
	cfg := DeviceConfig{
		DeviceID:     device.ID.String(),
		DeviceName:   device.Name,
		ServerHost:   serverHost,
		ServerPort:   serverPort,
		AssignedIP:   "10.8.0.2", // TODO: allocate dynamically
		PublicKey:    publicKeyB64,
		PrivateKey:   privateKeyB64,
		ServerPubKey: s.serverPubKey,
		DNS:          "1.1.1.1",
	}

	cfgJSON, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal config: %w", err)
	}

	return &CreateResponse{
		Device:     device,
		PrivateKey: privateKeyB64,
		PublicKey:  publicKeyB64,
		Config:     string(cfgJSON),
	}, nil
}

func (s *Service) List(ctx context.Context, userID uuid.UUID) ([]*Device, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *Service) Get(ctx context.Context, id, userID uuid.UUID) (*Device, error) {
	return s.repo.GetByID(ctx, id, userID)
}

func (s *Service) Delete(ctx context.Context, id, userID uuid.UUID) error {
	return s.repo.Delete(ctx, id, userID)
}

package instance_service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"terraform-provider-leaseweb/internal/core/domain/entity"
	"terraform-provider-leaseweb/internal/core/ports"
)

type Service struct {
	instanceRepository ports.InstanceRepository
}

func (srv Service) GetAllInstances(ctx context.Context) (
	entity.Instances,
	error,
) {
	instances, err := srv.instanceRepository.GetAllInstances(ctx)
	if err != nil {
		return entity.Instances{}, fmt.Errorf(
			"failed to retrieve instances from repository: %w",
			err,
		)
	}

	return instances, nil
}

func (srv Service) GetInstance(
	id uuid.UUID,
	ctx context.Context,
) (*entity.Instance, error) {
	instance, err := srv.instanceRepository.GetInstance(id, ctx)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to retrieve instance %q from repository of: %w",
			id,
			err,
		)
	}

	return instance, nil
}

func (srv Service) CreateInstance(
	instance entity.Instance,
	ctx context.Context,
) (*entity.Instance, error) {
	createdInstance, err := srv.instanceRepository.CreateInstance(instance, ctx)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create instance %q in repository: %w",
			instance.Id,
			err,
		)
	}

	return createdInstance, nil
}

func (srv Service) UpdateInstance(
	instance entity.Instance,
	ctx context.Context,
) (*entity.Instance, error) {
	updatedInstance, err := srv.instanceRepository.UpdateInstance(instance, ctx)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to update instance %q in repository: %w",
			instance.Id,
			err,
		)
	}

	return updatedInstance, nil
}

func (srv Service) DeleteInstance(id uuid.UUID, ctx context.Context) error {
	err := srv.instanceRepository.DeleteInstance(id, ctx)
	if err != nil {
		return fmt.Errorf(
			"failed to delete instance %q in repository: %w",
			id,
			err,
		)
	}

	return nil
}

func New(instanceRepository ports.InstanceRepository) Service {
	return Service{instanceRepository: instanceRepository}
}

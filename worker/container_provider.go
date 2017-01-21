package worker

import (
	"errors"
	"fmt"
	"os"
	"time"

	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/garden"
	"code.cloudfoundry.org/lager"
	"github.com/concourse/atc"
	"github.com/concourse/atc/db"
	"github.com/concourse/atc/dbng"
	"github.com/concourse/baggageclaim"
)

var ErrCreatedContainerNotFound = errors.New("container-in-created-state-not-found-in-garden")

const creatingContainerRetryDelay = 1 * time.Second

//go:generate counterfeiter . ContainerProviderFactory

type ContainerProviderFactory interface {
	ContainerProviderFor(Worker) ContainerProvider
}

type containerProviderFactory struct {
	gardenClient            garden.Client
	baggageclaimClient      baggageclaim.Client
	volumeClient            VolumeClient
	imageFactory            ImageFactory
	dbVolumeFactory         dbng.VolumeFactory
	dbResourceCacheFactory  dbng.ResourceCacheFactory
	dbResourceConfigFactory dbng.ResourceConfigFactory
	dbTeamFactory           dbng.TeamFactory

	db GardenWorkerDB

	httpProxyURL  string
	httpsProxyURL string
	noProxy       string

	clock clock.Clock
}

func NewContainerProviderFactory(
	gardenClient garden.Client,
	baggageclaimClient baggageclaim.Client,
	volumeClient VolumeClient,
	imageFactory ImageFactory,
	dbVolumeFactory dbng.VolumeFactory,
	dbResourceCacheFactory dbng.ResourceCacheFactory,
	dbResourceConfigFactory dbng.ResourceConfigFactory,
	dbTeamFactory dbng.TeamFactory,
	db GardenWorkerDB,
	httpProxyURL string,
	httpsProxyURL string,
	noProxy string,
	clock clock.Clock,
) ContainerProviderFactory {
	return &containerProviderFactory{
		gardenClient:            gardenClient,
		baggageclaimClient:      baggageclaimClient,
		volumeClient:            volumeClient,
		imageFactory:            imageFactory,
		dbVolumeFactory:         dbVolumeFactory,
		dbResourceCacheFactory:  dbResourceCacheFactory,
		dbResourceConfigFactory: dbResourceConfigFactory,
		dbTeamFactory:           dbTeamFactory,
		db:                      db,
		httpProxyURL:            httpProxyURL,
		httpsProxyURL:           httpsProxyURL,
		noProxy:                 noProxy,
		clock:                   clock,
	}
}

func (f *containerProviderFactory) ContainerProviderFor(
	worker Worker,
) ContainerProvider {
	return &containerProvider{
		gardenClient:            f.gardenClient,
		baggageclaimClient:      f.baggageclaimClient,
		volumeClient:            f.volumeClient,
		imageFactory:            f.imageFactory,
		dbVolumeFactory:         f.dbVolumeFactory,
		dbResourceCacheFactory:  f.dbResourceCacheFactory,
		dbResourceConfigFactory: f.dbResourceConfigFactory,
		dbTeamFactory:           f.dbTeamFactory,
		db:                      f.db,
		httpProxyURL:            f.httpProxyURL,
		httpsProxyURL:           f.httpsProxyURL,
		noProxy:                 f.noProxy,
		clock:                   f.clock,
		worker:                  worker,
	}
}

//go:generate counterfeiter . ContainerProvider

type ContainerProvider interface {
	FindContainerByHandle(
		logger lager.Logger,
		handle string,
		teamID int,
	) (Container, bool, error)

	FindOrCreateBuildContainer(
		logger lager.Logger,
		cancel <-chan os.Signal,
		delegate ImageFetchingDelegate,
		id Identifier,
		metadata Metadata,
		spec ContainerSpec,
		resourceTypes atc.ResourceTypes,
		outputPaths map[string]string,
	) (Container, error)

	FindOrCreateResourceCheckContainer(
		logger lager.Logger,
		cancel <-chan os.Signal,
		delegate ImageFetchingDelegate,
		id Identifier,
		metadata Metadata,
		spec ContainerSpec,
		resourceTypes atc.ResourceTypes,
		resourceType string,
		source atc.Source,
	) (Container, error)

	FindOrCreateResourceTypeCheckContainer(
		logger lager.Logger,
		cancel <-chan os.Signal,
		delegate ImageFetchingDelegate,
		id Identifier,
		metadata Metadata,
		spec ContainerSpec,
		resourceTypes atc.ResourceTypes,
		resourceTypeName string,
		source atc.Source,
	) (Container, error)

	FindOrCreateResourceGetContainer(
		logger lager.Logger,
		cancel <-chan os.Signal,
		delegate ImageFetchingDelegate,
		id Identifier,
		metadata Metadata,
		spec ContainerSpec,
		resourceTypes atc.ResourceTypes,
		outputPaths map[string]string,
		resourceTypeName string,
		version atc.Version,
		source atc.Source,
		params atc.Params,
	) (Container, error)
}

type containerProvider struct {
	gardenClient            garden.Client
	baggageclaimClient      baggageclaim.Client
	volumeClient            VolumeClient
	imageFactory            ImageFactory
	dbVolumeFactory         dbng.VolumeFactory
	dbResourceCacheFactory  dbng.ResourceCacheFactory
	dbResourceConfigFactory dbng.ResourceConfigFactory
	dbTeamFactory           dbng.TeamFactory

	db       GardenWorkerDB
	provider WorkerProvider

	worker        Worker
	httpProxyURL  string
	httpsProxyURL string
	noProxy       string

	clock clock.Clock
}

func (p *containerProvider) FindOrCreateBuildContainer(
	logger lager.Logger,
	cancel <-chan os.Signal,
	delegate ImageFetchingDelegate,
	id Identifier,
	metadata Metadata,
	spec ContainerSpec,
	resourceTypes atc.ResourceTypes,
	outputPaths map[string]string,
) (Container, error) {
	return p.findOrCreateContainer(
		logger,
		cancel,
		delegate,
		id,
		metadata,
		spec,
		resourceTypes,
		outputPaths,
		func() (dbng.CreatingContainer, dbng.CreatedContainer, error) {
			return p.dbTeamFactory.GetByID(spec.TeamID).FindBuildContainer(
				&dbng.Worker{
					Name:       p.worker.Name(),
					GardenAddr: p.worker.Address(),
				},
				id.BuildID,
				id.PlanID,
				dbng.ContainerMetadata{
					Name: metadata.StepName,
					Type: string(metadata.Type),
				},
			)
		},
		func() (dbng.CreatingContainer, error) {
			return p.dbTeamFactory.GetByID(spec.TeamID).CreateBuildContainer(
				&dbng.Worker{
					Name:       p.worker.Name(),
					GardenAddr: p.worker.Address(),
				},
				id.BuildID,
				id.PlanID,
				dbng.ContainerMetadata{
					Name: metadata.StepName,
					Type: string(metadata.Type),
				},
			)
		},
	)
}

func (p *containerProvider) FindOrCreateResourceCheckContainer(
	logger lager.Logger,
	cancel <-chan os.Signal,
	delegate ImageFetchingDelegate,
	id Identifier,
	metadata Metadata,
	spec ContainerSpec,
	resourceTypes atc.ResourceTypes,
	resourceType string,
	source atc.Source,
) (Container, error) {
	resourceConfig, err := p.dbResourceConfigFactory.FindOrCreateResourceConfigForResource(
		logger,
		id.ResourceID,
		resourceType,
		source,
		metadata.PipelineID,
		resourceTypes,
	)
	if err != nil {
		logger.Error("failed-to-get-resource-config", err)
		return nil, err
	}

	return p.findOrCreateContainer(
		logger,
		cancel,
		delegate,
		id,
		metadata,
		spec,
		resourceTypes,
		map[string]string{},
		func() (dbng.CreatingContainer, dbng.CreatedContainer, error) {
			return p.dbTeamFactory.GetByID(spec.TeamID).FindResourceCheckContainer(
				&dbng.Worker{
					Name:       p.worker.Name(),
					GardenAddr: p.worker.Address(),
				},
				resourceConfig,
			)
		},
		func() (dbng.CreatingContainer, error) {
			return p.dbTeamFactory.GetByID(spec.TeamID).CreateResourceCheckContainer(
				&dbng.Worker{
					Name:       p.worker.Name(),
					GardenAddr: p.worker.Address(),
				},
				resourceConfig,
			)
		},
	)
}

func (p *containerProvider) FindOrCreateResourceTypeCheckContainer(
	logger lager.Logger,
	cancel <-chan os.Signal,
	delegate ImageFetchingDelegate,
	id Identifier,
	metadata Metadata,
	spec ContainerSpec,
	resourceTypes atc.ResourceTypes,
	resourceTypeName string,
	source atc.Source,
) (Container, error) {
	resourceConfig, err := p.dbResourceConfigFactory.FindOrCreateResourceConfigForResourceType(
		logger,
		resourceTypeName,
		source,
		metadata.PipelineID,
		resourceTypes,
	)
	if err != nil {
		return nil, err
	}

	return p.findOrCreateContainer(
		logger,
		cancel,
		delegate,
		id,
		metadata,
		spec,
		resourceTypes,
		map[string]string{},
		func() (dbng.CreatingContainer, dbng.CreatedContainer, error) {
			return p.dbTeamFactory.GetByID(spec.TeamID).FindResourceCheckContainer(
				&dbng.Worker{
					Name:       p.worker.Name(),
					GardenAddr: p.worker.Address(),
				},
				resourceConfig,
			)
		},
		func() (dbng.CreatingContainer, error) {
			return p.dbTeamFactory.GetByID(spec.TeamID).CreateResourceCheckContainer(
				&dbng.Worker{
					Name:       p.worker.Name(),
					GardenAddr: p.worker.Address(),
				},
				resourceConfig,
			)
		},
	)
}

func (p *containerProvider) FindOrCreateResourceGetContainer(
	logger lager.Logger,
	cancel <-chan os.Signal,
	delegate ImageFetchingDelegate,
	id Identifier,
	metadata Metadata,
	spec ContainerSpec,
	resourceTypes atc.ResourceTypes,
	outputPaths map[string]string,
	resourceTypeName string,
	version atc.Version,
	source atc.Source,
	params atc.Params,
) (Container, error) {
	var resourceCache *dbng.UsedResourceCache

	if id.BuildID != 0 {
		var err error
		resourceCache, err = p.dbResourceCacheFactory.FindOrCreateResourceCacheForBuild(
			logger,
			id.BuildID,
			resourceTypeName,
			version,
			source,
			params,
			metadata.PipelineID,
			resourceTypes,
		)
		if err != nil {
			logger.Error("failed-to-get-resource-cache-for-build", err, lager.Data{"build-id": id.BuildID})
			return nil, err
		}
	} else if id.ResourceID != 0 {
		var err error
		resourceCache, err = p.dbResourceCacheFactory.FindOrCreateResourceCacheForResource(
			logger,
			id.ResourceID,
			resourceTypeName,
			version,
			source,
			params,
			metadata.PipelineID,
			resourceTypes,
		)
		if err != nil {
			logger.Error("failed-to-get-resource-cache-for-resource", err, lager.Data{"resource-id": id.ResourceID})
			return nil, err
		}
	} else {
		var err error
		resourceCache, err = p.dbResourceCacheFactory.FindOrCreateResourceCacheForResourceType(
			logger,
			resourceTypeName,
			version,
			source,
			params,
			metadata.PipelineID,
			resourceTypes,
		)
		if err != nil {
			logger.Error("failed-to-get-resource-cache-for-resource-type", err, lager.Data{"resource-type": resourceTypeName})
			return nil, err
		}
	}

	return p.findOrCreateContainer(
		logger,
		cancel,
		delegate,
		id,
		metadata,
		spec,
		resourceTypes,
		map[string]string{},
		func() (dbng.CreatingContainer, dbng.CreatedContainer, error) {
			return p.dbTeamFactory.GetByID(spec.TeamID).FindResourceGetContainer(
				&dbng.Worker{
					Name:       p.worker.Name(),
					GardenAddr: p.worker.Address(),
				},
				resourceCache,
				metadata.StepName,
			)
		},
		func() (dbng.CreatingContainer, error) {
			return p.dbTeamFactory.GetByID(spec.TeamID).CreateResourceGetContainer(
				&dbng.Worker{
					Name:       p.worker.Name(),
					GardenAddr: p.worker.Address(),
				},
				resourceCache,
				metadata.StepName,
			)
		},
	)
}

func (p *containerProvider) FindContainerByHandle(
	logger lager.Logger,
	handle string,
	teamID int,
) (Container, bool, error) {
	gardenContainer, err := p.gardenClient.Lookup(handle)
	if err != nil {
		if _, ok := err.(garden.ContainerNotFoundError); ok {
			logger.Info("container-not-found")
			return nil, false, nil
		}

		logger.Error("failed-to-lookup-on-garden", err)
		return nil, false, err
	}

	createdContainer, found, err := p.dbTeamFactory.GetByID(teamID).FindContainerByHandle(handle)
	if err != nil {
		logger.Error("failed-to-lookup-in-db", err)
		return nil, false, err
	}

	if !found {
		return nil, false, nil
	}

	createdVolumes, err := p.dbVolumeFactory.FindVolumesForContainer(createdContainer)
	if err != nil {
		return nil, false, err
	}

	container, err := newGardenWorkerContainer(
		logger,
		gardenContainer,
		createdContainer,
		createdVolumes,
		p.gardenClient,
		p.baggageclaimClient,
		p.db,
		p.worker.Name(),
	)

	if err != nil {
		logger.Error("failed-to-construct-container", err)
		return nil, false, err
	}

	return container, true, nil
}

func (p *containerProvider) findOrCreateContainer(
	logger lager.Logger,
	cancel <-chan os.Signal,
	delegate ImageFetchingDelegate,
	id Identifier,
	metadata Metadata,
	spec ContainerSpec,
	resourceTypes atc.ResourceTypes,
	outputPaths map[string]string,
	findContainerFunc func() (dbng.CreatingContainer, dbng.CreatedContainer, error),
	createContainerFunc func() (dbng.CreatingContainer, error),
) (Container, error) {
	var gardenContainer garden.Container

	creatingContainer, createdContainer, err := findContainerFunc()
	if err != nil {
		logger.Error("failed-to-find-container-in-db", err)
		return nil, err
	}

	if createdContainer != nil {
		gardenContainer, err = p.gardenClient.Lookup(createdContainer.Handle())
		if err != nil {
			logger.Error("failed-to-lookup-created-container-in-garden", err)
			return nil, err
		}

		createdVolumes, err := p.dbVolumeFactory.FindVolumesForContainer(createdContainer)
		if err != nil {
			logger.Error("failed-to-find-container-volumes", err)
			return nil, err
		}

		return newGardenWorkerContainer(
			logger,
			gardenContainer,
			createdContainer,
			createdVolumes,
			p.gardenClient,
			p.baggageclaimClient,
			p.db,
			p.worker.Name(),
		)
	} else {
		if creatingContainer != nil {
			gardenContainer, err = p.gardenClient.Lookup(creatingContainer.Handle())
			if err != nil {
				if _, ok := err.(garden.ContainerNotFoundError); !ok {
					logger.Error("failed-to-lookup-creating-container-in-garden", err)
					return nil, err
				}
			}
		}

		if gardenContainer == nil {
			image, err := p.imageFactory.GetImage(
				logger,
				p.worker,
				p.volumeClient,
				spec.ImageSpec,
				spec.TeamID,
				cancel,
				delegate,
				id,
				metadata,
				resourceTypes,
			)
			if err != nil {
				return nil, err
			}

			if creatingContainer == nil {
				creatingContainer, err = createContainerFunc()
				if err != nil {
					logger.Error("failed-to-create-container-in-db", err)
					return nil, err
				}
			}

			lock, acquired, err := p.db.AcquireContainerCreatingLock(logger, creatingContainer.ID())
			if err != nil {
				logger.Error("failed-to-acquire-volume-creating-lock", err)
				return nil, err
			}

			if !acquired {
				p.clock.Sleep(creatingContainerRetryDelay)
				return p.findOrCreateContainer(
					logger,
					cancel,
					delegate,
					id,
					metadata,
					spec,
					resourceTypes,
					outputPaths,
					findContainerFunc,
					createContainerFunc,
				)
			}

			defer lock.Release()

			fetchedImage, err := image.FetchForContainer(logger, creatingContainer)
			if err != nil {
				logger.Error("failed-to-fetch-image-for-container", err)
				return nil, err
			}

			gardenContainer, err = p.createGardenContainer(
				logger,
				creatingContainer,
				id,
				metadata,
				spec,
				outputPaths,
				fetchedImage.Metadata,
				fetchedImage.URL,
			)
			if err != nil {
				logger.Error("failed-to-create-container-in-garden", err)
				return nil, err
			}

			metadata.WorkerName = p.worker.Name()
			metadata.Handle = gardenContainer.Handle()

			metadata.User = fetchedImage.Metadata.User
			if spec.User != "" {
				metadata.User = spec.User
			}

			id.ResourceTypeVersion = fetchedImage.Version

			_, err = p.db.UpdateContainerTTLToBeRemoved(
				db.Container{
					ContainerIdentifier: db.ContainerIdentifier(id),
					ContainerMetadata:   db.ContainerMetadata(metadata),
				},
				p.maxContainerLifetime(metadata),
			)
			if err != nil {
				logger.Error("failed-to-update-container-ttl", err)
				return nil, err
			}
		}

		createdContainer, err = creatingContainer.Created()
		if err != nil {
			logger.Error("failed-to-mark-container-as-created", err)
			return nil, err
		}
	}

	createdVolumes, err := p.dbVolumeFactory.FindVolumesForContainer(createdContainer)
	if err != nil {
		logger.Error("failed-to-find-container-volumes", err)
		return nil, err
	}

	return newGardenWorkerContainer(
		logger,
		gardenContainer,
		createdContainer,
		createdVolumes,
		p.gardenClient,
		p.baggageclaimClient,
		p.db,
		p.worker.Name(),
	)
}

func (p *containerProvider) createGardenContainer(
	logger lager.Logger,
	creatingContainer dbng.CreatingContainer,
	id Identifier,
	metadata Metadata,
	spec ContainerSpec,
	outputPaths map[string]string,
	imageMetadata ImageMetadata,
	imageURL string,
) (garden.Container, error) {
	volumeMounts := []VolumeMount{}
	for name, outputPath := range outputPaths {
		outVolume, volumeErr := p.volumeClient.FindOrCreateVolumeForContainer(
			logger,
			VolumeSpec{
				Strategy:   OutputStrategy{Name: name},
				Privileged: bool(spec.ImageSpec.Privileged),
			},
			creatingContainer,
			spec.TeamID,
			outputPath,
		)
		if volumeErr != nil {
			return nil, volumeErr
		}

		volumeMounts = append(volumeMounts, VolumeMount{
			Volume:    outVolume,
			MountPath: outputPath,
		})
	}

	for _, mount := range spec.Mounts {
		volumeMounts = append(volumeMounts, mount)
	}

	for _, mount := range spec.Inputs {
		cowVolume, volumeErr := p.volumeClient.FindOrCreateVolumeForContainer(
			logger,
			VolumeSpec{
				Strategy: ContainerRootFSStrategy{
					Parent: mount.Volume,
				},
				Privileged: spec.ImageSpec.Privileged,
			},
			creatingContainer,
			spec.TeamID,
			mount.MountPath,
		)
		if volumeErr != nil {
			return nil, volumeErr
		}

		volumeMounts = append(volumeMounts, VolumeMount{
			Volume:    cowVolume,
			MountPath: mount.MountPath,
		})
	}

	bindMounts := []garden.BindMount{}

	volumeHandleMounts := map[string]string{}
	for _, mount := range volumeMounts {
		bindMounts = append(bindMounts, garden.BindMount{
			SrcPath: mount.Volume.Path(),
			DstPath: mount.MountPath,
			Mode:    garden.BindMountModeRW,
		})
		volumeHandleMounts[mount.Volume.Handle()] = mount.MountPath
	}

	gardenProperties := garden.Properties{userPropertyName: imageMetadata.User}
	if spec.User != "" {
		gardenProperties = garden.Properties{userPropertyName: spec.User}
	}

	if spec.Ephemeral {
		gardenProperties[ephemeralPropertyName] = "true"
	}

	env := append(imageMetadata.Env, spec.Env...)

	if p.httpProxyURL != "" {
		env = append(env, fmt.Sprintf("http_proxy=%s", p.httpProxyURL))
	}

	if p.httpsProxyURL != "" {
		env = append(env, fmt.Sprintf("https_proxy=%s", p.httpsProxyURL))
	}

	if p.noProxy != "" {
		env = append(env, fmt.Sprintf("no_proxy=%s", p.noProxy))
	}

	gardenSpec := garden.ContainerSpec{
		BindMounts: bindMounts,
		Privileged: spec.ImageSpec.Privileged,
		Properties: gardenProperties,
		RootFSPath: imageURL,
		Env:        env,
		Handle:     creatingContainer.Handle(),
	}

	return p.gardenClient.Create(gardenSpec)
}

func (p *containerProvider) maxContainerLifetime(metadata Metadata) time.Duration {
	if metadata.Type == db.ContainerTypeCheck {
		uptime := p.worker.Uptime()
		switch {
		case uptime < 5*time.Minute:
			return 5 * time.Minute
		case uptime > 1*time.Hour:
			return 1 * time.Hour
		default:
			return uptime
		}
	}

	return time.Duration(0)
}

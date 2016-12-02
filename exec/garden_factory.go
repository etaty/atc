package exec

import (
	"crypto/sha1"
	"fmt"
	"path/filepath"
	"time"

	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/lager"

	"github.com/concourse/atc"
	"github.com/concourse/atc/dbng"
	"github.com/concourse/atc/resource"
	"github.com/concourse/atc/worker"
)

type gardenFactory struct {
	workerClient            worker.Client
	resourceFetcher         resource.Fetcher
	resourceFactory         resource.ResourceFactory
	resourceInstanceFactory resource.ResourceInstanceFactory
}

func NewGardenFactory(
	workerClient worker.Client,
	resourceFetcher resource.Fetcher,
	resourceFactory resource.ResourceFactory,
	resourceInstanceFactory resource.ResourceInstanceFactory,
) Factory {
	return &gardenFactory{
		workerClient:            workerClient,
		resourceFetcher:         resourceFetcher,
		resourceFactory:         resourceFactory,
		resourceInstanceFactory: resourceInstanceFactory,
	}
}

func (factory *gardenFactory) DependentGet(
	logger lager.Logger,
	stepMetadata StepMetadata,
	sourceName worker.ArtifactName,
	id worker.Identifier,
	workerMetadata worker.Metadata,
	delegate GetDelegate,
	resourceConfig atc.ResourceConfig,
	tags atc.Tags,
	teamID int,
	params atc.Params,
	resourceTypes atc.ResourceTypes,
	containerSuccessTTL time.Duration,
	containerFailureTTL time.Duration,
) StepFactory {
	return newDependentGetStep(
		logger,
		sourceName,
		resourceConfig,
		params,
		stepMetadata,
		resource.Session{
			ID:        id,
			Ephemeral: false,
			Metadata:  workerMetadata,
		},
		tags,
		teamID,
		delegate,
		factory.resourceFetcher,
		resourceTypes,
		containerSuccessTTL,
		containerFailureTTL,
		factory.resourceInstanceFactory,
	)
}

func (factory *gardenFactory) Get(
	logger lager.Logger,
	stepMetadata StepMetadata,
	sourceName worker.ArtifactName,
	id worker.Identifier,
	workerMetadata worker.Metadata,
	delegate GetDelegate,
	resourceConfig atc.ResourceConfig,
	tags atc.Tags,
	teamID int,
	params atc.Params,
	version atc.Version,
	resourceTypes atc.ResourceTypes,
	containerSuccessTTL time.Duration,
	containerFailureTTL time.Duration,
) StepFactory {
	workerMetadata.WorkingDirectory = resource.ResourcesDir("get")
	return newGetStep(
		logger,
		sourceName,
		resourceConfig,
		version,
		params,
		factory.resourceInstanceFactory.NewBuildResourceInstance(
			resource.ResourceType(resourceConfig.Type),
			version,
			resourceConfig.Source,
			params,
			&dbng.Build{ID: id.BuildID},
			&dbng.Pipeline{ID: workerMetadata.PipelineID},
			resourceTypes,
		),
		stepMetadata,
		resource.Session{
			ID:        id,
			Metadata:  workerMetadata,
			Ephemeral: false,
		},
		tags,
		teamID,
		delegate,
		factory.resourceFetcher,
		resourceTypes,

		containerSuccessTTL,
		containerFailureTTL,
	)
}

func (factory *gardenFactory) Put(
	logger lager.Logger,
	stepMetadata StepMetadata,
	id worker.Identifier,
	workerMetadata worker.Metadata,
	delegate PutDelegate,
	resourceConfig atc.ResourceConfig,
	tags atc.Tags,
	teamID int,
	params atc.Params,
	resourceTypes atc.ResourceTypes,
	containerSuccessTTL time.Duration,
	containerFailureTTL time.Duration,
) StepFactory {
	workerMetadata.WorkingDirectory = resource.ResourcesDir("put")
	return newPutStep(
		logger,
		resourceConfig,
		params,
		stepMetadata,
		resource.Session{
			ID:        id,
			Ephemeral: false,
			Metadata:  workerMetadata,
		},
		tags,
		teamID,
		delegate,
		factory.resourceFactory,
		resourceTypes,
		containerSuccessTTL,
		containerFailureTTL,
	)
}

func (factory *gardenFactory) Task(
	logger lager.Logger,
	sourceName worker.ArtifactName,
	id worker.Identifier,
	workerMetadata worker.Metadata,
	delegate TaskDelegate,
	privileged Privileged,
	tags atc.Tags,
	teamID int,
	configSource TaskConfigSource,
	resourceTypes atc.ResourceTypes,
	inputMapping map[string]string,
	outputMapping map[string]string,
	imageArtifactName string,
	clock clock.Clock,
	containerSuccessTTL time.Duration,
	containerFailureTTL time.Duration,
) StepFactory {
	workingDirectory := factory.taskWorkingDirectory(sourceName)
	workerMetadata.WorkingDirectory = workingDirectory
	return newTaskStep(
		logger,
		id,
		workerMetadata,
		tags,
		teamID,
		delegate,
		privileged,
		configSource,
		factory.workerClient,
		workingDirectory,
		resourceTypes,
		inputMapping,
		outputMapping,
		imageArtifactName,
		clock,
		containerSuccessTTL,
		containerFailureTTL,
	)
}

func (factory *gardenFactory) taskWorkingDirectory(sourceName worker.ArtifactName) string {
	sum := sha1.Sum([]byte(sourceName))
	return filepath.Join("/tmp", "build", fmt.Sprintf("%x", sum[:4]))
}

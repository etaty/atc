package dbng

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"code.cloudfoundry.org/lager"

	sq "github.com/Masterminds/squirrel"
	"github.com/concourse/atc"
	"github.com/concourse/atc/db/lock"
	"github.com/lib/pq"
)

var ErrResourceCacheAlreadyExists = errors.New("resource-cache-already-exists")
var ErrResourceCacheDisappeared = errors.New("resource-cache-disappeared")

// ResourceCache represents an instance of a ResourceConfig's version.
//
// A ResourceCache is created by a `get`, an `image_resource`, or a resource
// type in a pipeline.
//
// ResourceCaches are garbage-collected by gc.ResourceCacheCollector.
type ResourceCache struct {
	ResourceConfig ResourceConfig // The resource configuration.
	Version        atc.Version    // The version of the resource.
	Params         atc.Params     // The params used when fetching the version.
}

// UsedResourceCache is created whenever a ResourceCache is Created and/or
// Used.
//
// So long as the UsedResourceCache exists, the underlying ResourceCache can
// not be removed.
//
// UsedResourceCaches become unused by the gc.ResourceCacheCollector, which may
// then lead to the ResourceCache being garbage-collected.
//
// See FindOrCreateForBuild, FindOrCreateForResource, and
// FindOrCreateForResourceType for more information on when it becomes unused.
type UsedResourceCache struct {
	ID             int
	ResourceConfig *UsedResourceConfig
	Version        atc.Version
}

func (cache ResourceCache) FindOrCreateForBuild(logger lager.Logger, tx Tx, lockFactory lock.LockFactory, buildID int) (*UsedResourceCache, error) {
	usedResourceConfig, err := cache.ResourceConfig.FindOrCreateForBuild(logger, tx, lockFactory, buildID)
	if err != nil {
		return nil, err
	}

	return cache.findOrCreate(logger, tx, lockFactory, usedResourceConfig, "build_id", buildID)
}

func (cache ResourceCache) FindOrCreateForResource(logger lager.Logger, tx Tx, lockFactory lock.LockFactory, resourceID int) (*UsedResourceCache, error) {
	usedResourceConfig, err := cache.ResourceConfig.FindOrCreateForResource(logger, tx, lockFactory, resourceID)
	if err != nil {
		return nil, err
	}

	return cache.findOrCreate(logger, tx, lockFactory, usedResourceConfig, "resource_id", resourceID)
}

func (cache ResourceCache) FindOrCreateForResourceType(logger lager.Logger, tx Tx, lockFactory lock.LockFactory, resourceType *UsedResourceType) (*UsedResourceCache, error) {
	usedResourceConfig, err := cache.ResourceConfig.FindOrCreateForResourceType(logger, tx, lockFactory, resourceType)
	if err != nil {
		return nil, err
	}

	return cache.findOrCreate(logger, tx, lockFactory, usedResourceConfig, "resource_type_id", resourceType.ID)
}

func (cache *UsedResourceCache) Destroy(tx Tx) (bool, error) {
	rows, err := psql.Delete("resource_caches").
		Where(sq.Eq{
			"id": cache.ID,
		}).
		RunWith(tx).
		Exec()
	if err != nil {
		return false, err
	}

	affected, err := rows.RowsAffected()
	if err != nil {
		return false, err
	}

	if affected == 0 {
		return false, ErrResourceCacheDisappeared
	}

	return true, nil
}

func (cache ResourceCache) findOrCreate(
	logger lager.Logger,
	tx Tx,
	lockFactory lock.LockFactory,
	resourceConfig *UsedResourceConfig,
	forColumnName string,
	forColumnID int,
) (*UsedResourceCache, error) {
	id, found, err := cache.findWithResourceConfig(tx, resourceConfig)
	if err != nil {
		return nil, err
	}

	if !found {
		err = psql.Insert("resource_caches").
			Columns(
				"resource_config_id",
				"version",
				"params_hash",
			).
			Values(
				resourceConfig.ID,
				cache.version(),
				paramsHash(cache.Params),
			).
			Suffix("RETURNING id").
			RunWith(tx).
			QueryRow().
			Scan(&id)
		if err != nil {
			if pqErr, ok := err.(*pq.Error); ok && pqErr.Code.Name() == "foreign_key_violation" {
				return nil, ErrSafeRetryFindOrCreate
			}

			if pqErr, ok := err.(*pq.Error); ok && pqErr.Code.Name() == "unique_violation" {
				return nil, ErrSafeRetryFindOrCreate
			}

			return nil, err
		}
	}

	rc := &UsedResourceCache{
		ID:             id,
		ResourceConfig: resourceConfig,
		Version:        cache.Version,
	}

	var resourceCacheUseExists int
	err = psql.Select("1").
		From("resource_cache_uses").
		Where(sq.Eq{
			"resource_cache_id": id,
			forColumnName:       forColumnID,
		}).
		RunWith(tx).
		QueryRow().
		Scan(&resourceCacheUseExists)
	if err != nil {
		if err == sql.ErrNoRows {
			_, err = psql.Insert("resource_cache_uses").
				Columns(
					"resource_cache_id",
					forColumnName,
				).
				Values(
					id,
					forColumnID,
				).
				RunWith(tx).
				Exec()
			if err != nil {
				if pqErr, ok := err.(*pq.Error); ok && pqErr.Code.Name() == "foreign_key_violation" {
					return nil, ErrSafeRetryFindOrCreate
				}

				return nil, err
			}

			return rc, nil
		}

		return nil, err
	}

	return rc, nil
}

func (cache ResourceCache) lockName() (string, error) {
	cacheJSON, err := json.Marshal(cache)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", sha256.Sum256(cacheJSON)), nil
}

func (cache ResourceCache) findWithResourceConfig(tx Tx, resourceConfig *UsedResourceConfig) (int, bool, error) {
	var id int
	err := psql.Select("id").From("resource_caches").Where(sq.Eq{
		"resource_config_id": resourceConfig.ID,
		"version":            cache.version(),
	}).RunWith(tx).QueryRow().Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, false, nil
		}

		return 0, false, err
	}

	return id, true, nil
}

func (cache ResourceCache) version() string {
	j, _ := json.Marshal(cache.Version)
	return string(j)
}

func paramsHash(p atc.Params) string {
	if p != nil {
		return mapHash(p)
	}

	return mapHash(atc.Params{})
}

func mapHash(m map[string]interface{}) string {
	j, _ := json.Marshal(m)
	return fmt.Sprintf("%x", sha256.Sum256(j))
}

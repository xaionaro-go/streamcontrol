package memoize

import (
	"context"
	"fmt"
	"time"

	"github.com/facebookincubator/go-belt/tool/experimental/errmon"
	"github.com/facebookincubator/go-belt/tool/logger"
	"github.com/immune-gmbh/attestation-sdk/pkg/lockmap"
	"github.com/immune-gmbh/attestation-sdk/pkg/objhash"
	"github.com/sasha-s/go-deadlock"
)

type MemoizeData struct {
	Cache         map[objhash.ObjHash]any
	CacheMetaLock deadlock.Mutex
	CacheLockMap  *lockmap.LockMap
}

func NewMemoizeData() *MemoizeData {
	return &MemoizeData{
		Cache:        map[objhash.ObjHash]any{},
		CacheLockMap: lockmap.NewLockMap(),
	}
}

const timeFormat = time.RFC3339

func Memoize[REQ any, REPLY any, T func(context.Context, REQ) (REPLY, error)](
	d *MemoizeData,
	fn T,
	ctx context.Context,
	req REQ,
	cacheDuration time.Duration,
) (_ret REPLY, _err error) {
	logger.Debugf(ctx, "memoize %T", req)
	defer logger.Debugf(ctx, "/memoize %T", req)

	if IsNoCache(ctx) {
		cacheDuration = 0
	}

	key, err := objhash.Build(fmt.Sprintf("%T", req), req)
	errmon.ObserveErrorCtx(ctx, err)
	if err != nil {
		return fn(ctx, req)
	}
	logger.Debugf(ctx, "cache key %X", key[:])

	d.CacheMetaLock.Lock()
	logger.Tracef(ctx, "grpc.CacheMetaLock.Lock()-ed")
	cache := d.Cache

	h := d.CacheLockMap.Lock(key)
	logger.Tracef(ctx, "grpc.CacheLockMap.Lock(%X)-ed", key[:])

	type cacheItem struct {
		Reply   REPLY
		Error   error
		SavedAt time.Time
	}

	if h.UserData != nil {
		if v, ok := h.UserData.(cacheItem); ok {
			d.CacheMetaLock.Unlock()
			logger.Tracef(ctx, "grpc.CacheMetaLock.Unlock()-ed")
			h.Unlock()
			logger.Tracef(ctx, "grpc.CacheLockMap.Unlock(%X)-ed", key[:])
			logger.Debugf(ctx, "re-using the value")
			return v.Reply, v.Error
		} else {
			logger.Errorf(ctx, "cache-failure: expected type %T, but got %T", (*cacheItem)(nil), h.UserData)
		}
	}

	cachedResult, ok := cache[key]

	if ok {
		if v, ok := cachedResult.(cacheItem); ok {
			cutoffTS := time.Now().Add(-cacheDuration)
			if cacheDuration > 0 && !v.SavedAt.Before(cutoffTS) {
				d.CacheMetaLock.Unlock()
				logger.Tracef(ctx, "grpc.CacheMetaLock.Unlock()-ed")
				h.UserData = nil
				h.Unlock()
				logger.Tracef(ctx, "grpc.CacheLockMap.Unlock(%X)-ed", key[:])
				logger.Debugf(ctx, "using the cached value")
				return v.Reply, v.Error
			}
			logger.Debugf(ctx, "the cached value expired: %s < %s", v.SavedAt.Format(timeFormat), cutoffTS.Format(timeFormat))
			delete(cache, key)
		} else {
			logger.Errorf(ctx, "cache-failure: expected type %T, but got %T", (*cacheItem)(nil), cachedResult)
		}
	}
	d.CacheMetaLock.Unlock()
	logger.Tracef(ctx, "grpc.CacheMetaLock.Unlock()-ed")

	var ts time.Time
	defer func() {
		cacheItem := cacheItem{
			Reply:   _ret,
			Error:   _err,
			SavedAt: ts,
		}
		h.UserData = cacheItem
		h.Unlock()
		logger.Tracef(ctx, "grpc.CacheLockMap.Unlock(%X)-ed", key[:])

		d.CacheMetaLock.Lock()
		logger.Tracef(ctx, "grpc.CacheMetaLock.Lock()-ed")
		defer logger.Tracef(ctx, "grpc.CacheMetaLock.Unlock()-ed")
		defer d.CacheMetaLock.Unlock()
		cache[key] = cacheItem
	}()

	logger.Tracef(ctx, "no cache")
	ts = time.Now()
	return fn(ctx, req)
}

func (d *MemoizeData) InvalidateCache(ctx context.Context) {
	logger.Debugf(ctx, "invalidateCache()")
	defer logger.Debugf(ctx, "/invalidateCache()")

	d.CacheMetaLock.Lock()
	defer d.CacheMetaLock.Unlock()

	d.CacheLockMap = lockmap.NewLockMap()
	d.Cache = map[objhash.ObjHash]any{}
}

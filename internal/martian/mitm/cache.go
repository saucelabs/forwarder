// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// Copyright 2015 Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mitm

import (
	"crypto/tls"
	"time"

	"github.com/cespare/xxhash/v2"
	"github.com/elastic/go-freelru"
)

func xxHashString(k string) uint32 {
	v := xxhash.Sum64String(k)
	return uint32(v)
}

type Cache struct {
	*freelru.ShardedLRU[string, *tls.Certificate]
}

type CacheConfig struct {
	Capacity uint32
	TTL      time.Duration
}

func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		Capacity: 1024,
		TTL:      6 * time.Hour,
	}
}

func NewCache(cfg CacheConfig) (Cache, error) {
	certs, err := freelru.NewSharded[string, *tls.Certificate](cfg.Capacity, xxHashString)
	if err != nil {
		return Cache{}, err
	}
	certs.SetLifetime(cfg.TTL)
	return Cache{certs}, nil
}

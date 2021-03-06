// Copyright (c) 2018 HyperHQ Inc.
//
// SPDX-License-Identifier: Apache-2.0
//

package containerdshim

import (
	"context"
	"path"

	"github.com/containerd/containerd/mount"
	"github.com/kata-containers/runtime/pkg/katautils"
	"github.com/kata-containers/runtime/virtcontainers/types"

	"github.com/sirupsen/logrus"
)

func deleteContainer(ctx context.Context, s *service, c *container) error {
	status, err := s.sandbox.StatusContainer(c.id)
	if err != nil && !isNotFound(err) {
		return err
	}
	if !c.cType.IsSandbox() && err == nil {
		if status.State.State != types.StateStopped {
			_, err = s.sandbox.StopContainer(c.id, false)
			if err != nil {
				return err
			}
		}

		if _, err = s.sandbox.DeleteContainer(c.id); err != nil {
			return err
		}
	}

	// Run post-stop OCI hooks.
	if err := katautils.PostStopHooks(ctx, *c.spec, s.sandbox.ID(), c.bundle); err != nil {
		// log warning and continue, as defined in oci runtime spec
		// https://github.com/opencontainers/runtime-spec/blob/master/runtime.md#lifecycle
		logrus.WithError(err).Warn("Failed to run post-stop hooks")
	}

	if c.mounted {
		rootfs := path.Join(c.bundle, "rootfs")
		if err := mount.UnmountAll(rootfs, 0); err != nil {
			logrus.WithError(err).Warn("failed to cleanup rootfs mount")
		}
	}

	delete(s.containers, c.id)

	return nil
}

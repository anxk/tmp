package daemon // import "github.com/docker/docker/daemon"

import (
	"context"
	"fmt"

	"github.com/docker/docker/container"
	"github.com/sirupsen/logrus"
)

// ContainerPause 的相反操作，可参考 ./pause.go
// ContainerUnpause unpauses a container
func (daemon *Daemon) ContainerUnpause(name string) error {
	container, err := daemon.GetContainer(name)
	if err != nil {
		return err
	}
	return daemon.containerUnpause(container)
}

// containerUnpause resumes the container execution after the container is paused.
func (daemon *Daemon) containerUnpause(container *container.Container) error {
	container.Lock()
	defer container.Unlock()

	// We cannot unpause the container which is not paused
	if !container.Paused {
		return fmt.Errorf("Container %s is not paused", container.ID)
	}

	if err := daemon.containerd.Resume(context.Background(), container.ID); err != nil {
		return fmt.Errorf("Cannot unpause container %s: %s", container.ID, err)
	}

	container.Paused = false // 设置容器状态
	daemon.setStateCounter(container) // 更新 stateCtr
	daemon.updateHealthMonitor(container) // 开启健康检查
	daemon.LogContainerEvent(container, "unpause") // 记录容器 unpause 事件

	// 将容器配置保存到磁盘，并且更新 daemon.containersReplica
	if err := container.CheckpointTo(daemon.containersReplica); err != nil {
		logrus.WithError(err).Warn("could not save container to disk")
	}

	return nil
}

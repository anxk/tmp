package daemon // import "github.com/docker/docker/daemon"

import (
	"context"
	"fmt"

	"github.com/docker/docker/container"
	"github.com/sirupsen/logrus"
)

// ContainerPause pauses a container
func (daemon *Daemon) ContainerPause(name string) error {
	container, err := daemon.GetContainer(name)
	if err != nil {
		return err
	}
	return daemon.containerPause(container)
}

// containerPause pauses the container execution without stopping the process.
// The execution can be resumed by calling containerUnpause.
func (daemon *Daemon) containerPause(container *container.Container) error {
	container.Lock() // 加锁，防止竞争，因为涉及到 container 状态的查询与修改
	defer container.Unlock()

	// We cannot Pause the container which is not running
	if !container.Running {
		return errNotRunning(container.ID)
	}

	// We cannot Pause the container which is already paused
	if container.Paused {
		return errNotPaused(container.ID)
	}

	// We cannot Pause the container which is restarting
	if container.Restarting {
		return errContainerIsRestarting(container.ID)
	}

	// 将实际对容器的操作代理给 containerd
	if err := daemon.containerd.Pause(context.Background(), container.ID); err != nil {
		return fmt.Errorf("Cannot pause container %s: %s", container.ID, err)
	}

	container.Paused = true // 设置容器为暂停状态
	daemon.setStateCounter(container) // 使用 ./monitor.go 中的 setStateCounter 设置 ./metrics.go 中的 stateCtr，将此容器设置为暂停状态
	daemon.updateHealthMonitor(container) // 关闭容器健康检查，使用了 ./heath.go 中的 updateHealthMonitor
	daemon.LogContainerEvent(container, "pause") // 使用 ./events.go 中的 LogContainerEvent 记录关闭容器 pause 事件

	// 将容器配置保存到磁盘，并且更新 daemon.containersReplica
	if err := container.CheckpointTo(daemon.containersReplica); err != nil {
		logrus.WithError(err).Warn("could not save container to disk")
	}

	return nil
}
